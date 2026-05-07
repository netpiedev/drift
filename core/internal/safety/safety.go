package safety

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/netpiedev/drift/core/internal/config"
)

func IsProductionEnv(env string) bool {
	env = strings.ToLower(strings.TrimSpace(env))
	return env == "prod" || env == "production"
}

func DBFingerprint(ctx context.Context, db *sql.DB) (string, error) {
	const q = `SELECT current_database(), current_user, inet_server_addr()::text, inet_server_port();`
	var database string
	var user string
	var host string
	var port int
	if err := db.QueryRowContext(ctx, q).Scan(&database, &user, &host, &port); err != nil {
		return "", fmt.Errorf("query db fingerprint: %w", err)
	}
	raw := fmt.Sprintf("%s|%s|%s|%d", database, user, host, port)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:]), nil
}

func EnforceSafety(ctx context.Context, db *sql.DB, cfg config.Config, command string) error {
	isProd := IsProductionEnv(cfg.Environment)
	if !isProd {
		return nil
	}

	if cfg.Safety.ReadonlyProd {
		if strings.HasPrefix(command, "migrate up") || strings.HasPrefix(command, "migrate down") || strings.HasPrefix(command, "migrate rollback") {
			return fmt.Errorf("production is readonly (safety.readonly_prod=true)")
		}
	}

	if cfg.Safety.ProdFingerprint != "" {
		fingerprint, err := DBFingerprint(ctx, db)
		if err != nil {
			return err
		}
		if fingerprint != cfg.Safety.ProdFingerprint {
			return fmt.Errorf("environment fingerprint mismatch: expected %s got %s", cfg.Safety.ProdFingerprint, fingerprint)
		}
	}

	if cfg.Safety.RequireConfirmProd && strings.Contains(command, "down") {
		if err := confirmProductionDanger(); err != nil {
			return err
		}
	}

	return nil
}

func confirmProductionDanger() error {
	fmt.Fprint(os.Stderr, "Production rollback requested. Type 'ROLLBACK' to continue: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read confirmation: %w", err)
	}
	if strings.TrimSpace(line) != "ROLLBACK" {
		return fmt.Errorf("production rollback aborted")
	}
	return nil
}
