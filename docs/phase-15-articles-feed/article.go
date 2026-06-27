//go:build ignore

// Template for Phase 15 — copy to: internal/domain/article.go
//
// First-party (and optionally cached external) articles feed. camelCase JSON per api.md §10.
//
package domain

// Article is a CPD Hub editorial / tutorial, or a cached external blog entry.
type Article struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Source      string   `json:"source"` // "Codeforces" | "LeetCode" | "CPD Hub"
	SourceURL   string   `json:"sourceUrl"`
	Excerpt     string   `json:"excerpt"`
	FullContent string   `json:"fullContent"`
	PublishedAt string   `json:"publishedAt"` // ISO-8601 (RFC3339)
	Tags        []string `json:"tags"`
	Rating      int      `json:"rating"`
}

// ArticleFilter carries the optional list query params so the handler stays thin.
type ArticleFilter struct {
	Limit  int    // default 10, clamp [1,100]
	Offset int    // default 0, clamp >=0
	Source string // "", "codeforces", "leetcode", "cpdhub"
	Tag    string // "" or a single topic tag
}

// ArticleRepository is implemented in infrastructure/databases.
type ArticleRepository interface {
	List(f ArticleFilter) ([]*Article, error)
}
