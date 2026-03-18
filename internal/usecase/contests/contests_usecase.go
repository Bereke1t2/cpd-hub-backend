package contests

import "github.com/bereket/cpd-hub-backend/internal/domain"

type Usecase struct {
	repo domain.ContestRepository
}

func New(repo domain.ContestRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) List() ([]*domain.Contest, error) {
	return u.repo.List()
}

func (u *Usecase) Leaderboard(id string) ([]*domain.LeaderboardEntry, error) {
	return u.repo.Leaderboard(id)
}
