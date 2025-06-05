package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrNotFound      = errors.New("no cached entry")
	ErrExpiredItem   = errors.New("entry has expired")
	ErrInvalidConfig = errors.New("invalid config value")

	DefaultItemTTL    = 2 * time.Minute
	DefaultGCInterval = 30 * time.Second

	DefaultCacheOptions = &Options{
		ItemTTL:    DefaultItemTTL,
		GCInterval: DefaultGCInterval,
	}
)

// Cache represents the in-memory cache with expiration
type Cache struct {
	entries map[string]*cacheEntry
	opts    *Options

	gc garbageCollector
	mu sync.RWMutex
}

// cacheEntry contains values of a single entry
type cacheEntry struct {
	value      any
	expiration time.Time
}

// garbageCollector implements cache eviction logic
type garbageCollector interface {
	Start()
	Stop()
	IsRunning() bool
}

// Options config for auto-cleaning cache
type Options struct {
	ItemTTL    time.Duration
	GCInterval time.Duration
}

// Validate sanity checks config values, and sets defaults values if not set
func (o *Options) Validate() error {
	// check for valid config
	if err := o.sanityCheck(); err != nil {
		return errors.WithMessage(ErrInvalidConfig, err.Error())
	}

	// set defaults
	o.backfillDefaults()
	return nil
}

// sanityCheck cache config values
func (o *Options) sanityCheck() error {
	// negative durations
	if o.ItemTTL < 0 {
		return fmt.Errorf("cacheEntry TTL must be positive: %d", o.ItemTTL)
	} else if o.GCInterval < 0 {
		return fmt.Errorf("GC interval must be positive: %d", o.GCInterval)
	}
	return nil
}

// backfillDefaults if not set
func (o *Options) backfillDefaults() {
	if o.ItemTTL == 0 {
		o.ItemTTL = DefaultItemTTL
	}

	if o.GCInterval == 0 {
		o.GCInterval = DefaultGCInterval
	}
}

// isExpired helper utility for cache entry
func (ce *cacheEntry) isExpired() bool {
	return time.Now().After(ce.expiration)
}
