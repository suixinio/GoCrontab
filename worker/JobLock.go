package worker

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gocrontab/common"
	"time"
)

//分布式锁（TXN事务）
type JobLock struct {
	//etcd客户端
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string             //任务锁
	cancelFunc context.CancelFunc //用于终止自动续租
	leaseId    clientv3.LeaseID   //租约ID
	isLock     bool               //是否上锁成功
}

func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	//创建租约（5秒）防止节点down机其他没有办法获取到这个锁
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	//租约ID
	leaseId = leaseGrantResp.ID
	//自动续租'
	if keepRespChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseId); err != nil {
		return
		goto FAIL
	}
	//处理续租应答的协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case keepResp = <-keepRespChan: // 自动续租的应答
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()
	//创建事务txn
	txn = jobLock.kv.Txn(context.TODO())
	//事务抢锁
	//锁路径
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))
	//提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	//成功返回，失败释放租约
	if !txnResp.Succeeded {
		//	锁占用
		fmt.Println("锁占用", jobLock.jobName, time.Now())
		err = common.ERR_LOCK_ALREAY_REQUIRED
		goto FAIL
	}
	//抢锁成功
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLock = true
	return
FAIL:
	cancelFunc() //取消自动续租
	jobLock.lease.Revoke(context.TODO(), leaseId)
	return
}

//释放锁
func (jobLock *JobLock) Unlock() {
	if jobLock.isLock {
		fmt.Println("释放锁", jobLock.jobName, time.Now())
		jobLock.cancelFunc()                                  // 取消程序的自动续租协程
		jobLock.lease.Revoke(context.TODO(), jobLock.leaseId) //释放租约
	}
}
