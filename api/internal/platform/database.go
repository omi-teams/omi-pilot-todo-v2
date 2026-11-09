package platform

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"todo/api/migrations"
)

func OpenDatabase(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	config.MaxConns = 4
	config.ConnConfig.RuntimeParams["statement_timeout"] = "10000"
	config.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = "10000"
	config.ConnConfig.RuntimeParams["lock_timeout"] = "5000"
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	tx, err := pool.Begin(ctx)
	if err == nil {
		// Serialize concurrent startup migrations; the lock lasts only for this transaction.
		_, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(73618002)")
		if err == nil {
			_, err = tx.Exec(ctx, migrations.Todos)
		}
		if err == nil {
			err = tx.Commit(ctx)
		}
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return pool, nil
}
