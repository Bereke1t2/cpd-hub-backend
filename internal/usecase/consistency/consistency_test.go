package consistency

import (
	"errors"
	"testing"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeConsistencyRepo struct {
	days        []string
	streak      *domain.Streak
	goal        *domain.Goal
	ladders     []*domain.Ladder
	solvedCount int
	savedTo     *domain.Streak
	savedGoal   *domain.Goal
	err         error
}

func (f *fakeConsistencyRepo) ActiveDays(string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.days, nil
}
func (f *fakeConsistencyRepo) GetStreak(string) (*domain.Streak, error) { return f.streak, nil }
func (f *fakeConsistencyRepo) SaveStreak(_ string, s *domain.Streak) error {
	f.savedTo = s
	return nil
}
func (f *fakeConsistencyRepo) GetGoal(string) (*domain.Goal, error) { return f.goal, nil }
func (f *fakeConsistencyRepo) SaveGoal(_ string, g *domain.Goal) error {
	f.savedGoal = g
	return f.err
}
func (f *fakeConsistencyRepo) GetLadders(string) ([]*domain.Ladder, error) {
	return f.ladders, f.err
}
func (f *fakeConsistencyRepo) SaveLadder(string, *domain.Ladder) error { return nil }
func (f *fakeConsistencyRepo) SolvedCountSince(string, string) (int, error) {
	return f.solvedCount, nil
}

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

func TestGetGoalDefaultsAndRecomputesProgress(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	repo := &fakeConsistencyRepo{solvedCount: 3}
	uc := New(repo)

	got, err := uc.GetGoal("alice", now)
	if err != nil {
		t.Fatalf("GetGoal() error = %v", err)
	}
	if got.ID != "weekly-problems" || got.Target != 5 || got.Progress != 3 || got.PeriodStart != "2026-06-21" {
		t.Fatalf("goal = %+v", got)
	}
	if repo.savedGoal != got {
		t.Fatal("expected goal snapshot to be saved")
	}
}

func TestSaveGoalValidatesTarget(t *testing.T) {
	uc := New(&fakeConsistencyRepo{})

	if _, err := uc.SaveGoal("alice", &domain.Goal{Target: 0}); err == nil {
		t.Fatal("SaveGoal() error = nil, want validation error")
	}

	goal := &domain.Goal{Target: 4}
	got, err := uc.SaveGoal("alice", goal)
	if err != nil {
		t.Fatalf("SaveGoal() valid error = %v", err)
	}
	if got != goal {
		t.Fatal("SaveGoal() did not return original goal")
	}
}

func TestGetLaddersDelegates(t *testing.T) {
	ladders := []*domain.Ladder{{ID: "starter"}}
	got, err := New(&fakeConsistencyRepo{ladders: ladders}).GetLadders("alice")
	if err != nil {
		t.Fatalf("GetLadders() error = %v", err)
	}
	if got[0].ID != "starter" {
		t.Fatalf("GetLadders()[0].ID = %q, want starter", got[0].ID)
	}
}

func TestGetStreakReturnsActiveDaysError(t *testing.T) {
	wantErr := errors.New("repo failed")
	_, err := New(&fakeConsistencyRepo{err: wantErr}).GetStreak("alice", time.Now())
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetStreak() error = %v, want %v", err, wantErr)
	}
}
