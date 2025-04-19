package limiter

import (
	"time"

	"go.uber.org/ratelimit"
)

type ConcurrencyLimiter struct {
	sem chan struct{}
}

func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		sem: make(chan struct{}, limit),
	}
}

func (l *ConcurrencyLimiter) Acquire() bool {
	select {
	case l.sem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (l *ConcurrencyLimiter) Release() {
	select {
	case <-l.sem:
	default:
	}
}

type RateLimiter struct {
	limiter ratelimit.Limiter
}

func NewRateLimiter(rps float64) *RateLimiter {
	return &RateLimiter{
		limiter: ratelimit.New(int(rps), ratelimit.Per(time.Second)),
	}
}

func (l *RateLimiter) Allow() bool {
	now := time.Now()
	next := l.limiter.Take()
	return next.Sub(now) <= 0
}

func (l *RateLimiter) Wait() {
	l.limiter.Take()
}
