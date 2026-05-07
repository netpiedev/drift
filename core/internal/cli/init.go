package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newInitCommand(configPath *string) *cobra.Command {
	_ = configPath
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize drift project structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			dirs := []string{"migrations", "seeds/dev", "seeds/staging", "snapshots"}
			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create dir %s: %w", dir, err)
				}
			}

			if _, err := os.Stat("drift.yaml"); os.IsNotExist(err) {
				content := `environment: dev

database:
  url: ${DATABASE_URL}

migrations:
  dir: ./migrations

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
`
				if err := os.WriteFile("drift.yaml", []byte(content), 0o644); err != nil {
					return fmt.Errorf("write drift.yaml: %w", err)
				}
			}

			readmePath := filepath.Join("migrations", "README.md")
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
}
