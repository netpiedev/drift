package cli

import (
	"fmt"

	"github.com/netpiedev/drift/core/internal/seed"
	"github.com/spf13/cobra"
)

func newSeedCommand(configPath *string) *cobra.Command {
	cmd := &cobra.Command{Use: "seed", Short: "Run seed data jobs"}
	cmd.AddCommand(newSeedEnvCommand(configPath, "dev"))
	cmd.AddCommand(newSeedEnvCommand(configPath, "staging"))
	return cmd
}

func newSeedEnvCommand(configPath *string, env string) *cobra.Command {
	return &cobra.Command{
		Use:   env,
		Short: fmt.Sprintf("Run %s seeds", env),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := BuildRuntime(cmd.Context(), *configPath, true)
			if err != nil {
				return err
			}
			defer closeRuntime(rt)

			seeder := seed.New(rt.DB, rt.Config.Seeds.Dir)
			if err := seeder.Run(rt.Ctx, env); err != nil {
				return err
			}
			cmd.Printf("%s seeds completed\n", env)
			return nil
		},
	}
}
