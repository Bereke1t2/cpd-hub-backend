//go:build ignore

// Template for Phase 14 — copy to: internal/domain/practice.go
//
// Smart Practice: spaced-repetition review queue (SM-2) + contest upsolves.
// JSON tags are snake_case to match api.md §9 and the Flutter ReviewItemModel/UpsolveItemModel.
//
package domain

// ReviewItem is one SM-2 spaced-repetition card, scoped to a user.
type ReviewItem struct {
	ProblemID   string  `json:"problem_id"`
	DueDate     string  `json:"due_date"` // ISO-8601 (RFC3339)
	Interval    int     `json:"interval"` // days until next review
	Ease        float64 `json:"ease"`     // SM-2 ease factor, >= 1.3, starts 2.5
	Repetitions int     `json:"repetitions"`
}

// UpsolveItem is a problem flagged from a past contest to retry, scoped to a user.
type UpsolveItem struct {
	ContestID    string `json:"contest_id"`
	ContestTitle string `json:"contest_title"`
	ProblemID    string `json:"problem_id"`
	ProblemTitle string `json:"problem_title"`
	Resolved     bool   `json:"resolved"`
}

// PracticeRepository is implemented in infrastructure/databases.
// Every method is scoped to the authenticated username.
type PracticeRepository interface {
	// Review queue
	ListReviewQueue(username string) ([]*ReviewItem, error)
	AddReview(username string, item *ReviewItem) (*ReviewItem, error)
	UpdateReview(username string, item *ReviewItem) (*ReviewItem, error)
	DeleteReview(username, problemID string) error

	// Upsolves
	ListUpsolves(username string) ([]*UpsolveItem, error)
	AddUpsolve(username string, item *UpsolveItem) (*UpsolveItem, error)
	UpdateUpsolve(username, problemID string, resolved bool) (*UpsolveItem, error)
}
