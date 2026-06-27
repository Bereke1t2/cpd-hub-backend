//go:build ignore

// Template for Phase 3 — replaces parts of: internal/infrastructure/security/jwt.go
//
// Access + refresh tokens with a `typ` claim guard so an access token can't be
// used to refresh and vice-versa. Reads the secret from config/env.
//
// NOTE: add this type to internal/domain (e.g. user.go), used by auth_usecase.go:
//
//	type UserRecord struct {
//		Username, Email, FullName, PasswordHash string
//	}
//
package security

import (
	"errors"
	"os"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"fullName,omitempty"`
	Typ      string `json:"typ"` // "access" | "refresh"
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("dev-secret-change-me")
}

// GenerateToken mints an access token.
func GenerateToken(u *domain.UserProfile, email string, ttl time.Duration) (string, error) {
	return sign(u, email, "access", ttl)
}

// GenerateRefreshToken mints a long-lived refresh token.
func GenerateRefreshToken(u *domain.UserProfile, email string, ttl time.Duration) (string, error) {
	return sign(u, email, "refresh", ttl)
}

func sign(u *domain.UserProfile, email, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Username: u.Username,
		Email:    email,
		FullName: u.FullName,
		Typ:      typ,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "cpd-hub",
			Subject:   u.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret())
}

// ParseToken validates a token and enforces the expected type ("access"/"refresh").
func ParseToken(tokenStr, expectedTyp string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret(), nil
	})
	if err != nil || !tok.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	if claims.Typ != expectedTyp {
		return nil, errors.New("wrong token type")
	}
	return claims, nil
}
