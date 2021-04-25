package master

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gocrontab/common"
)

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		client *mongo.Client
	)

	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(G_config.MongodbUri))
	if err != nil {
		//fmt.Println(err)
		return
	}
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

func (logMgr *LogMgr) ListLog(name string, skip, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
	)
	//len(logArr)
	logArr = make([]*common.JobLog, 0)
	//过滤条件
	filter = &common.JobLogFilter{JobName: name}
	//按照任务开始时间倒排
	logSort = &common.SortLogByStartTime{
		SortOrder: -1,
	}
	//ctx, cancel := context.WithTimeout(context.Background(), time.Duration(G_config.MongodbConnectTimeout)*time.Second)
	//defer cancel()

	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, &options.FindOptions{Sort: logSort, Skip: &[]int64{int64(skip)}[0], Limit: &[]int64{int64(limit)}[0]}); err != nil {
		fmt.Println(err)
		return
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}
		if err = cursor.Decode(jobLog); err != nil {
			continue
		}
		logArr = append(logArr, jobLog)
	}
	return
}
