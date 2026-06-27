package httpdelivery

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens   float64
	lastFill time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	capacity float64
	refill   float64 // tokens per second
}

// NewRateLimiter creates a token-bucket limiter: capacity requests per window, bursting to capacity.
func NewRateLimiter(capacity int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: float64(capacity),
		refill:   float64(capacity) / window.Seconds(),
	}
	go rl.pruneLoop()
	return rl
}

func (rl *RateLimiter) allow(key string, now time.Time) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	b, ok := rl.buckets[key]
	if !ok {
		rl.buckets[key] = &bucket{tokens: rl.capacity - 1, lastFill: now}
		return true
	}
	b.tokens += now.Sub(b.lastFill).Seconds() * rl.refill
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}
	b.lastFill = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware limits by client IP, or by username if already authenticated.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if u, ok := c.Get("username"); ok {
			if s, _ := u.(string); s != "" {
				key = s
			}
		}
		if !rl.allow(key, time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "too many requests, slow down",
			})
			return
		}
		c.Next()
	}
}

// pruneLoop removes idle buckets every 5 minutes to prevent unbounded map growth.
func (rl *RateLimiter) pruneLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-5 * time.Minute)
		rl.mu.Lock()
		for key, b := range rl.buckets {
			if b.lastFill.Before(cutoff) {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}
