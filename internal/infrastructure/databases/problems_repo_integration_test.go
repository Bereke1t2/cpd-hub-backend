//go:build integration

package databases

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func setupIntegrationDB(t *testing.T) *postgres.Client {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	migrationsDir, err := filepath.Abs(filepath.Join(filepath.Dir(file), "../../../migrations"))
	if err != nil {
		t.Fatalf("resolve migrations dir: %v", err)
	}
	m, err := migrate.New("file://"+migrationsDir, dsn)
	if err != nil {
		t.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}
	_, _ = m.Close()

	client, err := postgres.NewClient(context.Background(), dsn)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() { client.Close() })
	t.Cleanup(func() {
		_, _ = client.Pool.Exec(context.Background(), `TRUNCATE user_problems, problems, users RESTART IDENTITY CASCADE`)
	})
	return client
}

func seedProblemUser(t *testing.T, client *postgres.Client, problemID string) {
	t.Helper()
	ctx := context.Background()
	_, err := client.Pool.Exec(ctx, `
		INSERT INTO users (username, email, full_name, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (username) DO NOTHING`, "alice", "alice@example.com", "Alice Example", "hash")
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, err = client.Pool.Exec(ctx, `
		INSERT INTO problems (id, title, difficulty, topic_tags, likes, dislikes, deep_link)
		VALUES ($1, $2, $3, $4, 0, 0, $5)
		ON CONFLICT (id) DO NOTHING`, problemID, "Two Sum", "Easy", "arrays,hashmap", "https://example.com/problems/two-sum")
	if err != nil {
		t.Fatalf("seed problem: %v", err)
	}
}

func TestProblemsRepositoryLike_TogglesTransaction(t *testing.T) {
	client := setupIntegrationDB(t)
	seedProblemUser(t, client, "p1")
	repo := NewProblemsRepositoryDB(client)

	if err := repo.Like("alice", "p1"); err != nil {
		t.Fatalf("Like() error = %v", err)
	}
	var likes, dislikes int
	if err := client.Pool.QueryRow(context.Background(), `SELECT likes, dislikes FROM problems WHERE id=$1`, "p1").Scan(&likes, &dislikes); err != nil {
		t.Fatalf("query problem: %v", err)
	}
	if likes != 1 || dislikes != 0 {
		t.Fatalf("counters after like = (%d,%d), want (1,0)", likes, dislikes)
	}
	if err := repo.Like("alice", "p1"); err != nil {
		t.Fatalf("Like() toggle off error = %v", err)
	}
	if err := client.Pool.QueryRow(context.Background(), `SELECT likes, dislikes FROM problems WHERE id=$1`, "p1").Scan(&likes, &dislikes); err != nil {
		t.Fatalf("query problem after toggle off: %v", err)
	}
	if likes != 0 || dislikes != 0 {
		t.Fatalf("counters after toggle off = (%d,%d), want (0,0)", likes, dislikes)
	}
	var liked bool
	if err := client.Pool.QueryRow(context.Background(), `SELECT liked FROM user_problems WHERE username=$1 AND problem_id=$2`, "alice", "p1").Scan(&liked); err != nil {
		t.Fatalf("query user_problems: %v", err)
	}
	if liked {
		t.Fatal("expected final like state to be false")
	}
}

func TestProblemsRepositoryDailyPick_IsStable(t *testing.T) {
	client := setupIntegrationDB(t)
	seedProblemUser(t, client, "p1")
	seedProblemUser(t, client, "p2")
	repo := NewProblemsRepositoryDB(client)

	first, err := repo.GetDailyForUser("alice")
	if err != nil {
		t.Fatalf("GetDailyForUser() error = %v", err)
	}
	second, err := repo.GetDailyForUser("alice")
	if err != nil {
		t.Fatalf("GetDailyForUser() second call error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("daily pick changed within one run: %q vs %q", first.ID, second.ID)
	}
}
