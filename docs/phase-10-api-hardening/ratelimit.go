//go:build ignore
// Template for Phase 10 — copy to: internal/delivery/httpdelivery/ratelimit.go
//
// In-memory token-bucket rate limiter keyed by a caller identity (IP or username).
// Swap the store for Redis when you run more than one instance.
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

// NewRateLimiter: capacity tokens, refilled at capacity/window per second.
// e.g. NewRateLimiter(5, time.Minute) => 5 requests/min, bursting to 5.
func NewRateLimiter(capacity int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*bucket),
		capacity: float64(capacity),
		refill:   float64(capacity) / window.Seconds(),
	}
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

// Middleware limits by client IP. For authenticated writes, key by username
// instead (read it from the context after loadUser).
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
				"error": "rate_limited", "message": "too many requests, slow down",
			})
			return
		}
		c.Next()
	}
}

// Wiring:
//   authLimiter := NewRateLimiter(5, time.Minute)
//   authGroup.POST("/login",  authLimiter.Middleware(), h.Login)
//   authGroup.POST("/signup", authLimiter.Middleware(), h.Signup)
//
// NOTE: prune idle buckets periodically (a background goroutine deleting entries
// whose lastFill is older than a few minutes) so the map doesn't grow unbounded.
