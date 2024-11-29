//////////////////////////////////////////////////////////////////////
//
// Given is some code to cache key-value pairs from a database into
// the main memory (to reduce access time). Note that golang's map are
// not entirely thread safe. Multiple readers are fine, but multiple
// writers are not. Change the code to make this thread safe.
//

package main

import (
	"container/list"
	"sync"
	"testing"
)

// CacheSize determines how big the cache can grow
const CacheSize = 100

// KeyStoreCacheLoader is an interface for the KeyStoreCache
type KeyStoreCacheLoader interface {
	// Load implements a function where the cache should gets it's content from
	Load(string) string
}

type page struct {
	Key   string
	Value string
}

// KeyStoreCache is a LRU cache for string key-value pairs
type KeyStoreCache struct {
	mu    sync.RWMutex
	cache map[string]*list.Element
	pages list.List
	load  func(string) string
}

// New creates a new KeyStoreCache
func New(load KeyStoreCacheLoader) *KeyStoreCache {
	return &KeyStoreCache{
		load:  load.Load,
		cache: make(map[string]*list.Element),
	}
}

// Get gets the key from cache, loads it from the source if needed
func (k *KeyStoreCache) Get(key string) string {
	// Try to retrieve from cache using a read lock
	k.mu.RLock()
	val, ok := k.cache[key]
	if ok {
		k.pages.MoveToFront(val)
		value := val.Value.(*page).Value
		k.mu.RUnlock()
		return value
	}
	k.mu.RUnlock()

	// Cache miss - load from the database
	k.mu.Lock()
	defer k.mu.Unlock()

	// Re-check the cache after acquiring the write lock in case another
	// goroutine already added the key
	val, ok = k.cache[key]
	if ok {
		k.pages.MoveToFront(val)
		return val.Value.(*page).Value
	}

	valueString := k.load(key)

	// Evict least-used item if cache is full
	if len(k.cache) >= CacheSize {
		back := k.pages.Back()
		if back != nil {
			delete(k.cache, back.Value.(*page).Key)
			k.pages.Remove(back)
		}
	}

	// Add new item to cache
	pageEntry := &page{Key: key, Value: valueString}
	k.pages.PushFront(pageEntry)
	k.cache[key] = k.pages.Front()

	return valueString
}

// Loader implements KeyStoreLoader
type Loader struct {
	DB *MockDB
}

// Load gets the data from the database
func (l *Loader) Load(key string) string {
	val, err := l.DB.Get(key)
	if err != nil {
		panic(err)
	}

	return val
}

func run(t *testing.T) (*KeyStoreCache, *MockDB) {
	loader := Loader{
		DB: GetMockDB(),
	}
	cache := New(&loader)

	RunMockServer(cache, t)

	return cache, loader.DB
}

func main() {
	run(nil)
}
