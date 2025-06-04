package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries      map[string]cacheEntry
	lock         sync.Mutex
	reapInterval time.Duration
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{reapInterval: interval}
	go cache.reapLoop()
	return &cache
}

func (c *Cache) Add(key string, val []byte) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.entries[key]

	return val.val, ok
}

func (c *Cache) reapLoop() {
	for {
		c.lock.Lock()

		for key, entry := range c.entries {
			if time.Since(entry.createdAt) > c.reapInterval {
				delete(c.entries, key)
			}
		}
		c.lock.Unlock()
		time.Sleep(c.reapInterval)
	}
}
