package stockpile

import (
	"runtime"
	"sync"
	"time"
)

const (
	// No expiration duration
	NoExpiry time.Duration = -1

	// Convenience durations
	MinuteExpiry   time.Duration = time.Minute
	HalfHourExpiry time.Duration = time.Minute * 30
	HourExpiry     time.Duration = time.Hour
	HalfDayExpiry  time.Duration = time.Hour * 12
	DayExpiry      time.Duration = time.Hour * 24
	WeekExpiry     time.Duration = DayExpiry * 7
)

type Cache struct {
	*cache
}

type cache struct {
	mu      sync.RWMutex
	store   map[string]item
	janitor *janitor
}

// New creates a new cache with the provided cleanup interval.
func New(ci time.Duration) *Cache {
	c := &Cache{
		cache: &cache{
			store: make(map[string]item),
		},
	}

	if ci > 0 {
		j := &janitor{
			interval: ci,
			stop:     make(chan struct{}),
		}

		c.janitor = j

		go j.run(c.cache)

		runtime.SetFinalizer(c, func(c *Cache) {
			c.cache.janitor.stop <- struct{}{}
		})
	}

	return c
}

// Set adds an item to the cache, replacing any existing item, with the provided expiration.
func (c *cache) Set(k string, v interface{}, d time.Duration) {
	if d <= 0 {
		d = NoExpiry
	}
	c.set(k, v, d)
}

// Set adds an item to the cache, replacing any existing item, without an expiration.
func (c *cache) SetNoExpiry(k string, v interface{}) {
	c.set(k, v, NoExpiry)
}

// Get retrieves an item from the cache if it exists.
func (c *cache) Get(k string) (interface{}, bool) {
	c.mu.RLock()
	item, ok := c.store[k]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// handle eviction of expirable items on access.
	if item.exp > 0 {
		if item.isExpired(time.Now()) {
			go c.Delete(item.key)
			return nil, false
		}
	}

	return item.val, true
}

// Delete removes the items stored at a given key.
func (c *cache) Delete(k string) {
	c.mu.Lock()
	delete(c.store, k)
	c.mu.Unlock()
}

// Reset removes all items from the cache.
func (c *cache) Reset() {
	c.mu.Lock()
	c.store = make(map[string]item)
	c.mu.Unlock()
}

// Count provides the total number of items stored.
//
// Note: This may include items that have expired but have not yet been cleaned up.
func (c *cache) Count() int {
	c.mu.RLock()
	n := len(c.store)
	c.mu.RUnlock()
	return n
}

func (c *cache) set(k string, v interface{}, d time.Duration) {
	var e int64
	if d > 0 {
		e = time.Now().Add(d).UnixNano()
	}
	c.mu.Lock()
	c.store[k] = item{key: k, val: v, exp: e}
	c.mu.Unlock()
}

func (c *cache) evict() {
	now := time.Now()
	c.mu.Lock()
	for k, v := range c.store {
		if v.isExpired(now) {
			delete(c.store, k)
		}
	}
	c.mu.Unlock()
}

type item struct {
	key string
	val interface{}
	exp int64
}

func (i *item) isExpired(now time.Time) bool {
	if i.exp == 0 {
		return false
	}
	return now.UnixNano() > i.exp
}

type janitor struct {
	interval time.Duration
	stop     chan struct{}
}

func (j *janitor) run(c *cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.evict()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}
