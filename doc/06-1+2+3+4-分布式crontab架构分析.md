## 6-1 架构分析
### 传统crontab痛点
- 机器故障，任务停止调度，甚至crontab配置都找不回来
- 任务数量多，单机的硬件资源耗尽，需要人工迁移到其他机器
- 需要人工去机器上配置cron,任务执行状态不方便查看

### 分布式架构-核心要素
- 调度器：需要高可用，确保不会因为单点故障停止调度
- 执行器：需要扩展性，提供大量任务的并行处理能力

### 常见开源调度架构
- 调度master
- 调度standby
- 基于zookeeper
- 常见quartz
- 执行worker
- 基于RPC下方任务
- RPC状态上报

### 伪分布式设计
- 分布式网络环境不可靠，RPC异常属于常态
- Master下发任务RPC异常，导致Master与Worker状态不一致
- Worker上报任务RPC异常，导致Master状态信息落后

### 异常case举例
- 状态不一致: Master下发任务给node1异常，实际node1收到并开始执行
- 并发执行: Master重试下发任务给node2,结果node1与node2同时执行一个任务
- 状态丢失: Master更新zookeeper中任务状态异常，此时Master宕机切换Standby,任务仍旧处于旧状态

### 分布式伪命题
- 但凡需要经过网络的操作，都可能出现异常
- 将应用状态放在存储中，必然会出现内存与存储状态不一致
- 应用直接利用raft管理状态，可以确保最终一致，但是成本太高

### CAP理论(常用于分布式存储)
- C: 一致性，写入后立即读到新值
- A: 可用性，通常保障最终一致
- P: 分区容错性，分布式必须面对网络分区

### BASE理论 （常用于应用架构）
- 基本可用: 损失部分可用性，保证整体可用性
- 软状态: 允许状态同步延迟，不会影响系统即可
- 最终一致: 经过一段时间后，系统能够达到一致性

### 架构师能够做什么？
- 架构简化，减少有状态服务，秉承最终一致
- 架构折衷，确保异常可以被程序自我治愈

## 6-2 master-worker整体架构
### 整体架构
- Web控制台
- 无状态Master集群
- 保存任务（向下） 查询任务（向上）  查询日志（向上）
- Etcd任务存储                  MongoDB日志存储
- 同步任务（向下） 分布式锁（向上）  保存日志（向上)
- 可扩展Worker集群

### 架构思路
- 利用etcd同步全量任务列表到所有Worker节点
- 每个worker独立调度全量任务，无需与Master产生直接RPC
- 各个worker利用分布式锁抢占，解决并发调度相同任务的问题

## 6-3 master功能点与实现思路
### Master功能
- 任务管理HTTP接口: 新建、修改、查看、删除任务
- 任务日志HTTP接口: 查看任务执行历史日志
- 任务控制HTTP接口: 提供强制结束任务的接口
- 实现web管理界面: 基于jquery+bootstrap的Web控制台，前后端分离

### Master内部架构
- 任务管理Http接口    etcd /cron/jobs/
- 任务日志Http接口    mongodb
- 任务控制Http接口    etcd /cron/killer/

### 任务管理
- Etcd结构
```
/cron/jobs/任务名  -> {
    name,       //任务名
    command,    //shell命令
    cronExpr,   //cron表达式
}
```
- 保存到etcd的任务，会被实时同步到所有worker

### 任务日志
- MongoDB结构
```
{
    jobName,    //任务名
    command,    //shell命令
    err,        //执行报错
    output,     //执行输出
    startTime,  //开始时间
    endTime,    //结束时间
}
```
- 请求MongoDB,按任务名查看最近的执行日志

### 任务控制
- Etcd结构:

```
/cron/killer/任务名 -> ""
```
- worker监听/cron/killer/目录下put修改操作
- master将要结束的任务名put在/cron/killer/目录下，触发worker立即结束shell任务

## 6-4 Worker功能点与实现思路
### Worker功能
- 任务同步: 监听etcd中/cron/jobs/目录变化
- 任务调度: 基于cron表达式计算，触发过期任务
- 任务执行: 协程池并发执行多任务，基于etcd分布式锁抢占
- 日志保存: 捕获任务执行输出，保存到MongoDB
### Worker内部架构
- etcd /cron/jobs/
- 监听协程
- 任务变化
- cron调度协程
- 执行协程池
- 任务抢占
- etcd /cron/lock/
- 任务执行结果
- 存储日志
### 监听协程
- 利用watch API， 监听/cron/jobs/和/cron/killer/目录的变化
- 将变化事件通过channel推送给调度协程，更新内存中的任务信息

### 调度协程
- 监听任务变更event,更新内存中维护的任务列表
- 检查任务cron表达式，扫描到期任务，交给执行协程运行
- 监听任务控制event,强制中断正在执行中的子进程
- 监听任务执行result,更新内存中任务状态，投递执行日志

### 执行协程
- 在etcd中抢占分布式乐观锁: /cron/lock/任务名
- 抢占成功则通过Command类执行shell任务
- 捕获Command输出并等待子进程结束，将执行结果投递给调度协程

### 日志协程
- 监听调度发来的执行日志，放入一个batch中
- 对新batch启动定时器，超时未满自动提交
- 若batch被放满，那么立即提交，并取消自动提交定时器
