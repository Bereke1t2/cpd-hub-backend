package databases

import (
	"context"
	"fmt"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ActivityRepositoryDB struct {
	client *postgres.Client
}

func NewActivityRepositoryDB(client *postgres.Client) *ActivityRepositoryDB {
	return &ActivityRepositoryDB{client: client}
}

func (r *ActivityRepositoryDB) List() ([]*domain.Activity, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT id, username, action, type, timestamp FROM activity ORDER BY timestamp DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Activity{}
	for rows.Next() {
		var a domain.Activity
		if err := rows.Scan(&a.ID, &a.Username, &a.Action, &a.Type, &a.Timestamp); err != nil {
			continue
		}
		out = append(out, &a)
	}
	return out, nil
}
