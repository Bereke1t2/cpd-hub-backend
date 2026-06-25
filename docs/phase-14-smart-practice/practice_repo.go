//go:build ignore
// Template for Phase 14 — copy to: internal/infrastructure/databases/practice_repo.go
//
// CRUD for review cards and upsolves, every row keyed by (username, problem_id).
// Errors are *domain.AppError. Dates round-trip as RFC3339 strings.
package databases

import (
	"context"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type PracticeRepositoryDB struct{ client *postgres.Client }

func NewPracticeRepositoryDB(c *postgres.Client) *PracticeRepositoryDB {
	return &PracticeRepositoryDB{client: c}
}

// --- Review queue ---

func (r *PracticeRepositoryDB) ListReviewQueue(username string) ([]*domain.ReviewItem, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT problem_id, due_date, interval, ease, repetitions
		   FROM review_items WHERE username=$1 ORDER BY due_date`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not list review queue").Wrap(err)
	}
	defer rows.Close()
	out := []*domain.ReviewItem{}
	for rows.Next() {
		var it domain.ReviewItem
		var due time.Time
		if err := rows.Scan(&it.ProblemID, &due, &it.Interval, &it.Ease, &it.Repetitions); err != nil {
			continue
		}
		it.DueDate = due.Format(time.RFC3339)
		out = append(out, &it)
	}
	return out, nil
}

func (r *PracticeRepositoryDB) AddReview(username string, it *domain.ReviewItem) (*domain.ReviewItem, error) {
	return r.upsertReview(username, it)
}

func (r *PracticeRepositoryDB) UpdateReview(username string, it *domain.ReviewItem) (*domain.ReviewItem, error) {
	return r.upsertReview(username, it)
}

func (r *PracticeRepositoryDB) upsertReview(username string, it *domain.ReviewItem) (*domain.ReviewItem, error) {
	due := parseDate(it.DueDate)
	if it.Ease == 0 {
		it.Ease = 2.5
	}
	if it.Interval == 0 {
		it.Interval = 1
	}
	_, err := r.client.Pool.Exec(context.Background(),
		`INSERT INTO review_items (username, problem_id, due_date, interval, ease, repetitions)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (username, problem_id) DO UPDATE SET
		   due_date=EXCLUDED.due_date, interval=EXCLUDED.interval,
		   ease=EXCLUDED.ease, repetitions=EXCLUDED.repetitions`,
		username, it.ProblemID, due, it.Interval, it.Ease, it.Repetitions)
	if err != nil {
		return nil, domain.ErrInternal("could not save review item").Wrap(err)
	}
	it.DueDate = due.Format(time.RFC3339)
	return it, nil
}

func (r *PracticeRepositoryDB) DeleteReview(username, problemID string) error {
	ct, err := r.client.Pool.Exec(context.Background(),
		`DELETE FROM review_items WHERE username=$1 AND problem_id=$2`, username, problemID)
	if err != nil {
		return domain.ErrInternal("could not delete review item").Wrap(err)
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound("review item not found")
	}
	return nil
}

// --- Upsolves ---

func (r *PracticeRepositoryDB) ListUpsolves(username string) ([]*domain.UpsolveItem, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT contest_id, contest_title, problem_id, problem_title, resolved
		   FROM upsolve_items WHERE username=$1 ORDER BY contest_id, problem_id`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not list upsolves").Wrap(err)
	}
	defer rows.Close()
	out := []*domain.UpsolveItem{}
	for rows.Next() {
		var it domain.UpsolveItem
		if err := rows.Scan(&it.ContestID, &it.ContestTitle, &it.ProblemID, &it.ProblemTitle, &it.Resolved); err != nil {
			continue
		}
		out = append(out, &it)
	}
	return out, nil
}

func (r *PracticeRepositoryDB) AddUpsolve(username string, it *domain.UpsolveItem) (*domain.UpsolveItem, error) {
	_, err := r.client.Pool.Exec(context.Background(),
		`INSERT INTO upsolve_items (username, problem_id, contest_id, contest_title, problem_title, resolved)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (username, problem_id) DO UPDATE SET
		   contest_id=EXCLUDED.contest_id, contest_title=EXCLUDED.contest_title,
		   problem_title=EXCLUDED.problem_title, resolved=EXCLUDED.resolved`,
		username, it.ProblemID, it.ContestID, it.ContestTitle, it.ProblemTitle, it.Resolved)
	if err != nil {
		return nil, domain.ErrInternal("could not save upsolve").Wrap(err)
	}
	return it, nil
}

func (r *PracticeRepositoryDB) UpdateUpsolve(username, problemID string, resolved bool) (*domain.UpsolveItem, error) {
	row := r.client.Pool.QueryRow(context.Background(),
		`UPDATE upsolve_items SET resolved=$3 WHERE username=$1 AND problem_id=$2
		 RETURNING contest_id, contest_title, problem_id, problem_title, resolved`,
		username, problemID, resolved)
	var it domain.UpsolveItem
	if err := row.Scan(&it.ContestID, &it.ContestTitle, &it.ProblemID, &it.ProblemTitle, &it.Resolved); err != nil {
		return nil, domain.ErrNotFound("upsolve item not found")
	}
	return &it, nil
}

func parseDate(s string) time.Time {
	if s == "" {
		return time.Now()
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return time.Now()
}
