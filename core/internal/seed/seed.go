package seed

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Seeder struct {
	db  *sql.DB
	dir string
}

func New(db *sql.DB, dir string) *Seeder {
	return &Seeder{db: db, dir: dir}
}

func (s *Seeder) Run(ctx context.Context, environment string) error {
	seedDir := filepath.Join(s.dir, environment)
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		return fmt.Errorf("read seed dir: %w", err)
	}
	files := make([]string, 0)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(seedDir, e.Name()))
		}
	}
	sort.Strings(files)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("read seed file %s: %w", file, err)
		}
		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("exec seed file %s: %w", file, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	return nil
}

func DeterministicRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
