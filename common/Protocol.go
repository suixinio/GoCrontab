package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//定时任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

//调度计划
type JobSchedulePlan struct {
	Job      *Job                 //要调度的任务
	Expr     *cronexpr.Expression //解析好的cronexpr表达式
	NextTime time.Time            // 下次调度时间

}

//任务执行状态
type JobExecuteInfo struct {
	Job      *Job
	PlanTime time.Time //理论上的调度时间
	RealTime time.Time //实际上的调度时间
}

//http response
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//变化事件
type JobEvent struct {
	EventType int // save delete
	Job       *Job
}

//任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo //执行状态
	Output      []byte          // 脚本的输出
	Err         error
	StartTime   time.Time
	EndTime     time.Time
}

// resp func

func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)
	response.Errno = errno
	response.Msg = msg
	response.Data = data

	//序列化json
	resp, err = json.Marshal(response)
	return
}

// 反序列化Job
//func UnpackJob(value []byte) (ret *Job, err error) {
//	var (
//		job *Job
//	)
//
//	job = &Job{}
//	if err = json.Unmarshal(value, job); err != nil {
//		return
//	}
//	ret = job
//	return
//}

func UnpackJob(value []byte) (ret *Job, err error) {
	//fmt.Println(ret)
	if err = json.Unmarshal(value, &ret); err != nil {
		return
	}
	return
}

//从etcd的key中提取任务名
// /cron/jobs/job1 抹掉/cron/jobs
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

//任务变化时间有两种 1更新任务 2 删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

//构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	//解析job的cron表达式
	expr, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return
	}
	//生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInof *JobExecuteInfo) {
	jobExecuteInof = &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime, // 计划调度时间
		RealTime: time.Now(),               //真实调度时间
	}
	return
}
