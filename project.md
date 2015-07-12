Icarus
======

Icarus，简单来说，是一台刷课机。

工程结构
-------

### handler

handler 包主要用于建立 HTTP 服务器，处理用户的请求，即外部 API 部分。

### task

task 包主要管理目前的刷课任务。

一个*任务*指的是一门课程或者一些课程的集合，以及一个任务完成条件。例如「编号为 1, 2, 3 的课程，(1 or 2 or 3) 选上时退出」就是一个任务。

### dispatcher

dispatcher 包主要用于将任务分解为选课子任务，并维护子任务队列；同时负责响应 Satellite 获取子任务的请求，将子任务分配给 Satellite。

*子任务*分为两种：

+ 登录请求，即根据某个用户名和密码登录选课系统，获取会话 ID。
+ 具体的选课请求，即选某一门课一次。例如「试选编号为 1 的课程」就是一个子任务。此类子任务会返回「选课成功」，「课程已满」和「出错」等结果。如果出错，则根据实际情况选择重启*任务*。

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

