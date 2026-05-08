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
	var environment string
	var databaseURL string
	var migrationsDir string
	var sequenceType string
	var seedsDir string
	var snapshotsDir string
	var requireConfirmProd bool
	var readonlyProd bool
	var scaffoldSeeds bool
	var scaffoldSnapshots bool
	var force bool
	var nonInteractive bool

	command := &cobra.Command{
		Use:   "init",
		Short: "Initialize drift project structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			selectedEnvironment := strings.TrimSpace(environment)
			if selectedEnvironment == "" {
				selectedEnvironment = "dev"
			}

			selectedDatabaseURL := strings.TrimSpace(databaseURL)
			if selectedDatabaseURL == "" {
				selectedDatabaseURL = "${DATABASE_URL}"
			}

			selectedMigrationsDir := strings.TrimSpace(migrationsDir)
			if selectedMigrationsDir == "" {
				selectedMigrationsDir = "./migrations"
			}

			selectedSeedsDir := strings.TrimSpace(seedsDir)
			if selectedSeedsDir == "" {
				selectedSeedsDir = "./seeds"
			}

			selectedSnapshotsDir := strings.TrimSpace(snapshotsDir)
			if selectedSnapshotsDir == "" {
				selectedSnapshotsDir = "./snapshots"
			}

			selectedSequenceType, err := normalizeSequenceType(sequenceType)
			if err != nil {
				return err
			}

			interactive := !nonInteractive && isInteractiveStdin()
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				promptedEnvironment, promptErr := promptString(reader, "Environment name", selectedEnvironment)
				if promptErr != nil {
					return promptErr
				}
				selectedEnvironment = promptedEnvironment

				promptedDatabaseURL, promptErr := promptString(reader, "Database URL", selectedDatabaseURL)
				if promptErr != nil {
					return promptErr
				}
				selectedDatabaseURL = promptedDatabaseURL

				promptedMigrationsDir, promptErr := promptString(reader, "Where should Drift store migrations?", selectedMigrationsDir)
				if promptErr != nil {
					return promptErr
				}
				selectedMigrationsDir = promptedMigrationsDir

				promptedSequenceType, promptErr := promptString(reader, "Select migration sequence type [timestamp|date|serial|none]", selectedSequenceType)
				if promptErr != nil {
					return promptErr
				}
				selectedSequenceType, promptErr = normalizeSequenceType(promptedSequenceType)
				if promptErr != nil {
					return promptErr
				}

				promptedSeedsDir, promptErr := promptString(reader, "Where should Drift store seeds?", selectedSeedsDir)
				if promptErr != nil {
					return promptErr
				}
				selectedSeedsDir = promptedSeedsDir

				promptedScaffoldSeeds, promptErr := promptBool(reader, "Create seed folders (dev/staging) now?", scaffoldSeeds)
				if promptErr != nil {
					return promptErr
				}
				scaffoldSeeds = promptedScaffoldSeeds

				promptedSnapshotsDir, promptErr := promptString(reader, "Where should Drift store snapshots?", selectedSnapshotsDir)
				if promptErr != nil {
					return promptErr
				}
				selectedSnapshotsDir = promptedSnapshotsDir

				promptedScaffoldSnapshots, promptErr := promptBool(reader, "Create snapshots folder now?", scaffoldSnapshots)
				if promptErr != nil {
					return promptErr
				}
				scaffoldSnapshots = promptedScaffoldSnapshots

				promptedRequireConfirmProd, promptErr := promptBool(reader, "Require confirmation for production migrations?", requireConfirmProd)
				if promptErr != nil {
					return promptErr
				}
				requireConfirmProd = promptedRequireConfirmProd

				promptedReadonlyProd, promptErr := promptBool(reader, "Enable readonly production mode?", readonlyProd)
				if promptErr != nil {
					return promptErr
				}
				readonlyProd = promptedReadonlyProd
			}

			dirs := []string{selectedMigrationsDir}
			if scaffoldSeeds {
				dirs = append(dirs, filepath.Join(selectedSeedsDir, "dev"), filepath.Join(selectedSeedsDir, "staging"))
			}
			if scaffoldSnapshots {
				dirs = append(dirs, selectedSnapshotsDir)
			}

			for _, dir := range dirs {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					return fmt.Errorf("create dir %s: %w", dir, err)
				}
			}

			configExists := false
			if info, statErr := os.Stat("drift.yaml"); statErr == nil && !info.IsDir() {
				configExists = true
			} else if statErr != nil && !os.IsNotExist(statErr) {
				return fmt.Errorf("read drift.yaml: %w", statErr)
			}

			if configExists && !force {
				if interactive {
					reader := bufio.NewReader(os.Stdin)
					overwrite, promptErr := promptBool(reader, "drift.yaml already exists. Overwrite it?", false)
					if promptErr != nil {
						return promptErr
					}
					if !overwrite {
						cmd.Println("drift.yaml already exists, skipping write")
						cmd.Println("drift initialized")
						return nil
					}
				} else {
					cmd.Println("drift.yaml already exists, skipping write (use --force to overwrite)")
					cmd.Println("drift initialized")
					return nil
				}
			}

			content := fmt.Sprintf(`environment: %s

database:
  url: %s

migrations:
  dir: %s
  sequence_type: %s

safety:
  require_confirm_prod: %t
  readonly_prod: %t
  prod_fingerprint: ""

seeds:
  dir: %s

snapshots:
  dir: %s

runners:
  node: node
  bun: bun
  python: python3
  go: go

observability:
  enable_otel: false
  otel_service_name: drift
  prometheus_listen: ""
`, selectedEnvironment, selectedDatabaseURL, selectedMigrationsDir, selectedSequenceType, requireConfirmProd, readonlyProd, selectedSeedsDir, selectedSnapshotsDir)
			if err := os.WriteFile("drift.yaml", []byte(content), 0o644); err != nil {
				return fmt.Errorf("write drift.yaml: %w", err)
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

	command.Flags().StringVar(&environment, "environment", "dev", "default environment name to write into drift.yaml")
	command.Flags().StringVar(&databaseURL, "database-url", "${DATABASE_URL}", "database URL value to write into drift.yaml")
	command.Flags().StringVar(&migrationsDir, "migrations-dir", "./migrations", "migrations directory path to use in drift.yaml")
	command.Flags().StringVar(&sequenceType, "sequence-type", SequenceTypeTimestamp, "migration sequence type: timestamp|date|serial|none")
	command.Flags().StringVar(&seedsDir, "seeds-dir", "./seeds", "seed root directory path to use in drift.yaml")
	command.Flags().StringVar(&snapshotsDir, "snapshots-dir", "./snapshots", "snapshots directory path to use in drift.yaml")
	command.Flags().BoolVar(&requireConfirmProd, "require-confirm-prod", true, "require confirmation when running migrations in production")
	command.Flags().BoolVar(&readonlyProd, "readonly-prod", false, "enable readonly production mode in drift.yaml")
	command.Flags().BoolVar(&scaffoldSeeds, "scaffold-seeds", false, "create seed folders during init")
	command.Flags().BoolVar(&scaffoldSnapshots, "scaffold-snapshots", false, "create snapshots folder during init")
	command.Flags().BoolVar(&force, "force", false, "overwrite existing drift.yaml")
	command.Flags().BoolVar(&nonInteractive, "yes", false, "run init without interactive prompts")
	return command
}

func promptString(reader *bufio.Reader, label string, defaultValue string) (string, error) {
	fmt.Fprintf(os.Stdout, "%s [%s]: ", label, defaultValue)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read prompt input for %s: %w", strings.ToLower(label), err)
	}
	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func promptBool(reader *bufio.Reader, label string, defaultValue bool) (bool, error) {
	defaultSuffix := "y/N"
	if defaultValue {
		defaultSuffix = "Y/n"
	}
	fmt.Fprintf(os.Stdout, "%s [%s]: ", label, defaultSuffix)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read prompt input for %s: %w", strings.ToLower(label), err)
	}
	value := strings.ToLower(strings.TrimSpace(line))
	if value == "" {
		return defaultValue, nil
	}
	switch value {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid response %q for %s, use yes/no", value, strings.ToLower(label))
	}
}

func isInteractiveStdin() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
