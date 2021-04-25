package master

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gocrontab/common"
	"time"
)

type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	G_workerMgr *WorkerMgr
)

func (workerMgr *WorkerMgr) ListWorkers() (workerArr []string, err error) {
	var (
		getResp *clientv3.GetResponse
	)
	//初始化数组
	workerArr = make([]string, 0)
	workerKey := common.JOB_WORKER_DIR

	//获取目录下所有的kv
	if getResp, err = workerMgr.kv.Get(context.TODO(), workerKey, clientv3.WithPrefix()); err != nil {
		return
	} else {
		for _, kvPair := range getResp.Kvs {
			//fmt.Println(kvPair)
			workerIP := common.ExtractWorkerIP(string(kvPair.Key))

			workerArr = append(workerArr, workerIP)

		}
	}
	fmt.Println(workerArr)
	return
}

func InitWorkerMgr() (err error) {
	config := clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}
	//建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return
	}
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	G_workerMgr = &WorkerMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}
