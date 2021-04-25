package main

import (
	"flag"
	"fmt"
	"gocrontab/master"
	"runtime"
	"time"
)

var (
	confFile string //配置文件路径
)

//解析命令行参数
func initArgs() {
	//master -config ./master.json
	flag.StringVar(&confFile, "config", "./master.json", "指定master.json")
	flag.Parse()
}

//初始化线程
func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	//初始化命令行参数
	initArgs()
	//初始化线程
	initEnv()

	var (
		err error
	)
	//加载配置
	if err = master.InitConfig(confFile); err != nil {
		goto ERR
	}
	//日志管理器
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	//启动api服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
	//正常退出
	for {
		time.Sleep(1 * time.Second)
	}
	return
ERR:
	//异常退出
	fmt.Println(err)
}
