package ratelimit

import (
	"math"
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	windowSize time.Duration
	quota      uint
	clock      Colck

	lastWindowStart    int64
	lastWindowReqCount int

	currentWindowStart    int64
	currentWindowReqCount int

	mu sync.Mutex
}

func NewSlidingWindowCounter(windowSize time.Duration, quota uint, clock Colck) *SlidingWindowCounter {
	now := clock.Now().UnixNano()

	return &SlidingWindowCounter{
		windowSize:         windowSize,
		quota:              quota,
		lastWindowStart:    now - int64(windowSize),
		currentWindowStart: now,
		clock:              clock,
	}
}

func (swc *SlidingWindowCounter) Request() bool {
	now := swc.clock.Now().UnixNano()
	windowStart := now - int64(swc.windowSize)

	swc.mu.Lock()
	defer swc.mu.Unlock()

	if windowStart > swc.currentWindowStart {
		swc.lastWindowStart = swc.currentWindowStart
		swc.lastWindowReqCount = swc.currentWindowReqCount
		swc.currentWindowStart = windowStart
		swc.currentWindowReqCount = 0
	}

	swc.currentWindowReqCount++

	rollingAndPreviousWindowOverlap := int64(swc.windowSize) - (now - swc.currentWindowStart)
	slidingWindowCount := swc.currentWindowReqCount + int(math.Round(
		float64(swc.lastWindowReqCount)*float64(rollingAndPreviousWindowOverlap)/float64(swc.windowSize)),
	)

	return slidingWindowCount <= int(swc.quota)
}
