//go:build ignore

// Template for Phase 3 — copy to: internal/infrastructure/security/load_user.go
//
// LoadUser runs after AuthMiddleware. It reads the validated claims, loads the
// caller's profile, and stashes it on the context so every handler can do
// currentUsername(c) / current user.
//
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

// LoadUser returns middleware that hydrates c with the current user.
// Pass the profile repo. If loading fails it does not abort (the route may only
// need the username from claims) — but it always sets "username".
func LoadUser(loader ProfileLoader) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := GetClaims(c) // set by AuthMiddleware
		if !ok || claims.Username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized", "message": "missing or invalid token",
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

// --- add these accessors in the httpdelivery package (handler.go) ---
//
// func currentUsername(c *gin.Context) string {
// 	v, _ := c.Get("username")
// 	s, _ := v.(string)
// 	return s
// }
//
// func currentUser(c *gin.Context) *domain.UserProfile {
// 	v, _ := c.Get("user")
// 	u, _ := v.(*domain.UserProfile)
// 	return u
// }
