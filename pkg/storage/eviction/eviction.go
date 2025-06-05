package eviction

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrInvalidConfig = fmt.Errorf("bad GarbageCollector config")
)

// GarbageCollector manages automatic cache eviction
type GarbageCollector struct {
	interval  time.Duration
	cleanFunc func() (int, error)

	// go-routine controls
	stopCleanup chan struct{}
	stopped     chan struct{}
	onceStop    sync.Once

	// isRunning status
	mu        sync.RWMutex
	isRunning bool
}

// NewGarbageCollector with the given interval and cleanup function.
func NewGarbageCollector(interval time.Duration, cleanupFunc func() (int, error)) (*GarbageCollector, error) {
	if interval <= 0 {
		return nil, errors.WithMessage(ErrInvalidConfig, "interval must be >0")
	}

	return &GarbageCollector{
		interval:  interval,
		cleanFunc: cleanupFunc,

		stopCleanup: make(chan struct{}),
		stopped:     make(chan struct{}),
	}, nil
}

// Start automatic garbage collection!
func (e *GarbageCollector) Start() {
	e.mu.Lock()
	e.isRunning = true
	e.mu.Unlock()

	fmt.Printf("GarbageCollector: [%s] starting automatic clean-up...\n", time.Now())
	go func() {
		ticker := time.NewTicker(e.interval)
		defer ticker.Stop()
		defer close(e.stopped)

		for {
			select {
			case <-ticker.C:
				e.run()
			case <-e.stopCleanup:
				e.mu.Lock()
				e.isRunning = false
				e.mu.Unlock()
				fmt.Printf("GarbageCollector: [%s] stopped running...\n", time.Now())
				return
			}
		}
	}()
}

// Stop garbage collection
// note - onceStop prevents panics if this func is called multiple times
func (e *GarbageCollector) Stop() {
	e.onceStop.Do(func() {
		close(e.stopCleanup)
		<-e.stopped
	})
}

func (e *GarbageCollector) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isRunning
}

// run GC task w/ logging for visibility
func (e *GarbageCollector) run() {
	itemsDeleted, err := e.cleanFunc()
	if err != nil {
		fmt.Printf("GarbageCollector: [%s] - error: `%s`\n", time.Now(), err)
		return
	}
	fmt.Printf("GarbageCollector: [%s] - ran clean up - %d items deleted\n", time.Now(), itemsDeleted)
}
