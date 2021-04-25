## 7-7 job-list接口开发：从etcd获取所有任务
- 路由
```
mux.HandleFunc("/job/list", handleJobList)
```
- 实现handleJobList()
```
//列举所有crontab任务
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList []*common.Job
		err     error
	)

	//获取任务列表
	if jobList, err = G_jobMgr.ListJobs(); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
		return
	}

	//正常应答
	common.ResponseErr(resp, 0, "success", jobList)
	return
}

```
- 实现ListJobs()
```
//列举任务
func (jobMgr *JobMgr) ListJobs() (jobList []*common.Job, err error) {
	var (
		dirKey  string
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		job     *common.Job
	)

	//任务保存的目录
	dirKey = common.JOB_SAVE_DIR

	//获取目录下所有的任务信息
	if getResp, err = jobMgr.kv.Get(context.TODO(), dirKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	//初始化数组空间
	jobList = make([]*common.Job, 0)

	//遍历所有任务，进行反序列化
	for _, kvPair = range getResp.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(kvPair.Value, job); err != nil {
			err = nil
			continue
		}
		jobList = append(jobList, job)
	}
	return
}
```