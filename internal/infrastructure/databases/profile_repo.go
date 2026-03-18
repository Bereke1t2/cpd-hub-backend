package databases

import (
	"context"
	"fmt"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ProfileRepositoryDB struct {
	client *postgres.Client
}

func NewProfileRepositoryDB(client *postgres.Client) *ProfileRepositoryDB {
	return &ProfileRepositoryDB{client: client}
}

func (r *ProfileRepositoryDB) ListUsers() ([]*domain.UserProfile, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT username, COALESCE(full_name,''), COALESCE(bio,''), COALESCE(avatar_url,''), COALESCE(rating,0) FROM users LEFT JOIN profiles ON users.username=profiles.username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.UserProfile{}
	for rows.Next() {
		var p domain.UserProfile
		if err := rows.Scan(&p.Username, &p.FullName, &p.Bio, &p.AvatarURL, &p.Rating); err != nil {
			continue
		}
		out = append(out, &p)
	}
	return out, nil
}

func (r *ProfileRepositoryDB) GetProfile(username string) (*domain.UserProfile, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	row := r.client.Pool.QueryRow(ctx, "SELECT username, COALESCE(full_name,''), COALESCE(bio,''), COALESCE(avatar_url,''), COALESCE(rating,0) FROM users LEFT JOIN profiles ON users.username=profiles.username WHERE users.username=$1", username)
	var p domain.UserProfile
	if err := row.Scan(&p.Username, &p.FullName, &p.Bio, &p.AvatarURL, &p.Rating); err != nil {
		return nil, fmt.Errorf("not found")
	}
	return &p, nil
}

func (r *ProfileRepositoryDB) CreateUser(user *domain.UserProfile) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "INSERT INTO users (username, full_name) VALUES ($1,$2)", user.Username, user.FullName)
	if err != nil {
		return err
	}
	_, err = r.client.Pool.Exec(ctx, "INSERT INTO profiles (username, bio, rating, avatar_url) VALUES ($1,$2,$3,$4)", user.Username, user.Bio, user.Rating, user.AvatarURL)
	return err
}

func (r *ProfileRepositoryDB) UpdateUser(user *domain.UserProfile) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "UPDATE profiles SET bio=$2, rating=$3, avatar_url=$4 WHERE username=$1", user.Username, user.Bio, user.Rating, user.AvatarURL)
	return err
}

func (r *ProfileRepositoryDB) DeleteUser(username string) error {
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
