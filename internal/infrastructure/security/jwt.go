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
	Typ      string `json:"typ"` // "access" | "refresh"
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte("dev-secret-change-me")
}

// GenerateToken mints a short-lived access token (typ = "access").
func GenerateToken(u *domain.UserProfile, email string, ttl time.Duration) (string, error) {
	return sign(u, email, "access", ttl)
}

// GenerateRefreshToken mints a long-lived refresh token (typ = "refresh").
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

// ParseToken validates a token and enforces the expected type ("access" or "refresh").
// Pass expectedTyp="" to skip the type check (e.g. for testing).
func ParseToken(tokenStr, expectedTyp string) (*Claims, error) {
	if tokenStr == "" {
		return nil, errors.New("empty token")
	}
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
	if expectedTyp != "" && claims.Typ != expectedTyp {
		return nil, errors.New("wrong token type")
	}
	return claims, nil
}
