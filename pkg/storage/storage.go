package storage

import (
	"time"

	"github.com/garlandch/go-cache/pkg/storage/eviction"
	"github.com/pkg/errors"
)

// NewCache creates auto-cleaning cache w/ config
func NewCache[K comparable, V any](opts *Options) (*Cache[K, V], error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	cache := &Cache[K, V]{
		entries: make(map[K]*cacheEntry[V]),
		opts:    opts,
	}

	gc, err := eviction.NewGarbageCollector(opts.GCInterval, cache.runCleanUp)
	if err != nil {
		return nil, err
	}
	cache.gc = gc
	gc.Start()

	return cache, nil
}

// Set cache entry with default TTL
func (c *Cache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry[V]{
		value:      val,
		expiration: time.Now().Add(c.opts.ItemTTL),
	}
}

// SetWithTTL cache entry with custom TTL
func (c *Cache[K, V]) SetWithTTL(key K, val V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry[V]{
		value:      val,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves a value by key
func (c *Cache[K, V]) Get(key K) (V, error) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	var (
		zeroValue V
	)
	if !exists {
		return zeroValue, errors.WithMessagef(ErrNotFound, "key: %v", key)
	}

	if entry.isExpired() {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return zeroValue, errors.WithMessagef(ErrExpiredItem, "key: %v, expiration: %s", key, entry.expiration)
	}

	return entry.value, nil
}

// Delete single entry from cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear all entries
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[K]*cacheEntry[V])
}

// Keys returns all non-expired keys
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []K
	for k, v := range c.entries {
		if !v.isExpired() {
			keys = append(keys, k)
		}
	}
	return keys
}

// ContainsKey checks for the existence of a non-expired entry
func (c *Cache[K, V]) ContainsKey(key K) bool {
	_, err := c.Get(key)
	return err == nil
}

// Size returns the total number of entries (including expired)
func (c *Cache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Close stops the garbage collector
func (c *Cache[K, V]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gc.Stop()
}

// runCleanUp purges expired items
func (c *Cache[K, V]) runCleanUp() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var (
		now          = time.Now()
		itemsDeleted = 0
	)
	for key, item := range c.entries {
		if now.After(item.expiration) {
			delete(c.entries, key)
			itemsDeleted++
		}
	}
	return itemsDeleted, nil
}
