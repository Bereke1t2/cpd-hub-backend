package security

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware returns a Gin middleware that validates JWT bearer tokens.
// On success the parsed *Claims is stored in the Gin context under key "claims".
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "missing authorization header"})
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "invalid authorization header"})
			c.Abort()
			return
		}
		token := strings.TrimSpace(parts[1])
		claims, err := ParseToken(token, "access")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": err.Error()})
			c.Abort()
			return
		}
		// store claims for handlers
		c.Set("claims", claims)
		c.Next()
	}
}

// GetClaims retrieves parsed JWT claims from the Gin context.
func GetClaims(c *gin.Context) (*Claims, bool) {
	v, ok := c.Get("claims")
	if !ok {
		return nil, false
	}
	cl, ok := v.(*Claims)
	return cl, ok
}
