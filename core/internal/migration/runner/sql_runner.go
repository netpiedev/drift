package runner

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/netpiedev/drift/core/internal/config"
	"github.com/netpiedev/drift/core/internal/migration/analyzer"
	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type SQLRunner struct{}

func (r SQLRunner) Run(ctx context.Context, db *sql.DB, m parser.Migration, _ config.Config, dryRun bool) (Result, error) {
	sqlText := string(m.Content)
	report := analyzer.AnalyzeSQL(sqlText)
	stmts := splitSQLStatements(sqlText)

	if dryRun {
		return Result{
			SQL:               stmts,
			EstimatedLocks:    report.EstimatedLocks,
			AffectedTables:    report.AffectedTables,
			Warnings:          warningsToMessages(report),
			TransactionalSafe: report.TransactionalOK,
			RollbackSupported: true,
		}, nil
	}

	if len(stmts) == 0 {
		return Result{}, fmt.Errorf("no SQL statements found in %s", m.Path)
	}

	if report.TransactionalOK {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return Result{}, fmt.Errorf("begin tx: %w", err)
		}
		for _, stmt := range stmts {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				_ = tx.Rollback()
				return Result{}, fmt.Errorf("exec statement: %w", err)
			}
		}
		if err := tx.Commit(); err != nil {
			return Result{}, fmt.Errorf("commit tx: %w", err)
		}
	} else {
		for _, stmt := range stmts {
			if _, err := db.ExecContext(ctx, stmt); err != nil {
				return Result{}, fmt.Errorf("exec non-transactional statement: %w", err)
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

func splitSQLStatements(text string) []string {
	parts := strings.Split(text, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		stmt := strings.TrimSpace(p)
		if stmt == "" {
			continue
		}
		out = append(out, stmt+";")
	}
	return out
}

func warningsToMessages(report analyzer.Report) []string {
	out := make([]string, 0, len(report.Warnings))
	for _, w := range report.Warnings {
		out = append(out, w.Message)
	}
	return out
}
