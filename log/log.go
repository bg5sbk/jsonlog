package log

import (
	"github.com/funny/jsonlog"
	"time"
)

var gLogger *L

// 初始化全局日志
func Init(dir string) {
	l, err := New(dir)
	if err != nil {
		panic(err)
	}
	gLogger = l
}

// 关闭全局日志系统
func Close() {
	gLogger.Close()
}

// 在全局日志中输出信息
func Info(msg string, data ...interface{}) {
	gLogger.Info(msg, data...)
}

// 在全局日志中输出警告信息
func Warn(msg string, data ...interface{}) {
	gLogger.Warn(msg, data...)
}

// 在全局日志中输出错误信息
func Error(msg string, data ...interface{}) {
	gLogger.Error(msg, data...)
}

// 在全局日志中输出调试信息
func Debug(msg string, data ...interface{}) {
	gLogger.Debug(msg, data...)
}

// 全局日志开启或关闭调试信息的输出
func SetDebug(debug bool) {
	gLogger.SetDebug(debug)
}

type M map[string]interface{}

// 日志记录器
type L struct {
	l     *jsonlog.L
	debug bool
}

// 新建一个日志记录器
func New(dir string) (*L, error) {
	l, err := jsonlog.New(jsonlog.Config{
		Dir:      dir,
		Switcher: jsonlog.DAY_SWITCHER,
		FileType: ".log",
	})
	if err != nil {
		return nil, err
	}
	return &L{l, true}, nil
}

// 关闭日志系统
func (logger *L) Close() {
	logger.l.Close()
}

func (logger *L) Log(msg string, typ string, data ...interface{}) {
	m := jsonlog.M{
		"Time":    time.Now().UnixNano(),
		"Type":    typ,
		"Message": msg,
	}
	if data != nil {
		m["Data"] = data
	}
	logger.l.Log(m)
}

// 在日志文件中输出信息
func (logger *L) Info(msg string, data ...interface{}) {
	logger.Log(msg, "info", data...)
}

// 在日志文件中输出警告信息
func (logger *L) Warn(msg string, data ...interface{}) {
	logger.Log(msg, "warn", data...)
}

// 在日志文件中输出错误信息
func (logger *L) Error(msg string, data ...interface{}) {
	logger.Log(msg, "error", data...)
}

// 在日志文件中输出调试信息
func (logger *L) Debug(msg string, data ...interface{}) {
	if logger.debug {
		logger.Log(msg, "debug", data...)
	}
}

// 开启或关闭调试信息的输出
func (logger *L) SetDebug(debug bool) {
	logger.debug = debug
}
