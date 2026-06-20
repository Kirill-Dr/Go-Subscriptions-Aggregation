package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string, log *slog.Logger) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pingWithRetry(ctx, pool, log); err != nil {
		pool.Close()
		return nil, err
	}

	log.Info("connected to PostgreSQL")
	return pool, nil
}

func pingWithRetry(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	const attempts = 10
	var lastErr error
	for i := 1; i <= attempts; i++ {
		if err := pool.Ping(ctx); err == nil {
			return nil
		} else {
			lastErr = err
			log.Warn("waiting for database", "attempt", i, "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("database not ready after %d attempts: %w", attempts, lastErr)
}
