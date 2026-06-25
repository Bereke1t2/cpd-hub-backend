package problems

import "github.com/bereket/cpd-hub-backend/internal/domain"

// Usecase provides business logic for problems.
// All operations are scoped to a calling username.
type Usecase struct {
	repo domain.ProblemRepository
}

func New(repo domain.ProblemRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) ListForUser(username string) ([]*domain.Problem, error) {
	return u.repo.ListForUser(username)
}

func (u *Usecase) GetByIDForUser(username, id string) (*domain.Problem, error) {
	return u.repo.GetByIDForUser(username, id)
}

func (u *Usecase) GetDailyForUser(username string) (*domain.Problem, error) {
	return u.repo.GetDailyForUser(username)
}

func (u *Usecase) Like(username, id string) error {
	return u.repo.Like(username, id)
}

func (u *Usecase) Dislike(username, id string) error {
	return u.repo.Dislike(username, id)
}

func (u *Usecase) MarkSolved(username, id string) error {
	return u.repo.MarkSolved(username, id)
}

func (u *Usecase) UnmarkSolved(username, id string) error {
	return u.repo.UnmarkSolved(username, id)
}
