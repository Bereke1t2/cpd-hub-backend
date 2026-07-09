package domain

const (
	MinReviewEase     = 1.3
	DefaultReviewEase = 2.5
)

type ReviewItem struct {
	ProblemID   string  `json:"problem_id"`
	DueDate     string  `json:"due_date"`
	Interval    int     `json:"interval"`
	Ease        float64 `json:"ease"`
	Repetitions int     `json:"repetitions"`
}

type UpsolveItem struct {
	ContestID    string `json:"contest_id"`
	ContestTitle string `json:"contest_title"`
	ProblemID    string `json:"problem_id"`
	ProblemTitle string `json:"problem_title"`
	Resolved     bool   `json:"resolved"`
}

type PracticeRepository interface {
	ListReviewQueue(username string) ([]*ReviewItem, error)
	GetReview(username, problemID string) (*ReviewItem, error)
	AddReview(username string, item *ReviewItem) (*ReviewItem, error)
	UpdateReview(username string, item *ReviewItem) (*ReviewItem, error)
	DeleteReview(username, problemID string) error

	ListUpsolves(username string) ([]*UpsolveItem, error)
	AddUpsolve(username string, item *UpsolveItem) (*UpsolveItem, error)
	UpdateUpsolve(username, problemID string, resolved bool) (*UpsolveItem, error)
}
