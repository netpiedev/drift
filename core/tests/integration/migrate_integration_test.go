package integration

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestDatabaseConnectivity(t *testing.T) {
	dsn := os.Getenv("DRIFT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("DRIFT_TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}
