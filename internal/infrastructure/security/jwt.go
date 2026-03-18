package security

import (
	"errors"
	"os"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

// Claims holds custom JWT claims for the application.
type Claims struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullName,omitempty"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		// fallback to a default for local dev — change in production
		s = "replace-me-with-a-secure-secret"
	}
	return []byte(s)
}

// GenerateToken creates a signed JWT for the given user profile. expiresIn is the token lifetime.
func GenerateToken(user *domain.UserProfile, expiresIn time.Duration) (string, error) {
	claims := Claims{
		Username: user.Username,
		Email:    "",
		FullName: user.FullName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	// set email if user has it (UserProfile currently doesn't include email field, leave blank)
	if token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims); token != nil {
		signed, err := token.SignedString(jwtSecret())
		if err != nil {
			return "", err
		}
		return signed, nil
	}
	return "", errors.New("failed to create token")
}

// ParseToken validates the token and returns the claims.
func ParseToken(tokenStr string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}
	var claims Claims
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		// ensure signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	return &claims, nil
}
