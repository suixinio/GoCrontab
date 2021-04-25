package common

const (
	//任务保存目录
	JOB_SAVE_DIR = "/cron/jobs/"

	//杀死任
	JOB_KILLER_DIR = "/cron/killer/"

	//任务锁
	JOB_LOCK_DIR = "/cron/lock/"

	// 服务注册目录
	JOB_WORKER_DIR = "/cron/workers/"

	//保存任务事件
	JOB_EVENT_SAVE = 1

	//删除任务事件
	JOB_EVENT_DELETE = 2

	//杀死任务事件
	JOB_EVENT_KILL = 3
)
