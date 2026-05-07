package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func OpenPostgres(ctx context.Context, url string) (*sql.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("database url is empty")
	}

	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	if err := ensureMigrationTable(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureMigrationTable(ctx context.Context, db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS drift_migrations (
    id BIGSERIAL PRIMARY KEY,
    version TEXT NOT NULL,
    name TEXT NOT NULL,
    direction TEXT NOT NULL,
    checksum TEXT NOT NULL,
    execution_time_ms BIGINT NOT NULL,
    applied_by TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL,
    success BOOLEAN NOT NULL,
    rollback_supported BOOLEAN NOT NULL,
    UNIQUE(version, direction)
);`
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		return fmt.Errorf("ensure drift_migrations table: %w", err)
	}
	return nil
}
