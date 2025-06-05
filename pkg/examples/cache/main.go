package main

import (
	"fmt"
	"log"

	"github.com/garlandch/go-cache/pkg/storage"
)

type sampleData struct {
	foo string
	bar int
}

func main() {
	// initialize
	cache, err := storage.NewCache[string, sampleData](storage.DefaultCacheOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close() // stop GC

	// [Sample Usage] - Set + Get (w/ type unboxing via generics)
	cache.Set("someUUID", sampleData{foo: "bar", bar: 1})

	data, err := cache.Get("someUUID")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(data.foo, data.bar)
}
