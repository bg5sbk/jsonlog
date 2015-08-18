package jsonlog

import (
	"time"
)

var (
	SWITCH_BY_DAY   = switchByDay{}   // 按天切换文件
	SWITCH_BY_HOURS = switchByHours{} // 按小时切换文件
)

type Switcher interface {
	FirstSwitchTime() time.Duration
	NextSwitchTime() time.Duration
	DirAndFileName(base string) (dir, file string)
}

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
