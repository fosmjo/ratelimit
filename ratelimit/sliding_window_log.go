package ratelimit

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	windowSize time.Duration
	quota      uint
	clock      Colck
	timestamps []int64 // unix nanoseconds
	mu         sync.Mutex
}

func NewSlidingWindowLog(windowSize time.Duration, quota uint, clock Colck) *SlidingWindowLog {
	swl := &SlidingWindowLog{
		windowSize: windowSize,
		quota:      quota,
		clock:      clock,
		timestamps: make([]int64, quota),
	}

	return swl
}

func (swl *SlidingWindowLog) Request() bool {
	now := swl.clock.Now().UnixNano()
	windowStart := now - int64(swl.windowSize)

	swl.mu.Lock()
	defer swl.mu.Unlock()

	swl.timestamps = append(swl.timestamps, now)

	windowStartIndex := 0
	for i := len(swl.timestamps) - 1; i >= 0; i-- {
		if swl.timestamps[i] <= windowStart {
			windowStartIndex = i
			break
		}
	}

	swl.timestamps = swl.timestamps[windowStartIndex+1:]

	return len(swl.timestamps) <= int(swl.quota)
}
