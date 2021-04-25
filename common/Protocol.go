package common

import (
	"context"
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
	Job        *Job
	PlanTime   time.Time          //理论上的调度时间
	RealTime   time.Time          //实际上的调度时间
	CancelFunc context.CancelFunc //取消任务
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

// 任务执行日志
type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`           // 任务名字
	Command      string `json:"command" bson:"command"`           // 脚本命令
	Err          string `json:"err" bson:"err"`                   // 错误原因
	Output       string `json:"output" bson:"output"`             // 脚本输出
	PlanTime     int64  `json:"planTime" bson:"planTime"`         // 计划开始时间
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"` // 实际调度时间
	StartTime    int64  `json:"startTime" bson:"startTime"`       // 任务执行开始时间
	EndTime      int64  `json:"endTime" bson:"endTime"`           // 任务执行结束时间
}

// 日志批次
type LogBatch struct {
	Logs []interface{} // 多条日志
}

// 任务日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

// 任务日志排序规则
type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` // {startTime: -1}
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

// 提取worker的IP
func ExtractWorkerIP(regKey string) string {
	return strings.TrimPrefix(regKey, JOB_WORKER_DIR)
}
func ExtractKillerName(killKey string) string {
	return strings.TrimPrefix(killKey, JOB_KILLER_DIR)
}
