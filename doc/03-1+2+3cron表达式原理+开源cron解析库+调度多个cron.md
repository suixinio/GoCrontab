## Cron基本格式

```
分（0-59）    时(0-23)     日(1-31)     月(1-12)     星期(0-7)      Command 
*            *            *            *           *             Shell命令
```

## Cron常见用法
```
- 每隔5分钟执行1次:     */5 * * * * echo hello > /tmp/error.log
- 第1-5分钟执行5次:     1-5 * * * * /usr/bin/python /data/x.py
- 每天10点，22点整执行1次: 0 10,22 * * * echo bye | tail -1 
```

## Cron调度原理
- 当前时间戳 2018/10/1 10:40:00
- 已经Cron表达式. 30 * * * *
- 如何计算命令的下次调度时间？

### 计算逻辑
```
40 分 10时 1日
枚举: 
分钟:[30]
小时:[0,1,2...23]
日期:[1,2,3,...,31]
月份:[1,2,3..12]

更大的单位向更小粒度的单位
```

## 使用Cron解析库
### 为什么不自己实现？
- 时间计算容易写出bug
- 非关键技术环节，造轮子收益太低
### 开源Cronexpr库
- Parse(): 解析与校验Cron表达式
- Next(): 根据当前时间，计算下次调度时间


## 3-2 开源cron解析库
```
package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

var (
	expr     *cronexpr.Expression
	err      error
	now      time.Time
	nextTime time.Time
)

func main() {

	//每分钟执行1次
	if expr, err = cronexpr.Parse("* * * * *"); err != nil {
		fmt.Println(err)
		return
	}

	//每5分钟执行1次
	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil {
		fmt.Println(err)
		return
	}

	//当前时间
	now = time.Now()

	//下次调度时间
	nextTime = expr.Next(now)
	fmt.Println(now, nextTime)

	//等待定时器超时
	time.AfterFunc(nextTime.Sub(now), func() {
		fmt.Println("被调度了:", nextTime)
	})

	time.Sleep(10 * time.Second)
}

```


## 3-3 调度多个cron
```
package main

import (
	"github.com/gorhill/cronexpr"
	"time"
	"fmt"
)

//代表一个任务
type CronJob struct {
	expr *cronexpr.Expression
	nextTime time.Time
}

var (
	cronJob *CronJob
	expr *cronexpr.Expression
	now time.Time
	scheduleTable map[string]*CronJob //任务的名字
)

func main()  {
	//需要有1个调度协程，它定时检查所有的Cron任务，谁过期了就执行谁

	//当前时间
	now = time.Now()

	//Job时间表
	scheduleTable = make(map[string]*CronJob)

	//定义2个cronjob
	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr:expr,
		nextTime:expr.Next(now),
	}

	//任务注册到调度表
	scheduleTable["job1"] = cronJob

	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr:expr,
		nextTime:expr.Next(now),
	}
	scheduleTable["job2"] = cronJob


	//启动一个调度协程
	go func() {
		var (
			jobName string
			cronJob *CronJob
			now time.Time
		)
		//定时检查一下任务调度表
		for{
			now = time.Now()

			for jobName,cronJob = range scheduleTable{
				//判断是否过期
				if cronJob.nextTime.Before(now) || cronJob.nextTime.Equal(now){

					//启动一个协程，执行这个任务
					go func(jobName string) {
						fmt.Println(time.Now(),"执行:",jobName)
					}(jobName)


					cronJob.nextTime = cronJob.expr.Next(now)
					fmt.Println(jobName,"下次执行时间:",cronJob.nextTime)
				}
			}

			//睡眠100毫秒
			select {
			//将在100毫秒可读，返回
			case <- time.NewTimer(100*time.Millisecond).C:
			}
			//time.Sleep(100*time.Millisecond)
		}
	}()

	time.Sleep(100*time.Second)
}
```