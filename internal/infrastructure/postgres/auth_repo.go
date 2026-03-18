package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/security"
)

// AuthRepositoryPG implements domain.AuthRepository backed by Postgres.
type AuthRepositoryPG struct {
	client *Client
}

func NewAuthRepositoryPG(client *Client) *AuthRepositoryPG {
	// ensure users table exists
	_ = client.ensureUsersTable(context.Background())
	return &AuthRepositoryPG{client: client}
}

func (r *AuthRepositoryPG) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email and password required")
	}
	ctx := context.Background()
	row := r.client.Pool.QueryRow(ctx, "SELECT username, full_name, password_hash FROM users WHERE username=$1", req.Email)
	var username, fullName, passwordHash string
	if err := row.Scan(&username, &fullName, &passwordHash); err != nil {
		return nil, errors.New("invalid credentials")
	}
	if err := security.ComparePassword(passwordHash, req.Password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	// generate token
	tok, err := security.GenerateToken(&domain.UserProfile{Username: username, FullName: fullName}, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}
	return &domain.AuthResponse{Token: tok, User: domain.UserProfile{Username: username, FullName: fullName}}, nil
}

func (r *AuthRepositoryPG) Signup(req *domain.SignupRequest) (*domain.AuthResponse, error) {
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return nil, errors.New("missing fields")
	}
	if req.Password != req.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}
	ctx := context.Background()
	// check exists
	row := r.client.Pool.QueryRow(ctx, "SELECT username FROM users WHERE username=$1", req.Email)
	var existing string
	if err := row.Scan(&existing); err == nil {
		return nil, errors.New("user already exists")
	}
	hash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}
	_, err = r.client.Pool.Exec(ctx, "INSERT INTO users (username, full_name, password_hash, rating) VALUES ($1,$2,$3,$4)", req.Email, req.FullName, hash, 0)
	if err != nil {
		return nil, fmt.Errorf("could not insert user: %w", err)
	}
	tok, err := security.GenerateToken(&domain.UserProfile{Username: req.Email, FullName: req.FullName}, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}
	return &domain.AuthResponse{Token: tok, User: domain.UserProfile{Username: req.Email, FullName: req.FullName}}, nil
}
