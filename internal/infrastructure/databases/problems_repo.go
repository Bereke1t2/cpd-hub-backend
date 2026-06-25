package databases

import (
	"context"
	"strings"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

// ProblemsRepositoryDB implements domain.ProblemRepository against Postgres.
// Reads LEFT JOIN user_problems so the calling user's per-user state is always
// returned (false for users who have never interacted with the problem).
// Writes use transactions to keep the denormalized like/dislike counters consistent.
type ProblemsRepositoryDB struct {
	client *postgres.Client
}

func NewProblemsRepositoryDB(client *postgres.Client) *ProblemsRepositoryDB {
	return &ProblemsRepositoryDB{client: client}
}

// selectWithState is the base SELECT that joins the calling user's state.
// $1 = username. Append additional WHERE / ORDER BY / LIMIT as needed.
const selectWithState = `
SELECT p.id, p.title, p.difficulty, p.topic_tags, p.likes, p.dislikes, p.deep_link,
       COALESCE(up.liked,false), COALESCE(up.disliked,false), COALESCE(up.solved,false),
       (SELECT count(*) FROM user_problems s WHERE s.problem_id = p.id AND s.solved) AS solvers
FROM problems p
LEFT JOIN user_problems up ON up.problem_id = p.id AND up.username = $1
`

// scanRow scans one row produced by selectWithState.
func (r *ProblemsRepositoryDB) scanRow(row interface {
	Scan(...interface{}) error
}) (*domain.Problem, int, error) {
	var p domain.Problem
	var tags *string
	var solvers int
	if err := row.Scan(
		&p.ID, &p.Title, &p.Difficulty, &tags,
		&p.Likes, &p.Dislikes, &p.DeepLink,
		&p.IsLiked, &p.IsDisliked, &p.Solved, &solvers,
	); err != nil {
		return nil, 0, err
	}
	p.TopicTags = parseTags(tags)
	return &p, solvers, nil
}

// ListForUser returns all problems with the calling user's per-user state.
func (r *ProblemsRepositoryDB) ListForUser(username string) ([]*domain.Problem, error) {
	rows, err := r.client.Pool.Query(context.Background(), selectWithState, username)
	if err != nil {
		return nil, domain.ErrInternal("could not list problems").Wrap(err)
	}
	defer rows.Close()
	var out []*domain.Problem
	for rows.Next() {
		p, solvers, err := r.scanRow(rows)
		if err != nil {
			continue
		}
		p.SolverCount = solvers
		out = append(out, p)
	}
	return out, nil
}

// GetByIDForUser returns one problem with the calling user's state.
func (r *ProblemsRepositoryDB) GetByIDForUser(username, id string) (*domain.Problem, error) {
	row := r.client.Pool.QueryRow(
		context.Background(),
		selectWithState+" WHERE p.id = $2",
		username, id,
	)
	p, solvers, err := r.scanRow(row)
	if err != nil {
		return nil, domain.ErrNotFound("problem not found").Wrap(err)
	}
	p.SolverCount = solvers
	return p, nil
}

// GetDailyForUser picks a problem deterministically per calendar day (UTC).
// All users see the same problem on the same day; result rotates at midnight UTC.
func (r *ProblemsRepositoryDB) GetDailyForUser(username string) (*domain.Problem, error) {
	today := time.Now().UTC().Format("2006-01-02")
	q := selectWithState + " ORDER BY md5(p.id || $2) LIMIT 1"
	row := r.client.Pool.QueryRow(context.Background(), q, username, today)
	p, solvers, err := r.scanRow(row)
	if err != nil {
		return nil, domain.ErrNotFound("no daily problem").Wrap(err)
	}
	p.SolverCount = solvers
	return p, nil
}

// Like toggles the caller's like (mutually exclusive with dislike).
// One transaction: upsert user_problems row + adjust denormalized counters.
func (r *ProblemsRepositoryDB) Like(username, id string) error {
	ctx := context.Background()
	tx, err := r.client.Pool.Begin(ctx)
	if err != nil {
		return domain.ErrInternal("").Wrap(err)
	}
	defer tx.Rollback(ctx)

	// ensure problem exists
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT true FROM problems WHERE id=$1`, id).Scan(&exists); err != nil {
		return domain.ErrNotFound("problem not found")
	}

	// read previous state
	var liked, disliked bool
	_ = tx.QueryRow(ctx,
		`SELECT liked, disliked FROM user_problems WHERE username=$1 AND problem_id=$2`,
		username, id).Scan(&liked, &disliked)

	newLiked := !liked
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_problems (username, problem_id, liked, disliked, updated_at)
		VALUES ($1,$2,$3,false,now())
		ON CONFLICT (username, problem_id)
		DO UPDATE SET liked=$3, disliked=false, updated_at=now()`,
		username, id, newLiked); err != nil {
		return domain.ErrInternal("").Wrap(err)
	}

	likeDelta := boolToInt(newLiked) - boolToInt(liked)
	dislikeDelta := -boolToInt(disliked)
	if _, err := tx.Exec(ctx,
		`UPDATE problems SET likes = GREATEST(likes+$2,0), dislikes = GREATEST(dislikes+$3,0) WHERE id=$1`,
		id, likeDelta, dislikeDelta); err != nil {
		return domain.ErrInternal("").Wrap(err)
	}
	return tx.Commit(ctx)
}

// Dislike is the mirror-image of Like.
func (r *ProblemsRepositoryDB) Dislike(username, id string) error {
	ctx := context.Background()
	tx, err := r.client.Pool.Begin(ctx)
	if err != nil {
		return domain.ErrInternal("").Wrap(err)
	}
	defer tx.Rollback(ctx)

	// ensure problem exists
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT true FROM problems WHERE id=$1`, id).Scan(&exists); err != nil {
		return domain.ErrNotFound("problem not found")
	}

	var liked, disliked bool
	_ = tx.QueryRow(ctx,
		`SELECT liked, disliked FROM user_problems WHERE username=$1 AND problem_id=$2`,
		username, id).Scan(&liked, &disliked)

	newDisliked := !disliked
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_problems (username, problem_id, liked, disliked, updated_at)
		VALUES ($1,$2,false,$3,now())
		ON CONFLICT (username, problem_id)
		DO UPDATE SET disliked=$3, liked=false, updated_at=now()`,
		username, id, newDisliked); err != nil {
		return domain.ErrInternal("").Wrap(err)
	}

	dislikeDelta := boolToInt(newDisliked) - boolToInt(disliked)
	likeDelta := -boolToInt(liked)
	if _, err := tx.Exec(ctx,
		`UPDATE problems SET dislikes = GREATEST(dislikes+$2,0), likes = GREATEST(likes+$3,0) WHERE id=$1`,
		id, dislikeDelta, likeDelta); err != nil {
		return domain.ErrInternal("").Wrap(err)
	}
	return tx.Commit(ctx)
}

// MarkSolved sets solved=true. On the false→true transition it also bumps
// daily_solves(username, today) and records a submission.
func (r *ProblemsRepositoryDB) MarkSolved(username, id string) error {
	ctx := context.Background()
	tx, err := r.client.Pool.Begin(ctx)
	if err != nil {
		return domain.ErrInternal("").Wrap(err)
	}
	defer tx.Rollback(ctx)

	// fetch title for the submission record
	var title string
	if err := tx.QueryRow(ctx, `SELECT title FROM problems WHERE id=$1`, id).Scan(&title); err != nil {
		return domain.ErrNotFound("problem not found")
	}

	var wasSolved bool
	_ = tx.QueryRow(ctx,
		`SELECT solved FROM user_problems WHERE username=$1 AND problem_id=$2`,
		username, id).Scan(&wasSolved)

	if _, err := tx.Exec(ctx, `
		INSERT INTO user_problems (username, problem_id, solved, solved_at, updated_at)
		VALUES ($1,$2,true,now(),now())
		ON CONFLICT (username, problem_id)
		DO UPDATE SET solved=true, solved_at=COALESCE(user_problems.solved_at,now()), updated_at=now()`,
		username, id); err != nil {
		return domain.ErrInternal("").Wrap(err)
	}

	// bump heatmap and record submission only on false→true transition
	if !wasSolved {
		if _, err := tx.Exec(ctx, `
			INSERT INTO daily_solves (username, day, count) VALUES ($1, CURRENT_DATE, 1)
			ON CONFLICT (username, day) DO UPDATE SET count = daily_solves.count + 1`,
			username); err != nil {
			return domain.ErrInternal("").Wrap(err)
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO submissions (id, username, problem_id, problem_title, status, language, created_at)
			VALUES (gen_random_uuid(), $1, $2, $3, 'Accepted', 'C++', now())`,
			username, id, title); err != nil {
			return domain.ErrInternal("").Wrap(err)
		}
	}
	return tx.Commit(ctx)
}

// UnmarkSolved sets solved=false. Historical daily_solves rows are left intact.
func (r *ProblemsRepositoryDB) UnmarkSolved(username, id string) error {
	_, err := r.client.Pool.Exec(context.Background(),
		`UPDATE user_problems SET solved=false, updated_at=now() WHERE username=$1 AND problem_id=$2`,
		username, id)
	if err != nil {
		return domain.ErrInternal("").Wrap(err)
	}
	return nil
}

// CountSolvers returns the number of users who have solved this problem.
func (r *ProblemsRepositoryDB) CountSolvers(id string) (int, error) {
	var n int
	err := r.client.Pool.QueryRow(context.Background(),
		`SELECT count(*) FROM user_problems WHERE problem_id=$1 AND solved`, id).Scan(&n)
	if err != nil {
		return 0, domain.ErrInternal("").Wrap(err)
	}
	return n, nil
}

// parseTags splits a nullable comma-separated tag string into a slice.
func parseTags(t *string) []string {
	if t == nil || *t == "" {
		return nil
	}
	var out []string
	for _, s := range strings.Split(*t, ",") {
		out = append(out, strings.TrimSpace(s))
	}
	return out
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
