package runner

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/netpiedev/drift/core/internal/config"
	"github.com/netpiedev/drift/core/internal/migration/analyzer"
	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type ScriptRunner struct{}

type scriptResult struct {
	SQL []string `json:"sql"`
}

func (r ScriptRunner) Run(ctx context.Context, db *sql.DB, m parser.Migration, cfg config.Config, dryRun bool) (Result, error) {
	stmts, err := runScriptExtraction(ctx, m, cfg)
	if err != nil {
		return Result{}, err
	}
	combined := strings.Join(stmts, "\n")
	report := analyzer.AnalyzeSQL(combined)

	if !dryRun {
		if report.TransactionalOK {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				return Result{}, fmt.Errorf("begin tx: %w", err)
			}
			for _, stmt := range stmts {
				if _, err := tx.ExecContext(ctx, stmt); err != nil {
					_ = tx.Rollback()
					return Result{}, fmt.Errorf("exec statement in tx: %w", err)
				}
			}
			if err := tx.Commit(); err != nil {
				return Result{}, fmt.Errorf("commit tx: %w", err)
			}
		} else {
			for _, stmt := range stmts {
				if _, err := db.ExecContext(ctx, stmt); err != nil {
					return Result{}, fmt.Errorf("exec statement: %w", err)
				}
			}
		}
	}

	return Result{
		SQL:               stmts,
		EstimatedLocks:    report.EstimatedLocks,
		AffectedTables:    report.AffectedTables,
		Warnings:          warningsToMessages(report),
		TransactionalSafe: report.TransactionalOK,
		RollbackSupported: true,
	}, nil
}

func runScriptExtraction(ctx context.Context, m parser.Migration, cfg config.Config) ([]string, error) {
	switch m.Ext {
	case "js", "ts":
		return runJSLike(ctx, m, cfg)
	case "py":
		return runPython(ctx, m, cfg)
	case "go":
		return runGo(ctx, m, cfg)
	default:
		return nil, fmt.Errorf("unsupported script migration extension: %s", m.Ext)
	}
}

func runJSLike(ctx context.Context, m parser.Migration, cfg config.Config) ([]string, error) {
	runtime := cfg.Runners.Bun
	if runtime == "" {
		runtime = "bun"
	}
	launcher := `
import path from "node:path";
import { pathToFileURL } from "node:url";
const migrationPath = process.argv[2];
const direction = process.argv[3];
const mod = await import(pathToFileURL(path.resolve(migrationPath)).href);
const fn = mod[direction];
if (typeof fn !== "function") {
  throw new Error("missing exported function: " + direction);
}
const captured = [];
const db = {
  exec: async (sql) => { if (sql) captured.push(sql); },
  query: async (sql) => { if (sql) captured.push(sql); return { rows: [] }; },
  sql: (sql) => { if (sql) captured.push(sql); }
};
const returned = await fn(db);
if (typeof returned === "string") captured.push(returned);
if (Array.isArray(returned)) captured.push(...returned);
process.stdout.write(JSON.stringify({ sql: captured }));
`
	launcherPath, err := writeTempScript("drift-js-runner-*.mjs", launcher)
	if err != nil {
		return nil, err
	}
	defer os.Remove(launcherPath)

	cmd := exec.CommandContext(ctx, runtime, launcherPath, m.Path, string(m.Direction))
	out, err := cmd.CombinedOutput()
	if err != nil {
		fallback := cfg.Runners.Node
		if runtime != fallback && fallback != "" {
			cmd = exec.CommandContext(ctx, fallback, launcherPath, m.Path, string(m.Direction))
			out, err = cmd.CombinedOutput()
		}
		if err != nil {
			return nil, fmt.Errorf("run %s migration: %w (%s)", m.Ext, err, strings.TrimSpace(string(out)))
		}
	}
	return parseScriptJSON(out)
}

func runPython(ctx context.Context, m parser.Migration, cfg config.Config) ([]string, error) {
	runtime := cfg.Runners.Python
	if runtime == "" {
		runtime = "python3"
	}
	launcher := `
import importlib.util
import json
import os
import sys

path = sys.argv[1]
direction = sys.argv[2]
spec = importlib.util.spec_from_file_location("drift_migration", path)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
fn = getattr(module, direction, None)
if fn is None:
    raise Exception(f"missing function: {direction}")

captured = []
class DB:
    def exec(self, sql):
        if sql:
            captured.append(sql)
    def query(self, sql):
        if sql:
            captured.append(sql)
        return []

db = DB()
returned = fn(db)
if isinstance(returned, str):
    captured.append(returned)
if isinstance(returned, list):
    captured.extend(returned)
print(json.dumps({"sql": captured}))
`
	launcherPath, err := writeTempScript("drift-py-runner-*.py", launcher)
	if err != nil {
		return nil, err
	}
	defer os.Remove(launcherPath)

	cmd := exec.CommandContext(ctx, runtime, launcherPath, m.Path, string(m.Direction))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run python migration: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return parseScriptJSON(out)
}

func runGo(ctx context.Context, m parser.Migration, cfg config.Config) ([]string, error) {
	runtime := cfg.Runners.Go
	if runtime == "" {
		runtime = "go"
	}
	harness := `
package main

import (
  "encoding/json"
  "fmt"
  "os"
)

func main() {
  // Contract: go migration files are executable and should print JSON: {"sql": ["..."]}
  // This shim exists for interface parity and clearer errors.
  fmt.Fprintln(os.Stderr, "go migration contract: executable file must print JSON {\"sql\": [...]}")
  os.Exit(2)
}
`
	_ = harness
	cmd := exec.CommandContext(ctx, runtime, "run", m.Path)
	cmd.Env = append(os.Environ(), "DRIFT_DIRECTION="+string(m.Direction))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run go migration: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return parseScriptJSON(out)
}

func parseScriptJSON(out []byte) ([]string, error) {
	var parsed scriptResult
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("parse script output as JSON: %w (output=%s)", err, strings.TrimSpace(string(out)))
	}
	clean := make([]string, 0, len(parsed.SQL))
	for _, stmt := range parsed.SQL {
		s := strings.TrimSpace(stmt)
		if s == "" {
			continue
		}
		if !strings.HasSuffix(s, ";") {
			s += ";"
		}
		clean = append(clean, s)
	}
	if len(clean) == 0 {
		return nil, fmt.Errorf("script migration returned no SQL statements")
	}
	return clean, nil
}

func writeTempScript(pattern string, content string) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("create temp runner: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(strings.TrimSpace(content)); err != nil {
		return "", fmt.Errorf("write temp runner: %w", err)
	}
	return filepath.Clean(f.Name()), nil
}
