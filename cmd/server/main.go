package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/bereket/cpd-hub-backend/internal/delivery/httpdelivery"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/config"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/databases"
	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

// loadDotEnv loads a simple .env file with key=value pairs into process environment.
// It ignores empty lines and lines starting with '#' or '//'.
func loadDotEnv(path string) {
	bs, err := os.ReadFile(path)
	if err != nil {
		// not fatal; proceed relying on actual environment
		return
	}
	for _, raw := range strings.Split(string(bs), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		if !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// remove surrounding quotes if present
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if key != "" {
			os.Setenv(key, val)
		}
	}
}

func main() {
	// load .env before reading environment variables
	loadDotEnv(".env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Require DATABASE_URL for DB-backed operation
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set. The server requires a Postgres database to run.")
	}

	ctx := context.Background()
	client, err := postgres.NewClient(ctx, dsn)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}
	defer client.Close()

	if err := postgres.RunMigrations(cfg.Database.URL, "migrations"); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// Wire DB-backed repositories
	repos := httpdelivery.Repos{
		Auth:     postgres.NewAuthRepositoryPG(client),
		Problem:  databases.NewProblemsRepositoryDB(client),
		Contest:  databases.NewContestsRepositoryDB(client),
		Profile:  databases.NewProfileRepositoryDB(client),
		Activity: databases.NewActivityRepositoryDB(client),
		Info:     databases.NewInfoRepositoryDB(client),
	}

	h := httpdelivery.NewHandler(repos, client, cfg.CORS)

	srv := httpdelivery.NewServer(cfg.Server.Address, h.Router())
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
