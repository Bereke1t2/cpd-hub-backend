package practice

import (
	"math"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

const (
	MinEase     = domain.MinReviewEase
	DefaultEase = domain.DefaultReviewEase
)

func NewCard(problemID string, now time.Time) *domain.ReviewItem {
	return &domain.ReviewItem{
		ProblemID:   problemID,
		DueDate:     now.UTC().Format(time.RFC3339),
		Interval:    1,
		Ease:        DefaultEase,
		Repetitions: 0,
	}
}

func Schedule(item *domain.ReviewItem, quality int, now time.Time) *domain.ReviewItem {
	out := *item
	if out.Ease == 0 {
		out.Ease = DefaultEase
	}
	if out.Interval <= 0 {
		out.Interval = 1
	}
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}

	if quality < 3 {
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
			if out.Interval < 1 {
				out.Interval = 1
			}
		}
	}

	q := float64(quality)
	out.Ease = out.Ease + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	if out.Ease < MinEase {
		out.Ease = MinEase
	}
	out.DueDate = now.UTC().AddDate(0, 0, out.Interval).Format(time.RFC3339)
	return &out
}
