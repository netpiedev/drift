package cli

import (
	"github.com/netpiedev/drift/core/internal/migration/graph"
	"github.com/netpiedev/drift/core/internal/migration/parser"
	"github.com/spf13/cobra"
)

func newGraphCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "graph", Short: "Dependency graph outputs"}
	cmd.AddCommand(newGraphTablesCommand(configPath))
	cmd.AddCommand(newGraphMigrationsCommand(configPath))
	return cmd
}

func newGraphTablesCommand(configPath *string) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "tables",
		Short: "Show table dependency graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := BuildRuntime(cmd.Context(), *configPath, true)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			g, err := graph.BuildTableDependencyGraph(rt.Ctx, rt.DB)
			if err != nil {
				return err
			}
			if jsonOut {
				j, err := graph.ToJSON(g)
				if err != nil {
					return err
				}
				cmd.Println(j)
				return nil
			}
			cmd.Println(graph.ToTerminal(g))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}

func newGraphMigrationsCommand(configPath *string) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "migrations",
		Short: "Show migration dependency graph",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := BuildRuntime(cmd.Context(), *configPath, false)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			files, err := parser.ParseDir(rt.Config.Migrations.Dir)
			if err != nil {
				return err
			}
			g := graph.BuildMigrationGraph(files)
			if jsonOut {
				j, err := graph.ToJSON(g)
				if err != nil {
					return err
				}
				cmd.Println(j)
				return nil
			}
			cmd.Println(graph.ToTerminal(g))
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output as JSON")
	return cmd
}
