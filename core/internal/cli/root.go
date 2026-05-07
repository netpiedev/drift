package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	var configPath string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "drift",
		Short: "Drift is a production-grade migration CLI",
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "", "path to drift config file")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview SQL/actions without executing")

	cmd.AddCommand(newInitCommand(&configPath))
	cmd.AddCommand(newMakeCommand(&configPath))
	cmd.AddCommand(newMigrateCommand(&configPath, &dryRun))
	cmd.AddCommand(newSeedCommand(&configPath))
	cmd.AddCommand(newSnapshotCommand(&configPath))
	cmd.AddCommand(newGraphCommand(&configPath))
	cmd.AddCommand(newDiffAliasCommand(&configPath))

	return cmd
}
