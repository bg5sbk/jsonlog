package jsonlog

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
)

type SwitchMode int

const (
	SWITCH_BY_DAY   SwitchMode = 1 // 按天切换文件
	SWITCH_BY_HOURS SwitchMode = 2 // 按小时切换文件
)

type M map[string]interface{}

// 日志记录器
type L struct {
	dir       string
	logChan   chan M
	closeChan chan int
	out       *bufio.Writer
	encoder   *json.Encoder
	file      *os.File
}

// 新建一个日志记录器
func New(dir string, switchMode SwitchMode) (*L, error) {
	// 目录不存在就创建一个
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0644); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 日志切换的时间模式
	var fileTimer *time.Timer
	now := time.Now()
	switch switchMode {
	case SWITCH_BY_DAY:
		// 计算此刻到第二天零点的时间
		fileTimer = time.NewTimer(time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0, now.Location(),
		).Add(24 * time.Hour).Sub(now))
	case SWITCH_BY_HOURS:
		// 计算此刻到下一个小时的时间
		fileTimer = time.NewTimer(time.Date(
			now.Year(), now.Month(), now.Day(),
			now.Hour(), 0, 0, 0, now.Location(),
		).Add(time.Hour).Sub(now))
	}

	logger := &L{
		dir:       dir,
		closeChan: make(chan int),
		logChan:   make(chan M, 1000),
	}
	logger.switchFile(switchMode)

	go func() {
		// 每两秒刷新一次
		flushTicker := time.NewTicker(2 * time.Second)
		defer flushTicker.Stop()
		for {
			select {
			case r := <-logger.logChan:
				if err := logger.encoder.Encode(r); err != nil {
					log.Println("log failed:", err.Error())
				}
			case <-flushTicker.C:
				logger.out.Flush()
			case <-fileTimer.C:
				logger.switchFile(switchMode)
				switch switchMode {
				case SWITCH_BY_DAY:
					fileTimer = time.NewTimer(24 * time.Hour)
				case SWITCH_BY_HOURS:
					fileTimer = time.NewTimer(time.Hour)
				}
			case <-logger.closeChan:
				logger.out.Flush()
				logger.file.Close()
				return
			}
		}
	}()

	return logger, nil
}

// 切换文件
func (logger *L) switchFile(switchMode SwitchMode) {
	if logger.file != nil {
		logger.out.Flush()
		logger.file.Close()
	}

	var fileName string

	switch switchMode {
	case SWITCH_BY_DAY:
		fileName = logger.dir + "/" + time.Now().Format("2006-01-02") + ".log"
	case SWITCH_BY_HOURS:
		fileName = logger.dir + "/" + time.Now().Format("2006-01-02") + "/"
		if _, err := os.Stat(fileName); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(fileName, 0644); err != nil {
					panic(err)
				}
			} else {
				panic(err)
			}
		}
		fileName += strconv.Itoa(time.Now().Hour()) + ".log"
	}

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	logger.file = file
	logger.out = bufio.NewWriter(logger.file)
	logger.encoder = json.NewEncoder(logger.out)
}

// 关闭日志系统
func (logger *L) Close() {
	close(logger.closeChan)
}

// 在日志文件中输出信息
func (logger *L) Log(r M) {
	logger.logChan <- r
}
