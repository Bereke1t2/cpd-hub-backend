package activity

import (
	"fmt"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type Recorder struct{ w domain.ActivityRepository }

func NewRecorder(w domain.ActivityRepository) *Recorder { return &Recorder{w: w} }

func (r *Recorder) RecordSolve(username, problemTitle string, now time.Time) {
	_ = r.w.InsertActivity(&domain.Activity{
		ID:        genID("act"),
		Username:  username,
		Action:    fmt.Sprintf("solved '%s'", problemTitle),
		Type:      "Solve",
		Timestamp: now.UTC().Format(time.RFC3339),
	})
}

func (r *Recorder) RecordLike(username, problemTitle string, now time.Time) {
	_ = r.w.InsertActivity(&domain.Activity{
		ID:        genID("act"),
		Username:  username,
		Action:    fmt.Sprintf("liked '%s'", problemTitle),
		Type:      "Like",
		Timestamp: now.UTC().Format(time.RFC3339),
	})
}

// HumanizeSince converts a stored RFC3339 timestamp into a human-readable string
// like "2 min ago". Call at read time so it's always fresh.
func HumanizeSince(ts string, now time.Time) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hr ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

func genID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
