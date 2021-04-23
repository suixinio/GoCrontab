package master

import "go.etcd.io/etcd/client/v3"

type JobMgr struct {
	client *clientv3.Client
}
