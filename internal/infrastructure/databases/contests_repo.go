package databases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ContestsRepositoryDB struct {
	client *postgres.Client
}

func NewContestsRepositoryDB(client *postgres.Client) *ContestsRepositoryDB {
	return &ContestsRepositoryDB{client: client}
}

func (r *ContestsRepositoryDB) List() ([]*domain.Contest, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT id, title, contest_url, start_time, duration, platform, number_of_problems, number_of_contestants, date, is_past, is_participating FROM contests")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.Contest{}
	for rows.Next() {
		var c domain.Contest
		var start time.Time
		if err := rows.Scan(&c.ID, &c.Title, &c.ContestURL, &start, &c.Duration, &c.Platform, &c.NumberOfProblems, &c.NumberOfContestants, &c.Date, &c.IsPast, &c.IsParticipating); err != nil {
			continue
		}
		c.StartTime = start
		out = append(out, &c)
	}
	return out, nil
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
