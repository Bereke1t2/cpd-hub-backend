package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	dbURL := flag.String("database", "", "Postgres DSN (overrides DATABASE_URL)")
	dir := flag.String("dir", "migrations", "migrations directory")
	flag.Parse()

	if *dbURL == "" {
		*dbURL = os.Getenv("DATABASE_URL")
	}
	if *dbURL == "" {
		log.Fatal("missing -database flag or DATABASE_URL env var")
	}
	cmd := flag.Arg(0)
	if cmd == "" {
		log.Fatal("usage: migrate [up|down N|version|force V]")
	}

	m, err := migrate.New("file://"+*dir, *dbURL)
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	defer m.Close()

	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		n := 1
		if flag.Arg(1) != "" {
			n, _ = strconv.Atoi(flag.Arg(1))
		}
		err = m.Steps(-n)
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			log.Fatalf("version: %v", verr)
		}
		log.Printf("version=%d dirty=%v", v, dirty)
		return
	case "force":
		v, _ := strconv.Atoi(flag.Arg(1))
		err = m.Force(v)
	default:
		log.Fatalf("unknown command: %s", cmd)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("%s: %v", cmd, err)
	}
	log.Printf("%s: ok", cmd)
}
