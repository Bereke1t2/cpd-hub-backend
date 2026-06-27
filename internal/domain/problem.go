package domain

// Problem is the API-facing problem entity.
// SolverCount is carried internally and mapped to numberOfSolvedPeople in the handler;
// it is not exposed as a JSON field directly (the handler builds the api shape).
type Problem struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Difficulty  string   `json:"difficulty"`
	TopicTags   []string `json:"topic_tags,omitempty"`
	Likes       int      `json:"likes"`
	Dislikes    int      `json:"dislikes"`
	DeepLink    string   `json:"deep_link,omitempty"`
	IsLiked     bool     `json:"isLiked"`
	IsDisliked  bool     `json:"isDisliked"`
	Solved      bool     `json:"solved"`
	SolverCount int      `json:"-"` // real solve count; mapped to numberOfSolvedPeople by handler
}

// ProblemRepository is user-aware: every read and write scoped to a username.
type ProblemRepository interface {
	// Reads — all return the calling user's isLiked/isDisliked/solved state.
	ListForUser(username string, limit, offset int) ([]*Problem, error)
	GetByIDForUser(username, id string) (*Problem, error)
	GetDailyForUser(username string) (*Problem, error)

	// Writes — all require the calling user's username.
	Like(username, id string) error
	Dislike(username, id string) error
	MarkSolved(username, id string) error
	UnmarkSolved(username, id string) error

	// Aggregate query used by the handler when needed separately.
	CountSolvers(id string) (int, error)
}
