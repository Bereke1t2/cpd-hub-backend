//go:build ignore

// Template for Phase 11 — copy to: internal/usecase/consistency/consistency_test.go
//
// Table-driven tests for the pure streak math. No DB: a fake repo returns canned
// active days. Demonstrates the pattern for every usecase test.
//
package consistency

import (
	"testing"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

// fakeRepo implements domain.ConsistencyRepository with canned data.
type fakeRepo struct {
	days    []string
	streak  *domain.Streak
	savedTo *domain.Streak
}

func (f *fakeRepo) ActiveDays(string) ([]string, error)          { return f.days, nil }
func (f *fakeRepo) GetStreak(string) (*domain.Streak, error)     { return f.streak, nil }
func (f *fakeRepo) SaveStreak(_ string, s *domain.Streak) error  { f.savedTo = s; return nil }
func (f *fakeRepo) GetGoal(string) (*domain.Goal, error)         { return nil, nil }
func (f *fakeRepo) SaveGoal(string, *domain.Goal) error          { return nil }
func (f *fakeRepo) GetLadders(string) ([]*domain.Ladder, error)  { return nil, nil }
func (f *fakeRepo) SaveLadder(string, *domain.Ladder) error      { return nil }
func (f *fakeRepo) SolvedCountSince(string, string) (int, error) { return 0, nil }

func TestGetStreak_Current(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC) // a Sunday
	d := func(s string) string { return s }

	cases := []struct {
		name        string
		days        []string
		freezes     int
		wantCurrent int
	}{
		{"empty", nil, 2, 0},
		{"today only", []string{d("2026-06-21")}, 2, 1},
		{"three in a row ending today", []string{"2026-06-19", "2026-06-20", "2026-06-21"}, 2, 3},
		{"ending yesterday (today not done yet)", []string{"2026-06-19", "2026-06-20"}, 2, 2},
		{"one gap bridged by a freeze", []string{"2026-06-18", "2026-06-20", "2026-06-21"}, 2, 3},
		{"two gaps break it", []string{"2026-06-15", "2026-06-16", "2026-06-21"}, 2, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{days: tc.days, streak: &domain.Streak{FreezesAvailable: tc.freezes}}
			uc := New(repo)
			got, err := uc.GetStreak("alice", now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Current != tc.wantCurrent {
				t.Errorf("current = %d, want %d", got.Current, tc.wantCurrent)
			}
		})
	}
}
