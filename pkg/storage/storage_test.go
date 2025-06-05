package storage

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCacheHappyPath for basic flow of -> Set, Get, Delete
func TestCacheDefaultOptions(t *testing.T) {
	var (
		validate   = assert.New(t)
		cache, err = NewCache[string, string](DefaultCacheOptions)
	)
	defer cache.Close()

	validate.NoError(err)
	validate.NotNil(cache)
}

// TestCacheHappyPath for basic flow of -> Set, Get, Delete
func TestCacheHappyPath(t *testing.T) {
	var (
		validate = assert.New(t)
		cache, _ = NewCache[string, string](DefaultCacheOptions)
	)
	defer cache.Close()

	// Set -> Get
	cache.Set("1", "a")
	var (
		getVal, getErr = cache.Get("1")
		cacheSize      = cache.Size()
	)
	validate.NoError(getErr)
	validate.Equal("a", getVal)
	validate.Equal(1, cacheSize)

	// Delete
	cache.Delete("1")
	validate.Equal(0, cache.Size())

	_, getErr = cache.Get("1")
	validate.ErrorIs(getErr, ErrNotFound)
}

// TestCacheExpiry for garbage collection on expired items
func TestCacheExpiry(t *testing.T) {
	var (
		validate = assert.New(t)

		opts = &Options{
			ItemTTL:    100 * time.Millisecond,
			GCInterval: 1 * time.Second,
		}
		cache, _ = NewCache[string, string](opts)
	)
	defer cache.Close()

	// Add items
	cache.Set("key1", "value1")
	cache.SetWithTTL("key2", "value2", 2*time.Second)

	// wait - GC has not run yet
	time.Sleep(300 * time.Millisecond)
	var (
		val1, err1 = cache.Get("key1")
		val2, err2 = cache.Get("key2")
	)
	validate.ErrorIs(err1, ErrExpiredItem)
	validate.NoError(err2)

	validate.Empty(val1)
	validate.NotNil(val2)

	// wait -> GC ran a cycle
	time.Sleep(1 * time.Second)
	validate.False(cache.ContainsKey("key1"))
	validate.True(cache.ContainsKey("key2"))
	validate.Equal(1, cache.Size())

	// wait -> GC ran another cycle
	time.Sleep(1 * time.Second)
	validate.False(cache.ContainsKey("key1"))
	validate.False(cache.ContainsKey("key2"))
	validate.Equal(0, cache.Size())
}

// TestCacheConcurrentAccess basic thread safety
func TestCacheConcurrentAccess(t *testing.T) {
	var (
		validate = assert.New(t)

		opts = &Options{
			ItemTTL:    500 * time.Millisecond,
			GCInterval: 1 * time.Second,
		}
		cache, _ = NewCache[string, any](opts)
	)
	defer cache.Close()

	// seed some data
	for i := 0; i < 1000; i++ {
		go cache.Set(fmt.Sprintf("key-%d", i), i)
	}

	// run ops during GC
	for i := 0; i < 300; i++ {
		// randomly perform ops over (3s timeframe)
		if i%100 == 0 {
			time.Sleep(1 * time.Second)
		}

		var (
			key = fmt.Sprintf("key-%d", i)
		)
		if randomBool() {
			go cache.Delete(key)
		}
		if randomBool() {
			go cache.Set(key, i)
		}
		if randomBool() {
			go cache.Get(key)
		}
	}
	cache.Set("1", "a")

	// No panic – means test has passed!
	validate.NotZero(cache.Size())
}

// randomBool util for race-y test
func randomBool() bool {
	return rand.Intn(2) == 0
}
