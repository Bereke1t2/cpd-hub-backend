package contests

import (
	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/external"
)

type Usecase struct {
	repo   domain.ContestRepository
	client *external.KontestsClient
}

func New(repo domain.ContestRepository) *Usecase {
	return &Usecase{repo: repo}
}

func NewWithClient(repo domain.ContestRepository, client *external.KontestsClient) *Usecase {
	return &Usecase{repo: repo, client: client}
}

func (u *Usecase) List() ([]*domain.Contest, error) {
	var final []*domain.Contest
	// fetch platform contests if client available
	if u.client != nil {
		// use a convenience helper to get latest N Codeforces upcoming contests (5)
		cf, err := u.client.FetchUpcomingCodeforces(5)
		if err == nil {
			for i := range cf {
				c := cf[i]
				final = append(final, &c)
			}
		}

		// also fetch recent (past) Codeforces contests (3)
		recent, err := u.client.FetchRecentCodeforces(3)
		if err == nil {
			for i := range recent {
				c := recent[i]
				final = append(final, &c)
			}
		}

		lc, err := u.client.FetchPlatform("leetcode")
		if err == nil {
			for i := range lc {
				c := lc[i]
				final = append(final, &c)
			}
		}
	}
	// merge repo contests
	if u.repo != nil {
		list, err := u.repo.List()
		if err != nil {
			return final, err
		}
		// dedupe by ID
		seen := map[string]struct{}{}
		for _, c := range final {
			seen[c.ID] = struct{}{}
		}
		for _, rc := range list {
			if _, ok := seen[rc.ID]; ok {
				continue
			}
			final = append(final, rc)
		}
	}
	return final, nil
}

func (u *Usecase) Leaderboard(id string) ([]*domain.LeaderboardEntry, error) {
	return u.repo.Leaderboard(id)
}
