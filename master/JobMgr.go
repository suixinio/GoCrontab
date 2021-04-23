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
	oldJobObj := common.Job{}
	if putResp.PrevKv != nil {
		//对旧值做一个反序列化
		if err := json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
		}
		oldJob = &oldJobObj
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
	oldJobObj := common.Job{}

	//oldJobObjs := []common.Job{}
	if len(delResp.PrevKvs) != 0 {
		//对旧值做一个反序列化
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return

}
