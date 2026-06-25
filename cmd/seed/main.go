package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/bereket/cpd-hub-backend/internal/infrastructure/postgres"
)

func main() {
	dsnFlag := flag.String("dsn", "", "Postgres DSN (overrides DATABASE_URL env)")
	seedPath := flag.String("seed", "cmd/seed/seed.sql", "Path to a seed .sql file or a directory of .sql files (executed in alphabetical order)")
	flag.Parse()

	dsn := *dsnFlag
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set. Provide -dsn or set DATABASE_URL env var")
	}

	// connect once
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := postgres.NewClient(ctx, dsn)
	if err != nil {
		log.Fatalf("could not connect to postgres: %v", err)
	}
	defer client.Close()

	info, err := os.Stat(*seedPath)
	if err == nil && info.IsDir() {
		// gather .sql files in directory
		entries := []string{}
		err = filepath.WalkDir(*seedPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if filepath.Ext(d.Name()) == ".sql" {
				entries = append(entries, path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("error scanning seed directory: %v", err)
		}
		if len(entries) == 0 {
			log.Fatalf("no .sql files found in %s", *seedPath)
		}
		sort.Strings(entries)
		for _, p := range entries {
			log.Printf("applying seed %s", p)
			bs, err := os.ReadFile(p)
			if err != nil {
				log.Fatalf("could not read seed file %s: %v", p, err)
			}
			if _, err := client.Pool.Exec(ctx, string(bs)); err != nil {
				log.Fatalf("failed to execute seed %s: %v", p, err)
			}
		}
		fmt.Println("all seeds applied successfully")
		return
	}

	// fallback: single file
	bs, err := os.ReadFile(*seedPath)
	if err != nil {
		log.Fatalf("could not read seed file: %v", err)
	}

	log.Printf("connected to DB, applying seed from %s", *seedPath)
	if _, err := client.Pool.Exec(ctx, string(bs)); err != nil {
		log.Fatalf("failed to execute seed: %v", err)
	}
	fmt.Println("seed applied successfully")
}
