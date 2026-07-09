package practice

import (
	"testing"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

func TestSchedulePassingGradesGrowInterval(t *testing.T) {
	now := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	card := NewCard("cf-1234A", now)

	card = Schedule(card, 4, now)
	if card.Repetitions != 1 || card.Interval != 1 {
		t.Fatalf("rep1: got reps=%d interval=%d, want 1 and 1", card.Repetitions, card.Interval)
	}

	card = Schedule(card, 4, now)
	if card.Repetitions != 2 || card.Interval != 6 {
		t.Fatalf("rep2: got reps=%d interval=%d, want 2 and 6", card.Repetitions, card.Interval)
	}

	card = Schedule(card, 5, now)
	if card.Interval <= 6 {
		t.Fatalf("rep3: interval should grow past 6, got %d", card.Interval)
	}
	if card.Ease < MinEase {
		t.Fatalf("ease fell below floor: %v", card.Ease)
	}
}

func TestScheduleFailResets(t *testing.T) {
	now := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	card := &domain.ReviewItem{ProblemID: "cf-1A", Interval: 15, Ease: 2.6, Repetitions: 4}

	card = Schedule(card, 1, now)
	if card.Repetitions != 0 || card.Interval != 1 {
		t.Fatalf("fail should reset: got reps=%d interval=%d", card.Repetitions, card.Interval)
	}
	wantDue := now.AddDate(0, 0, 1).Format(time.RFC3339)
	if card.DueDate != wantDue {
		t.Fatalf("due_date: got %s want %s", card.DueDate, wantDue)
	}
}

func TestScheduleEaseFloor(t *testing.T) {
	now := time.Date(2026, 6, 24, 0, 0, 0, 0, time.UTC)
	card := &domain.ReviewItem{ProblemID: "cf-1A", Interval: 1, Ease: MinEase, Repetitions: 1}

	for i := 0; i < 10; i++ {
		card = Schedule(card, 3, now)
		if card.Ease < MinEase {
			t.Fatalf("ease dropped below floor: %v", card.Ease)
		}
	}
}
