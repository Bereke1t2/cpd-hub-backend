//go:build ignore

// Template for Phase 14 — copy to: internal/usecase/practice/sm2.go
//
// The SuperMemo-2 scheduling algorithm. Pure, gin-free, sql-free, time-injectable
// (pass `now` so it is deterministically unit-testable). The server recomputes the
// schedule from the user's recall `quality` (0..5) so the client can't game it.
//
package practice

import (
	"math"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

const (
	minEase     = 1.3
	defaultEase = 2.5
)

// Schedule applies one SM-2 review to `item` given a recall quality 0..5 and the
// current time, returning the updated card. It mutates a copy, not the input.
//
//	quality: 5 perfect · 4 correct(hesitation) · 3 correct(hard) · <3 failed recall
func Schedule(item *domain.ReviewItem, quality int, now time.Time) *domain.ReviewItem {
	out := *item
	if out.Ease == 0 {
		out.Ease = defaultEase
	}
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}

	if quality < 3 {
		// failed recall — relearn from the start
		out.Repetitions = 0
		out.Interval = 1
	} else {
		out.Repetitions++
		switch out.Repetitions {
		case 1:
			out.Interval = 1
		case 2:
			out.Interval = 6
		default:
			out.Interval = int(math.Round(float64(out.Interval) * out.Ease))
		}
	}

	// ease update (clamped at 1.3); applied for every grade.
	q := float64(quality)
	out.Ease = out.Ease + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if out.Ease < minEase {
		out.Ease = minEase
	}

	out.DueDate = now.AddDate(0, 0, out.Interval).Format(time.RFC3339)
	return &out
}

// NewCard builds a first-time card with SM-2 defaults, due immediately.
func NewCard(problemID string, now time.Time) *domain.ReviewItem {
	return &domain.ReviewItem{
		ProblemID:   problemID,
		DueDate:     now.Format(time.RFC3339),
		Interval:    1,
		Ease:        defaultEase,
		Repetitions: 0,
	}
}
