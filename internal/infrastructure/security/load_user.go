package security

import (
	"net/http"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

// ProfileLoader is satisfied by the profile repository.
type ProfileLoader interface {
	GetProfile(username string) (*domain.UserProfile, error)
}

// LoadUser returns middleware that runs after AuthMiddleware. It reads the
// validated claims (set by AuthMiddleware), loads the caller's profile, and
// stores it in the Gin context:
//
//	c.Get("username") → string
//	c.Get("user")     → *domain.UserProfile  (nil if load fails)
//
// The handler can use currentUsername(c) + currentUser(c) helpers.
func LoadUser(loader ProfileLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c)
		if !ok || claims.Username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing or invalid token",
			})
			return
		}
		c.Set("username", claims.Username)
		if loader != nil {
			if p, err := loader.GetProfile(claims.Username); err == nil {
				c.Set("user", p)
			}
		}
		c.Next()
	}
}
