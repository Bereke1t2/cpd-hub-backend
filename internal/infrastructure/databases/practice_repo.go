package databases

import (
	"context"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type PracticeRepositoryDB struct {
	client *postgres.Client
}

func NewPracticeRepositoryDB(client *postgres.Client) *PracticeRepositoryDB {
	return &PracticeRepositoryDB{client: client}
}

func (r *PracticeRepositoryDB) ListReviewQueue(username string) ([]*domain.ReviewItem, error) {
	rows, err := r.client.Pool.Query(context.Background(), `
		SELECT problem_id, due_date, interval, ease, repetitions
		FROM review_items
		WHERE username = $1
		ORDER BY due_date, problem_id`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not list review queue").Wrap(err)
	}
	defer rows.Close()

	out := []*domain.ReviewItem{}
	for rows.Next() {
		item, err := scanReview(rows)
		if err == nil {
			out = append(out, item)
		}
	}
	return out, nil
}

func (r *PracticeRepositoryDB) GetReview(username, problemID string) (*domain.ReviewItem, error) {
	row := r.client.Pool.QueryRow(context.Background(), `
		SELECT problem_id, due_date, interval, ease, repetitions
		FROM review_items
		WHERE username = $1 AND problem_id = $2`, username, problemID)
	item, err := scanReview(row)
	if err != nil {
		return nil, domain.ErrNotFound("review item not found").Wrap(err)
	}
	return item, nil
}

func (r *PracticeRepositoryDB) AddReview(username string, item *domain.ReviewItem) (*domain.ReviewItem, error) {
	return r.upsertReview(username, item)
}

func (r *PracticeRepositoryDB) UpdateReview(username string, item *domain.ReviewItem) (*domain.ReviewItem, error) {
	return r.upsertReview(username, item)
}

func (r *PracticeRepositoryDB) DeleteReview(username, problemID string) error {
	tag, err := r.client.Pool.Exec(context.Background(), `
		DELETE FROM review_items
		WHERE username = $1 AND problem_id = $2`, username, problemID)
	if err != nil {
		return domain.ErrInternal("could not delete review item").Wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound("review item not found")
	}
	return nil
}

func (r *PracticeRepositoryDB) ListUpsolves(username string) ([]*domain.UpsolveItem, error) {
	rows, err := r.client.Pool.Query(context.Background(), `
		SELECT contest_id, contest_title, problem_id, problem_title, resolved
		FROM upsolve_items
		WHERE username = $1
		ORDER BY resolved, contest_id, problem_id`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not list upsolves").Wrap(err)
	}
	defer rows.Close()

	out := []*domain.UpsolveItem{}
	for rows.Next() {
		var item domain.UpsolveItem
		if err := rows.Scan(&item.ContestID, &item.ContestTitle, &item.ProblemID, &item.ProblemTitle, &item.Resolved); err == nil {
			out = append(out, &item)
		}
	}
	return out, nil
}

func (r *PracticeRepositoryDB) AddUpsolve(username string, item *domain.UpsolveItem) (*domain.UpsolveItem, error) {
	_, err := r.client.Pool.Exec(context.Background(), `
		INSERT INTO upsolve_items (username, problem_id, contest_id, contest_title, problem_title, resolved)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (username, problem_id) DO UPDATE SET
			contest_id = EXCLUDED.contest_id,
			contest_title = EXCLUDED.contest_title,
			problem_title = EXCLUDED.problem_title,
			resolved = EXCLUDED.resolved`,
		username, item.ProblemID, item.ContestID, item.ContestTitle, item.ProblemTitle, item.Resolved)
	if err != nil {
		return nil, domain.ErrInternal("could not save upsolve").Wrap(err)
	}
	return item, nil
}

func (r *PracticeRepositoryDB) UpdateUpsolve(username, problemID string, resolved bool) (*domain.UpsolveItem, error) {
	row := r.client.Pool.QueryRow(context.Background(), `
		UPDATE upsolve_items
		SET resolved = $3
		WHERE username = $1 AND problem_id = $2
		RETURNING contest_id, contest_title, problem_id, problem_title, resolved`,
		username, problemID, resolved)

	var item domain.UpsolveItem
	if err := row.Scan(&item.ContestID, &item.ContestTitle, &item.ProblemID, &item.ProblemTitle, &item.Resolved); err != nil {
		return nil, domain.ErrNotFound("upsolve item not found").Wrap(err)
	}
	return &item, nil
}

func (r *PracticeRepositoryDB) upsertReview(username string, item *domain.ReviewItem) (*domain.ReviewItem, error) {
	normalizeReviewItem(item)
	due, err := parseReviewDueDate(item.DueDate)
	if err != nil {
		return nil, domain.ErrValidation("due_date must be RFC3339")
	}

	row := r.client.Pool.QueryRow(context.Background(), `
		INSERT INTO review_items (username, problem_id, due_date, interval, ease, repetitions)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (username, problem_id) DO UPDATE SET
			due_date = EXCLUDED.due_date,
			interval = EXCLUDED.interval,
			ease = EXCLUDED.ease,
			repetitions = EXCLUDED.repetitions
		RETURNING problem_id, due_date, interval, ease, repetitions`,
		username, item.ProblemID, due, item.Interval, item.Ease, item.Repetitions)
	saved, err := scanReview(row)
	if err != nil {
		return nil, domain.ErrInternal("could not save review item").Wrap(err)
	}
	return saved, nil
}

type reviewScanner interface {
	Scan(...interface{}) error
}

func scanReview(row reviewScanner) (*domain.ReviewItem, error) {
	var item domain.ReviewItem
	var due time.Time
	if err := row.Scan(&item.ProblemID, &due, &item.Interval, &item.Ease, &item.Repetitions); err != nil {
		return nil, err
	}
	item.DueDate = due.UTC().Format(time.RFC3339)
	return &item, nil
}

func normalizeReviewItem(item *domain.ReviewItem) {
	if item.Ease == 0 {
		item.Ease = domain.DefaultReviewEase
	}
	if item.Ease < domain.MinReviewEase {
		item.Ease = domain.MinReviewEase
	}
	if item.Interval <= 0 {
		item.Interval = 1
	}
	if item.Repetitions < 0 {
		item.Repetitions = 0
	}
	if item.DueDate == "" {
		item.DueDate = time.Now().UTC().Format(time.RFC3339)
	}
}

func parseReviewDueDate(value string) (time.Time, error) {
	if value == "" {
		return time.Now().UTC(), nil
	}
	due, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	return due.UTC(), nil
}
