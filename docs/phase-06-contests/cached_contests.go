//go:build ignore

// Template for Phase 6 — copy to: internal/infrastructure/external/cached_contests.go
//
// Wraps the live Kontests/Codeforces fetch with a TTL cache + a graceful-
// degradation ladder: fresh cache -> upstream -> stale cache -> DB list -> [].
//
package external

import (
	"log"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

const contestsCacheKey = "contests:list"

// dbLister is the repo fallback (domain.ContestRepository.List).
type dbLister interface {
	List() ([]*domain.Contest, error)
}

type CachedContests struct {
	live  *KontestsClient
	db    dbLister
	cache *TTLCache
}

func NewCachedContests(live *KontestsClient, db dbLister, ttl time.Duration) *CachedContests {
	return &CachedContests{live: live, db: db, cache: NewTTLCache(ttl)}
}

// List returns contests, never erroring out the request when upstream is down.
func (c *CachedContests) List() ([]*domain.Contest, error) {
	if v, fresh, ok := c.cache.Get(contestsCacheKey); ok && fresh {
		return v.([]*domain.Contest), nil
	}

	fetched, err := c.fetchAndNormalize()
	if err == nil {
		c.cache.Set(contestsCacheKey, fetched)
		return fetched, nil
	}
	log.Printf("contests: upstream failed: %v — falling back", err)

	// stale cache
	if v, _, ok := c.cache.Get(contestsCacheKey); ok {
		return v.([]*domain.Contest), nil
	}
	// DB-backed list
	if c.db != nil {
		if list, derr := c.db.List(); derr == nil {
			return list, nil
		}
	}
	// last resort: empty, but a 200
	return []*domain.Contest{}, nil
}

// fetchAndNormalize pulls from the live client and guarantees the countdown
// fields are set (startTime non-zero, isPast correct), sorted upcoming-first.
func (c *CachedContests) fetchAndNormalize() ([]*domain.Contest, error) {
	list, err := c.live.FetchContests() // adapt to your client's method
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, ct := range list {
		if !ct.StartTime.IsZero() {
			ct.IsPast = ct.StartTime.Before(now)
		}
	}
	// sort: upcoming (soonest first), then past (most recent first)
	sortContests(list, now)
	return list, nil
}
