package pokecache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

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
	if _, ok := cache.Get("missing"); ok {
		t.Error("expected missing key to not be found")
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 10*time.Millisecond

	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	if _, ok := cache.Get("https://example.com"); !ok {
		t.Fatal("expected to find key right after adding it")
	}

	time.Sleep(waitTime)

	if _, ok := cache.Get("https://example.com"); ok {
		t.Error("expected key to have been reaped")
	}
}

func TestReapKeepsFreshEntries(t *testing.T) {
	cache := &Cache{entries: make(map[string]cacheEntry)}
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
