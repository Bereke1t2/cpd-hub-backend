package databases

import (
	"context"
	"fmt"
	"strings"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ContestsRepositoryDB struct {
	client *postgres.Client
}

func NewContestsRepositoryDB(client *postgres.Client) *ContestsRepositoryDB {
	return &ContestsRepositoryDB{client: client}
}

const selectContest = `
SELECT 
	id, title, contest_url, start_time, duration, platform, 
	number_of_problems, number_of_contestants, date, is_past,
	EXISTS(SELECT 1 FROM contest_participants p WHERE p.contest_id = contests.id AND p.username = $1) as is_participating
FROM contests
`

func (r *ContestsRepositoryDB) ListForUser(username string) ([]*domain.Contest, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, selectContest, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Contest{}
	for rows.Next() {
		var c domain.Contest
		if err := rows.Scan(&c.ID, &c.Title, &c.ContestURL, &c.StartTime, &c.Duration, &c.Platform, &c.NumberOfProblems, &c.NumberOfContestants, &c.Date, &c.IsPast, &c.IsParticipating); err != nil {
			continue
		}
		out = append(out, &c)
	}
	return out, nil
}

func (r *ContestsRepositoryDB) GetByID(id string) (*domain.Contest, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	row := r.client.Pool.QueryRow(ctx, "SELECT id, title, contest_url, start_time, duration, platform, number_of_problems, number_of_contestants, date, is_past FROM contests WHERE id=$1", id)
	var c domain.Contest
	if err := row.Scan(&c.ID, &c.Title, &c.ContestURL, &c.StartTime, &c.Duration, &c.Platform, &c.NumberOfProblems, &c.NumberOfContestants, &c.Date, &c.IsPast); err != nil {
		return nil, domain.ErrNotFound("contest not found").Wrap(err)
	}
	return &c, nil
}

func (r *ContestsRepositoryDB) Leaderboard(contestID string) ([]*domain.LeaderboardEntry, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT rank, username, rating, score, penalty, problems_solved FROM contest_leaderboard WHERE contest_id=$1 ORDER BY rank", contestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.LeaderboardEntry{}
	for rows.Next() {
		var e domain.LeaderboardEntry
		var solved string
		if err := rows.Scan(&e.Rank, &e.Username, &e.Rating, &e.Score, &e.Penalty, &solved); err != nil {
			continue
		}
		if solved != "" {
			e.ProblemsSolved = strings.Split(solved, ",")
		}
		out = append(out, &e)
	}
	return out, nil
}

func (r *ContestsRepositoryDB) Participate(username, contestID string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "INSERT INTO contest_participants (username, contest_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", username, contestID)
	return err
}

func (r *ContestsRepositoryDB) Unparticipate(username, contestID string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "DELETE FROM contest_participants WHERE username=$1 AND contest_id=$2", username, contestID)
	return err
}
