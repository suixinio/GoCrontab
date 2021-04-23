package master

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/client/v3"
	"gocrontab/common"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	G_jobMgr *JobMgr
)

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
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

func (JobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	//把任务保存到/corn/jobs/任务名 -》json

	//etc的保存key
	jobKey := common.JOB_SAVE_DIR + job.Name
	//任务信息json
	jobValue, err := json.Marshal(job)
	if err != nil {
		return
	}
	//保存到etcd
	putResp, err := JobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return
	}
	//如果是更新，那么返回旧值
	if putResp.PrevKv != nil {
		//对旧值做一个反序列化
		if err := json.Unmarshal(putResp.PrevKv.Value, &oldJob); err != nil {
			err = nil
		}
	}
	return

}

func (JobMgr *JobMgr) DelJob(jobName string) (oldJob *common.Job, err error) {
	//把任务保存到/corn/jobs/任务名 -》json

	//etc的保存key
	jobKey := common.JOB_SAVE_DIR + jobName
	//任务信息json
	//jobValue, err := json.Marshal(job)
	if err != nil {
		return
	}
	//保存到etcd
	delResp, err := JobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV())
	if err != nil {
		return
	}
	//如果是更新，那么返回旧值
	if len(delResp.PrevKvs) != 0 {
		//对旧值做一个反序列化
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJob); err != nil {
			err = nil
			return
		}
	}
	return

}

//列举任务
func (JobMgr *JobMgr) ListJob() (jobList []*common.Job, err error) {
	var (
		getResp *clientv3.GetResponse
	)
	dirKey := common.JOB_SAVE_DIR

	if getResp, err = JobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	} else {
		//jobList := make([]*common.Job, 0)
		for _, kvPair := range getResp.Kvs {
			job := &common.Job{}
			if err := json.Unmarshal(kvPair.Value, job); err != nil {
				err = nil
				continue
			}
			jobList = append(jobList, job)
		}
	}
	return
}

//杀死任务
func (JobMgr *JobMgr) KillJob(jobName string) (err error) {
	//更新一下key/cron/killer/任务名
	killerKey := common.JOB_KILLER_DIR + jobName
	leaseGrantResp, err := JobMgr.lease.Grant(context.TODO(), 1)
	if err != nil {
		return
	}
	leaseId := leaseGrantResp.ID
	if _, err = JobMgr.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
