package storage

import (
	"time"

	"github.com/garlandch/go-cache/pkg/storage/eviction"
	"github.com/pkg/errors"
)

// NewCache creates auto-cleaning cache w/ config
func NewCache(opts *Options) (*Cache, error) {
	// sanity check cache config
	err := opts.Validate()
	if err != nil {
		return nil, err
	}

	// plumb together instance
	var (
		cache = &Cache{
			entries: make(map[string]*cacheEntry),
			opts:    opts,
		}
	)

	// init garbage collection
	gc, err := eviction.NewGarbageCollector(opts.GCInterval, cache.runCleanUp)
	if err != nil {
		return nil, err
	}
	cache.gc = gc
	gc.Start()

	// ready to rumble!
	return cache, nil
}

// Set cache entry with default TTL
func (c *Cache) Set(key string, val any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		value:      val,
		expiration: time.Now().Add(c.opts.ItemTTL),
	}
}

// SetWithTTL cache entry with custom TTL
func (c *Cache) SetWithTTL(key string, val any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &cacheEntry{
		value:      val,
		expiration: time.Now().Add(ttl),
	}
}

// Get retrieves a value by key
// if no valid entry, then returns ErrNotFound or ErrExpiredItem
func (c *Cache) Get(key string) (any, error) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, errors.WithMessagef(ErrNotFound, "key: %s", key)
	}

	// delete expired cacheEntry and return err
	if entry.isExpired() {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, errors.WithMessagef(ErrExpiredItem, "key: %s, expiration: %s", key, entry.expiration)
	}

	return entry.value, nil
}

// Delete single entry from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Clear all entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// Keys returns all non-expired keys in the cache
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var (
		results = make([]string, 0, len(c.entries))
	)
	for key, val := range c.entries {
		if !val.isExpired() {
			results = append(results, key)
		}
	}
	return results
}

// ContainsKey checks for the existence of a non-expired entry
func (c *Cache) ContainsKey(key string) bool {
	_, err := c.Get(key)
	return err == nil
}

// Size returns total num of entries (count can include expired items)
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Close stops R/W actions and the async garbage collector
func (c *Cache) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.gc.Stop()
}

// runCleanUp implements interface for automatic garbage collection
func (c *Cache) runCleanUp() (int, error) {
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
