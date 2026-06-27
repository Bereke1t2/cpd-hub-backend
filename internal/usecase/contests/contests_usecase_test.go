package contests

import (
	"errors"
	"testing"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeContestRepo struct {
	listUsername string
	getID        string
	boardID      string
	joinUser     string
	joinID       string
	leaveUser    string
	leaveID      string
	err          error
}

func (f *fakeContestRepo) ListForUser(username string) ([]*domain.Contest, error) {
	f.listUsername = username
	return []*domain.Contest{{ID: "c1"}}, f.err
}

func (f *fakeContestRepo) GetByID(id string) (*domain.Contest, error) {
	f.getID = id
	return &domain.Contest{ID: id}, f.err
}

func (f *fakeContestRepo) Leaderboard(id string) ([]*domain.LeaderboardEntry, error) {
	f.boardID = id
	return []*domain.LeaderboardEntry{{Username: "alice", Rank: 1}}, f.err
}

func (f *fakeContestRepo) Participate(username, id string) error {
	f.joinUser = username
	f.joinID = id
	return f.err
}

func (f *fakeContestRepo) Unparticipate(username, id string) error {
	f.leaveUser = username
	f.leaveID = id
	return f.err
}

func TestUsecaseDelegatesContests(t *testing.T) {
	repo := &fakeContestRepo{}
	uc := New(repo, nil)

	if got, err := uc.ListForUser("alice"); err != nil || got[0].ID != "c1" {
		t.Fatalf("ListForUser() = (%v, %v), want c1 nil", got, err)
	}
	if got, err := uc.GetByID("c2"); err != nil || got.ID != "c2" {
		t.Fatalf("GetByID() = (%v, %v), want c2 nil", got, err)
	}
	if got, err := uc.Leaderboard("c3"); err != nil || got[0].Username != "alice" {
		t.Fatalf("Leaderboard() = (%v, %v), want alice nil", got, err)
	}
	if err := uc.Participate("alice", "c4"); err != nil {
		t.Fatalf("Participate() error = %v", err)
	}
	if err := uc.Unparticipate("alice", "c5"); err != nil {
		t.Fatalf("Unparticipate() error = %v", err)
	}

	if repo.listUsername != "alice" || repo.getID != "c2" || repo.boardID != "c3" {
		t.Fatalf("read args = (%q,%q,%q)", repo.listUsername, repo.getID, repo.boardID)
	}
	if repo.joinUser != "alice" || repo.joinID != "c4" {
		t.Fatalf("join args = (%q,%q)", repo.joinUser, repo.joinID)
	}
	if repo.leaveUser != "alice" || repo.leaveID != "c5" {
		t.Fatalf("leave args = (%q,%q)", repo.leaveUser, repo.leaveID)
	}
}

func TestUsecaseReturnsContestRepositoryErrors(t *testing.T) {
	wantErr := errors.New("repo failed")
	uc := New(&fakeContestRepo{err: wantErr}, nil)

	if _, err := uc.ListForUser("alice"); !errors.Is(err, wantErr) {
		t.Fatalf("ListForUser() error = %v, want %v", err, wantErr)
	}
	if err := uc.Participate("alice", "c1"); !errors.Is(err, wantErr) {
		t.Fatalf("Participate() error = %v, want %v", err, wantErr)
	}
}
