说明
====

jsonlog保存下来的json日志文件是一行一条json数据，当日志数据结构复杂的时候不方便肉眼阅读，所以我做了这个工具用来格式化json日志文件。

用法跟cat一样，可以一次输出多个文件也可以从标准输入读取数据并格式化输出。

用法示例：

```
logfmt file1.log file2.log

logfmt < file1.log
```

也可以配合gzdog命令一起使用：

```
gzdog file1.log.gz file2.log.gz | logfmt
```
