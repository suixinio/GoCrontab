package worker

import (
	"context"
	"gocrontab/common"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	G_executor *Executor
)

func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd    *exec.Cmd
			err    error
			output []byte
			result *common.JobExecuteResult
		)
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}
		//记录开始时间
		result.StartTime = time.Now()
		//执行shell命令
		//执行并捕获输出
		//任务执行完成后，把执行的结果返回给Scheduler，Scheduler会从executingtable中删除执行记录
		cmd = exec.CommandContext(context.TODO(), "/bin/bash", "-c", info.Job.Command)
		output, err = cmd.CombinedOutput()
		//记录任务结束时间
		result.EndTime = time.Now()
		result.Output = output
		result.Err = err

		//任务执行完成后，把执行的结果返回给Scheduler ， Sched会从executingTable中删除掉执行记录
		G_scheduler.PushJobResult(result)
	}()
	return
}

func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
