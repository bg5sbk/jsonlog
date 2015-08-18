package jsonlog

import (
	"time"
)

var (
	DAY_SWITCHER   = daySwitch{}   // 按天切换文件
	HOURS_SWITCHER = hoursSwitch{} // 按小时切换文件
)

type Switcher interface {
	FirstSwitchTime() time.Duration
	NextSwitchTime() time.Duration
	DirAndFileName(base string) (dir, file string)
}

type daySwitch struct{}

func (_ daySwitch) FirstSwitchTime() time.Duration {
	// 到明天凌晨间隔多长时间
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location(),
	).Add(24 * time.Hour).Sub(now)
}

func (_ daySwitch) NextSwitchTime() time.Duration {
	return 24 * time.Hour
}

func (_ daySwitch) DirAndFileName(base string) (dir, file string) {
	now := time.Now()
	dir = base + "/" + now.Format("2006-01/")
	file = dir + now.Format("2006-01-02")
	return
}

type hoursSwitch struct{}

func (_ hoursSwitch) FirstSwitchTime() time.Duration {
	// 到下一个整点间隔多长时间
	now := time.Now()
	return time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), 0, 0, 0, now.Location(),
	).Add(time.Hour).Sub(now)
}

func (_ hoursSwitch) NextSwitchTime() time.Duration {
	return time.Hour
}

func (_ hoursSwitch) DirAndFileName(base string) (dir, file string) {
	now := time.Now()
	dir = base + "/" + now.Format("2006-01/2006-01-02/")
	file = dir + now.Format("2006-01-02_03")
	return
}
