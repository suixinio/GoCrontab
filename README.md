# go-crontab

```
doc
common  
master  master程序
worker  worker程序
```

## 前置
- 启动mongodb 
```
mongodb://127.0.0.1:27017
```
- 启动etcd  
```
127.0.0.1:2379
```
- 下载依赖包
```
go get go.mongodb.org/mongo-driver
go get go.etcd.io/etcd/clientv3
```
## 安装&运行
- 安装运行
```
cd $GOPATH/src
git clone  xx.git
cd go-crontab
cd master/main
go run master.go

cd worker/main
go run worker.go
```
- 访问 localhost:8070
- 测试 新建任务
```
test
echo "test"
5 * * * * * * 
``` 
## 知识点
- github.com/gorhill/cronexpr cron解析库的使用
- etcd的使用
- mongodb的使用