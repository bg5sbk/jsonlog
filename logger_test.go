package jsonlog

import (
	"testing"
	"time"

	"github.com/funny/utest"
)

func Test_SwitchByDay(t *testing.T) {
	log, err := New(Config{
		Dir:      ".",
		Switcher: DAY_SWITCHER,
		FileType: ".log",
	})
	utest.IsNilNow(t, err)
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Close()
}

func Test_SwitchByHours(t *testing.T) {
	log, err := New(Config{
		Dir:      ".",
		Switcher: HOURS_SWITCHER,
		FileType: ".log",
	})
	utest.IsNilNow(t, err)
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Close()
}
