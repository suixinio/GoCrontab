package worker

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gocrontab/common"
	"net"
	"time"
)

//注册节点到etcd /cron/workers/ip地址
type Register struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	localIP string
}

var (
	G_register *Register
)

func getLocalIp() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}
	for _, addr = range addrs {
		//ipv4  ipv6
		//这个网络地址是ip地址 ipv4  ipv6
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet {
			//跳过ipv6
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUND
	return
}

//注册到/cron/worker/ip,并且自动续租
func (register *Register) keepOnline() {
	var (
		regKey         string
		leaseGrantResp *clientv3.LeaseGrantResponse
		err            error
		keepAliveChan  <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp  *clientv3.LeaseKeepAliveResponse
		//putResp        *clientv3.PutResponse
		cancelFunc context.CancelFunc
		cancelCtx  context.Context
	)
	for {
		//注册路径
		regKey = common.JOB_WORKER_DIR + register.localIP
		cancelFunc = nil
		if leaseGrantResp, err = register.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}
		//自动续租
		if keepAliveChan, err = register.lease.KeepAlive(context.TODO(), leaseGrantResp.ID); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())

		//注册到etcd
		if _, err = register.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGrantResp.ID)); err != nil {
			goto RETRY
		}
		//处理续租应答
		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil {
					goto RETRY
				}
			}
		}

	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

}

func InitRegister() (err error) {
	config := clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}
	//建立连接
	client, err := clientv3.New(config)
	if err != nil {
		return
	}
	localIP, err := getLocalIp()
	if err != nil {
		return
	}
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIP,
	}
	go G_register.keepOnline()
	return
}
