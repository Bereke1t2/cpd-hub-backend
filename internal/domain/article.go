package domain

type Article struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Source      string   `json:"source"`
	SourceURL   string   `json:"sourceUrl"`
	Excerpt     string   `json:"excerpt"`
	FullContent string   `json:"fullContent"`
	PublishedAt string   `json:"publishedAt"`
	Tags        []string `json:"tags"`
	Rating      int      `json:"rating"`
}

type ArticleFilter struct {
	Limit  int
	Offset int
	Source string
	Tag    string
}

type ArticleRepository interface {
	List(filter ArticleFilter) ([]*Article, error)
}
