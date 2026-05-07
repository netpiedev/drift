package integration

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/netpiedev/drift/core/internal/config"
	driftdb "github.com/netpiedev/drift/core/internal/db"
	"github.com/netpiedev/drift/core/internal/migration/executor"
	"github.com/netpiedev/drift/core/internal/migration/parser"
	"github.com/netpiedev/drift/core/internal/migration/plan"
	"github.com/netpiedev/drift/core/internal/migration/store"
	"github.com/netpiedev/drift/core/internal/telemetry"
)

func TestMigrateUpAndRollbackFlow(t *testing.T) {
	dsn := os.Getenv("DRIFT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("DRIFT_TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	db, err := driftdb.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	resetPublicSchema(t, ctx, db)
	_ = db.Close()

	db, err = driftdb.OpenPostgres(ctx, dsn)
	if err != nil {
		t.Fatalf("reopen db: %v", err)
	}
	defer db.Close()

	migrationDir := t.TempDir()
	writeMigrationFile(t, migrationDir, "001_create_users.up.sql", "CREATE TABLE users(id BIGSERIAL PRIMARY KEY, email TEXT NOT NULL);")
	writeMigrationFile(t, migrationDir, "001_create_users.down.sql", "DROP TABLE IF EXISTS users;")
	writeMigrationFile(t, migrationDir, "002_add_note.up.sql", "ALTER TABLE users ADD COLUMN note TEXT;")
	writeMigrationFile(t, migrationDir, "002_add_note.down.sql", "ALTER TABLE users DROP COLUMN IF EXISTS note;")

	migrations, err := parser.ParseDir(migrationDir)
	if err != nil {
		t.Fatalf("parse migrations: %v", err)
	}
	if err := parser.ValidatePairs(migrations); err != nil {
		t.Fatalf("validate pairs: %v", err)
	}

	metrics := telemetry.NewMetrics()
	s := store.New(db)
	exec := executor.New(db, s, metrics, config.Default())

	applied, err := s.ListApplied(ctx)
	if err != nil {
		t.Fatalf("list applied before run: %v", err)
	}

	pending := plan.PendingUp(migrations, applied)
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending migrations, got %d", len(pending))
	}

	if _, err := exec.ApplyMany(ctx, pending, executor.Options{}); err != nil {
		t.Fatalf("apply up migrations: %v", err)
	}

	if !tableExists(t, ctx, db, "users") {
		t.Fatal("expected users table to exist after migrate up")
	}
	if !columnExists(t, ctx, db, "users", "note") {
		t.Fatal("expected users.note column to exist after migrate up")
	}

	applied, err = s.ListApplied(ctx)
	if err != nil {
		t.Fatalf("list applied after up: %v", err)
	}

	rollbackOne, err := plan.NextDown(migrations, applied, 1)
	if err != nil {
		t.Fatalf("plan rollback one: %v", err)
	}
	if len(rollbackOne) != 1 {
		t.Fatalf("expected 1 rollback migration, got %d", len(rollbackOne))
	}
	if _, err := exec.ApplyMany(ctx, rollbackOne, executor.Options{}); err != nil {
		t.Fatalf("apply rollback one: %v", err)
	}
	for _, m := range rollbackOne {
		if err := s.DeleteByVersionDirection(ctx, m.Version, "up"); err != nil {
			t.Fatalf("delete migration state: %v", err)
		}
	}

	if columnExists(t, ctx, db, "users", "note") {
		t.Fatal("expected users.note column to be removed after one rollback")
	}
	if !tableExists(t, ctx, db, "users") {
		t.Fatal("expected users table to still exist after one rollback")
	}

	applied, err = s.ListApplied(ctx)
	if err != nil {
		t.Fatalf("list applied before final rollback: %v", err)
	}

	rollbackFinal, err := plan.NextDown(migrations, applied, 1)
	if err != nil {
		t.Fatalf("plan final rollback: %v", err)
	}
	if len(rollbackFinal) != 1 {
		t.Fatalf("expected 1 final rollback migration, got %d", len(rollbackFinal))
	}
	if _, err := exec.ApplyMany(ctx, rollbackFinal, executor.Options{}); err != nil {
		t.Fatalf("apply final rollback: %v", err)
	}
	for _, m := range rollbackFinal {
		if err := s.DeleteByVersionDirection(ctx, m.Version, "up"); err != nil {
			t.Fatalf("delete final migration state: %v", err)
		}
	}

	if tableExists(t, ctx, db, "users") {
		t.Fatal("expected users table to be removed after final rollback")
	}
}

func writeMigrationFile(t *testing.T, dir string, name string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content+"\n"), 0o644); err != nil {
		t.Fatalf("write migration %s: %v", name, err)
	}
}

func tableExists(t *testing.T, ctx context.Context, db *sql.DB, table string) bool {
	t.Helper()
	var exists bool
	err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name=$1)", table).Scan(&exists)
	if err != nil {
		t.Fatalf("query table exists: %v", err)
	}
	return exists
}

func columnExists(t *testing.T, ctx context.Context, db *sql.DB, table string, column string) bool {
	t.Helper()
	var exists bool
	err := db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 AND column_name=$2)", table, column).Scan(&exists)
	if err != nil {
		t.Fatalf("query column exists: %v", err)
	}
	return exists
}

func resetPublicSchema(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	_, err := db.ExecContext(ctx, `
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
        EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
    END LOOP;
END $$;
`)
	if err != nil {
		t.Fatalf("reset schema tables: %v", err)
	}
}
