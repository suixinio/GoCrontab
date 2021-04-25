package master

import (
	"encoding/json"
	"fmt"
	"gocrontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

//任务的http接口
type ApiServer struct {
	httpServer *http.Server
}

var (
	G_apiServer *ApiServer
)

//保存任务接口
// post job = {"name":"job1","command":"echo hello","cronExpr":"* * * * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	//任务保存到ETCD中
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)

	//1、解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//2取出表单中的job字段
	postJob = req.PostForm.Get("job")
	//3反序列化
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}
	//save etcd
	oldJob, err = G_jobMgr.SaveJob(&job)
	if err != nil {
		goto ERR
	}
	// response
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	// response
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 删除任务
// post name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err  error
		name string
		//job     common.Job
		oldJob *common.Job
		bytes  []byte
	)
	//1、解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//2取出表单中的job字段
	name = req.PostForm.Get("name")

	oldJob, err = G_jobMgr.DelJob(name)
	if err != nil {
		goto ERR
	}
	// response
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
	return
}

// 任务列表
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList []*common.Job
		err     error
		bytes   []byte
	)
	jobList, err = G_jobMgr.ListJob()
	if err != nil {
		goto ERR
	}

	// response
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
	return
}

//强制杀死某个任务
//POST /job/kill name=job1
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		killerName string
		bytes      []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	killerName = req.PostForm.Get("name")
	//kill job
	if err = G_jobMgr.KillJob(killerName); err != nil {
		goto ERR
	}
	// response
	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
	return
}

//查询任务日志
func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		name       string
		skipParam  string
		limitParam string
		skip       int
		limit      int
		logArr     []*common.JobLog
		bytes      []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//获取请求参数
	name = req.Form.Get("name")
	skipParam = req.Form.Get("skip")
	limitParam = req.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}
	if logArr, err = G_logMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}
	// response
	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		fmt.Println(string(bytes))
		resp.Write(bytes)
	}
	return
}

//获取所有的节点
func handleWorkerList(resp http.ResponseWriter, req *http.Request) {
	var (
		workerArr []string
		err       error
		bytes     []byte
	)
	if workerArr, err = G_workerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	// response
	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		fmt.Println(string(bytes))
		resp.Write(bytes)
	}
	return
}

//初始化服务
func InitApiServer() (err error) {
	//配置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)

	staticDir := http.Dir(G_config.WebRoot)
	staticHandler := http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

	//启动TCP监听
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort))
	if err != nil {
		return
	}
	//创建http服务
	httpServer := &http.Server{
		ReadHeaderTimeout: time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout:      time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:           mux,
	}
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}
	go httpServer.Serve(listener)

	return
}
