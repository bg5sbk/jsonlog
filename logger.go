package jsonlog

import (
	"log"
	"os"
	"sync"
	"time"
)

type Config struct {
	Dir         string
	Switcher    Switcher
	FileType    string
	Compress    bool
	FlushTick   time.Duration
	LogChanSize int
}

// 日志记录器
type L struct {
	config    Config
	logChan   chan M
	closeChan chan int
	closeWait sync.WaitGroup
	file      *File
}

// 新建一个日志记录器
func New(config Config) (*L, error) {
	if config.FileType[0] != '.' {
		config.FileType = "." + config.FileType
	}

	if config.Compress {
		config.FileType += ".gz"
	}

	if config.FlushTick <= 0 {
		config.FlushTick = 2 * time.Second
	}

	if config.LogChanSize <= 0 {
		config.LogChanSize = 2000
	}

	logger := &L{
		config:    config,
		closeChan: make(chan int),
		logChan:   make(chan M, config.LogChanSize),
	}

	if err := logger.switchFile(); err != nil {
		return nil, err
	}

	logger.closeWait.Add(1)
	go logger.loop()

	return logger, nil
}

func (logger *L) loop() {
	defer func() {
		logger.closeWait.Done()
		if logger.file != nil {
			logger.file.Close()
		}
	}()

	fileTimer := time.NewTimer(logger.config.Switcher.FirstSwitchTime())

	// 定时刷新
	flushTicker := time.NewTicker(logger.config.FlushTick)
	defer flushTicker.Stop()

	for {
		select {
		case r := <-logger.logChan:
			logger.file.Write(r)
		case <-flushTicker.C:
			if err := logger.file.Flush(); err != nil {
				log.Println("log flush failed:", err.Error())
			}
		case <-fileTimer.C:
			if err := logger.switchFile(); err != nil {
				println("jsonlog switch file failed: " + err.Error())
				panic(err)
			}
			fileTimer.Reset(logger.config.Switcher.NextSwitchTime())
		case <-logger.closeChan:
			for {
				select {
				case r := <-logger.logChan:
					logger.file.Write(r)
				default:
					return
				}
			}
		}
	}
}

// 切换文件
func (logger *L) switchFile() error {
	dir, fileName := logger.config.Switcher.DirAndFileName(logger.config.Dir)

	// 确认目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 先关闭旧文件再切换
	if logger.file != nil {
		if err := logger.file.Close(); err != nil {
			return err
		}
	}

	// 创建或者打开已存在文件
	file, err := NewFile(fileName, logger.config.FileType, logger.config.Compress)
	if err != nil {
		return err
	}
	logger.file = file

	return nil
}

// 关闭日志系统
func (logger *L) Close() {
	close(logger.closeChan)
	logger.closeWait.Wait()
}

// 在日志文件中输出信息
func (logger *L) Log(r M) {
	logger.logChan <- r
}
