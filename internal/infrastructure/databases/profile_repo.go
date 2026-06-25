package databases

import (
	"context"
	"fmt"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/domain"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

type ProfileRepositoryDB struct {
	client *postgres.Client
}

func NewProfileRepositoryDB(client *postgres.Client) *ProfileRepositoryDB {
	return &ProfileRepositoryDB{client: client}
}

func (r *ProfileRepositoryDB) ListUsers() ([]*domain.UserProfile, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT users.username, COALESCE(users.full_name,''), COALESCE(profiles.bio,''), COALESCE(profiles.avatar_url,''), COALESCE(profiles.rating, users.rating, 0) FROM users LEFT JOIN profiles ON users.username=profiles.username")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []*domain.UserProfile{}
	for rows.Next() {
		var p domain.UserProfile
		if err := rows.Scan(&p.Username, &p.FullName, &p.Bio, &p.AvatarURL, &p.Rating); err != nil {
			continue
		}
		out = append(out, &p)
	}
	return out, nil
}

func (r *ProfileRepositoryDB) GetProfile(username string) (*domain.UserProfile, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	row := r.client.Pool.QueryRow(ctx, "SELECT users.username, COALESCE(users.full_name,''), COALESCE(profiles.bio,''), COALESCE(profiles.avatar_url,''), COALESCE(profiles.rating, users.rating, 0) FROM users LEFT JOIN profiles ON users.username=profiles.username WHERE users.username=$1", username)
	var p domain.UserProfile
	if err := row.Scan(&p.Username, &p.FullName, &p.Bio, &p.AvatarURL, &p.Rating); err != nil {
		return nil, fmt.Errorf("not found: %w", err)
	}
	return &p, nil
}

func (r *ProfileRepositoryDB) CreateUser(user *domain.UserProfile) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "INSERT INTO users (username, full_name) VALUES ($1,$2)", user.Username, user.FullName)
	if err != nil {
		return err
	}
	_, err = r.client.Pool.Exec(ctx, "INSERT INTO profiles (username, bio, rating, avatar_url) VALUES ($1,$2,$3,$4)", user.Username, user.Bio, user.Rating, user.AvatarURL)
	return err
}

func (r *ProfileRepositoryDB) UpdateUser(user *domain.UserProfile) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "UPDATE profiles SET bio=$2, rating=$3, avatar_url=$4 WHERE username=$1", user.Username, user.Bio, user.Rating, user.AvatarURL)
	return err
}

func (r *ProfileRepositoryDB) DeleteUser(username string) error {
	if r.client == nil || r.client.Pool == nil {
		return fmt.Errorf("no db client")
	}
	ctx := context.Background()
	_, err := r.client.Pool.Exec(ctx, "DELETE FROM profiles WHERE username=$1", username)
	if err != nil {
		return err
	}
	_, err = r.client.Pool.Exec(ctx, "DELETE FROM users WHERE username=$1", username)
	return err
}

// GetProfileHeatmap returns daily solve counts for a user from the daily_solves table.
func (r *ProfileRepositoryDB) GetProfileHeatmap(username string) ([]domain.HeatmapEntry, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT day, count FROM daily_solves WHERE username=$1 ORDER BY day", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.HeatmapEntry{}
	for rows.Next() {
		var d time.Time
		var count int
		if err := rows.Scan(&d, &count); err != nil {
			continue
		}
		out = append(out, domain.HeatmapEntry{
			Date:       d.Format("2006-01-02"),
			SolveCount: count,
		})
	}
	return out, nil
}

// GetProfileRatingHistory returns rating history for a user from the rating_history table.
func (r *ProfileRepositoryDB) GetProfileRatingHistory(username string) ([]domain.RatingEntry, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT day, rating FROM rating_history WHERE username=$1 ORDER BY day", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.RatingEntry{}
	for rows.Next() {
		var d time.Time
		var rating int
		if err := rows.Scan(&d, &rating); err != nil {
			continue
		}
		out = append(out, domain.RatingEntry{
			Date:   d.Format("2006-01-02"),
			Rating: rating,
		})
	}
	return out, nil
}

// GetProfileAttendance returns attendance entries for a user from the attendance table.
func (r *ProfileRepositoryDB) GetProfileAttendance(username string) ([]domain.AttendanceEntry, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx, "SELECT day, status FROM attendance WHERE username=$1 ORDER BY day", username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.AttendanceEntry{}
	for rows.Next() {
		var d time.Time
		var status string
		if err := rows.Scan(&d, &status); err != nil {
			continue
		}
		out = append(out, domain.AttendanceEntry{
			Date:   d.Format("2006-01-02"),
			Status: status,
		})
	}
	return out, nil
}

// GetProfileSubmissions returns recent submissions for a user.
// Reads from the submissions table created by migration 0002.
func (r *ProfileRepositoryDB) GetProfileSubmissions(username string) ([]domain.Submission, error) {
	if r.client == nil || r.client.Pool == nil {
		return nil, fmt.Errorf("no db client")
	}
	ctx := context.Background()
	rows, err := r.client.Pool.Query(ctx,
		"SELECT id, problem_id, problem_title, status, language, "+
			"COALESCE(execution_time,''), COALESCE(memory_used,''), created_at "+
			"FROM submissions WHERE username=$1 ORDER BY created_at DESC LIMIT 100",
		username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Submission{}
	for rows.Next() {
		var id, pid, title, status, lang, execTime, mem string
		var ts time.Time
		if err := rows.Scan(&id, &pid, &title, &status, &lang, &execTime, &mem, &ts); err != nil {
			continue
		}
		out = append(out, domain.Submission{
			ID:            id,
			ProblemID:     pid,
			ProblemTitle:  title,
			Status:        status,
			Language:      lang,
			ExecutionTime: execTime,
			MemoryUsed:    mem,
			Timestamp:     ts.Format(time.RFC3339),
		})
	}
	return out, nil
}
