//go:build ignore

// Template for Phase 10 — copy to: internal/delivery/httpdelivery/pagination.go
//
// Parses ?limit=&offset= with safe defaults and a hard cap. Pass the result into
// repo List queries as LIMIT/OFFSET.
//
package httpdelivery

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Page struct {
	Limit  int
	Offset int
}

func parsePage(c *gin.Context) Page {
	limit := defaultLimit
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return Page{Limit: limit, Offset: offset}
}

// Optional metadata envelope — use ONLY on new endpoints, or update the Flutter
// parser when adopting it on an existing one.
type Paginated struct {
	Items  interface{} `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}
