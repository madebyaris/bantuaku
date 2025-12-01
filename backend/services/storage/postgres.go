package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres wraps a PostgreSQL connection pool
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a new PostgreSQL connection pool
func NewPostgres(databaseURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Postgres{pool: pool}, nil
}

// Close closes the connection pool
func (p *Postgres) Close() {
	p.pool.Close()
}

// Pool returns the underlying connection pool for direct queries
func (p *Postgres) Pool() *pgxpool.Pool {
	return p.pool
}

// Exec executes a query without returning rows
func (p *Postgres) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := p.pool.Exec(ctx, sql, args...)
	return err
}

// QueryRow executes a query that returns at most one row
func (p *Postgres) QueryRow(ctx context.Context, sql string, args ...interface{}) *pgxpool.Pool {
	return p.pool
}
