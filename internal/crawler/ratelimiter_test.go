package crawler

import (
	"testing"
	"time"
)

func TestRateLimiterFirstRequestImmediate(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	stopCh := make(chan struct{})

	start := time.Now()
	ok := rl.Wait("example.com", stopCh)
	elapsed := time.Since(start)

	if !ok {
		t.Error("expected ok=true")
	}
	if elapsed > 20*time.Millisecond {
		t.Errorf("first request should be immediate, took %v", elapsed)
	}
}

func TestRateLimiterSecondRequestDelayed(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	stopCh := make(chan struct{})

	rl.Wait("example.com", stopCh)

	start := time.Now()
	ok := rl.Wait("example.com", stopCh)
	elapsed := time.Since(start)

	if !ok {
		t.Error("expected ok=true")
	}
	if elapsed < 50*time.Millisecond {
		t.Errorf("second request should be delayed, took only %v", elapsed)
	}
}

func TestRateLimiterDifferentDomainsIndependent(t *testing.T) {
	rl := NewRateLimiter(100 * time.Millisecond)
	stopCh := make(chan struct{})

	rl.Wait("example.com", stopCh)

	start := time.Now()
	ok := rl.Wait("other.com", stopCh)
	elapsed := time.Since(start)

	if !ok {
		t.Error("expected ok=true")
	}
	if elapsed > 20*time.Millisecond {
		t.Errorf("different domain should be immediate, took %v", elapsed)
	}
}

func TestRateLimiterCancellation(t *testing.T) {
	rl := NewRateLimiter(500 * time.Millisecond)
	stopCh := make(chan struct{})

	rl.Wait("example.com", stopCh)

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(stopCh)
	}()

	start := time.Now()
	ok := rl.Wait("example.com", stopCh)
	elapsed := time.Since(start)

	if ok {
		t.Error("expected ok=false after cancellation")
	}
	if elapsed > 200*time.Millisecond {
		t.Errorf("should have cancelled quickly, took %v", elapsed)
	}
}
