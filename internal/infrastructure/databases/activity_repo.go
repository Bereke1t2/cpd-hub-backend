package databases

import (
	"context"
	"fmt"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ActivityRepositoryDB struct {
	client *postgres.Client
}

func NewActivityRepositoryDB(client *postgres.Client) *ActivityRepositoryDB {
	return &ActivityRepositoryDB{client: client}
}

func (r *ActivityRepositoryDB) List(limit, offset int) ([]*domain.Activity, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx,
		"SELECT id, username, action, type, created_at FROM activity ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Activity{}
	for rows.Next() {
		var a domain.Activity
		var ts time.Time
		if err := rows.Scan(&a.ID, &a.Username, &a.Action, &a.Type, &ts); err != nil {
			continue
		}
		a.Timestamp = ts.UTC().Format(time.RFC3339)
		out = append(out, &a)
	}
	return out, nil
}

func (r *ActivityRepositoryDB) InsertActivity(a *domain.Activity) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	_, err := r.client.Pool.Exec(context.Background(),
		`INSERT INTO activity (id, username, action, type, timestamp) VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (id) DO NOTHING`,
		a.ID, a.Username, a.Action, a.Type, a.Timestamp)
	return err
}
