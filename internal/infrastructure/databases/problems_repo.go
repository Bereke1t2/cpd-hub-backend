package databases

import (
	"context"
	"fmt"
	"strings"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ProblemsRepositoryDB struct {
	client *postgres.Client
}

func NewProblemsRepositoryDB(client *postgres.Client) *ProblemsRepositoryDB {
	return &ProblemsRepositoryDB{client: client}
}

func (r *ProblemsRepositoryDB) List() ([]*domain.Problem, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT id, title, difficulty, topic_tags, likes, dislikes, deep_link, is_liked, is_disliked, solved FROM problems")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Problem{}
	for rows.Next() {
		var p domain.Problem
		var tags *string
		if err := rows.Scan(&p.ID, &p.Title, &p.Difficulty, &tags, &p.Likes, &p.Dislikes, &p.DeepLink, &p.IsLiked, &p.IsDisliked, &p.Solved); err != nil {
			// log or continue on scan error
			continue
		}
		p.TopicTags = parseTags(tags)
		out = append(out, &p)
	}
	return out, nil
}

func (r *ProblemsRepositoryDB) GetDaily() (*domain.Problem, error) {
	// For simplicity return the first problem as daily
	list, err := r.List()
	if err != nil || len(list) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return list[0], nil
}

func (r *ProblemsRepositoryDB) Like(id string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	ct, err := r.client.Pool.Exec(ctx, `
		UPDATE problems SET
			is_liked = CASE WHEN COALESCE(is_liked,false) THEN false ELSE true END,
			likes = CASE WHEN COALESCE(is_liked,false) THEN GREATEST(COALESCE(likes,0)-1,0) ELSE COALESCE(likes,0)+1 END
		WHERE id=$1
	`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *ProblemsRepositoryDB) Dislike(id string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	ct, err := r.client.Pool.Exec(ctx, `
		UPDATE problems SET
			is_disliked = CASE WHEN COALESCE(is_disliked,false) THEN false ELSE true END,
			dislikes = CASE WHEN COALESCE(is_disliked,false) THEN GREATEST(COALESCE(dislikes,0)-1,0) ELSE COALESCE(dislikes,0)+1 END
		WHERE id=$1
	`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *ProblemsRepositoryDB) MarkSolved(id string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	ct, err := r.client.Pool.Exec(ctx, "UPDATE problems SET solved = true WHERE id=$1", id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

func (r *ProblemsRepositoryDB) UnmarkSolved(id string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	ct, err := r.client.Pool.Exec(ctx, "UPDATE problems SET solved = false WHERE id=$1", id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("not found")
	}
	return nil
}

// helper types and funcs

func parseTags(t *string) []string {
	if t == nil || *t == "" {
		return nil
	}
	// naive comma-split
	var out []string
	for _, s := range strings.Split(*t, ",") {
		out = append(out, strings.TrimSpace(s))
	}
	return out
}
