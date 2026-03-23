package problems

import "github.com/bereket/cpd-hub-backend/internal/domain"

// Usecase provides business logic for problems.
type Usecase struct {
	repo domain.ProblemRepository
}

func New(repo domain.ProblemRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) List() ([]*domain.Problem, error) {
	return u.repo.List()
}

func (u *Usecase) GetDaily() (*domain.Problem, error) {
	return u.repo.GetDaily()
}

func (u *Usecase) Like(id string) error {
	return u.repo.Like(id)
}

func (u *Usecase) Dislike(id string) error {
	return u.repo.Dislike(id)
}

func (u *Usecase) MarkSolved(id string) error {
	return u.repo.MarkSolved(id)
}

func (u *Usecase) UnmarkSolved(id string) error {
	return u.repo.UnmarkSolved(id)
}

func (u *Usecase) GetById(id string)  (*domain.Problem, error){
	return u.repo.GetById(id)
}