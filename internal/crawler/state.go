package crawler

import "sync/atomic"

// CrawlState represents the lifecycle state of the crawler.
type CrawlState int32

const (
	StateIdle      CrawlState = iota // Not started
	StateRunning                     // Actively crawling
	StatePaused                      // Paused by user
	StateStopped                     // Stopped by user
	StateCompleted                   // Finished naturally
)

// String returns a human-readable state name.
func (s CrawlState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateStopped:
		return "stopped"
	case StateCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

// atomicState provides atomic CrawlState operations.
type atomicState struct {
	val int32
}

func (a *atomicState) Load() CrawlState {
	return CrawlState(atomic.LoadInt32(&a.val))
}

func (a *atomicState) Store(s CrawlState) {
	atomic.StoreInt32(&a.val, int32(s))
}

func (a *atomicState) CompareAndSwap(old, new CrawlState) bool {
	return atomic.CompareAndSwapInt32(&a.val, int32(old), int32(new))
}
