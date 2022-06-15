package ratelimit

import (
	"sync"
	"time"
)

type FixedWindowCounter struct {
	windowSize time.Duration
	quota      uint
	current    uint
	closed     chan struct{}
	mu         sync.Mutex
}

func NewFixedWindowCounter(windowSize time.Duration, quota uint) *FixedWindowCounter {
	fwc := &FixedWindowCounter{
		windowSize: windowSize,
		quota:      quota,
		closed:     make(chan struct{}),
	}

	go fwc.reset()

	return fwc
}

func (fwc *FixedWindowCounter) Request() bool {
	fwc.mu.Lock()
	defer fwc.mu.Unlock()

	if fwc.current < fwc.quota {
		fwc.current++
		return true
	}

	return false
}

func (tb *FixedWindowCounter) Close() {
	close(tb.closed)
}

func (fwc *FixedWindowCounter) reset() {
	ticker := time.NewTicker(fwc.windowSize)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fwc._reset()
		case <-fwc.closed:
			return
		}
	}
}

func (fwc *FixedWindowCounter) _reset() {
	fwc.mu.Lock()
	defer fwc.mu.Unlock()

	fwc.current = 0
}
