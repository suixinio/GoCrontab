## 4-3 搭建与连接etcd
### 下载

-  ```https://github.com/etcd-io/etcd/releases```
- 解压

### 启动
- 服务器启动

```
nohup ./etcd --listen-client-urls 'http://0.0.0.0:2379' --advertise-client-urls 'http://0.0.0.0:2379' &

//查看nohup输出
less nohup.out
```

- 客户端启动

```
//指定etcd api版本
ETCDCTL_API=3 ./etcdctl
```
- 客户端操作

```
ETCDCTL_API=3 ./etcdctl put "name" "sam"
ETCDCTL_API=3 ./etcdctl get "name"
ETCDCTL_API=3 ./etcdctl del "name"




ETCDCTL_API=3 ./etcdctl put "/cron/jobs/job1" "{...job1}"
ETCDCTL_API=3 ./etcdctl put "/cron/jobs/job2" "{...job2}"
ETCDCTL_API=3 ./etcdctl get "/cron/jobs/job1"
ETCDCTL_API=3 ./etcdctl get "/cron/jobs/job2"

ETCDCTL_API=3 ./etcdctl get "/cron/jobs/" --prefix

//watch机制
ETCDCTL_API=3 ./etcdctl watch "/cron/jobs/" --prefix
```

### etcd客户端
- 下载

```
go get go.etcd.io/etcd/clientv3
```

- 连接客户端

```
package main

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"fmt"
)

var (
	config clientv3.Config
	client *clientv3.Client
	err error
)

func main(){
	//客户端配置
	config = clientv3.Config{
		Endpoints:[]string{"127.0.0.1:2379"},
		DialTimeout:5*time.Second,
	}

	//建立连接
	if client, err = clientv3.New(config);err != nil{
		fmt.Println(err)
		return
	}

	fmt.Println("连接成功")
}
```