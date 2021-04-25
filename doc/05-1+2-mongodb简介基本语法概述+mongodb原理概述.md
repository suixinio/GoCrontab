## 5-1 mongodb简介&基础语法概述
### 核心特性
- 文档数据库(schema free),基于二进制Json存储文档
- 高性能、高可用、直接加机器即可解决扩展性问题
- 支持丰富的CURD操作，例如：聚合统计、全文检索、坐标检索

### 文档数据库schema free
- field: value

### 概念类比
- mysql         mongoDb
- database      database
- table         collection
- row           document(bson)
- column        field
- index         index
- table joins   $lookup
- primary key   _id
- group by      aggregation pipeline

### 用法示例 选择database

- 列举数据库: show databases
- 选择数据库: use  db
- 结论: 数据库无需创建，只是一个命令空间

### 用法示例 创建collection
- 列举数据表: show collections
- 建立数据表: db.createCollection("my_collection")
- 结论: 数据表schema free,无需定义字段

### 用法示例 插入document
- db.my_collection.insertOne({uid:10000,name:"xiaoming",likes:["football","game"]})
- 结论1: 任意嵌套层级的BSON(二进制的Json)
- 结论2: 文档ID是自动生成的，通常无需自己指定


### 用法示例 查询document
- db.my_collection.find({likes:"football",name:{$in:['xiaoming','libai']}}.sort({uid:1})
- 结论1: 可以基于任意Bson层级过滤
- 结论2: 支持的功能与Mysql相当

### 用法示例 更新document
- db.my_collection.updateMany({likes:'football'},{$set:{name:'libai'}})
- 结论1: 第一个参数是过滤条件
- 结论2: 第二个参数是更新操作

### 用法示例 删除document
- db.my_collection.deleteMany({name:'xiaoming'})
- 结论: 参数是过滤条件

### 用法示例 创建index
- db.my_collection.createIndex({uid:1,name:-1})
- 结论: 可以指定建立索引时的正反顺序

### 聚合类比
- Mysql         MongoDB
- where         $match
- group by      $group
- having        $match
- select        $project
- order by      $sort
- limit         $limit
- sum           $sum
- count         $sum

### 聚合统计示例
- db.my_collection.aggregate([$unwind:'$likes'},{$group:{_id:{likes:'$likes'}}},{$project:{_id:0, like:"$_id.likes",total:{$sum:1}}}]
- 输出: {"like":"game","total":1}{"like":"football","total":1}
- 结论: pipeline流式计算，功能复杂，使用时多查手册

## 5-2 mongodb原理概述
### 整体存储架构
- Mongod: 单机版数据库
- Replica Set: 复制集，由多个Mongod组成的高可用存储单元
- Sharding: 分布式集群，由多个Replica Set组成的可扩展集群

### Mongod架构
- client    client
- MongoDB Query Language
- MongoDB Data Model
- Wired Tiger、MMAPv1、In Memory、Encrypted、3rd Party Engine

### Mongod架构
- 默认采用WiredTiger高性能存储引擎
- 基于journaling log宕机恢复(类比mysql的redo log)

### Replica Set架构
- Client Application
- Driver
- Writes Reads
- Primary
- Secondary Secondary

### Replica Set架构
- 至少3个节点组成，其中1个可以只充当arbiter
- 主从基于oplog复制同步（类比mysql binlog)
- 客户端默认读写primary节点

### Sharding架构
- App Server Router mongos   1 or more Routers 路由层
- Shard (replica set)          2 or more Shards
- Config Servers (replica set)


### Sharding架构
- mongos作为代理，路由请求到特定shard
- 3个mongod节点组成config server,保存数据元信息
- 每个shard是一个replica set, 可以无限扩容

### Collection分片
- Router -> Collection1
- Collection自动分裂成多个chunk
- 每个chunk被自动负载均衡到不同的shard
- 每个shard可以保证其上的chunk高可用

### 按range拆分chunk
- Shard key可以是单索引字段或者联合索引字段
- 超过16MB的chunk被一分为二
- 一个collection的所以的chunk首位相连，构成整个表

### 按hash拆分chunk
- Shard key必须是哈希索引字段
- Shard key经过hash函数打散，避免写热点
- 支持预分配chunk,避免运行时分裂影响写入

### shard用法示例
- 为DB激活特性: sh.enableSharding("my_db")
- 配置hash分片: sh.shardCollection("my_db.my_collection",{_id:"hashed"},false,{numInitialChunks:10240})

### 重要提示
- 按非shard key查询，请求被扇出给所有shard