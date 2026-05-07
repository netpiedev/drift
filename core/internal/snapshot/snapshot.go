package snapshot

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Manager struct {
	dir string
}

func New(dir string) *Manager {
	return &Manager{dir: dir}
}

func (m *Manager) Create(ctx context.Context, dbURL string) (string, error) {
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return "", fmt.Errorf("create snapshot dir: %w", err)
	}
	name := time.Now().UTC().Format("20060102150405") + ".dump"
	path := filepath.Join(m.dir, name)
	cmd := exec.CommandContext(ctx, "pg_dump", "--format=custom", "--file", path, dbURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump failed: %w (%s)", err, string(out))
	}
	return path, nil
}

func (m *Manager) Restore(ctx context.Context, dbURL string, file string) error {
	cmd := exec.CommandContext(ctx, "pg_restore", "--clean", "--if-exists", "--dbname", dbURL, file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore failed: %w (%s)", err, string(out))
	}
	return nil
}
