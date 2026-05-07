package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDirAndValidatePairs(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"001_init.up.sql":   "CREATE TABLE users(id INT);",
		"001_init.down.sql": "DROP TABLE users;",
		"002_add.up.sql":    "ALTER TABLE users ADD COLUMN name TEXT;",
		"002_add.down.sql":  "ALTER TABLE users DROP COLUMN name;",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	migrations, err := ParseDir(dir)
	if err != nil {
		t.Fatalf("ParseDir failed: %v", err)
	}
	if len(migrations) != 4 {
		t.Fatalf("expected 4 migrations, got %d", len(migrations))
	}
	if err := ValidatePairs(migrations); err != nil {
		t.Fatalf("ValidatePairs failed: %v", err)
	}
}
