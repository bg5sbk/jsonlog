package jsonlog

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type M map[string]interface{}

type LogFile struct {
	mutex   sync.Mutex
	f       *os.File
	bufio   *bufio.Writer
	gzip    *gzip.Writer
	json    *json.Encoder
	changed bool
}

func NewLogFile(fileName, fileType string, compress bool) (*LogFile, error) {
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
	log := &LogFile{f: f}
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

func (file *LogFile) Write(r M) {
	file.mutex.Lock()
	defer file.mutex.Unlock()

	if err := file.json.Encode(r); err != nil {
		log.Println("log write failed:", err.Error())
	}
	file.changed = true
}

func (file *LogFile) flush(isClose bool) error {
	file.mutex.Lock()
	defer file.mutex.Unlock()

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

func (file *LogFile) Flush() error {
	return file.flush(false)
}

func (file *LogFile) Close() error {
	if err := file.flush(true); err != nil {
		return err
	}
	return file.f.Close()
}
