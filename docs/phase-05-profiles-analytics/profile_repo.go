//go:build ignore

// Template for Phase 5 — copy to: internal/infrastructure/databases/profile_repo.go
//
// Real profile analytics from the Phase-2 tables. Each list method returns an
// empty slice (not an error) when there's no data, so the client renders an
// empty heatmap/list instead of getting a 500.
//
package databases

import (
	"context"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ProfileRepositoryDB struct{ client *postgres.Client }

func NewProfileRepositoryDB(c *postgres.Client) *ProfileRepositoryDB {
	return &ProfileRepositoryDB{client: c}
}

func (r *ProfileRepositoryDB) GetProfile(username string) (*domain.UserProfile, error) {
	row := r.client.Pool.QueryRow(context.Background(), `
		SELECT u.username, u.full_name, COALESCE(p.bio,''), COALESCE(p.avatar_url,''),
		       COALESCE(p.rating,0), COALESCE(p.rank,''), COALESCE(p.division,''),
		       (SELECT count(*) FROM user_problems WHERE username=u.username AND solved),
		       (SELECT count(*) FROM attendance WHERE username=u.username AND status='Present')
		FROM users u LEFT JOIN profiles p ON p.username = u.username
		WHERE u.username = $1`, username)

	var up domain.UserProfile
	if err := row.Scan(&up.Username, &up.FullName, &up.Bio, &up.AvatarURL,
		&up.Rating, &up.Rank, &up.Division, &up.SolvedProblems, &up.AttendedContestsCount); err != nil {
		return nil, domain.ErrNotFound("profile not found").Wrap(err)
	}
	return &up, nil
}

func (r *ProfileRepositoryDB) ListUsers() ([]*domain.UserProfile, error) {
	rows, err := r.client.Pool.Query(context.Background(), `
		SELECT u.username, u.full_name, COALESCE(p.bio,''), COALESCE(p.avatar_url,''), COALESCE(p.rating,0)
		FROM users u LEFT JOIN profiles p ON p.username = u.username
		ORDER BY COALESCE(p.rating,0) DESC`)
	if err != nil {
		return nil, domain.ErrInternal("could not list users").Wrap(err)
	}
	defer rows.Close()
	out := []*domain.UserProfile{}
	for rows.Next() {
		var u domain.UserProfile
		if err := rows.Scan(&u.Username, &u.FullName, &u.Bio, &u.AvatarURL, &u.Rating); err == nil {
			out = append(out, &u)
		}
	}
	return out, nil
}

func (r *ProfileRepositoryDB) GetProfileHeatmap(username string) ([]domain.HeatmapEntry, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT to_char(day,'YYYY-MM-DD'), count FROM daily_solves WHERE username=$1 ORDER BY day`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not get heatmap").Wrap(err)
	}
	defer rows.Close()
	out := []domain.HeatmapEntry{}
	for rows.Next() {
		var e domain.HeatmapEntry
		if err := rows.Scan(&e.Date, &e.SolveCount); err == nil {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *ProfileRepositoryDB) GetProfileRatingHistory(username string) ([]domain.RatingEntry, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT to_char(day,'YYYY-MM-DD'), rating FROM rating_history WHERE username=$1 ORDER BY day`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not get rating history").Wrap(err)
	}
	defer rows.Close()
	out := []domain.RatingEntry{}
	for rows.Next() {
		var e domain.RatingEntry
		if err := rows.Scan(&e.Date, &e.Rating); err == nil {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *ProfileRepositoryDB) GetProfileAttendance(username string) ([]domain.AttendanceEntry, error) {
	rows, err := r.client.Pool.Query(context.Background(),
		`SELECT to_char(day,'YYYY-MM-DD'), status FROM attendance WHERE username=$1 ORDER BY day DESC`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not get attendance").Wrap(err)
	}
	defer rows.Close()
	out := []domain.AttendanceEntry{}
	for rows.Next() {
		var e domain.AttendanceEntry
		if err := rows.Scan(&e.Date, &e.Status); err == nil {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *ProfileRepositoryDB) GetProfileSubmissions(username string) ([]domain.Submission, error) {
	rows, err := r.client.Pool.Query(context.Background(), `
		SELECT id, problem_id, problem_title, status, language,
		       COALESCE(execution_time,''), COALESCE(memory_used,''),
		       to_char(created_at,'YYYY-MM-DD"T"HH24:MI:SS"Z"')
		FROM submissions WHERE username=$1 ORDER BY created_at DESC LIMIT 100`, username)
	if err != nil {
		return nil, domain.ErrInternal("could not get submissions").Wrap(err)
	}
	defer rows.Close()
	out := []domain.Submission{}
	for rows.Next() {
		var s domain.Submission
		if err := rows.Scan(&s.ID, &s.ProblemID, &s.ProblemTitle, &s.Status,
			&s.Language, &s.ExecutionTime, &s.MemoryUsed, &s.Timestamp); err == nil {
			out = append(out, s)
		}
	}
	return out, nil
}

// CreateUser / UpdateUser / DeleteUser: implement the remaining ProfileRepository
// methods as straightforward INSERT/UPDATE/DELETE against users + profiles.
