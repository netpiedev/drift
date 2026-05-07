package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newInitCommand(configPath *string) *cobra.Command {
	_ = configPath
	var migrationsDir string
	var nonInteractive bool

	command := &cobra.Command{
		Use:   "init",
		Short: "Initialize drift project structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedMigrationsDir := strings.TrimSpace(migrationsDir)
			if selectedMigrationsDir == "" {
				selectedMigrationsDir = "./migrations"
			}

			if !nonInteractive && isInteractiveStdin() && migrationsDir == "" {
				prompted, err := promptMigrationsDir(selectedMigrationsDir)
				if err != nil {
					return err
				}
				selectedMigrationsDir = prompted
			}

			dirs := []string{selectedMigrationsDir, "seeds/dev", "seeds/staging", "snapshots"}
			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create dir %s: %w", dir, err)
				}
			}

			if _, err := os.Stat("drift.yaml"); os.IsNotExist(err) {
				content := fmt.Sprintf(`environment: dev

database:
  url: ${DATABASE_URL}

migrations:
  dir: %s

safety:
  require_confirm_prod: true
  readonly_prod: false
  prod_fingerprint: ""

seeds:
  dir: ./seeds

snapshots:
  dir: ./snapshots

runners:
  node: node
  bun: bun
  python: python3
  go: go

observability:
  enable_otel: false
  otel_service_name: drift
  prometheus_listen: ""
`, selectedMigrationsDir)
				if err := os.WriteFile("drift.yaml", []byte(content), 0o644); err != nil {
					return fmt.Errorf("write drift.yaml: %w", err)
				}
			}

			readmePath := filepath.Join(selectedMigrationsDir, "README.md")
			if _, err := os.Stat(readmePath); os.IsNotExist(err) {
				msg := "# Migrations\n\nFiles use <version>_<name>.<up|down>.<sql|ts|js|py|go>.\n"
				if err := os.WriteFile(readmePath, []byte(msg), 0o644); err != nil {
					return fmt.Errorf("write migration readme: %w", err)
				}
			}

			cmd.Println("drift initialized")
			return nil
		},
	}

	command.Flags().StringVar(&migrationsDir, "migrations-dir", "", "migrations directory path to use in drift.yaml")
	command.Flags().BoolVar(&nonInteractive, "yes", false, "run init without interactive prompts")
	return command
}

func promptMigrationsDir(defaultPath string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Fprintf(os.Stdout, "Where should Drift store migrations? [%s]: ", defaultPath)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read migrations directory input: %w", err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return defaultPath, nil
	}
	return value, nil
}

func isInteractiveStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
