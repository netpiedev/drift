package cli

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/netpiedev/drift/core/internal/migration/diff"
	"github.com/netpiedev/drift/core/internal/migration/executor"
	"github.com/netpiedev/drift/core/internal/migration/lint"
	"github.com/netpiedev/drift/core/internal/migration/parser"
	"github.com/netpiedev/drift/core/internal/migration/plan"
	"github.com/netpiedev/drift/core/internal/migration/status"
	"github.com/netpiedev/drift/core/internal/migration/store"
	"github.com/netpiedev/drift/core/internal/migration/verify"
	"github.com/netpiedev/drift/core/internal/safety"
	"github.com/netpiedev/drift/core/internal/ui"
	"github.com/spf13/cobra"
)

type diffOptions struct {
	FromURL string
	ToURL   string
	Schema  string
	Write   bool
	Name    string
}

func newMigrateCommand(configPath *string, dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{Use: "migrate", Short: "Migration operations"}
	cmd.AddCommand(newMigrateUpCommand(configPath, dryRun))
	cmd.AddCommand(newMigrateDownCommand(configPath, dryRun))
	cmd.AddCommand(newMigrateRollbackCommand(configPath, dryRun))
	cmd.AddCommand(newMigrateStatusCommand(configPath))
	cmd.AddCommand(newMigrateVerifyCommand(configPath))
	cmd.AddCommand(newMigrateDoctorCommand(configPath))
	cmd.AddCommand(newMigrateLintCommand(configPath))
	cmd.AddCommand(newMigrateDiffCommand(configPath))
	return cmd
}

func newMigrateUpCommand(configPath *string, dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Apply all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, s, files, applied, err := loadMigrationContext(*configPath, true)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			if err := safety.EnforceSafety(rt.Ctx, rt.DB, rt.Config, "migrate up"); err != nil {
				return err
			}

			if err := abortOnValidationIssues(files, applied); err != nil {
				return err
			}

			pending := plan.PendingUp(files, applied)
			if len(pending) == 0 {
				cmd.Println("no pending migrations")
				return nil
			}

			exec := executor.New(rt.DB, s, rt.Metrics, rt.Config)
			results, err := exec.ApplyMany(rt.Ctx, pending, executor.Options{DryRun: *dryRun})
			if err != nil {
				return err
			}
			cmd.Println(ui.RenderMigrationResults(results, *dryRun))
			return nil
		},
	}
}

func newMigrateDownCommand(configPath *string, dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Rollback one migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeRollback(cmd.Context(), *configPath, *dryRun, 1, cmd)
		},
	}
}

func newMigrateRollbackCommand(configPath *string, dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "rollback <n>",
		Short: "Rollback the last n migrations",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var steps int
			if _, err := fmt.Sscanf(args[0], "%d", &steps); err != nil || steps <= 0 {
				return fmt.Errorf("invalid rollback steps: %s", args[0])
			}
			return executeRollback(cmd.Context(), *configPath, *dryRun, steps, cmd)
		},
	}
}

func newMigrateStatusCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, _, files, applied, err := loadMigrationContext(*configPath, true)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			rows := status.BuildRows(files, applied)
			for _, r := range rows {
				state := "pending"
				if r.Applied {
					state = "applied"
				}
				cmd.Printf("%s %-28s %s\n", r.Version, r.Name, state)
			}
			return nil
		},
	}
}

func newMigrateVerifyCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Verify migration ordering and checksums",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, _, files, applied, err := loadMigrationContext(*configPath, true)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			issues := append(verify.ValidateOrdered(files), verify.ValidateChecksums(files, applied)...)
			if len(issues) == 0 {
				cmd.Println("verification passed")
				return nil
			}
			for _, i := range issues {
				cmd.Printf("[%s] %s\n", i.Severity, i.Message)
			}
			return fmt.Errorf("verification failed")
		},
	}
}

func newMigrateDoctorCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics and safety checks",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, _, files, applied, err := loadMigrationContext(*configPath, true)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			issues := make([]string, 0)
			if err := parser.ValidatePairs(files); err != nil {
				issues = append(issues, "pair validation: "+err.Error())
			}
			for _, i := range verify.ValidateOrdered(files) {
				issues = append(issues, i.Message)
			}
			for _, i := range verify.ValidateChecksums(files, applied) {
				issues = append(issues, i.Message)
			}
			for _, i := range lint.Analyze(files) {
				issues = append(issues, i.Message+" ("+i.File+")")
			}
			cmd.Println(ui.RenderWarnings(issues))
			if len(issues) > 0 {
				return fmt.Errorf("doctor detected %d issue(s)", len(issues))
			}
			return nil
		},
	}
}

func newMigrateLintCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Lint migration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, _, files, _, err := loadMigrationContext(*configPath, false)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			issues := lint.Analyze(files)
			cmd.Println(lint.FormatIssues(issues))
			for _, issue := range issues {
				if issue.Severity == "error" {
					return fmt.Errorf("lint errors found")
				}
			}
			return nil
		},
	}
}

func newMigrateDiffCommand(configPath *string) *cobra.Command {
	opts := diffOptions{}
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff source DB schema against target DB schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(opts.ToURL) == "" {
				return errors.New("--to-url is required")
			}
			return runDiff(cmd.Context(), *configPath, opts, cmd)
		},
	}
	cmd.Flags().StringVar(&opts.FromURL, "from-url", "", "source database URL (defaults to database.url from config)")
	cmd.Flags().StringVar(&opts.ToURL, "to-url", "", "target database URL")
	cmd.Flags().StringVar(&opts.Schema, "schema", "public", "schema name")
	cmd.Flags().BoolVar(&opts.Write, "write", false, "generate a migration pair into migrations.dir")
	cmd.Flags().StringVar(&opts.Name, "name", "schema_diff", "name used when generating migration files")
	return cmd
}

func executeRollback(ctx context.Context, configPath string, dryRun bool, steps int, cmd *cobra.Command) error {
	rt, s, files, applied, err := loadMigrationContext(configPath, true)
	if err != nil {
		return err
	}
	defer closeRuntime(rt)

	if err := safety.EnforceSafety(rt.Ctx, rt.DB, rt.Config, "migrate rollback"); err != nil {
		return err
	}

	downMigrations, err := plan.NextDown(files, applied, steps)
	if err != nil {
		return err
	}
	if len(downMigrations) == 0 {
		cmd.Println("no applied migrations to rollback")
		return nil
	}

	exec := executor.New(rt.DB, s, rt.Metrics, rt.Config)
	results, err := exec.ApplyMany(ctx, downMigrations, executor.Options{DryRun: dryRun})
	if err != nil {
		return err
	}
	if !dryRun {
		for _, migration := range downMigrations {
			if err := s.DeleteByVersionDirection(ctx, migration.Version, "up"); err != nil {
				return err
			}
		}
	}
	cmd.Println(ui.RenderMigrationResults(results, dryRun))
	return nil
}

func loadMigrationContext(configPath string, needsDB bool) (*Runtime, *store.Store, []parser.Migration, []parser.AppliedMigration, error) {
	rt, err := BuildRuntime(context.Background(), configPath, needsDB)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	files, err := parser.ParseDir(rt.Config.Migrations.Dir)
	if err != nil {
		closeRuntime(rt)
		return nil, nil, nil, nil, err
	}
	if err := parser.ValidatePairs(files); err != nil {
		closeRuntime(rt)
		return nil, nil, nil, nil, err
	}

	if !needsDB {
		return rt, nil, files, nil, nil
	}

	s := store.New(rt.DB)
	applied, err := s.ListApplied(rt.Ctx)
	if err != nil {
		closeRuntime(rt)
		return nil, nil, nil, nil, err
	}
	return rt, s, files, applied, nil
}

func abortOnValidationIssues(files []parser.Migration, applied []parser.AppliedMigration) error {
	issues := append(verify.ValidateOrdered(files), verify.ValidateChecksums(files, applied)...)
	for _, issue := range issues {
		if issue.Severity == "error" {
			return fmt.Errorf("validation failed: %s", issue.Message)
		}
	}
	return nil
}

func runDiff(ctx context.Context, configPath string, opts diffOptions, cmd *cobra.Command) error {
	rt, err := BuildRuntime(ctx, configPath, false)
	if err != nil {
		return err
	}
	defer closeRuntime(rt)

	fromURL := strings.TrimSpace(opts.FromURL)
	if fromURL == "" {
		fromURL = strings.TrimSpace(rt.Config.Database.URL)
	}
	if fromURL == "" {
		return errors.New("source database URL is empty: set --from-url or database.url")
	}

	fromDB, err := openSecondaryDB(ctx, fromURL)
	if err != nil {
		return fmt.Errorf("open source database: %w", err)
	}
	defer fromDB.Close()

	toDB, err := openSecondaryDB(ctx, opts.ToURL)
	if err != nil {
		return fmt.Errorf("open target database: %w", err)
	}
	defer toDB.Close()

	fromSchema, err := diff.Introspect(ctx, fromDB, opts.Schema)
	if err != nil {
		return err
	}
	toSchema, err := diff.Introspect(ctx, toDB, opts.Schema)
	if err != nil {
		return err
	}

	changes := diff.Compare(fromSchema, toSchema)
	if len(changes) == 0 {
		cmd.Println("no schema differences")
		return nil
	}

	for _, c := range changes {
		cmd.Printf("[%s] %s\n%s\n", c.Type, c.Description, c.SQL)
	}

	if opts.Write {
		upPath, downPath, err := generateDiffMigrationFiles(rt.Config.Migrations.Dir, opts.Name, changes)
		if err != nil {
			return err
		}
		cmd.Printf("generated migration files:\n- %s\n- %s\n", upPath, downPath)
	}

	return nil
}

func generateDiffMigrationFiles(dir string, name string, changes []diff.Change) (string, string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", fmt.Errorf("create migrations dir: %w", err)
	}

	version := time.Now().UTC().Format("200601021504")
	safeName := sanitizeDiffName(name)
	prefix := version + "_" + safeName

	upPath := filepath.Join(dir, prefix+".up.sql")
	downPath := filepath.Join(dir, prefix+".down.sql")

	upLines := make([]string, 0, len(changes)+2)
	downLines := make([]string, 0, len(changes)+4)
	upLines = append(upLines, "-- Generated by drift diff")
	downLines = append(downLines, "-- Generated by drift diff")
	downLines = append(downLines, "-- Review this rollback manually before production use")

	for _, c := range changes {
		upLines = append(upLines, c.SQL)
		reverse := strings.TrimSpace(c.ReverseSQL)
		if reverse == "" {
			reverse = "-- manual rollback required"
		}
		downLines = append(downLines, reverse)
	}

	if err := os.WriteFile(upPath, []byte(strings.Join(upLines, "\n")+"\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("write up migration: %w", err)
	}
	if err := os.WriteFile(downPath, []byte(strings.Join(downLines, "\n")+"\n"), 0o644); err != nil {
		return "", "", fmt.Errorf("write down migration: %w", err)
	}

	return upPath, downPath, nil
}

func sanitizeDiffName(name string) string {
	value := strings.ToLower(strings.TrimSpace(name))
	value = strings.ReplaceAll(value, " ", "_")
	re := regexp.MustCompile(`[^a-z0-9_\-]+`)
	value = re.ReplaceAllString(value, "")
	if value == "" {
		return "schema_diff"
	}
	return value
}

func openSecondaryDB(ctx context.Context, url string) (*sql.DB, error) {
	secondary, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if err := secondary.PingContext(ctx); err != nil {
		return nil, err
	}
	return secondary, nil
}

func closeRuntime(rt *Runtime) {
	if rt == nil {
		return
	}
	if rt.DB != nil {
		_ = rt.DB.Close()
	}
	if rt.Shutdown != nil {
		_ = rt.Shutdown(context.Background())
	}
}
