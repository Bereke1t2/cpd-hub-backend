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

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		reqID, _ := c.Get("request_id")

		accessLogger.Info("http_request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("route", route),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("latency", time.Since(start)),
			slog.String("ip", c.ClientIP()),
			slog.Any("request_id", reqID),
		)
	}
}
