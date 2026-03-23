package domain

type Problem struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Difficulty string   `json:"difficulty"`
	TopicTags  []string `json:"topic_tags,omitempty"`
	Likes      int      `json:"likes"`
	Dislikes   int      `json:"dislikes"`
	DeepLink   string   `json:"deep_link,omitempty"`
	IsLiked    bool     `json:"isLiked"`
	IsDisliked bool     `json:"isDisliked"`
	Solved     bool     `json:"solved"`
}

type ProblemRepository interface {
	List() ([]*Problem, error)
	GetDaily() (*Problem, error)
	GetById(id string) (*Problem, error)
	Like(id string) error
	Dislike(id string) error
	MarkSolved(id string) error
	UnmarkSolved(id string) error
}
