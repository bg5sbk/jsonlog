介绍
====

基于jsonlog实现的系统信息日志模块，在指定目录记录日志文件，每天凌晨自动切换到新文件写入，日志以JSON格式记录，方便使用工具分析。

用法
====

log模块可以以全局的方式使用也可以以多个实例的方式使用。

全局化使用log模块需要在程序启动的时候调用`log.Init(dir)`来初始化全局日志系统，程序退出时调用`log.Close()`来关闭全局日志系统。

实例化的方式则是根据不同情况的需要，调用`log.New(dir)`来实例化日志记录器，不在使用时需要调用具体实例的`Close()`方法关闭记录器。

全局用法和实例用法接口是一致的，所以以下内容使用全局用法做示例。

记录日志的接口分为以下几种：

1. Info() -- 用于记录消息，通常是一些跟系统运行情况相关的数据，比如程序启动时间
2. Warn() -- 用于记录警告，通常是一些不影响业务但是需要注意的消息，比如数据库连接失败但重试成功
3. Error() -- 用于记录错误，通常是一些业务失败或操作失败，比如非法请求或请求处理失败
4. Debug() -- 用于记录调试信息，通常是开发时才关心的一些调试数据

以上接口都接受两个参数，第一个参数为所要记录的消息，第二个参数则是变长的日志数据列表，可以传任意多个对象，日志数据会一起被序列化为JSON格式。

在很多时候我们需要用到key-value的形式来记录日志数据，所以log模块内置了一个M类型，用来做这件事情。

M类型是一个用字符串做key，存放`interface{}`类型的map，所以它可以存放任何可以参与JSON序列化的数据。

以下是一些用法示例：

```go
log.Info("Hello World")

log.Info("This is data", 123, 456, "string data")

log.Info("This is key-value", log.M{
	"error": error.Error(),
	"user": username,
	"email": email,
})
```

有一些项目需要对不同模块的日志做划分，可以为每个模块的日志单独实例化日志系统：

```go
logger1 := log.New("dir1", true)
logger2 := log.New("dir2", true)
```

在平时我们需要关闭生产环境上的调试信息输出，在出现一些异常情况时我们又会需要动态开启，所以log模块可以动态设置调试信息的输出与否：

```go
log.SetDebug(true)

logger1 := log.New("dir1", true)
logger1.SetDebug(true)
```