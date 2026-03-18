package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

func main() {
	dsnFlag := flag.String("dsn", "", "Postgres DSN (overrides DATABASE_URL env)")
	seedPath := flag.String("seed", "internal/infrastructure/postgres/seed.sql", "Path to seed SQL file")
	flag.Parse()

	dsn := *dsnFlag
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set. Provide -dsn or set DATABASE_URL env var")
	}

	bs, err := os.ReadFile(*seedPath)
	if err != nil {
		log.Fatalf("could not read seed file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := postgres.NewClient(ctx, dsn)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}
	defer client.Close()

	log.Printf("connected to DB, applying seed from %s", *seedPath)
	if _, err := client.Pool.Exec(ctx, string(bs)); err != nil {
		log.Fatalf("failed to execute seed: %v", err)
	}
	fmt.Println("seed applied successfully")
}
