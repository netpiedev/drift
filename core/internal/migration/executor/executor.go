package executor

import (
	"context"
	"database/sql"
	"fmt"
	"os/user"
	"time"

	"github.com/netpiedev/drift/core/internal/config"
	"github.com/netpiedev/drift/core/internal/migration/parser"
	"github.com/netpiedev/drift/core/internal/migration/runner"
	"github.com/netpiedev/drift/core/internal/migration/store"
	"github.com/netpiedev/drift/core/internal/telemetry"
)

type Options struct {
	DryRun bool
}

type ApplyResult struct {
	Migration         parser.Migration
	Duration          time.Duration
	Statements        []string
	AffectedTables    []string
	EstimatedLocks    []string
	Warnings          []string
	TransactionalSafe bool
}

type Executor struct {
	db      *sql.DB
	store   *store.Store
	metrics *telemetry.Metrics
	cfg     config.Config
}

func New(db *sql.DB, s *store.Store, metrics *telemetry.Metrics, cfg config.Config) *Executor {
	return &Executor{db: db, store: s, metrics: metrics, cfg: cfg}
}

func (e *Executor) ApplyMany(ctx context.Context, migrations []parser.Migration, opts Options) ([]ApplyResult, error) {
	results := make([]ApplyResult, 0, len(migrations))
	for _, m := range migrations {
		res, err := e.ApplyOne(ctx, m, opts)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}

func (e *Executor) ApplyOne(ctx context.Context, m parser.Migration, opts Options) (ApplyResult, error) {
	r, err := runner.ForMigration(m)
	if err != nil {
		return ApplyResult{}, err
	}

	start := time.Now()
	runResult, err := r.Run(ctx, e.db, m, e.cfg, opts.DryRun)
	duration := time.Since(start)

	if err != nil {
		e.metrics.ObserveMigration(m.Name, string(m.Direction), false, duration)
		return ApplyResult{}, fmt.Errorf("apply migration %s: %w", m.Path, err)
	}

	if !opts.DryRun {
		appliedBy := "unknown"
		if current, lookupErr := user.Current(); lookupErr == nil {
			appliedBy = current.Username
		}
		if err := e.store.Record(ctx, m, duration.Milliseconds(), appliedBy, true, runResult.RollbackSupported); err != nil {
			return ApplyResult{}, fmt.Errorf("record migration: %w", err)
		}
	}

	e.metrics.ObserveMigration(m.Name, string(m.Direction), true, duration)

	return ApplyResult{
		Migration:         m,
		Duration:          duration,
		Statements:        runResult.SQL,
		AffectedTables:    runResult.AffectedTables,
		EstimatedLocks:    runResult.EstimatedLocks,
		Warnings:          runResult.Warnings,
		TransactionalSafe: runResult.TransactionalSafe,
	}, nil
}
