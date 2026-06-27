package consistency

import (
	"testing"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeConsistencyRepo struct {
	days    []string
	streak  *domain.Streak
	savedTo *domain.Streak
}

func (f *fakeConsistencyRepo) ActiveDays(string) ([]string, error)          { return f.days, nil }
func (f *fakeConsistencyRepo) GetStreak(string) (*domain.Streak, error)     { return f.streak, nil }
func (f *fakeConsistencyRepo) SaveStreak(_ string, s *domain.Streak) error  { f.savedTo = s; return nil }
func (f *fakeConsistencyRepo) GetGoal(string) (*domain.Goal, error)         { return nil, nil }
func (f *fakeConsistencyRepo) SaveGoal(string, *domain.Goal) error          { return nil }
func (f *fakeConsistencyRepo) GetLadders(string) ([]*domain.Ladder, error)  { return nil, nil }
func (f *fakeConsistencyRepo) SaveLadder(string, *domain.Ladder) error      { return nil }
func (f *fakeConsistencyRepo) SolvedCountSince(string, string) (int, error) { return 0, nil }

func TestGetStreak_TableDriven(t *testing.T) {
	now := time.Date(2026, 6, 21, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name        string
		days        []string
		freezes     int
		wantCurrent int
	}{
		{name: "empty history", days: nil, freezes: 2, wantCurrent: 0},
		{name: "single day", days: []string{"2026-06-21"}, freezes: 2, wantCurrent: 1},
		{name: "today not yet solved", days: []string{"2026-06-19", "2026-06-20"}, freezes: 2, wantCurrent: 2},
		{name: "freeze bridges one gap", days: []string{"2026-06-18", "2026-06-20", "2026-06-21"}, freezes: 1, wantCurrent: 3},
		{name: "two gaps break streak", days: []string{"2026-06-15", "2026-06-16", "2026-06-21"}, freezes: 2, wantCurrent: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeConsistencyRepo{days: tc.days, streak: &domain.Streak{FreezesAvailable: tc.freezes}}
			uc := New(repo)

			got, err := uc.GetStreak("alice", now)
			if err != nil {
				t.Fatalf("GetStreak() error = %v", err)
			}
			if got.Current != tc.wantCurrent {
				t.Fatalf("Current = %d, want %d", got.Current, tc.wantCurrent)
			}
			if repo.savedTo == nil {
				t.Fatal("expected recomputed streak to be saved")
			}
			if repo.savedTo.Current != tc.wantCurrent {
				t.Fatalf("saved Current = %d, want %d", repo.savedTo.Current, tc.wantCurrent)
			}
		})
	}
}
