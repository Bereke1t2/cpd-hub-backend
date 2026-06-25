package httpdelivery

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger is satisfied by *postgres.Client.
type Pinger interface {
	Ping(ctx context.Context) error
}

func (h *handlerImpl) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

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
