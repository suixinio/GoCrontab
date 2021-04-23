package master

import (
	"encoding/json"
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

//初始化服务
func InitApiServer() (err error) {
	//配置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
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
