package common

import (
	"encoding/json"
	"strings"
)

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
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
	if err = json.Unmarshal(value, ret); err != nil {
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
