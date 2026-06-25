// Package auth implements the auth business rules.
// The usecase owns: validation, handle derivation, bcrypt comparing/hashing,
// and token issuance. The repository does only DB I/O.
package auth

import (
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/security"
)

const (
	accessTTL  = 24 * time.Hour
	refreshTTL = 30 * 24 * time.Hour
)

// UseCase holds the auth business logic.
type UseCase struct {
	store domain.AuthRepository
}

// New creates a UseCase backed by the given store (AuthRepository).
func New(store domain.AuthRepository) *UseCase {
	return &UseCase{store: store}
}

// Login accepts email or username, verifies bcrypt, and issues tokens.
func (uc *UseCase) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	rec, err := uc.store.FindByEmailOrUsername(strings.TrimSpace(req.Email))
	if err != nil {
		// Same message for missing account and wrong password.
		return nil, domain.ErrUnauthorized("invalid credentials")
	}
	if err := security.ComparePassword(rec.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrUnauthorized("invalid credentials")
	}
	return uc.issue(rec)
}

// Signup creates a new account. Username is optional — derived from email if blank.
func (uc *UseCase) Signup(req *domain.SignupRequest) (*domain.AuthResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, domain.ErrValidation("passwords do not match")
	}
	taken, err := uc.store.ExistsEmail(req.Email)
	if err != nil {
		return nil, domain.ErrInternal("").Wrap(err)
	}
	if taken {
		return nil, domain.ErrConflict("an account with this email already exists")
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		username, err = uc.deriveUniqueHandle(req.Email)
		if err != nil {
			return nil, err
		}
	} else {
		// explicit handle — check it is free
		if t, e := uc.store.UsernameTaken(username); e != nil {
			return nil, domain.ErrInternal("").Wrap(e)
		} else if t {
			return nil, domain.ErrConflict("username is already taken")
		}
	}

	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, domain.ErrInternal("could not hash password").Wrap(err)
	}
	rec := &domain.UserRecord{
		Username:     username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: hash,
	}
	if err := uc.store.Insert(rec); err != nil {
		return nil, domain.ErrInternal("could not create user").Wrap(err)
	}
	return uc.issue(rec)
}

// Refresh validates a refresh token and mints a new access token.
func (uc *UseCase) Refresh(refreshToken string) (*domain.AuthResponse, error) {
	claims, err := security.ParseToken(refreshToken, "refresh")
	if err != nil {
		return nil, domain.ErrUnauthorized("invalid refresh token")
	}
	rec, err := uc.store.FindByEmailOrUsername(claims.Username)
	if err != nil {
		return nil, domain.ErrUnauthorized("invalid credentials")
	}
	return uc.issue(rec)
}

// issue mints both tokens and returns the AuthResponse.
func (uc *UseCase) issue(rec *domain.UserRecord) (*domain.AuthResponse, error) {
	profile := &domain.UserProfile{Username: rec.Username, FullName: rec.FullName}
	access, err := security.GenerateToken(profile, rec.Email, accessTTL)
	if err != nil {
		return nil, domain.ErrInternal("could not generate access token").Wrap(err)
	}
	refresh, err := security.GenerateRefreshToken(profile, rec.Email, refreshTTL)
	if err != nil {
		return nil, domain.ErrInternal("could not generate refresh token").Wrap(err)
	}
	return &domain.AuthResponse{
		Token:        access,
		RefreshToken: refresh,
		User:         *profile,
	}, nil
}

// deriveUniqueHandle slugifies the email local-part and appends a counter until free.
func (uc *UseCase) deriveUniqueHandle(email string) (string, error) {
	base := slug(strings.SplitN(email, "@", 2)[0])
	if base == "" {
		base = "user"
	}
	candidate := base
	for i := 2; i < 1000; i++ {
		taken, err := uc.store.UsernameTaken(candidate)
		if err != nil {
			return "", domain.ErrInternal("").Wrap(err)
		}
		if !taken {
			return candidate, nil
		}
		candidate = base + itoa(i)
	}
	return "", domain.ErrInternal("could not allocate a username")
}

// slug keeps only [a-z0-9] characters (safe handle subset).
func slug(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}
