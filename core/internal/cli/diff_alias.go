package cli

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

func newDiffAliasCommand(configPath *string) *cobra.Command {
	opts := diffOptions{}
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Alias for migrate diff",
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
