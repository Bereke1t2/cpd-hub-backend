//go:build ignore

// Template for Phase 15 — copy to: internal/infrastructure/databases/articles_repo.go
//
// Read-only feed with optional source/tag filters and limit/offset pagination.
// Tags are aggregated with array_agg to avoid an N+1. Errors are *domain.AppError.
//
package databases

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ArticlesRepositoryDB struct{ client *postgres.Client }

func NewArticlesRepositoryDB(c *postgres.Client) *ArticlesRepositoryDB {
	return &ArticlesRepositoryDB{client: c}
}

func (r *ArticlesRepositoryDB) List(f domain.ArticleFilter) ([]*domain.Article, error) {
	// clamp pagination
	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	// build WHERE dynamically with positional args
	var where []string
	var args []interface{}
	add := func(clause string, val interface{}) {
		args = append(args, val)
		where = append(where, strings.Replace(clause, "?", "$"+strconv.Itoa(len(args)), 1))
	}
	if f.Source != "" {
		add("LOWER(a.source) = LOWER(?)", f.Source)
	}
	if f.Tag != "" {
		// EXISTS keeps the row-per-article shape while filtering by tag
		add("EXISTS (SELECT 1 FROM article_tags t WHERE t.article_id = a.id AND t.tag = ?)", f.Tag)
	}

	q := `SELECT a.id, a.title, a.author, a.source, a.source_url, a.excerpt,
	             a.full_content, a.published_at, a.rating,
	             COALESCE(array_agg(at.tag) FILTER (WHERE at.tag IS NOT NULL), '{}') AS tags
	        FROM articles a
	        LEFT JOIN article_tags at ON at.article_id = a.id`
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " GROUP BY a.id ORDER BY a.published_at DESC"
	args = append(args, f.Limit)
	q += " LIMIT $" + strconv.Itoa(len(args))
	args = append(args, f.Offset)
	q += " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.client.Pool.Query(context.Background(), q, args...)
	if err != nil {
		return nil, domain.ErrInternal("could not list articles").Wrap(err)
	}
	defer rows.Close()

	out := []*domain.Article{}
	for rows.Next() {
		var a domain.Article
		var published time.Time
		if err := rows.Scan(&a.ID, &a.Title, &a.Author, &a.Source, &a.SourceURL,
			&a.Excerpt, &a.FullContent, &published, &a.Rating, &a.Tags); err != nil {
			continue
		}
		a.PublishedAt = published.Format(time.RFC3339)
		if a.Tags == nil {
			a.Tags = []string{}
		}
		out = append(out, &a)
	}
	return out, nil
}
