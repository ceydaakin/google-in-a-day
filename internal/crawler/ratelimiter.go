package crawler

import (
	"sync"
	"time"
)

// RateLimiter enforces per-domain request spacing.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[string]time.Time // host -> last request time
	interval time.Duration
}

// NewRateLimiter creates a rate limiter with the given minimum interval between
// requests to the same domain.
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]time.Time),
		interval: interval,
	}
}

// Wait blocks until a request to the given host is allowed.
// Returns false if stopCh is closed during the wait (caller should abort).
func (rl *RateLimiter) Wait(host string, stopCh <-chan struct{}) bool {
	rl.mu.Lock()
	lastReq := rl.limiters[host]
	now := time.Now()
	elapsed := now.Sub(lastReq)

	if elapsed >= rl.interval {
		rl.limiters[host] = now
		rl.mu.Unlock()
		return true
	}

	remaining := rl.interval - elapsed
	rl.limiters[host] = now.Add(remaining)
	rl.mu.Unlock()

	// Wait for the remaining interval, but allow cancellation
	timer := time.NewTimer(remaining)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-stopCh:
		return false
	}
}
