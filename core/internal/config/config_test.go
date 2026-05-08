package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDriftYAMLFromCurrentWorkingDirectory(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, "drift.yaml")
	content := `environment: staging
database:
  url: postgres://localhost:5432/drift
migrations:
  dir: ./db/migrations
  sequence_type: serial
`
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write drift.yaml: %v", err)
	}

	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Environment != "staging" {
		t.Fatalf("environment = %q, want %q", cfg.Environment, "staging")
	}
	if cfg.Database.URL != "postgres://localhost:5432/drift" {
		t.Fatalf("database.url = %q, want %q", cfg.Database.URL, "postgres://localhost:5432/drift")
	}
	if cfg.Migrations.Dir != "./db/migrations" {
		t.Fatalf("migrations.dir = %q, want %q", cfg.Migrations.Dir, "./db/migrations")
	}
	if cfg.Migrations.SequenceType != "serial" {
		t.Fatalf("migrations.sequence_type = %q, want %q", cfg.Migrations.SequenceType, "serial")
	}
}

func TestLoadReturnsErrorForMalformedDriftYAMLInCurrentWorkingDirectory(t *testing.T) {
	tmp := t.TempDir()
	configFile := filepath.Join(tmp, "drift.yaml")
	content := "environment: [bad\n"
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write malformed drift.yaml: %v", err)
	}

	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previousWD)
	})
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}

	if _, err := Load(""); err == nil {
		t.Fatalf("expected malformed drift.yaml error, got nil")
	}
}
