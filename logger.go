package jsonlog

import (
	"log"
	"os"
	"sync"
	"sync/atomic"
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
	closeWait sync.WaitGroup
	closeFlag uint32
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
		config:  config,
		logChan: make(chan M, config.LogChanSize),
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
		if logger.file != nil {
			logger.file.Close()
		}
		logger.closeWait.Done()
	}()

	// 定时切换文件
	switchTimer := time.NewTimer(logger.config.Switcher.FirstSwitchTime())

	// 定时刷新文件
	flushTicker := time.NewTicker(logger.config.FlushTick)
	defer flushTicker.Stop()

	for {
		select {
		case r, ok := <-logger.logChan:
			if ok {
				logger.file.Write(r)
			} else {
				for r := range logger.logChan {
					logger.file.Write(r)
				}
				return
			}
		case <-flushTicker.C:
			if err := logger.file.Flush(); err != nil {
				log.Println("jsonlog flush failed:", err.Error())
				panic(err)
			}
		case <-switchTimer.C:
			if err := logger.switchFile(); err != nil {
				log.Println("jsonlog switch failed:", err.Error())
				panic(err)
			}
			switchTimer.Reset(logger.config.Switcher.NextSwitchTime())
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
	if atomic.CompareAndSwapUint32(&logger.closeFlag, 0, 1) {
		close(logger.logChan)
		logger.closeWait.Wait()
	}
}

// 在日志文件中输出信息
func (logger *L) Log(r M) {
	if atomic.LoadUint32(&logger.closeFlag) == 0 {
		logger.logChan <- r
	}
}
