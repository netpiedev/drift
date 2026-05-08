package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Environment   string `mapstructure:"environment"`
	Database      DatabaseConfig
	Migrations    MigrationConfig
	Safety        SafetyConfig
	Seeds         SeedConfig
	Snapshots     SnapshotConfig
	Runners       RunnerConfig
	Observability ObservabilityConfig
}

type DatabaseConfig struct {
	URL string `mapstructure:"url"`
}

type MigrationConfig struct {
	Dir          string `mapstructure:"dir"`
	SequenceType string `mapstructure:"sequence_type"`
}

type SafetyConfig struct {
	RequireConfirmProd bool   `mapstructure:"require_confirm_prod"`
	ReadonlyProd       bool   `mapstructure:"readonly_prod"`
	ProdFingerprint    string `mapstructure:"prod_fingerprint"`
}

type SeedConfig struct {
	Dir string `mapstructure:"dir"`
}

type SnapshotConfig struct {
	Dir string `mapstructure:"dir"`
}

type RunnerConfig struct {
	Node   string `mapstructure:"node"`
	Bun    string `mapstructure:"bun"`
	Python string `mapstructure:"python"`
	Go     string `mapstructure:"go"`
}

type ObservabilityConfig struct {
	EnableOTel       bool   `mapstructure:"enable_otel"`
	OTelServiceName  string `mapstructure:"otel_service_name"`
	PrometheusListen string `mapstructure:"prometheus_listen"`
}

func Default() Config {
	return Config{
		Environment: "dev",
		Database: DatabaseConfig{
			URL: "",
		},
		Migrations: MigrationConfig{
			Dir:          "./migrations",
			SequenceType: "timestamp",
		},
		Safety: SafetyConfig{
			RequireConfirmProd: true,
			ReadonlyProd:       false,
			ProdFingerprint:    "",
		},
		Seeds: SeedConfig{
			Dir: "./seeds",
		},
		Snapshots: SnapshotConfig{
			Dir: "./snapshots",
		},
		Runners: RunnerConfig{
			Node:   "node",
			Bun:    "bun",
			Python: "python3",
			Go:     "go",
		},
		Observability: ObservabilityConfig{
			EnableOTel:       false,
			OTelServiceName:  "drift",
			PrometheusListen: "",
		},
	}
}

func Load(configPath string) (Config, error) {
	cfg := Default()
	v := viper.New()
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		if defaultConfig := findDefaultConfigFile(); defaultConfig != "" {
			v.SetConfigFile(defaultConfig)
		} else {
			v.SetConfigName("drift")
			v.SetConfigType("yaml")
			v.AddConfigPath(".")
		}
	}

	v.SetEnvPrefix("DRIFT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.SetDefault("environment", cfg.Environment)
	v.SetDefault("database.url", cfg.Database.URL)
	v.SetDefault("migrations.dir", cfg.Migrations.Dir)
	v.SetDefault("migrations.sequence_type", cfg.Migrations.SequenceType)
	v.SetDefault("safety.require_confirm_prod", cfg.Safety.RequireConfirmProd)
	v.SetDefault("safety.readonly_prod", cfg.Safety.ReadonlyProd)
	v.SetDefault("safety.prod_fingerprint", cfg.Safety.ProdFingerprint)
	v.SetDefault("seeds.dir", cfg.Seeds.Dir)
	v.SetDefault("snapshots.dir", cfg.Snapshots.Dir)
	v.SetDefault("runners.node", cfg.Runners.Node)
	v.SetDefault("runners.bun", cfg.Runners.Bun)
	v.SetDefault("runners.python", cfg.Runners.Python)
	v.SetDefault("runners.go", cfg.Runners.Go)
	v.SetDefault("observability.enable_otel", cfg.Observability.EnableOTel)
	v.SetDefault("observability.otel_service_name", cfg.Observability.OTelServiceName)
	v.SetDefault("observability.prometheus_listen", cfg.Observability.PrometheusListen)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) || configPath != "" {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.URL == "" {
		if envURL := v.GetString("database.url"); envURL != "" {
			cfg.Database.URL = envURL
		}
	}

	return cfg, nil
}

func findDefaultConfigFile() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for _, name := range []string{"drift.yaml", "drift.yml"} {
		candidate := filepath.Join(cwd, name)
		info, statErr := os.Stat(candidate)
		if statErr != nil {
			continue
		}
		if !info.IsDir() {
			return candidate
		}
	}

	return ""
}
