##  4-10    op取代get,put,delete方法
```
package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

var (
	config clientv3.Config
	client *clientv3.Client
	err    error
	kv     clientv3.KV
	putOp  clientv3.Op
	getOp  clientv3.Op
	opResp clientv3.OpResponse
)

func main() {
	//客户端配置
	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	//建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	//创建Op: operation
	putOp = clientv3.OpPut("/cron/jobs/job8", "12345678")

	//执行Op
	if opResp, err = kv.Do(context.TODO(), putOp); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("写入Revision:", opResp.Put().Header.Revision)

	//创建Op
	getOp = clientv3.OpGet("/cron/jobs/job8")

	//执行Op
	if opResp, err = kv.Do(context.TODO(), getOp); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("数据Revision:", opResp.Get().Kvs[0].ModRevision)
	fmt.Println("数据value:", string(opResp.Get().Kvs[0].Value))
}

```

## 4-11/12 事务tnx实现分布式锁
```
package main

import (
	"go.etcd.io/etcd/clientv3"
	"context"
	"time"
	"fmt"
)

var (
	config clientv3.Config
	client *clientv3.Client
	err error
	lease clientv3.Lease
	leaseGrantResp *clientv3.LeaseGrantResponse
	leaseId clientv3.LeaseID
	keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
	keepResp *clientv3.LeaseKeepAliveResponse
	ctx context.Context
	cancelFunc context.CancelFunc
	kv clientv3.KV
	txn clientv3.Txn
	txnResp *clientv3.TxnResponse
)


func main()  {
	//客户端配置
	config = clientv3.Config{
		Endpoints:[]string{"127.0.0.1:2379"},
		DialTimeout:5 * time.Second,
	}

	//建立连接
	if client,err = clientv3.New(config); err != nil{
		fmt.Println(err)
		return
	}

	//lease实现锁自动过期
	//op操作
	//txn事务: if else then

	//1,上锁 (创建租约，自动续租，拿着租约去抢占一个key)
	lease = clientv3.NewLease(client)

	//申请1个5秒的租约
	if leaseGrantResp, err = lease.Grant(context.TODO(),5); err != nil{
		fmt.Println(err)
		return
	}

	//拿到租约ID
	leaseId = leaseGrantResp.ID

	//准备一个用于取消自动续租的context
	ctx, cancelFunc = context.WithCancel(context.TODO())

	//确保函数退出后，自动续租会停止
	defer cancelFunc()
	defer lease.Revoke(context.TODO(),leaseId)

	//5秒后会取消自动续租
	if keepRespChan, err = lease.KeepAlive(ctx,leaseId); err != nil{
		fmt.Println(err)
		return
	}

	//处理续约应答的协程
	go func() {
		for{
			select{
			case keepResp = <- keepRespChan:
				if keepRespChan == nil{
					fmt.Println("租约已经失效了")
					goto END
				}else {
					fmt.Println(time.Now().Format("2006-01-02 15:04:05"),"收到自动续租应答:",keepResp.ID)
				}
			}
		}
		END:
	}()


	//if 不存在key, then 设置它， else抢锁失败
	kv = clientv3.NewKV(client)

	//创建事务
	txn = kv.Txn(context.TODO())


	//定义事务
	//如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"),"=",0)).
		Then(clientv3.OpPut("/cron/lock/job9","xxx123",clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/lock/job9")) //否则抢锁失败

	//提交事务
	if txnResp, err = txn.Commit(); err != nil{
		fmt.Println(err)
		return
	}

	//判断是否抢到了锁
	if !txnResp.Succeeded{
		fmt.Println("锁被占用:", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	//2, 处理业务
	fmt.Println("处理任务")
	time.Sleep(50*time.Second)

	//3, 释放锁（取消自动续租，释放租约
	//defer 会把租约释放掉，关联的KV就被删除了
}
```