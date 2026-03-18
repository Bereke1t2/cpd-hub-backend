package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Client wraps a pgx connection pool.
type Client struct {
	Pool *pgxpool.Pool
}

// NewClient creates a new Postgres client from a DSN (DATABASE_URL).
// If dsn is empty, it reads from the DATABASE_URL environment variable.
func NewClient(ctx context.Context, dsn string) (*Client, error) {
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		return nil, fmt.Errorf("missing DATABASE_URL")
	}

	// Connect using pgxpool.Connect (v4 compatible)
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return &Client{Pool: pool}, nil
}

// Close closes the underlying pool.
func (c *Client) Close() {
	if c == nil || c.Pool == nil {
		return
	}
	c.Pool.Close()
}

func (c *Client) ensureUsersTable(ctx context.Context) error {
	if c == nil || c.Pool == nil {
		return fmt.Errorf("no db client")
	}
	_, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		full_name TEXT,
		password_hash TEXT,
		rating INT DEFAULT 0
	);
	`)
	return err
}

func (c *Client) EnsureAllTables(ctx context.Context) error {
	if c == nil || c.Pool == nil {
		return fmt.Errorf("no db client")
	}
	// users table
	if _, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		full_name TEXT,
		password_hash TEXT,
		rating INT DEFAULT 0,
		bio TEXT,
		avatar_url TEXT
	);
	`); err != nil {
		return err
	}
	// problems
	if _, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS problems (
		id TEXT PRIMARY KEY,
		title TEXT,
		difficulty TEXT,
		topic_tags TEXT,
		likes INT DEFAULT 0,
		dislikes INT DEFAULT 0,
		deep_link TEXT,
		is_liked BOOLEAN DEFAULT false,
		is_disliked BOOLEAN DEFAULT false,
		solved BOOLEAN DEFAULT false
	);
	`); err != nil {
		return err
	}
	// contests
	if _, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS contests (
		id TEXT PRIMARY KEY,
		title TEXT,
		contest_url TEXT,
		start_time TIMESTAMP,
		duration TEXT,
		platform TEXT,
		number_of_problems INT DEFAULT 0,
		number_of_contestants INT DEFAULT 0,
		date TEXT,
		is_past BOOLEAN DEFAULT false,
		is_participating BOOLEAN DEFAULT false
	);
	`); err != nil {
		return err
	}
	// profiles - lightweight user profile info
	if _, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS profiles (
		username TEXT PRIMARY KEY REFERENCES users(username),
		bio TEXT,
		rating INT DEFAULT 0,
		avatar_url TEXT
	);
	`); err != nil {
		return err
	}
	// activity
	if _, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS activity (
		id TEXT PRIMARY KEY,
		username TEXT,
		action TEXT,
		type TEXT,
		timestamp TEXT
	);
	`); err != nil {
		return err
	}
	// info
	if _, err := c.Pool.Exec(ctx, `
	CREATE TABLE IF NOT EXISTS info (
		title TEXT PRIMARY KEY,
		description TEXT
	);
	`); err != nil {
		return err
	}
	return nil
}
