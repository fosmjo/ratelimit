package ratelimit

import (
	"sync"
	"time"
)

type TokenBucket struct {
	capacity     uint
	tokens       uint
	refillCount  uint
	refillPeriod time.Duration
	closed       chan struct{}
	mu           sync.Mutex
}

func NewTokenBucket(cap uint, refillCount uint, refillPeriod time.Duration) *TokenBucket {
	if refillCount > cap {
		refillCount = cap
	}

	tb := &TokenBucket{
		capacity:     cap,
		refillCount:  refillCount,
		refillPeriod: refillPeriod,
		closed:       make(chan struct{}),
	}

	go tb.refill()

	return tb
}

func (tb *TokenBucket) Request() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *TokenBucket) Close() {
	close(tb.closed)
}

func (tb *TokenBucket) refill() {
	ticker := time.NewTicker(tb.refillPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			tb.fill()
		case <-tb.closed:
			return
		}
	}
}

func (tb *TokenBucket) fill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	free := tb.capacity - tb.tokens
	delta := minUint(free, tb.refillCount)
	tb.tokens += delta
}

func minUint(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}
