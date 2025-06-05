# go-cache

Simple in-memory, auto-expiring key-value store 

---

## ✨ Features
- Thread-safe
- Automatic Cleanup w/ background Garbage Collector
- Per-entry TTL (time-to-live)
- Generics for Key/Value Typing

---

## 🚀 Example Usage

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/garlandch/go-cache/pkg/storage"
)

func useCache(cache *storage.Cache[string, string]) {
	// Set
	cache.Set("foo", "bar")
	cache.SetWithTTL("foo2", "bar2", 1*time.Minute)

	// Get
	val, err := cache.Get("foo")
	fmt.Println(val, err)

	// Misc
	fmt.Println(cache.Size())
	fmt.Println(cache.Keys())
	fmt.Println(cache.ContainsKey("foo"))

	// Delete
	cache.Delete("foo")
	cache.Clear()
}

func main() {
	// initialize
	var (
		opts = &storage.Options{
			ItemTTL:    5 * time.Second,
			GCInterval: 1 * time.Second,
		}
	)
	cache, err := storage.NewCache[string, string](opts) // note – can use `storage.DefaultCacheOptions`
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close() // stop GC

	// sample usage
	useCache(cache)
}
```
