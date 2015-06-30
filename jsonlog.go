package jsonlog

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type SwitchMode int

const (
	SWITCH_BY_DAY   SwitchMode = 1 // 按天切换文件
	SWITCH_BY_HOURS SwitchMode = 2 // 按小时切换文件
)

type M map[string]interface{}

type logFile struct {
	f       *os.File
	bufio   *bufio.Writer
	gzip    *gzip.Writer
	json    *json.Encoder
	changed bool
}

func openLogFile(fileName, fileType string, compress bool) (*logFile, error) {
	fullName := fileName + fileType
	if _, err := os.Stat(fullName); err == nil {
		os.Rename(fullName, fileName+".01"+fileType)
		fullName = fileName + ".02" + fileType
	} else if _, err := os.Stat(fileName + ".01" + fileType); err == nil {
		for fileId := 1; true; fileId++ {
			fullName = fileName + fmt.Sprintf(".%02d", fileId) + fileType
			if _, err := os.Stat(fullName); err != nil {
				break
			}
		}
	}
	f, err := os.OpenFile(fullName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}
	log := &logFile{f: f}
	if compress {
		log.bufio = bufio.NewWriter(log.f)
		log.gzip = gzip.NewWriter(log.bufio)
		log.json = json.NewEncoder(log.gzip)
	} else {
		log.bufio = bufio.NewWriter(log.f)
		log.json = json.NewEncoder(log.bufio)
	}
	return log, nil
}

func (file *logFile) Write(r M) {
	if err := file.json.Encode(r); err != nil {
		log.Println("log write failed:", err.Error())
	}
	file.changed = true
}

func (file *logFile) Flush() error {
	if !file.changed {
		return nil
	}
	if file.gzip != nil {
		if err := file.gzip.Flush(); err != nil {
			return err
		}
	}
	if err := file.bufio.Flush(); err != nil {
		return err
	}
	if err := file.f.Sync(); err != nil {
		return err
	}

	file.changed = false
	return nil
}

func (file *logFile) Close() error {
	if file.gzip != nil {
		if err := file.gzip.Flush(); err != nil {
			return err
		}
		if err := file.gzip.Close(); err != nil {
			return err
		}
	}
	if err := file.bufio.Flush(); err != nil {
		return err
	}
	if err := file.f.Sync(); err != nil {
		return err
	}
	return file.f.Close()
}

// 日志记录器
type L struct {
	dir       string
	logChan   chan M
	closeChan chan int
	closeWait sync.WaitGroup
	file      *logFile
}

// 新建一个日志记录器
func New(dir string, switchMode SwitchMode, fileType string, compress bool) (*L, error) {
	if compress {
		fileType += ".gz"
	}

	logger := &L{
		dir:       dir,
		closeChan: make(chan int),
		logChan:   make(chan M, 1000),
	}
	if err := logger.switchFile(switchMode, fileType, compress); err != nil {
		return nil, err
	}

	logger.closeWait.Add(1)
	go func() {
		var (
			fileTimer *time.Timer
			now       = time.Now()
		)
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

		// 每两秒刷新一次
		flushTicker := time.NewTicker(2 * time.Second)
		defer func() {
			flushTicker.Stop()
			logger.file.Close()
			logger.closeWait.Done()
		}()

		for {
			select {
			case r := <-logger.logChan:
				logger.file.Write(r)
			case <-flushTicker.C:
				if err := logger.file.Flush(); err != nil {
					log.Println("log flush failed:", err.Error())
				}
			case <-fileTimer.C:
				if err := logger.switchFile(switchMode, fileType, compress); err != nil {
					panic(err)
				}
				switch switchMode {
				case SWITCH_BY_DAY:
					fileTimer = time.NewTimer(24 * time.Hour)
				case SWITCH_BY_HOURS:
					fileTimer = time.NewTimer(time.Hour)
				}
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
	}()

	return logger, nil
}

// 切换文件
func (logger *L) switchFile(switchMode SwitchMode, fileType string, compress bool) error {
	var (
		dirName  string
		fileName string
		now      = time.Now()
	)

	// 确定目录名和文件名
	switch switchMode {
	case SWITCH_BY_DAY:
		dirName = logger.dir + "/" + now.Format("2006-01/")
		fileName = dirName + now.Format("2006-01-02")
	case SWITCH_BY_HOURS:
		dirName = logger.dir + "/" + now.Format("2006-01/2006-01-02/")
		fileName = dirName + now.Format("2006-01-02_03")
	}

	// 确认目录存在
	if err := os.MkdirAll(dirName, 0755); err != nil {
		return err
	}

	// 先关闭旧文件再切换
	if logger.file != nil {
		if err := logger.file.Close(); err != nil {
			return err
		}
	}

	// 创建或者打开已存在文件
	file, err := openLogFile(fileName, fileType, compress)
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
