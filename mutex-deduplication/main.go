package main

import (
	"fmt"
	"sync"
)

// work deduplication mech using mutexes (similar to singleflight)

type Cache struct {
	mu    sync.Mutex
	data  []byte
	locks map[string]*sync.Mutex // instances calling for same key will share the same mutex
}

func NewCache() *Cache {
	return &Cache{
		data:  nil,
		locks: make(map[string]*sync.Mutex),
	}
}

func (t *Cache) GetKeyLock(key string) *sync.Mutex {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.locks[key]; !ok {
		t.locks[key] = &sync.Mutex{}
	}
	return t.locks[key]
}

func (t *Cache) Get(key string) []byte {
	t.mu.Lock()
	if t.data != nil {
		t.mu.Unlock()
		return t.data
	}
	t.mu.Unlock()

	keylock := t.GetKeyLock(key)
	keylock.Lock() // only one gorountine acquires
	defer keylock.Unlock()

	// double check (as when keylock.Unlock() occurers, other gorountines may recompute cache data again.)
	if t.data != nil {
		return t.data
	}

	t.data = expensiveCalc()
	return t.data
}

func expensiveCalc() []byte {
	return []byte("some data")
}

// testing it
func main() {
	cache := NewCache()
	var wg sync.WaitGroup

	for range 10 { // 10 workers
		wg.Add(1)
		wg.Go(func() {
			data := cache.Get("someKey")
			fmt.Println(string(data))
		})
	}
	wg.Wait() // wait until workers finish before exitting
}
