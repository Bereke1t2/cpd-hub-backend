//go:build ignore
// Template for Phase 1 — copy to: internal/delivery/httpdelivery/middleware.go
//
// Cross-cutting gin middleware: panic recovery as JSON, request IDs, CORS.
package httpdelivery

import (
	"crypto/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecoveryJSON converts a panic into a 500 JSON error instead of dropping the
// connection, and keeps the worker alive.
func RecoveryJSON() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "internal",
			"message": "something went wrong",
		})
	})
}

// RequestID attaches an X-Request-Id to the context and echoes it back. If the
// client sent one, it is preserved (useful for tracing across services).
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			id = newRequestID()
		}
		c.Set("request_id", id)
		c.Writer.Header().Set("X-Request-Id", id)
		c.Next()
	}
}

// CORS allows the configured origins. Pass []string{"*"} in dev. Required for the
// Flutter web build / any browser client; harmless for native mobile.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*")
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if _, ok := allowed[origin]; ok {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// newRequestID returns a short random hex id. Uses crypto/rand so it works even
// when math/rand isn't seeded; falls back to a counter if rand fails.
func newRequestID() string {
	const hextable = "0123456789abcdef"
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-fallback"
	}
	out := make([]byte, 32)
	for i, v := range b {
		out[i*2] = hextable[v>>4]
		out[i*2+1] = hextable[v&0x0f]
	}
	return string(out)
}
