package pokecache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// fakeClock is a manually advanced clock, safe for use from several goroutines.
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// eventually polls cond until it is true, failing the test after a generous
// deadline. Polling keeps the tests fast without depending on exact timing.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for: %s", what)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{key: "https://example.com", val: []byte("testdata")},
		{key: "https://example.com/path", val: []byte("moretestdata")},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			t.Cleanup(cache.Stop)
			cache.Add(c.key, c.val)

			val, ok := cache.Get(c.key)
			if !ok {
				t.Fatalf("expected to find key %q", c.key)
			}
			if string(val) != string(c.val) {
				t.Errorf("expected %q, got %q", c.val, val)
			}
		})
	}
}

func TestGetMissing(t *testing.T) {
	cache := NewCache(5 * time.Second)
	t.Cleanup(cache.Stop)

	if _, ok := cache.Get("missing"); ok {
		t.Error("expected missing key to not be found")
	}
}

func TestNewCachePanicsOnNonPositiveInterval(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected NewCache(0) to panic")
		}
	}()
	NewCache(0)
}

func TestReapLoopRemovesOnlyExpiredEntries(t *testing.T) {
	// The interval is both the tick period and the maximum entry age. Whether
	// an entry is expired is decided by the fake clock, never by real time.
	clock := newFakeClock()
	cache := newCache(time.Millisecond, clock.Now)
	t.Cleanup(cache.Stop)

	cache.Add("old", []byte("1"))
	clock.Advance(10 * time.Millisecond) // "old" is now older than the interval
	cache.Add("fresh", []byte("2"))      // created at the new time, so not expired

	eventually(t, "old entry to be reaped", func() bool {
		_, ok := cache.Get("old")
		return !ok
	})
	if _, ok := cache.Get("fresh"); !ok {
		t.Error("expected the fresh entry to still be cached")
	}
}

func TestReapLoopKeepsEntriesWhileClockStandsStill(t *testing.T) {
	clock := newFakeClock()
	cache := newCache(time.Millisecond, clock.Now)
	t.Cleanup(cache.Stop)

	cache.Add("key", []byte("val"))
	time.Sleep(20 * time.Millisecond) // many ticks, but no time passes on the fake clock

	if _, ok := cache.Get("key"); !ok {
		t.Error("expected entry to survive while the clock has not advanced")
	}
}

func TestStopEndsReaping(t *testing.T) {
	clock := newFakeClock()
	cache := newCache(time.Millisecond, clock.Now)

	cache.Add("key", []byte("val"))
	cache.Stop() // returns only once the reap loop has exited

	clock.Advance(time.Hour)
	time.Sleep(20 * time.Millisecond) // plenty of ticks, if the loop were still running

	if _, ok := cache.Get("key"); !ok {
		t.Error("expected no reaping after Stop")
	}
}

func TestStopIsIdempotent(t *testing.T) {
	cache := NewCache(time.Second)
	cache.Stop()
	cache.Stop()
}

func TestReapKeepsFreshEntries(t *testing.T) {
	cache := &Cache{entries: make(map[string]cacheEntry), now: time.Now}
	cache.Add("fresh", []byte("a"))
	cache.entries["stale"] = cacheEntry{
		createdAt: time.Now().Add(-time.Minute),
		val:       []byte("b"),
	}

	cache.reap(time.Now(), 5*time.Second)

	if _, ok := cache.Get("fresh"); !ok {
		t.Error("expected fresh entry to be kept")
	}
	if _, ok := cache.Get("stale"); ok {
		t.Error("expected stale entry to be removed")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := NewCache(time.Millisecond)
	t.Cleanup(cache.Stop)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			for j := 0; j < 100; j++ {
				cache.Add(key, []byte("val"))
				cache.Get(key)
			}
		}(i)
	}
	wg.Wait()
}
