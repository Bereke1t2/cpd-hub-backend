package external

import (
	"sync"
	"time"
)

type entry struct {
	val       interface{}
	expiresAt time.Time
}

type TTLCache struct {
	mu  sync.RWMutex
	m   map[string]entry
	ttl time.Duration
}

func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{m: make(map[string]entry), ttl: ttl}
}

func (c *TTLCache) Get(key string) (val interface{}, fresh bool, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.m[key]
	if !ok {
		return nil, false, false
	}
	return e.val, time.Now().Before(e.expiresAt), true
}

func (c *TTLCache) Set(key string, val interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = entry{val: val, expiresAt: time.Now().Add(c.ttl)}
}
