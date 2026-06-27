package consistency

import (
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
)

const dateFmt = "2006-01-02"

type UseCase struct{ repo domain.ConsistencyRepository }

func New(repo domain.ConsistencyRepository) *UseCase { return &UseCase{repo: repo} }

// GetStreak recomputes current/longest/active_days from daily solves, honoring
// one freeze for a single gap day.
func (uc *UseCase) GetStreak(username string, now time.Time) (*domain.Streak, error) {
	days, err := uc.repo.ActiveDays(username) // ascending "YYYY-MM-DD"
	if err != nil {
		return nil, err
	}
	stored, _ := uc.repo.GetStreak(username) // for freezes_available; ignore error -> defaults
	freezes := 2
	if stored != nil {
		freezes = stored.FreezesAvailable
	}

	set := make(map[string]bool, len(days))
	for _, d := range days {
		set[d] = true
	}

	current := currentRun(set, now, &freezes)
	longest := longestRun(days)

	var last *string
	if len(days) > 0 {
		l := days[len(days)-1]
		last = &l
	}
	s := &domain.Streak{
		Current:          current,
		Longest:          max(longest, current),
		LastActiveDay:    last,
		FreezesAvailable: freezes,
		ActiveDays:       days,
	}
	_ = uc.repo.SaveStreak(username, s) // persist the recomputed snapshot
	return s, nil
}

// currentRun counts consecutive days ending today (or yesterday if today has no
// solve yet). A single missing day can be bridged by spending a freeze.
func currentRun(set map[string]bool, now time.Time, freezes *int) int {
	day := now
	if !set[day.Format(dateFmt)] {
		day = day.AddDate(0, 0, -1) // allow "today not done yet"
	}
	run := 0
	for {
		key := day.Format(dateFmt)
		switch {
		case set[key]:
			run++
			day = day.AddDate(0, 0, -1)
		case *freezes > 0 && run > 0:
			*freezes-- // bridge one gap
			day = day.AddDate(0, 0, -1)
			if !set[day.Format(dateFmt)] {
				return run // two gaps in a row -> stop
			}
		default:
			return run
		}
	}
}

func longestRun(daysAsc []string) int {
	if len(daysAsc) == 0 {
		return 0
	}
	best, run := 1, 1
	prev, _ := time.Parse(dateFmt, daysAsc[0])
	for _, d := range daysAsc[1:] {
		cur, _ := time.Parse(dateFmt, d)
		if cur.Sub(prev).Hours() == 24 {
			run++
		} else if cur.Sub(prev).Hours() > 24 {
			run = 1
		}
		if run > best {
			best = run
		}
		prev = cur
	}
	return best
}

// GetGoal returns the stored goal (or a default) with progress recomputed for the
// current period.
func (uc *UseCase) GetGoal(username string, now time.Time) (*domain.Goal, error) {
	g, err := uc.repo.GetGoal(username)
	if err != nil || g == nil {
		g = domain.DefaultGoal(startOfWeek(now).Format(dateFmt))
	}
	progress, _ := uc.repo.SolvedCountSince(username, g.PeriodStart)
	g.Progress = progress
	_ = uc.repo.SaveGoal(username, g)
	return g, nil
}

func (uc *UseCase) SaveGoal(username string, g *domain.Goal) (*domain.Goal, error) {
	if g.Target <= 0 {
		return nil, domain.ErrValidation("target must be positive")
	}
	if err := uc.repo.SaveGoal(username, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (uc *UseCase) GetLadders(username string) ([]*domain.Ladder, error) {
	return uc.repo.GetLadders(username)
}

func startOfWeek(t time.Time) time.Time {
	wd := int(t.Weekday()) // Sunday=0
	return time.Date(t.Year(), t.Month(), t.Day()-wd, 0, 0, 0, 0, t.Location())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
