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

type SwitchMode interface {
	FirstSwitchTime() time.Duration
	NextSwitchTime() time.Duration
	DirAndFileName(base string) (dir, file string)
}

var (
	SWITCH_BY_DAY   = switchByDay{}   // 按天切换文件
	SWITCH_BY_HOURS = switchByHours{} // 按小时切换文件
)

type switchByDay struct{}

func (_ switchByDay) FirstSwitchTime() time.Duration {
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location(),
	).Add(24 * time.Hour).Sub(now)
}

func (_ switchByDay) NextSwitchTime() time.Duration {
	return 24 * time.Hour
}

func (_ switchByDay) DirAndFileName(base string) (dir, file string) {
	now := time.Now()
	dir = base + "/" + now.Format("2006-01/")
	file = dir + now.Format("2006-01-02")
	return
}

type switchByHours struct{}

func (_ switchByHours) FirstSwitchTime() time.Duration {
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), 0, 0, 0, now.Location(),
	).Add(time.Hour).Sub(now)
}

func (_ switchByHours) NextSwitchTime() time.Duration {
	return time.Hour
}

func (_ switchByHours) DirAndFileName(base string) (dir, file string) {
	now := time.Now()
	dir = base + "/" + now.Format("2006-01/2006-01-02/")
	file = dir + now.Format("2006-01-02_03")
	return
}

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

func (file *logFile) flush(isClose bool) error {
	if !file.changed {
		return nil
	}
	if file.gzip != nil {
		if err := file.gzip.Flush(); err != nil {
			return err
		}
		if isClose {
			if err := file.gzip.Close(); err != nil {
				return err
			}
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

func (file *logFile) Flush() error {
	return file.flush(false)
}

func (file *logFile) Close() error {
	if err := file.flush(true); err != nil {
		return err
	}
	return file.f.Close()
}

type Config struct {
	Dir        string
	SwitchMode SwitchMode
	FileType   string
	Compress   bool
	FlushTick  time.Duration
}

// 日志记录器
type L struct {
	config    Config
	logChan   chan M
	closeChan chan int
	closeWait sync.WaitGroup
	file      *logFile
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

	logger := &L{
		config:    config,
		closeChan: make(chan int),
		logChan:   make(chan M, 1000),
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

	fileTimer := time.NewTimer(logger.config.SwitchMode.FirstSwitchTime())

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
			fileTimer.Reset(logger.config.SwitchMode.NextSwitchTime())
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
	dir, fileName := logger.config.SwitchMode.DirAndFileName(logger.config.Dir)

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
	file, err := openLogFile(fileName, logger.config.FileType, logger.config.Compress)
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
