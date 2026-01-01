package util

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"
)

// some func that fetches a value given a key
// used by the cache to get values when not present
// for example fetching awards for a team from the API
type FetchFunc[V any] func(key string) (V, error)

// Cache stores values for a certain duration and fetches them using a predefined function when not present
type Cache[V any] struct {
	// lru cache  that expires with time
	// while I'd hoped I could have the keys be generic, this doesn't work with singleflight so it's string keys
	lru *expirable.LRU[string, V]

	// this stops duplicate fetches for the same team bc if two people press the button at the same time it would make two requests which is dumb
	flight singleflight.Group

	// btw this is bc the functions are goroutines so we don't want race conditions
	// mutex locks cache when r/w
	mu  *sync.Mutex

	// function that fetches value when not present in cache
	fetch FetchFunc[V]
}

// NewCache creates a new Cache with the given max size, persistence duration, and fetch function
func NewCache[V any](maxSize int, persistenceDuration time.Duration, fetch FetchFunc[V]) *Cache[V] {
	return &Cache[V]{
		lru:   expirable.NewLRU[string, V](maxSize, nil, persistenceDuration),
		flight: singleflight.Group{},
		mu:    &sync.Mutex{},
		fetch: fetch,
	}
}

// Get gets the value for the given key from the cache, fetching it if not present
func (c *Cache[V]) Get(key string) (V, error) {
	c.mu.Lock()
	if val, exists := c.lru.Get(key); exists {
		c.mu.Unlock()
		return val, nil
	}
	c.mu.Unlock()

	// note: we let go of the lock while fetching to avoid blocking other operations
	// also here we basically index by team num, so if it sees one team num is there it doesn't repeat the request
	result, err, _ := c.flight.Do(key, func() (any, error) {
        return c.fetch(key)
    })
	if err != nil {
		var zero V
		return zero, err
	}

	val := result.(V)
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.lru.Add(key, val)
	
	return val, nil
}

func (c *Cache[V]) Set(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Add(key, value)
}
