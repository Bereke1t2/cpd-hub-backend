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

func assertAcyclic(prereqs map[string][]string) error {
	const (
		white = 0 // unvisited
		gray  = 1 // on the current DFS stack
		black = 2 // done
	)
	color := map[string]int{}
	for n := range prereqs {
		color[n] = white
	}

	var visit func(n string) error
	visit = func(n string) error {
		color[n] = gray
		for _, p := range prereqs[n] {
			switch color[p] {
			case gray:
				return fmt.Errorf("prerequisite cycle through %q -> %q", n, p)
			case white:
				if err := visit(p); err != nil {
					return err
				}
			}
		}
		color[n] = black
		return nil
	}

	for n := range prereqs {
		if color[n] == white {
			if err := visit(n); err != nil {
				return err
			}
		}
	}
	return nil
}

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

		// Verify acyclicity after seed
		rows, err := client.Pool.Query(ctx, "SELECT topic_id, prerequisite_id FROM topic_prerequisites")
		if err == nil {
			defer rows.Close()
			graph := map[string][]string{}
			for rows.Next() {
				var tid, pid string
				if err := rows.Scan(&tid, &pid); err == nil {
					graph[tid] = append(graph[tid], pid)
				}
			}
			if err := assertAcyclic(graph); err != nil {
				log.Fatalf("Critical: topic graph has a cycle! %v", err)
			}
			fmt.Println("Graph acyclicity verified.")
		}
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

	// Verify acyclicity after seed
	rows, err := client.Pool.Query(ctx, "SELECT topic_id, prerequisite_id FROM topic_prerequisites")
	if err == nil {
		defer rows.Close()
		graph := map[string][]string{}
		for rows.Next() {
			var tid, pid string
			if err := rows.Scan(&tid, &pid); err == nil {
				graph[tid] = append(graph[tid], pid)
			}
		}
		if err := assertAcyclic(graph); err != nil {
			log.Fatalf("Critical: topic graph has a cycle! %v", err)
		}
		fmt.Println("Graph acyclicity verified.")
	}
}
