package main

import (
	"flag"
	"fmt"
	"gocrontab/worker"
	"runtime"
	"time"
)

var (
	confFile string //配置文件路径
)

//解析命令行参数
func initArgs() {
	//worker -config ./worker.json
	flag.StringVar(&confFile, "config", "./worker.json", "指定worker.json")
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
	fmt.Println("InitConfig")
	if err = worker.InitConfig(confFile); err != nil {
		goto ERR
	}
	//启动执行器
	fmt.Println("InitExecutor")
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}
	//启动调度器
	fmt.Println("InitScheduler")
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}
	fmt.Println("InitJobMgr")
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}
	//正常退出
	fmt.Println("Loop")
	for {
		time.Sleep(1 * time.Second)
	}
	return
ERR:
	//异常退出
	fmt.Println("异常")
	fmt.Println(err)
}
