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

// Ping checks DB connectivity. Used by the readiness probe.
func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.Pool == nil {
		return fmt.Errorf("no db client")
	}
	return c.Pool.Ping(ctx)
}

// Close closes the underlying pool.
func (c *Client) Close() {
	if c == nil || c.Pool == nil {
		return
	}
	c.Pool.Close()
}

