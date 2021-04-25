package worker

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"gocrontab/common"
	"time"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

var (
	G_jobMgr *JobMgr
)

//监听任务变化
func (JobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp *clientv3.GetResponse
	)
	//get /cron/jobs/   get current revision
	if getResp, err = JobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		fmt.Println(err)
		return
	}
	// 当前有哪些任务
	for _, kvpair := range getResp.Kvs {
		//反序列化
		if job, err := common.UnpackJob(kvpair.Value); err == nil {
			//	todo ，同步给scheduler（调度协程）
			jobEvent := common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
			//fmt.Println(*jobEvent)
			G_scheduler.PushJobEvent(jobEvent)
			//job = job
		} else {
			fmt.Println(err)
		}
	}
	//watch event
	go func() { //watch
		var (
			waitChan   clientv3.WatchChan
			watchResp  clientv3.WatchResponse
			watchEvent *clientv3.Event
			job        *common.Job
			jobEvent   *common.JobEvent
			jobName    string
		)
		watchStartRevision := getResp.Header.Revision + 1
		//start watcher
		waitChan = JobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
		//处理监听事件
		for watchResp = range waitChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //save
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
					//jobEvent = jobEvent
				case mvccpb.DELETE: //delete
					//delete
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					//构造删除event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				//推送给scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

// 监听强杀任务通知
func (jobMgr *JobMgr) watchKiller() {
	var (
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *common.JobEvent
		jobName    string
		job        *common.Job
	)

	// 监听协程 监听/cron/killer目录
	go func() {
		// 监听/cron/killer/目录的变化
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_DIR, clientv3.WithPrefix())
		// 处理监听事件
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: // 杀死任务事件
					jobName = common.ExtractKillerName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					// 事件推送给scheduler
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
					// killer标记过期，被自动删除
				}
			}
		}
	}()
}

func InitJobMgr() (err error) {
	config := clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}
	//建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return
	}
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)
	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	// 启动任务监听
	G_jobMgr.watchJobs()
	G_jobMgr.watchKiller()
	return
}

//创建任务执行锁
func (jobMgr *JobMgr) CreateJobLock(jobName string) (jobLock *JobLock) {
	jobLock = InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
	return
}
