// Package pokecache provides a small in-memory cache whose entries expire.
package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

// Cache is a concurrency-safe in-memory cache. Entries older than the
// interval given to NewCache are removed periodically.
type Cache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
}

// NewCache creates a Cache and starts a background goroutine that removes
// entries older than interval, checking once per interval.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
	}
	go c.reapLoop(interval)
	return c
}

// Add stores val under key, replacing any existing entry.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

// Get returns the value stored under key and whether it was found.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	return entry.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.reap(time.Now(), interval)
	}
}

// reap removes every entry created more than interval before now.
func (c *Cache) reap(now time.Time, interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if now.Sub(entry.createdAt) > interval {
			delete(c.entries, key)
		}
	}
}
