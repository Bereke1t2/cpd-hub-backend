package users

import (
	"context"
	"errors"
	"testing"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

type fakeUserRepo struct {
	username string
	err      error
}

func (f *fakeUserRepo) Create(*domain.User) error { return nil }

func (f *fakeUserRepo) GetByUsername(username string) (*domain.User, error) {
	f.username = username
	return &domain.User{Username: username}, f.err
}

func (f *fakeUserRepo) List() ([]*domain.User, error) { return nil, nil }

type fakeProfileRepo struct {
	listLimit       int
	listOffset      int
	getUsername     string
	createdProfile  *domain.UserProfile
	updatedProfile  *domain.UserProfile
	deletedUsername string
	err             error
}

func (f *fakeProfileRepo) ListUsers(limit, offset int) ([]*domain.UserProfile, error) {
	f.listLimit = limit
	f.listOffset = offset
	return []*domain.UserProfile{{Username: "alice"}}, f.err
}

func (f *fakeProfileRepo) GetProfile(username string) (*domain.UserProfile, error) {
	f.getUsername = username
	return &domain.UserProfile{Username: username}, f.err
}

func (f *fakeProfileRepo) CreateUser(profile *domain.UserProfile) error {
	f.createdProfile = profile
	return f.err
}

func (f *fakeProfileRepo) UpdateUser(profile *domain.UserProfile) error {
	f.updatedProfile = profile
	return f.err
}

func (f *fakeProfileRepo) DeleteUser(username string) error {
	f.deletedUsername = username
	return f.err
}

func (f *fakeProfileRepo) GetProfileHeatmap(string) ([]domain.HeatmapEntry, error) {
	return nil, nil
}

func (f *fakeProfileRepo) GetProfileRatingHistory(string) ([]domain.RatingEntry, error) {
	return nil, nil
}

func (f *fakeProfileRepo) GetProfileAttendance(string) ([]domain.AttendanceEntry, error) {
	return nil, nil
}

func (f *fakeProfileRepo) GetProfileSubmissions(string) ([]domain.Submission, error) {
	return nil, nil
}

func TestUsecaseHandlesNilRepositories(t *testing.T) {
	uc := New(nil, nil)
	ctx := context.Background()

	if got, err := uc.GetUser(ctx, "alice"); err != nil || got != nil {
		t.Fatalf("GetUser() = (%v, %v), want nil nil", got, err)
	}
	if err := uc.CreateUser(ctx, &domain.UserProfile{Username: "alice"}); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if err := uc.DeleteUser(ctx, "alice"); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if got, err := uc.GetProfile(ctx, "alice"); err != nil || got != nil {
		t.Fatalf("GetProfile() = (%v, %v), want nil nil", got, err)
	}
	if err := uc.UpdateProfile(ctx, &domain.UserProfile{Username: "alice"}); err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got, err := uc.List(ctx, 10, 0); err != nil || got != nil {
		t.Fatalf("List() = (%v, %v), want nil nil", got, err)
	}
}

func TestUsecaseDelegatesUsersAndProfiles(t *testing.T) {
	userRepo := &fakeUserRepo{}
	profileRepo := &fakeProfileRepo{}
	uc := New(userRepo, profileRepo)
	ctx := context.Background()

	if got, err := uc.GetUser(ctx, "alice"); err != nil || got.Username != "alice" {
		t.Fatalf("GetUser() = (%v, %v), want alice nil", got, err)
	}
	profile := &domain.UserProfile{Username: "alice"}
	if err := uc.CreateUser(ctx, profile); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if err := uc.DeleteUser(ctx, "alice"); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if got, err := uc.GetProfile(ctx, "alice"); err != nil || got.Username != "alice" {
		t.Fatalf("GetProfile() = (%v, %v), want alice nil", got, err)
	}
	if err := uc.UpdateProfile(ctx, profile); err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if got, err := uc.List(ctx, 10, 5); err != nil || got[0].Username != "alice" {
		t.Fatalf("List() = (%v, %v), want alice nil", got, err)
	}

	if userRepo.username != "alice" {
		t.Fatalf("user username = %q, want alice", userRepo.username)
	}
	if profileRepo.createdProfile != profile || profileRepo.updatedProfile != profile {
		t.Fatal("profile create/update did not receive the original profile pointer")
	}
	if profileRepo.deletedUsername != "alice" || profileRepo.getUsername != "alice" {
		t.Fatalf("profile usernames = (%q,%q)", profileRepo.deletedUsername, profileRepo.getUsername)
	}
	if profileRepo.listLimit != 10 || profileRepo.listOffset != 5 {
		t.Fatalf("list args = (%d,%d), want (10,5)", profileRepo.listLimit, profileRepo.listOffset)
	}
}

func TestUsecaseReturnsUserRepositoryErrors(t *testing.T) {
	wantErr := errors.New("repo failed")
	ctx := context.Background()

	if _, err := New(&fakeUserRepo{err: wantErr}, nil).GetUser(ctx, "alice"); !errors.Is(err, wantErr) {
		t.Fatalf("GetUser() error = %v, want %v", err, wantErr)
	}
	if err := New(nil, &fakeProfileRepo{err: wantErr}).CreateUser(ctx, &domain.UserProfile{}); !errors.Is(err, wantErr) {
		t.Fatalf("CreateUser() error = %v, want %v", err, wantErr)
	}
}
