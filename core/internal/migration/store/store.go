package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) ListApplied(ctx context.Context) ([]parser.AppliedMigration, error) {
	const q = `
SELECT version, name, direction, checksum, execution_time_ms, applied_by, applied_at, success, rollback_supported
FROM drift_migrations
ORDER BY version ASC, direction ASC;`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list applied: %w", err)
	}
	defer rows.Close()

	out := make([]parser.AppliedMigration, 0)
	for rows.Next() {
		var m parser.AppliedMigration
		if err := rows.Scan(&m.Version, &m.Name, &m.Direction, &m.Checksum, &m.ExecutionTimeMS, &m.AppliedBy, &m.AppliedAt, &m.Success, &m.RollbackSupported); err != nil {
			return nil, fmt.Errorf("scan applied: %w", err)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied: %w", err)
	}
	return out, nil
}

func (s *Store) Record(ctx context.Context, m parser.Migration, durationMS int64, appliedBy string, success bool, rollbackSupported bool) error {
	const q = `
INSERT INTO drift_migrations (version, name, direction, checksum, execution_time_ms, applied_by, applied_at, success, rollback_supported)
VALUES ($1,$2,$3,$4,$5,$6, NOW(), $7, $8)
ON CONFLICT (version, direction)
DO UPDATE SET
  checksum = EXCLUDED.checksum,
  execution_time_ms = EXCLUDED.execution_time_ms,
  applied_by = EXCLUDED.applied_by,
  applied_at = EXCLUDED.applied_at,
  success = EXCLUDED.success,
  rollback_supported = EXCLUDED.rollback_supported;`
	_, err := s.db.ExecContext(ctx, q, m.Version, m.Name, string(m.Direction), m.Checksum, durationMS, appliedBy, success, rollbackSupported)
	if err != nil {
		return fmt.Errorf("record migration: %w", err)
	}
	return nil
}

func (s *Store) DeleteByVersionDirection(ctx context.Context, version string, direction string) error {
	const q = `DELETE FROM drift_migrations WHERE version = $1 AND direction = $2;`
	if _, err := s.db.ExecContext(ctx, q, version, direction); err != nil {
		return fmt.Errorf("delete migration record: %w", err)
	}
	return nil
}
