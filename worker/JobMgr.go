package worker

import (
	"context"
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
func (JobMgr *JobMgr) watchJobs(err error) {
	var (
		getResp *clientv3.GetResponse
	)
	//get /cron/jobs/   get current revision
	if getResp, err = JobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	for _, kvpair := range getResp.Kvs {
		//反序列化
		if job, err := common.UnpackJob(kvpair.Value); err == nil {
			//	todo ，同步给scheduler（调度协程）
			job = job
		}
	}
	//watch event
	go func() { //watch
		var (
			waitChan   clientv3.WatchChan
			watchResp  clientv3.WatchResponse
			watchEvent clientv3.Event
			job        *common.Job
			jobEvent   *common.JobEvent
			jobName    string
		)
		watchStartRevision := getResp.Header.Revision + 1
		//start watcher
		waitChan = JobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision))
		//处理监听事件
		for watchResp = range waitChan {
			for watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //save
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
					jobEvent = jobEvent
				//todo unpack，推送一个更新时间给scheduler
				case mvccpb.DELETE: //delete
					//delete
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}
					//构造删除event
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)

				}
				//	todo G_scheduler.PushJobEvent(jobEvent)
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
	return
}
