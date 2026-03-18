package databases

import (
	"context"
	"fmt"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type InfoRepositoryDB struct {
	client *postgres.Client
}

func NewInfoRepositoryDB(client *postgres.Client) *InfoRepositoryDB {
	return &InfoRepositoryDB{client: client}
}

func (r *InfoRepositoryDB) List() ([]*domain.Info, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT title, description FROM info ORDER BY title")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Info{}
	for rows.Next() {
		var i domain.Info
		if err := rows.Scan(&i.Title, &i.Description); err != nil {
			continue
		}
		out = append(out, &i)
	}
	return out, nil
}
