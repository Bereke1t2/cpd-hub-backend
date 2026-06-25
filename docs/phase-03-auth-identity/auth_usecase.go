//go:build ignore
// Template for Phase 3 — copy to: internal/usecase/auth/auth_usecase.go
//
// Auth business rules live here (validation, handle derivation, token issuance).
// The repository (postgres/auth_repo.go) becomes pure data access: lookups + insert.
package auth

import (
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/security"
)

// Store is the data-access surface the usecase needs. Implement it in postgres.
type Store interface {
	FindByEmailOrUsername(login string) (*domain.UserRecord, error) // returns hash too
	ExistsEmail(email string) (bool, error)
	UsernameTaken(username string) (bool, error)
	Insert(rec *domain.UserRecord) error
}

type UseCase struct {
	store     Store
	accessTTL time.Duration
}

func New(store Store) *UseCase {
	return &UseCase{store: store, accessTTL: 24 * time.Hour}
}

func (uc *UseCase) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, domain.ErrValidation("email and password are required")
	}
	rec, err := uc.store.FindByEmailOrUsername(strings.TrimSpace(req.Email))
	if err != nil {
		// same message whether the account is missing or the password is wrong
		return nil, domain.ErrUnauthorized("invalid credentials")
	}
	if err := security.ComparePassword(rec.PasswordHash, req.Password); err != nil {
		return nil, domain.ErrUnauthorized("invalid credentials")
	}
	return uc.issue(rec)
}

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

func (uc *UseCase) issue(rec *domain.UserRecord) (*domain.AuthResponse, error) {
	profile := &domain.UserProfile{Username: rec.Username, FullName: rec.FullName}
	tok, err := security.GenerateToken(profile, rec.Email, uc.accessTTL)
	if err != nil {
		return nil, domain.ErrInternal("could not generate token").Wrap(err)
	}
	return &domain.AuthResponse{Token: tok, User: *profile}, nil
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
