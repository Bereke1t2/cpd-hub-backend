package databases

import (
	"context"
	"fmt"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type UsersRepositoryDB struct {
	client *postgres.Client
}

func NewUsersRepositoryDB(client *postgres.Client) *UsersRepositoryDB {
	return &UsersRepositoryDB{client: client}
}

func (r *UsersRepositoryDB) Create(user *domain.User) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "INSERT INTO users (username, full_name, rating) VALUES ($1,$2,$3)", user.Username, user.FullName, user.Rating)
	return err
}

func (r *UsersRepositoryDB) GetByUsername(username string) (*domain.User, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	row := r.client.Pool.QueryRow(ctx, "SELECT username, full_name, bio, avatar_url, rating FROM users LEFT JOIN profiles ON users.username=profiles.username WHERE users.username=$1", username)
	var u domain.User
	if err := row.Scan(&u.Username, &u.FullName, &u.Bio, &u.AvatarURL, &u.Rating); err != nil {
		return nil, fmt.Errorf("not found")
	}
	return &u, nil
}

func (r *UsersRepositoryDB) List() ([]*domain.User, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT users.username, users.full_name, COALESCE(profiles.bio,''), COALESCE(profiles.avatar_url,''), COALESCE(profiles.rating,0) FROM users LEFT JOIN profiles ON users.username=profiles.username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.User{}
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.Username, &u.FullName, &u.Bio, &u.AvatarURL, &u.Rating); err != nil {
			continue
		}
		out = append(out, &u)
	}
	return out, nil
}

// UpdateProfile updates both users and profiles table when applicable.
func (r *UsersRepositoryDB) UpdateProfile(user *domain.User) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "UPDATE users SET full_name=$2 WHERE username=$1", user.Username, user.FullName)
	if err != nil {
		return err
	}
	_, err = r.client.Pool.Exec(ctx, "UPDATE profiles SET bio=$2, avatar_url=$3, rating=$4 WHERE username=$1", user.Username, user.Bio, user.AvatarURL, user.Rating)
	return err
}

// DeleteUser removes a user and related profile.
func (r *UsersRepositoryDB) DeleteUser(username string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "DELETE FROM profiles WHERE username=$1", username)
	if err != nil {
		return err
	}
	_, err = r.client.Pool.Exec(ctx, "DELETE FROM users WHERE username=$1", username)
	return err
}
