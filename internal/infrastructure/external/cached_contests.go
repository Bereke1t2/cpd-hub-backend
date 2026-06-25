package external

import (
	"log"
	"sort"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

const contestsCacheKey = "contests:list"

// dbLister is the repo fallback (domain.ContestRepository.ListForUser).
type dbLister interface {
	ListForUser(username string) ([]*domain.Contest, error)
}

type CachedContests struct {
	live  *KontestsClient
	db    dbLister
	cache *TTLCache
}

func NewCachedContests(live *KontestsClient, db dbLister, ttl time.Duration) *CachedContests {
	return &CachedContests{live: live, db: db, cache: NewTTLCache(ttl)}
}

// ListForUser returns contests, never erroring out when upstream is down.
// It merges live contests with the DB contests.
func (c *CachedContests) ListForUser(username string) ([]*domain.Contest, error) {
	var fetched []*domain.Contest

	if v, fresh, ok := c.cache.Get(contestsCacheKey); ok && fresh {
		fetched = v.([]*domain.Contest)
	} else {
		var err error
		fetched, err = c.fetchAndNormalize()
		if err == nil {
			c.cache.Set(contestsCacheKey, fetched)
		} else {
			log.Printf("contests: upstream failed: %v — falling back", err)
			if v, _, ok := c.cache.Get(contestsCacheKey); ok {
				fetched = v.([]*domain.Contest)
			}
		}
	}

	// Merge with DB-backed list (which contains local contests + participation state)
	var dbList []*domain.Contest
	if c.db != nil {
		dbList, _ = c.db.ListForUser(username)
	}

	return mergeAndDedupe(fetched, dbList), nil
}

func (c *CachedContests) fetchAndNormalize() ([]*domain.Contest, error) {
	// Fetch from multiple platforms as the usecase used to do
	platforms := []string{"codeforces", "leetcode", "code_chef", "at_coder"}
	var all []*domain.Contest

	for _, p := range platforms {
		list, err := c.live.FetchPlatform(p)
		if err != nil {
			continue
		}
		for i := range list {
			all = append(all, &list[i])
		}
	}

	now := time.Now()
	for _, ct := range all {
		if !ct.StartTime.IsZero() {
			ct.IsPast = ct.StartTime.Before(now)
		}
	}

	sortContests(all, now)
	return all, nil
}

func sortContests(list []*domain.Contest, now time.Time) {
	sort.Slice(list, func(i, j int) bool {
		// Upcoming contests by start time ascending
		if !list[i].IsPast && !list[j].IsPast {
			return list[i].StartTime.Before(list[j].StartTime)
		}
		// Upcoming first
		if !list[i].IsPast && list[j].IsPast {
			return true
		}
		if list[i].IsPast && !list[j].IsPast {
			return false
		}
		// Past contests by start time descending
		return list[i].StartTime.After(list[j].StartTime)
	})
}

func mergeAndDedupe(external, db []*domain.Contest) []*domain.Contest {
	seen := make(map[string]*domain.Contest)
	var final []*domain.Contest

	// Add external contests
	for _, c := range external {
		seen[c.ID] = c
		final = append(final, c)
	}

	// Add DB contests, merging participation state if IDs match
	for _, dc := range db {
		if ec, ok := seen[dc.ID]; ok {
			ec.IsParticipating = dc.IsParticipating
			continue
		}
		final = append(final, dc)
	}

	return final
}
