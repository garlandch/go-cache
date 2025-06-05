package eviction

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGarbageCollectorStop ensures that the go-routine running the cleanFunc closes
func TestGarbageCollectorStop(t *testing.T) {
	var (
		validate = assert.New(t)

		// dummy cleanUp task
		cleanFunc = func() (int, error) {
			return 0, nil
		}
		gcInterval    = 50 * time.Millisecond
		gc, gcInitErr = NewGarbageCollector(gcInterval, cleanFunc)
	)
	validate.NoError(gcInitErr)

	// Let GC task run a few times
	gc.Start()
	time.Sleep(500 * time.Millisecond)

	// Shutdown
	gc.Stop()
	time.Sleep(1 * time.Second)
	validate.False(gc.IsRunning())
}
