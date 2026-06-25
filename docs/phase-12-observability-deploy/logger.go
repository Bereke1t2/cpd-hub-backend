//go:build ignore
// Template for Phase 12 — copy to: internal/delivery/httpdelivery/logger.go
//
// Structured JSON access log, one line per request, including the request_id set
// by the Phase-1 RequestID middleware. Add to the middleware stack in place of
// gin.Logger().
package httpdelivery

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var accessLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID, _ := c.Get("request_id")
		accessLogger.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
			slog.Any("request_id", reqID),
		)
	}
}
