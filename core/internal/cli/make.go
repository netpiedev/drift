package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/netpiedev/drift/core/internal/config"
	"github.com/spf13/cobra"
)

func newMakeCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "make <name>",
		Short: "Create new migration pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(*configPath)
			if err != nil {
				return err
			}
			name := sanitizeName(args[0])
			version := time.Now().UTC().Format("200601021504")
			base := fmt.Sprintf("%s_%s", version, name)

			if err := os.MkdirAll(cfg.Migrations.Dir, 0o755); err != nil {
				return fmt.Errorf("mkdir migrations dir: %w", err)
			}

			upPath := filepath.Join(cfg.Migrations.Dir, base+".up.sql")
			downPath := filepath.Join(cfg.Migrations.Dir, base+".down.sql")
			if err := os.WriteFile(upPath, []byte("-- Write forward migration SQL here\n"), 0o644); err != nil {
				return err
			}
			if err := os.WriteFile(downPath, []byte("-- Write rollback migration SQL here\n"), 0o644); err != nil {
				return err
			}
			cmd.Printf("created %s and %s\n", upPath, downPath)
			return nil
		},
	}
}

func sanitizeName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "_")
	re := regexp.MustCompile(`[^a-z0-9_\-]+`)
	name = re.ReplaceAllString(name, "")
	if name == "" {
		return "migration"
	}
	return name
}
