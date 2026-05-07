package cli

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/netpiedev/drift/core/internal/config"
	"github.com/netpiedev/drift/core/internal/db"
	"github.com/netpiedev/drift/core/internal/logging"
	"github.com/netpiedev/drift/core/internal/telemetry"
	"github.com/rs/zerolog"
)

type Runtime struct {
	Ctx      context.Context
	Config   config.Config
	Logger   zerolog.Logger
	DB       *sql.DB
	Metrics  *telemetry.Metrics
	Shutdown func(context.Context) error
}

func BuildRuntime(ctx context.Context, configPath string, needsDB bool) (*Runtime, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	logger := logging.New()
	metrics := telemetry.NewMetrics()
	metrics.Register()
	if err := telemetry.StartPrometheusServer(cfg.Observability.PrometheusListen); err != nil {
		return nil, fmt.Errorf("start metrics server: %w", err)
	}
	shutdown := telemetry.InitOTel(cfg.Observability.EnableOTel)

	rt := &Runtime{
		Ctx:      ctx,
		Config:   cfg,
		Logger:   logger,
		Metrics:  metrics,
		Shutdown: shutdown,
	}
	if !needsDB {
		return rt, nil
	}
	database, err := db.OpenPostgres(ctx, cfg.Database.URL)
	if err != nil {
		return nil, err
	}
	rt.DB = database
	return rt, nil
}
