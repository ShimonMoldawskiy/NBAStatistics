package db

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/ShimonMoldawskiy/NBAStatistics/common"
)

type PostgresDatabase struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func NewPostgresDatabase(ctx context.Context, connString string) (*PostgresDatabase, error) {
	pool, err := pgxpool.Connect(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &PostgresDatabase{
		pool: pool,
		ctx:  ctx,
	}, nil
}

func (p *PostgresDatabase) Exec(query string, args ...interface{}) error {
	_, err := p.pool.Exec(p.ctx, query, args...)
	return err
}

func (p *PostgresDatabase) QueryRow(query string, args ...interface{}) common.Row {
	return p.pool.QueryRow(p.ctx, query, args...)
}

func (p *PostgresDatabase) Query(query string, args ...interface{}) (common.Rows, error) {
	return p.pool.Query(p.ctx, query, args...)
}

func (p *PostgresDatabase) Close() {
	p.pool.Close()
}
