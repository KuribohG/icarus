Icarus
======

Icarus，简单来说，是一台刷课机。

工程结构
-------

### handler

handler 包主要用于建立 HTTP 服务器，处理用户的请求，即外部 API 部分。

### task

task 包主要管理目前的刷课任务。

一个*任务*指的是一个用户，外加一门课程或者一些课程的集合。任一课程选上时任务完成。例如「编号为 1, 2, 3 的课程」就是一个任务。

在实际操作中，任务会被分解为*子任务*。

*子任务*分为两种：

+ 登录请求，即根据某个用户名和密码登录选课系统，获取会话 ID。
+ 具体的选课请求，即选某一门课一次。例如「试选编号为 1 的课程」就是一个子任务。此类子任务会返回「选课成功」，「课程已满」和「出错」等结果。如果出错，则根据实际情况选择重启*任务*。

### dispatcher

dispatcher 包主要负责响应 Satellite 获取子任务的请求，将子任务分配给 Satellite；并从 Satellite 回收结果。

进入 dispatcher 包的子任务请求除了 task 中所述的两种之外，还有一种特殊的子任务：

+ 列出课程请求，根据某个用户名和密码，获取他/她可补选的课程。

### client

提供选课客户端（即与学校选课系统对接的部分）应有的抽象，并提供一个通用接口供 Satellite 调用。

#### - pku

提供北京大学选课客户端的具体实现，并将自身注册到通用客户端。

### cmd

提供可执行文件。

#### - icarus

启动 Icarus 服务器端，即接收用户刷课请求，维护刷课任务，分发刷课子任务的部分。

#### - icarus-satellite

启动 Icarus Satellite，用于调用客户端，执行子任务。

依赖
----

本项目现有如下依赖：

```
github.com/Sirupsen/logrus
github.com/julienschmidt/httprouter
github.com/thinxer/semikami
```

