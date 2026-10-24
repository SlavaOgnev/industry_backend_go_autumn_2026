package main

import (
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	var last time.Time
	if clock != nil {
		last = clock.Now()
	}
	return &Limiter{mu: sync.Mutex{}, clock: clock, rate: ratePerSec, burst: burst, tokens: float64(burst), last: last}
}
func (l *Limiter) AllowN(n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.clock == nil || n <= 0 || n > l.burst {
		return false
	}
	timeNow := l.clock.Now()
	deferince := timeNow.Sub(l.last)

	if l.rate > 0 && deferince > 0 {
		newTokens := l.tokens + l.rate*deferince.Seconds()
		l.tokens = min(float64(l.burst), newTokens)
	}
	if deferince > 0 {

		l.last = timeNow
	}
	if float64(n) <= l.tokens {
		l.tokens -= float64(n)
		return true
	}
	return false
}
