package contests

import (
	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/external"
)

type Usecase struct {
	repo         domain.ContestRepository
	cachedClient *external.CachedContests
}

func New(repo domain.ContestRepository, cachedClient *external.CachedContests) *Usecase {
	return &Usecase{repo: repo, cachedClient: cachedClient}
}

func (u *Usecase) ListForUser(username string) ([]*domain.Contest, error) {
	if u.cachedClient != nil {
		return u.cachedClient.ListForUser(username)
	}
	return u.repo.ListForUser(username)
}

func (u *Usecase) GetByID(id string) (*domain.Contest, error) {
	return u.repo.GetByID(id)
}

func (u *Usecase) Leaderboard(id string) ([]*domain.LeaderboardEntry, error) {
	return u.repo.Leaderboard(id)
}

func (u *Usecase) Participate(username, id string) error {
	return u.repo.Participate(username, id)
}

func (u *Usecase) Unparticipate(username, id string) error {
	return u.repo.Unparticipate(username, id)
}
