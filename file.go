package jsonlog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type M map[string]interface{}

func fexists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

type File struct {
	mutex   sync.Mutex
	f       *os.File
	w       *bufio.Writer
	json    *json.Encoder
	changed bool
}

func NewFile(fileName, fileType string, writeBufferSize int) (*File, error) {
	fullName := fileName + ".01" + fileType

	for fileID := 2; fexists(fullName); fileID++ {
		fullName = fileName + fmt.Sprintf(".%02d", fileID) + fileType
	}

	f, err := os.OpenFile(fullName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	w := bufio.NewWriterSize(f, writeBufferSize)
	return &File{
		f:    f,
		w:    w,
		json: json.NewEncoder(w),
	}, nil
}

func (file *File) Write(r M) {
	file.mutex.Lock()
	defer file.mutex.Unlock()

	if err := file.json.Encode(r); err != nil {
		log.Println("jsonlog encode failed:", err.Error())
	}
	file.changed = true
}

func (file *File) Flush() error {
	file.mutex.Lock()
	defer file.mutex.Unlock()

	if !file.changed {
		return nil
	}

	if err := file.w.Flush(); err != nil {
		return err
	}

	if err := file.f.Sync(); err != nil {
		return err
	}

	file.changed = false
	return nil
}

func (file *File) Close() error {
	if err := file.Flush(); err != nil {
		return err
	}
	return file.f.Close()
}
