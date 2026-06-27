package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ConsistencyRepositoryDB struct{ client *postgres.Client }

func NewConsistencyRepositoryDB(c *postgres.Client) *ConsistencyRepositoryDB {
	return &ConsistencyRepositoryDB{client: c}
}

func (r *ConsistencyRepositoryDB) GetStreak(username string) (*domain.Streak, error) {
	row := r.client.Pool.QueryRow(context.Background(),
		`SELECT current, longest, to_char(last_active_day,'YYYY-MM-DD'), freezes_available
		 FROM streaks WHERE username=$1`, username)
	var s domain.Streak
	var last *string
	if err := row.Scan(&s.Current, &s.Longest, &last, &s.FreezesAvailable); err != nil {
		return nil, domain.ErrNotFound("no streak").Wrap(err)
	}
	s.LastActiveDay = last
	s.ActiveDays = []string{} // filled by usecase via ActiveDays()
	return &s, nil
}

func (r *ConsistencyRepositoryDB) SaveStreak(username string, s *domain.Streak) error {
	_, err := r.client.Pool.Exec(context.Background(), `
		INSERT INTO streaks (username, current, longest, last_active_day, freezes_available)
		VALUES ($1,$2,$3,$4::date,$5)
		ON CONFLICT (username) DO UPDATE SET
			current=$2, longest=$3, last_active_day=$4::date, freezes_available=$5`,
		username, s.Current, s.Longest, nullableDate(s.LastActiveDay), s.FreezesAvailable)
	if err != nil {
		return domain.ErrInternal("could not save streak").Wrap(err)
	}
	return nil
}

func (r *ConsistencyRepositoryDB) GetGoal(username string) (*domain.Goal, error) {
	row := r.client.Pool.QueryRow(context.Background(),
		`SELECT id, type, target, progress, to_char(period_start,'YYYY-MM-DD')
		 FROM goals WHERE username=$1 ORDER BY period_start DESC LIMIT 1`, username)
	var g domain.Goal
	if err := row.Scan(&g.ID, &g.Type, &g.Target, &g.Progress, &g.PeriodStart); err != nil {
		return nil, domain.ErrNotFound("no goal").Wrap(err)
	}
	return &g, nil
}

func (r *ConsistencyRepositoryDB) SaveGoal(username string, g *domain.Goal) error {
	_, err := r.client.Pool.Exec(context.Background(), `
		INSERT INTO goals (username, id, type, target, progress, period_start)
		VALUES ($1,$2,$3,$4,$5,$6::date)
		ON CONFLICT (username, id) DO UPDATE SET
			type=$3, target=$4, progress=$5, period_start=$6::date`,
		username, g.ID, g.Type, g.Target, g.Progress, g.PeriodStart)
	if err != nil {
		return domain.ErrInternal("could not save goal").Wrap(err)
	}
	return nil
}

// ActiveDays returns the distinct days the user solved at least one problem, asc.
func (r *ConsistencyRepositoryDB) ActiveDays(username string) ([]string, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT to_char(day,'YYYY-MM-DD') FROM daily_solves WHERE username=$1 AND count>0 ORDER BY day`, username)
	if err != nil {
		return nil, domain.ErrInternal("").Wrap(err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var d string
		if rows.Scan(&d) == nil {
			out = append(out, d)
		}
	}
	return out, nil
}

func (r *ConsistencyRepositoryDB) SolvedCountSince(username, sinceDay string) (int, error) {
	var n int
	err := r.client.Pool.QueryRow(context.Background(),
		`SELECT COALESCE(sum(count),0) FROM daily_solves WHERE username=$1 AND day >= $2::date`,
		username, sinceDay).Scan(&n)
	if err != nil {
		return 0, domain.ErrInternal("").Wrap(err)
	}
	return n, nil
}

func (r *ConsistencyRepositoryDB) GetLadders(username string) ([]*domain.Ladder, error) {
	rows, err := r.client.Pool.Query(context.Background(), `
		SELECT l.id, l.title, l.from_rating, l.to_rating,
		       r.problem_id, r.rating, r.topic_id,
		       COALESCE(up.solved,false)
		FROM ladders l
		JOIN ladder_rungs r ON r.ladder_id = l.id
		LEFT JOIN user_problems up ON up.problem_id = r.problem_id AND up.username = $1
		ORDER BY l.from_rating, r.ord`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not load ladders").Wrap(err)
	}
	defer rows.Close()

	byID := map[string]*domain.Ladder{}
	order := []string{}
	for rows.Next() {
		var id, title string
		var from, to, rating int
		var pid string
		var topic *string
		var solved bool
		if err := rows.Scan(&id, &title, &from, &to, &pid, &rating, &topic, &solved); err != nil {
			continue
		}
		l, ok := byID[id]
		if !ok {
			l = &domain.Ladder{ID: id, Title: title, FromRating: from, ToRating: to}
			byID[id] = l
			order = append(order, id)
		}
		l.Rungs = append(l.Rungs, domain.LadderRung{ProblemID: pid, Rating: rating, TopicID: topic, Solved: solved})
	}
	out := make([]*domain.Ladder, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out, nil
}

func (r *ConsistencyRepositoryDB) SaveLadder(username string, l *domain.Ladder) error {
	// Optional: persist per-user rung overrides.
	return nil
}

func nullableDate(s *string) interface{} {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
