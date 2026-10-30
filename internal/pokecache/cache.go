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
	now     func() time.Time // the clock; replaced in tests

	stopOnce sync.Once
	stop     chan struct{} // closed by Stop to end the reap loop
	done     chan struct{} // closed by the reap loop when it has exited
}

// NewCache creates a Cache and starts a background goroutine that removes
// entries older than interval, checking once per interval. It panics if
// interval is not positive. Call Stop to end the goroutine when the cache is no
// longer needed.
func NewCache(interval time.Duration) *Cache {
	return newCache(interval, time.Now)
}

func newCache(interval time.Duration, now func() time.Time) *Cache {
	if interval <= 0 {
		panic("pokecache: interval must be positive")
	}

	c := &Cache{
		entries: make(map[string]cacheEntry),
		now:     now,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	go c.reapLoop(interval)
	return c
}

// Add stores val under key, replacing any existing entry.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: c.now(),
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

// Stop ends the background reap loop and waits for it to exit. Entries remain
// readable afterwards but are no longer removed. It is safe to call Stop more
// than once.
func (c *Cache) Stop() {
	c.stopOnce.Do(func() { close(c.stop) })
	<-c.done
}

func (c *Cache) reapLoop(interval time.Duration) {
	defer close(c.done)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.reap(c.now(), interval)
		case <-c.stop:
			return
		}
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
