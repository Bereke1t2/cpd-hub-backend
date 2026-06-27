//go:build ignore

// Template for Phase 12 — copy to: internal/delivery/httpdelivery/metrics.go
//
// Prometheus request count + latency histogram, plus the /metrics endpoint.
// Requires: go get github.com/prometheus/client_golang/prometheus{,/promhttp}
//
package httpdelivery

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by route, method and status.",
	}, []string{"route", "method", "status"})

	httpLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"route", "method"})
)

// Metrics records counters + latency per request.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		httpRequests.WithLabelValues(route, c.Request.Method, strconv.Itoa(c.Writer.Status())).Inc()
		httpLatency.WithLabelValues(route, c.Request.Method).Observe(time.Since(start).Seconds())
	}
}

// MetricsHandler exposes /metrics for Prometheus to scrape.
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) { h.ServeHTTP(c.Writer, c.Request) }
}

// Wiring in NewHandler:
//   g.Use(Metrics())
//   g.GET("/metrics", MetricsHandler())
