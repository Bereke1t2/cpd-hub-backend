package users

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

// Usecase coordinates user-related operations across user and profile repositories.
type Usecase struct {
	userRepo    domain.UserRepository
	profileRepo domain.ProfileRepository
}

// New creates a users usecase. Either repo may be nil for limited runtimes.
func New(userRepo domain.UserRepository, profileRepo domain.ProfileRepository) *Usecase {
	return &Usecase{userRepo: userRepo, profileRepo: profileRepo}
}

// GetUser returns the lightweight User record (from users table).
func (u *Usecase) GetUser(ctx context.Context, username string) (*domain.User, error) {
	if u.userRepo == nil {
		return nil, nil
	}
	return u.userRepo.GetByUsername(username)
}

// CreateUser creates a new user and profile (profileRepo handles both inserts when implemented).
func (u *Usecase) CreateUser(ctx context.Context, profile *domain.UserProfile) error {
	if u.profileRepo == nil {
		return nil
	}
	return u.profileRepo.CreateUser(profile)
}

// DeleteUser removes a user/profile.
func (u *Usecase) DeleteUser(ctx context.Context, username string) error {
	if u.profileRepo == nil {
		return nil
	}
	return u.profileRepo.DeleteUser(username)
}

// GetProfile returns the UserProfile for a username.
func (u *Usecase) GetProfile(ctx context.Context, username string) (*domain.UserProfile, error) {
	if u.profileRepo == nil {
		return nil, nil
	}
	return u.profileRepo.GetProfile(username)
}

// UpdateProfile updates an existing profile record.
func (u *Usecase) UpdateProfile(ctx context.Context, profile *domain.UserProfile) error {
	if u.profileRepo == nil {
		return nil
	}
	return u.profileRepo.UpdateUser(profile)
}

// List returns all user profiles.
func (u *Usecase) List(ctx context.Context) ([]*domain.UserProfile, error) {
	if u.profileRepo == nil {
		return nil, nil
	}
	return u.profileRepo.ListUsers()
}
