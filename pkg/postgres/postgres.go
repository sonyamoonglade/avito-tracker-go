package postgres

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Config struct {
}

type Postgres struct {
	pool *pgxpool.Pool
}

type ReleaseFunc func()

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Exec(ctx context.Context, sql string, args ...interface{}) (*pgconn.CommandTag, ReleaseFunc, error) {
	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("connection acquire error: %w", err)
	}

	tag, err := conn.Exec(ctx, sql, args)
	if err != nil {
		return nil, nil, fmt.Errorf("query error: %w", err)
	}

	return &tag, func() {
		conn.Release()
	}, nil
}

func (p *Postgres) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, ReleaseFunc, error) {

	conn, err := p.pool.Acquire(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("connection acquire error: %w", err)
	}

	rows, err := conn.Query(ctx, sql, args)
	if err != nil {
		return nil, nil, fmt.Errorf("query error: %w", err)
	}

	return rows, func() {
		conn.Release()
	}, nil
}

func (p *Postgres) ScanAll(rows pgx.Rows, dst interface{}) error {
	return pgxscan.ScanAll(dst, rows)
}

func (p *Postgres) ScanOne(rows pgx.Rows, dst interface{}) error {
	return pgxscan.ScanOne(dst, rows)
}

func FromConfig(ctx context.Context, config *Config) (*Postgres, error) {

	panic("implement")

}

func FromConnectionString(ctx context.Context, connString string) (*Postgres, error) {

	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Postgres{
		pool: pool,
	}, nil
}
