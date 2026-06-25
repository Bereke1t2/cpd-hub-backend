package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

// AuthRepositoryPG implements domain.AuthRepository backed by Postgres.
// This is pure data access — all business rules live in usecase/auth.
type AuthRepositoryPG struct {
	client *Client
}

func NewAuthRepositoryPG(client *Client) *AuthRepositoryPG {
	return &AuthRepositoryPG{client: client}
}

// FindByEmailOrUsername looks up a user by either the email column or the
// username handle — used by login so the client can use either.
func (r *AuthRepositoryPG) FindByEmailOrUsername(login string) (*domain.UserRecord, error) {
	ctx := context.Background()
	row := r.client.Pool.QueryRow(ctx,
		`SELECT username, COALESCE(email,''), full_name, password_hash
		   FROM users
		  WHERE email = $1 OR username = $1
		  LIMIT 1`,
		login)
	var rec domain.UserRecord
	if err := row.Scan(&rec.Username, &rec.Email, &rec.FullName, &rec.PasswordHash); err != nil {
		return nil, errors.New("not found")
	}
	return &rec, nil
}

// ExistsEmail reports whether a row with this email already exists.
func (r *AuthRepositoryPG) ExistsEmail(email string) (bool, error) {
	ctx := context.Background()
	var count int
	err := r.client.Pool.QueryRow(ctx,
		`SELECT COUNT(1) FROM users WHERE email = $1`, email).Scan(&count)
	return count > 0, err
}

// UsernameTaken reports whether a row with this username already exists.
func (r *AuthRepositoryPG) UsernameTaken(username string) (bool, error) {
	ctx := context.Background()
	var count int
	err := r.client.Pool.QueryRow(ctx,
		`SELECT COUNT(1) FROM users WHERE username = $1`, username).Scan(&count)
	return count > 0, err
}

// Insert creates a new user row. password_hash must already be bcrypt-hashed.
func (r *AuthRepositoryPG) Insert(rec *domain.UserRecord) error {
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx,
		`INSERT INTO users (username, email, full_name, password_hash, rating)
		 VALUES ($1, $2, $3, $4, 0)`,
		rec.Username, rec.Email, rec.FullName, rec.PasswordHash)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}
