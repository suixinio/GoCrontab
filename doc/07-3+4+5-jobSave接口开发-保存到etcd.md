## 7-1/2 创建项目与搭建基本框架
- 新建Common包
- 新建Protocol.go 存储Job struct
```
master
    main
        master.go      //主程序，初始化配置
        master.json    //配置文件
    ApiServer.go
    Config.go          //读取系统配置文件master.json
    JobMgr.go          //任务管理，建立etcd连接，返回任务管理单例
common
    Protocol.go    
```
- func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) 

```
package master

import (
	"context"
	"encoding/json"
	"go-crontab/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
)

//任务管理
type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	//单例
	G_jobMgr *JobMgr
)

//初始化管理器
func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	//得到KV和elase的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//赋值单例
	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

//保存任务
func (jobMgr *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	//把任务保存到/cron/jobs/任务名 -> json
	var (
		jobKey    string
		jobValue  []byte
		putResp   *clientv3.PutResponse
		oldJobObj common.Job
	)

	//etcd的保存key
	jobKey = common.JOB_SAVE_DIR + job.Name
	//任务信息json
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}

	//保存到etcd
	if putResp, err = jobMgr.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}

	//如果是更新，那么返回旧值
	if putResp.PrevKv != nil {
		//对旧值做一个反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}
```

```
//保存任务接口
//POST job={"name":"job1","command":"echo hello", "cronExpr":"* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
	)

	//1,解析POST表单
	if err = req.ParseForm(); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
		return
	}

	//2,取表单中的job字段
	postJob = req.PostForm.Get("job")

	//3,反序列化job
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
		return
	}

	//4,保存到etcd
	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
		return
	}

	//正常应答
	common.ResponseErr(resp, 0, "success", oldJob)
	return

}
```