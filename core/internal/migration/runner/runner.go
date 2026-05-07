package runner

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/netpiedev/drift/core/internal/config"
	"github.com/netpiedev/drift/core/internal/migration/parser"
)

type Result struct {
	SQL               []string
	EstimatedLocks    []string
	AffectedTables    []string
	Warnings          []string
	TransactionalSafe bool
	RollbackSupported bool
}

type Runner interface {
	Run(ctx context.Context, db *sql.DB, m parser.Migration, cfg config.Config, dryRun bool) (Result, error)
}

func ForMigration(m parser.Migration) (Runner, error) {
	switch m.Ext {
	case "sql":
		return SQLRunner{}, nil
	case "ts", "js", "py", "go":
		return ScriptRunner{}, nil
	default:
		return nil, fmt.Errorf("unsupported migration extension: %s", m.Ext)
	}
}
