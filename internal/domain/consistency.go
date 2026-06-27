package domain

// Streak — last_active_day is nullable (pointer).
type Streak struct {
	Current          int      `json:"current" binding:"gte=0"`
	Longest          int      `json:"longest" binding:"gte=0"`
	LastActiveDay    *string  `json:"last_active_day"`                   // "2026-06-21" or null
	FreezesAvailable int      `json:"freezes_available" binding:"gte=0"` // default 2
	ActiveDays       []string `json:"active_days"`
}

// Goal — type mirrors the client's GoalType enum name (e.g. "problemsPerWeek").
type Goal struct {
	ID          string `json:"id" binding:"required"`   // default "weekly-problems"
	Type        string `json:"type" binding:"required"` // default "problemsPerWeek"
	Target      int    `json:"target" binding:"gte=1"`  // default 5
	Progress    int    `json:"progress" binding:"gte=0"`
	PeriodStart string `json:"period_start" binding:"required"` // "2026-06-15"
}

type LadderRung struct {
	ProblemID string  `json:"problem_id"`
	Rating    int     `json:"rating"`
	Solved    bool    `json:"solved"`
	TopicID   *string `json:"topic_id"`
}

type Ladder struct {
	ID         string       `json:"id"`
	Title      string       `json:"title"`
	FromRating int          `json:"from_rating"`
	ToRating   int          `json:"to_rating"`
	Rungs      []LadderRung `json:"rungs"`
}

// ConsistencyRepository — implemented in infrastructure/databases.
type ConsistencyRepository interface {
	GetStreak(username string) (*Streak, error)
	SaveStreak(username string, s *Streak) error

	GetGoal(username string) (*Goal, error)
	SaveGoal(username string, g *Goal) error

	// Base ladders with the caller's solved state overlaid per rung.
	GetLadders(username string) ([]*Ladder, error)
	SaveLadder(username string, l *Ladder) error

	// Raw signal used by the usecase to recompute streak/goal progress.
	ActiveDays(username string) ([]string, error) // distinct solve days, ascending
	SolvedCountSince(username, sinceDay string) (int, error)
}

func DefaultGoal(periodStart string) *Goal {
	return &Goal{ID: "weekly-problems", Type: "problemsPerWeek", Target: 5, Progress: 0, PeriodStart: periodStart}
}
