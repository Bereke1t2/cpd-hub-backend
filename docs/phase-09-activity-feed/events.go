//go:build ignore
// Template for Phase 9 — copy to: internal/usecase/activity/events.go
//
// A single place that turns a user action into its side effects (activity row,
// and for solves the submission + heatmap rows). Call from the problems usecase
// after a successful action — NOT from the repository.
package activity

import (
	"fmt"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

// Writer is the data-access surface the recorder needs.
type Writer interface {
	InsertActivity(a *domain.Activity) error
	InsertSubmission(s *domain.Submission) error
}

type Recorder struct{ w Writer }

func NewRecorder(w Writer) *Recorder { return &Recorder{w: w} }

// RecordSolve writes the feed line + an Accepted submission. daily_solves is
// handled inside MarkSolved (Phase 4) on the false->true transition.
func (r *Recorder) RecordSolve(username string, p *domain.Problem, now time.Time) {
	_ = r.w.InsertActivity(&domain.Activity{
		ID:        genID("act"),
		Username:  username,
		Action:    fmt.Sprintf("solved '%s'", p.Title),
		Type:      "Solve",
		Timestamp: now.Format(time.RFC3339), // stored; humanized at read time
	})
	_ = r.w.InsertSubmission(&domain.Submission{
		ID:           genID("sub"),
		ProblemID:    p.ID,
		ProblemTitle: p.Title,
		Status:       "Accepted",
		Language:     "",
		Timestamp:    now.Format(time.RFC3339),
	})
}

func (r *Recorder) RecordLike(username string, p *domain.Problem, now time.Time) {
	_ = r.w.InsertActivity(&domain.Activity{
		ID:        genID("act"),
		Username:  username,
		Action:    fmt.Sprintf("liked '%s'", p.Title),
		Type:      "Like",
		Timestamp: now.Format(time.RFC3339),
	})
}

// humanizeSince turns a stored RFC3339 timestamp into "2 min ago" for the client.
// Compute at READ time in the activity handler so it's always current.
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

// genID is a placeholder; use a real id source (uuid, or a sequence). Kept simple
// here because cmd seeds and the recorder both need unique-enough ids.
func genID(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("20060102T150405.000000000")
}
