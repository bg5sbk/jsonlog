package jsonlog

import (
	"github.com/funny/unitest"
	"testing"
	"time"
)

func Test_SwitchByDay(t *testing.T) {
	log, err := New(Config{
		Dir:      ".",
		Switcher: DAY_SWITCHER,
		FileType: ".log",
		Compress: false,
	})
	unitest.NotError(t, err)
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Close()
}

func Test_SwitchByDay_Compress(t *testing.T) {
	log, err := New(Config{
		Dir:      ".",
		Switcher: DAY_SWITCHER,
		FileType: ".log",
		Compress: true,
	})
	unitest.NotError(t, err)
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
		Compress: false,
	})
	unitest.NotError(t, err)
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Close()
}

func Test_SwitchByHours_Compress(t *testing.T) {
	log, err := New(Config{
		Dir:      ".",
		Switcher: HOURS_SWITCHER,
		FileType: ".log",
		Compress: true,
	})
	unitest.NotError(t, err)
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Log(M{"Time": time.Now()})
	log.Close()
}
