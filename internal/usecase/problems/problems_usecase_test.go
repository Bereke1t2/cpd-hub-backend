package problems

import (
	"errors"
	"testing"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeProblemRepo struct {
	listUsername   string
	listLimit      int
	listOffset     int
	getUsername    string
	getID          string
	dailyUsername  string
	likeUsername   string
	likeID         string
	dislikeUser    string
	dislikeID      string
	solvedUsername string
	solvedID       string
	unsolveUser    string
	unsolveID      string
	err            error
}

func (f *fakeProblemRepo) ListForUser(username string, limit, offset int) ([]*domain.Problem, error) {
	f.listUsername = username
	f.listLimit = limit
	f.listOffset = offset
	return []*domain.Problem{{ID: "p1"}}, f.err
}

func (f *fakeProblemRepo) GetByIDForUser(username, id string) (*domain.Problem, error) {
	f.getUsername = username
	f.getID = id
	return &domain.Problem{ID: id}, f.err
}

func (f *fakeProblemRepo) GetDailyForUser(username string) (*domain.Problem, error) {
	f.dailyUsername = username
	return &domain.Problem{ID: "daily"}, f.err
}

func (f *fakeProblemRepo) Like(username, id string) error {
	f.likeUsername = username
	f.likeID = id
	return f.err
}

func (f *fakeProblemRepo) Dislike(username, id string) error {
	f.dislikeUser = username
	f.dislikeID = id
	return f.err
}

func (f *fakeProblemRepo) MarkSolved(username, id string) error {
	f.solvedUsername = username
	f.solvedID = id
	return f.err
}

func (f *fakeProblemRepo) UnmarkSolved(username, id string) error {
	f.unsolveUser = username
	f.unsolveID = id
	return f.err
}

func (f *fakeProblemRepo) CountSolvers(string) (int, error) { return 0, nil }

func TestUsecaseDelegatesProblemReads(t *testing.T) {
	repo := &fakeProblemRepo{}
	uc := New(repo)

	if got, err := uc.ListForUser("alice", 10, 20); err != nil || got[0].ID != "p1" {
		t.Fatalf("ListForUser() = (%v, %v), want p1 nil", got, err)
	}
	if repo.listUsername != "alice" || repo.listLimit != 10 || repo.listOffset != 20 {
		t.Fatalf("list args = (%q,%d,%d)", repo.listUsername, repo.listLimit, repo.listOffset)
	}
	if got, err := uc.GetByIDForUser("alice", "p2"); err != nil || got.ID != "p2" {
		t.Fatalf("GetByIDForUser() = (%v, %v), want p2 nil", got, err)
	}
	if repo.getUsername != "alice" || repo.getID != "p2" {
		t.Fatalf("get args = (%q,%q)", repo.getUsername, repo.getID)
	}
	if got, err := uc.GetDailyForUser("alice"); err != nil || got.ID != "daily" {
		t.Fatalf("GetDailyForUser() = (%v, %v), want daily nil", got, err)
	}
	if repo.dailyUsername != "alice" {
		t.Fatalf("daily username = %q, want alice", repo.dailyUsername)
	}
}

func TestUsecaseDelegatesProblemWrites(t *testing.T) {
	repo := &fakeProblemRepo{}
	uc := New(repo)

	if err := uc.Like("alice", "p1"); err != nil {
		t.Fatalf("Like() error = %v", err)
	}
	if err := uc.Dislike("alice", "p2"); err != nil {
		t.Fatalf("Dislike() error = %v", err)
	}
	if err := uc.MarkSolved("alice", "p3"); err != nil {
		t.Fatalf("MarkSolved() error = %v", err)
	}
	if err := uc.UnmarkSolved("alice", "p4"); err != nil {
		t.Fatalf("UnmarkSolved() error = %v", err)
	}

	if repo.likeUsername != "alice" || repo.likeID != "p1" {
		t.Fatalf("like args = (%q,%q)", repo.likeUsername, repo.likeID)
	}
	if repo.dislikeUser != "alice" || repo.dislikeID != "p2" {
		t.Fatalf("dislike args = (%q,%q)", repo.dislikeUser, repo.dislikeID)
	}
	if repo.solvedUsername != "alice" || repo.solvedID != "p3" {
		t.Fatalf("solved args = (%q,%q)", repo.solvedUsername, repo.solvedID)
	}
	if repo.unsolveUser != "alice" || repo.unsolveID != "p4" {
		t.Fatalf("unsolve args = (%q,%q)", repo.unsolveUser, repo.unsolveID)
	}
}

func TestUsecaseReturnsProblemRepositoryErrors(t *testing.T) {
	wantErr := errors.New("repo failed")
	uc := New(&fakeProblemRepo{err: wantErr})

	if _, err := uc.ListForUser("alice", 1, 0); !errors.Is(err, wantErr) {
		t.Fatalf("ListForUser() error = %v, want %v", err, wantErr)
	}
	if err := uc.Like("alice", "p1"); !errors.Is(err, wantErr) {
		t.Fatalf("Like() error = %v, want %v", err, wantErr)
	}
}
