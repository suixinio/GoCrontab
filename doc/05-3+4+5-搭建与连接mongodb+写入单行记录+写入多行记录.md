## 搭建与连接mongodb
### 安装步骤
- 下载 tgz文件
- 解压

```
tar -zxvf mongodb-linux-x86_64-rhe170-4.0.0.tgz
cd mongodb-linux-x86_64-rhe170-4.0.0
```
- 创建数据库目录
```
mkdir data
```
- 启动单机版mongod
```
nohup bin/mongod --dbpath=./data --bind_ip=0.0.0.0 &
```
- 启动mongo
```
bin/mongo
```
### 操作
- 查看,使用数据库
```
show databases;
use my_db;
```
- 查看帮助
```
db.help();
```
- 建表,查看表
```
db.createCollection("my_collection");
show collections;

```

- 插入数据
```
db.my_collection.insertOne({uid:1000,name:"xiaoming"})
```
- 查找
```
db.my_collection.find();
db.my_collection.find({uid:1000});
```
- 建立索引
```
db.my_collection.createIndex({uid:1});
```

- 连接数据库
```
package main

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"context"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
	"fmt"
)

var (
	client *mongo.Client
	err error
	database *mongo.Database
	collection *mongo.Collection
)

func main()  {
	//1, 建立连接
	if client, err = mongo.Connect(context.TODO(), "mongodb://127.0.0.1:27017", clientopt.ConnectTimeout(5*time.Second)); err != nil{
		fmt.Println(err)
		return
	}

	//2, 选择数据库my_db
	database = client.Database("my_db")

	//3, 选择表my_collection
	collection = database.Collection("my_collection")

	//4, 输出表名
	fmt.Println(collection.Name())
}
```

###
5-4 InsertOne写入单行记录
```
package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

// 任务的执行时间
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

// 一条日志
type LogRecord struct {
	JobName   string    `bson:"jobName"`   //任务名
	Command   string    `bson:"command"`   //shell命令
	Err       string    `bson:"err"`       //脚本错误
	Content   string    `bson:"content"`   //脚本输出
	TimePoint TimePoint `bson:"timePoint"` //执行时间点
}

var (
	client     *mongo.Client
	err        error
	database   *mongo.Database
	collection *mongo.Collection
	record     *LogRecord
	result     *mongo.InsertOneResult
	docId      objectid.ObjectID
)

func main() {
	//1, 建立连接
	if client, err = mongo.Connect(context.TODO(), "mongodb://127.0.0.1:27017", clientopt.ConnectTimeout(5*time.Second)); err != nil {
		fmt.Println(err)
		return
	}

	//2, 选择数据库my_db
	database = client.Database("cron")

	//3, 选择表my_collection
	collection = database.Collection("log")

	//4, 插入记录(bson)
	record = &LogRecord{
		JobName: "job10",
		Command: "echo hello",
		Err:     "",
		Content: "hello",
		TimePoint: TimePoint{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 10,
		},
	}

	if result, err = collection.InsertOne(context.TODO(), record); err != nil {
		fmt.Println(err)
		return
	}

	//_id: 默认生成一个全局唯一ID,ObjectID: 12字节的二进制
	docId = result.InsertedID.(objectid.ObjectID)
	fmt.Println("自增ID:", docId.Hex())
}

```

### 5-5 InsertMany写入多行记录
```
package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

var (
	client     *mongo.Client
	err        error
	database   *mongo.Database
	collection *mongo.Collection
	record     *LogRecord
	logArr     []interface{}
	result     *mongo.InsertManyResult
	insertId   interface{}
	docId      objectid.ObjectID
)

// 任务的执行时间点
type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

// 一条日志
type LogRecord struct {
	JobName   string    `bson:"jobName"`   //任务名
	Command   string    `bson:"command"`   //shell命令
	Err       string    `bson:"err"`       //脚本错误
	Content   string    `bson:"content"`   //脚本输出
	TimePoint TimePoint `bson:"timePoint"` //执行时间点

}

func main() {
	//1, 建立连接
	if client, err = mongo.Connect(context.TODO(), "mongodb://127.0.0.1:27017", clientopt.ConnectTimeout(5*time.Second)); err != nil {
		fmt.Println(err)
		return
	}

	//2, 选择数据库cron
	database = client.Database("cron")

	//3, 选择表log
	collection = database.Collection("log")

	//4, 插入记录bson
	record = &LogRecord{
		JobName: "job10",
		Command: "echo 10",
		Err:     "",
		Content: "hello 10",
		TimePoint: TimePoint{
			StartTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + 20,
		},
	}

	//5, 批量插入多条document
	logArr = []interface{}{record, record, record}

	//发起插入
	if result, err = collection.InsertMany(context.TODO(), logArr); err != nil {
		fmt.Println(err)
		return
	}

	//
	for _, insertId = range result.InsertedIDs {
		docId = insertId.(objectid.ObjectID)
		fmt.Println("自增ID:", docId.Hex())
	}
}
```