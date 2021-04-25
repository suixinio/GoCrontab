## 7-8 job-kill接口开发: 在etcd中标记结束任务
- 路由
```
mux.HandleFunc("/job/kill", handleJobKill)
```
- 实现
```
//强制杀死某个任务
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err   error
		name  string
		bytes []byte
	)

	//解析POST表单
	if err = req.ParseForm(); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
	}

	//要杀死的任务名
	name = req.PostForm.Get("name")

	//杀死任务
	if err = G_jobMgr.KillJob(name); err != nil {
		common.ResponseErr(resp, -1, err.Error(), nil)
	}

	//正常应答
	common.ResponseErr(resp, 0, "success", nil)
	return
}
```
- 实现KillJob
```
//  杀死任务
func (jobMgr *JobMgr) KillJob(name string) (err error) {
	//更新一下key= /cron/killer/任务名
	var (
		killerKey      string
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	//通知worker杀死对应任务
	killerKey = common.JOB_KILLER_DIR + name

	//让worker监听到一次put操作，创建一个租约让其稍后自动过期即可
	if leaseGrantResp, err = jobMgr.lease.Grant(context.TODO(), 1); err != nil {
		return
	}

	//租约ID
	leaseId = leaseGrantResp.ID

	//设置killer标记
	if _, err = jobMgr.kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseId)); err != nil {
		return
	}
	return
}
```