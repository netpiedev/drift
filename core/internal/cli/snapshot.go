package cli

import (
	"fmt"

	"github.com/netpiedev/drift/core/internal/snapshot"
	"github.com/spf13/cobra"
)

func newSnapshotCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "snapshot", Short: "Create/restore DB snapshots"}
	cmd.AddCommand(newSnapshotCreateCommand(configPath))
	cmd.AddCommand(newSnapshotRestoreCommand(configPath))
	return cmd
}

func newSnapshotCreateCommand(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create PostgreSQL snapshot using pg_dump",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := BuildRuntime(cmd.Context(), *configPath, false)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			mgr := snapshot.New(rt.Config.Snapshots.Dir)
			path, err := mgr.Create(cmd.Context(), rt.Config.Database.URL)
			if err != nil {
				return err
			}
			cmd.Printf("snapshot created: %s\n", path)
			return nil
		},
	}
}

func newSnapshotRestoreCommand(configPath *string) *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore PostgreSQL snapshot using pg_restore",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			rt, err := BuildRuntime(cmd.Context(), *configPath, false)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			mgr := snapshot.New(rt.Config.Snapshots.Dir)
			if err := mgr.Restore(cmd.Context(), rt.Config.Database.URL, file); err != nil {
				return err
			}
			cmd.Printf("snapshot restored from %s\n", file)
			return nil
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "snapshot file path")
	return cmd
}
