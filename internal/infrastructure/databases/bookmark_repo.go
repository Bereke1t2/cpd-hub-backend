package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type BookmarkRepositoryDB struct{ client *postgres.Client }

func NewBookmarkRepositoryDB(c *postgres.Client) *BookmarkRepositoryDB {
	return &BookmarkRepositoryDB{client: c}
}

func (r *BookmarkRepositoryDB) Add(username, problemID string) error {
	_, err := r.client.Pool.Exec(context.Background(),
		`INSERT INTO bookmarks (username, problem_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		username, problemID)
	return err
}

func (r *BookmarkRepositoryDB) Remove(username, problemID string) error {
	_, err := r.client.Pool.Exec(context.Background(),
		`DELETE FROM bookmarks WHERE username=$1 AND problem_id=$2`,
		username, problemID)
	return err
}

func (r *BookmarkRepositoryDB) ListProblemIDs(username string) ([]string, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT problem_id FROM bookmarks WHERE username=$1 ORDER BY created_at DESC`,
		username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			out = append(out, id)
		}
	}
	return out, nil
}
