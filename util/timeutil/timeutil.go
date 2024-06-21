package timeutil

import (
	"sync/atomic"
	"time"
)

var currentTime int64

func init() {
	updateTime()
	go func() {
		for {
			updateTime()
			time.Sleep(time.Second)
		}
	}()
}

func updateTime() {
	now := time.Now().Unix()
	atomic.StoreInt64(&currentTime, now)
}

// High-performance acquisition of current time with accuracy in seconds
func NowSecond() int64 {
	return atomic.LoadInt64(&currentTime)
}
