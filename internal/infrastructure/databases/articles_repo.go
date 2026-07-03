package databases

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ArticlesRepositoryDB struct {
	client *postgres.Client
}

func NewArticlesRepositoryDB(client *postgres.Client) *ArticlesRepositoryDB {
	return &ArticlesRepositoryDB{client: client}
}

func (r *ArticlesRepositoryDB) List(filter domain.ArticleFilter) ([]*domain.Article, error) {
	filter = normalizeArticleFilter(filter)

	where := []string{}
	args := []interface{}{}
	addWhere := func(clause string, value interface{}) {
		args = append(args, value)
		where = append(where, strings.Replace(clause, "?", "$"+strconv.Itoa(len(args)), 1))
	}
	if filter.Source != "" {
		addWhere(`LOWER(a.source) = LOWER(?)`, filter.Source)
	}
	if filter.Tag != "" {
		addWhere(`EXISTS (SELECT 1 FROM article_tags ft WHERE ft.article_id = a.id AND LOWER(ft.tag) = LOWER(?))`, filter.Tag)
	}

	query := `
		SELECT a.id, a.title, a.author, a.source, a.source_url, a.excerpt,
		       a.full_content, a.published_at, a.rating,
		       COALESCE(array_agg(t.tag ORDER BY t.tag) FILTER (WHERE t.tag IS NOT NULL), '{}') AS tags
		FROM articles a
		LEFT JOIN article_tags t ON t.article_id = a.id`
	if len(where) > 0 {
		query += ` WHERE ` + strings.Join(where, ` AND `)
	}
	query += ` GROUP BY a.id ORDER BY a.published_at DESC, a.id`
	args = append(args, filter.Limit)
	query += ` LIMIT $` + strconv.Itoa(len(args))
	args = append(args, filter.Offset)
	query += ` OFFSET $` + strconv.Itoa(len(args))

	rows, err := r.client.Pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, domain.ErrInternal("could not list articles").Wrap(err)
	}
	defer rows.Close()

	out := []*domain.Article{}
	for rows.Next() {
		var article domain.Article
		var published time.Time
		if err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Author,
			&article.Source,
			&article.SourceURL,
			&article.Excerpt,
			&article.FullContent,
			&published,
			&article.Rating,
			&article.Tags,
		); err != nil {
			continue
		}
		article.PublishedAt = published.UTC().Format(time.RFC3339)
		if article.Tags == nil {
			article.Tags = []string{}
		}
		out = append(out, &article)
	}
	return out, nil
}

func normalizeArticleFilter(filter domain.ArticleFilter) domain.ArticleFilter {
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	filter.Source = strings.TrimSpace(filter.Source)
	filter.Tag = strings.TrimSpace(filter.Tag)
	return filter
}
