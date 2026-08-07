// Single place that shapes every HTTP response. Success returns the raw value the
// Flutter client parses (bare entity / array — no wrapper). Errors always use the
// {error, message} shape derived from *domain.AppError.
package httpdelivery

import (
	"fmt"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// respondOK writes a 200 with the raw payload. Do NOT wrap in {"data": ...} —
// the client's fromJson parsers expect the bare entity/array.
func respondOK(c *gin.Context, payload interface{}) {
	c.JSON(200, payload)
}

// respondCreated writes a 201 with the raw payload.
func respondCreated(c *gin.Context, payload interface{}) {
	c.JSON(201, payload)
}

// respondNoContent writes a 200 {success:true} for write actions that return no body.
func respondSuccess(c *gin.Context) {
	c.JSON(200, gin.H{"success": true})
}

// respondError maps a domain.AppError (or any error) to the consistent error shape.
// Usage in handlers:
//
//	if err != nil { respondError(c, err); return }
func respondError(c *gin.Context, err error) {
	ae := domain.AsAppError(err)
	if ae.Err != nil {
		fmt.Printf("[HTTP ERROR] code=%s status=%d message=%s cause=%v\n", ae.Code, ae.Status, ae.Message, ae.Err)
	} else {
		fmt.Printf("[HTTP ERROR] code=%s status=%d message=%s\n", ae.Code, ae.Status, ae.Message)
	}
	c.JSON(ae.Status, gin.H{
		"error":   ae.Code,
		"message": ae.Message,
	})
}
