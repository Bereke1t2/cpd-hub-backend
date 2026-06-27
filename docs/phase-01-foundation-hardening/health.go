//go:build ignore

// Template for Phase 1 — copy to: internal/delivery/httpdelivery/health.go
//
// Liveness and readiness probes. Register OUTSIDE the /api auth group so probes
// don't need a token. Pass the postgres client into the handler for readiness.
//
package httpdelivery

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger is satisfied by *postgres.Client (add a Ping method that calls Pool.Ping).
type Pinger interface {
	Ping(ctx context.Context) error
}

// Healthz is liveness: the process is up and serving. Always 200.
func (h *handlerImpl) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz is readiness: dependencies (DB) are reachable. 503 if not, so a load
// balancer stops routing traffic to a node that can't serve.
func (h *handlerImpl) Readyz(c *gin.Context) {
	if h.db == nil {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.db.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "database"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// --- wiring notes ---
// 1. Add `db Pinger` to handlerImpl and accept it in NewHandler.
// 2. Add to internal/infrastructure/postgres/db.go:
//        func (c *Client) Ping(ctx context.Context) error { return c.Pool.Ping(ctx) }
// 3. Register routes (in NewHandler, before RegisterRoutes):
//        g.GET("/healthz", h.Healthz)
//        g.GET("/readyz", h.Readyz)
