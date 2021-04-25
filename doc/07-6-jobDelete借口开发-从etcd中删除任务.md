## 7-6 job delete接口开发: 从etcd中删除任务
- 路由
```
mux.HandleFunc("/job/delete", handleJobDelete)
```
- 实现
```
//删除任务
func (jobMgr *JobMgr) DeleteJob(name string) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		delResp   *clientv3.DeleteResponse
		oldJobObj common.Job
	)

	//etcd中保存任务的key
	jobKey = common.JOB_SAVE_DIR + name

	//从etcd中删除它
	if delResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	//返回被删除的任务信息
	if len(delResp.PrevKvs) != 0 {
		//解析一下旧值，返回它
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}
```
```
//删除任务接口
//POST /job/delete name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error
		name   string
		oldJob *common.Job
	)
	//POST: a=1&b=2&c=3
	if err = req.ParseForm(); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
		return
	}

	//删除的任务名
	name = req.PostForm.Get("name")

	//去删除任务
	if oldJob, err = G_jobMgr.DeleteJob(name); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
		return
	}

	//正常应答
	common.ResponseErr(resp, 0, "success", oldJob)
	return
}
```