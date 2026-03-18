package auth

import "github.com/bereket/cpd-hub-backend/internal/domain"

// Usecase for authentication. For now delegates to AuthRepository.

type Usecase struct {
	repo domain.AuthRepository
}

func New(repo domain.AuthRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	return u.repo.Login(req)
}

func (u *Usecase) Signup(req *domain.SignupRequest) (*domain.AuthResponse, error) {
	return u.repo.Signup(req)
}
