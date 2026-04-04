package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/services"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dbURL := envOrDefault("DATABASE_URL", "postgres://loco:loco_secret@localhost:5432/loco_twin?sslmode=disable")
	locoID := envOrDefault("LOCO_ID", "loco-001")
	metricsConfig := envOrDefault("METRICS_CONFIG", "metrics.yaml")
	intervalStr := envOrDefault("TICK_INTERVAL", "1s")

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		slog.Error("invalid TICK_INTERVAL", "value", intervalStr, "err", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		slog.Error("connect to postgres", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		slog.Error("ping postgres", "err", err)
		os.Exit(1)
	}
	slog.Info("connected to postgres", "url", dbURL)

	migrationsPath := envOrDefault("MIGRATIONS_PATH", "db/migrations")
	if err = runMigrations(dbURL, migrationsPath); err != nil {
		slog.Error("run migrations", "err", err)
		os.Exit(1)
	}

	q := sqlc.New(pool)

	if err = seedMetrics(ctx, q, metricsConfig); err != nil {
		slog.Error("seed metrics", "err", err)
		os.Exit(1)
	}

	svc := services.NewSimulationService(q, locoID, interval)
	slog.Info("starting simulation", "loco_id", locoID, "interval", interval)
	if err = svc.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("simulation stopped with error", "err", err)
		os.Exit(1)
	}
	slog.Info("simulation stopped")
}

// runMigrations applies all pending UP migrations.
func runMigrations(dbURL, migrationsPath string) error {
	// golang-migrate pgx/v5 driver expects the pgx5:// scheme
	migrateURL := "pgx5://" + dbURL[len("postgres://"):]
	if len(dbURL) >= 14 && dbURL[:14] == "postgresql://" {
		migrateURL = "pgx5://" + dbURL[len("postgresql://"):]
	}

	m, err := migrate.New("file://"+migrationsPath, migrateURL)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	slog.Info("migrations applied")
	return nil
}

// seedMetrics reads metrics.yaml and upserts all metric definitions into the DB.
func seedMetrics(ctx context.Context, q *sqlc.Queries, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read metrics config %q: %w", path, err)
	}

	var cfg metricsConfig
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse metrics config: %w", err)
	}

	for _, m := range cfg.Metrics {
		displayOpts, _ := json.Marshal(m.DisplayOpts)

		params := sqlc.UpsertMetricDefinitionParams{
			Name:         m.Name,
			Display:      m.Display,
			Description:  m.Description,
			Unit:         m.Unit,
			PhysicalMin:  ptr32(m.Physical.Min),
			PhysicalMax:  ptr32(m.Physical.Max),
			NormalMin:    ptr32(m.Normal.Min),
			NormalMax:    ptr32(m.Normal.Max),
			WarnAbove:    m.Thresholds.WarnAbove,
			WarnBelow:    m.Thresholds.WarnBelow,
			CritAbove:    m.Thresholds.CriticalAbove,
			CritBelow:    m.Thresholds.CriticalBelow,
			HealthWeight: m.HealthWeight,
			EmaAlpha:     m.EmaAlpha,
			DisplayOpts:  displayOpts,
		}

		if _, err = q.UpsertMetricDefinition(ctx, params); err != nil {
			return fmt.Errorf("upsert metric %q: %w", m.Name, err)
		}
		slog.Info("seeded metric", "name", m.Name)
	}
	return nil
}

// --------------------------------------------------------------------------
// YAML structs
// --------------------------------------------------------------------------

type metricsConfig struct {
	Metrics []yamlMetric `yaml:"metrics"`
}

type yamlMetric struct {
	Name         string         `yaml:"name"`
	Display      string         `yaml:"display"`
	Description  string         `yaml:"description"`
	Unit         string         `yaml:"unit"`
	Group        string         `yaml:"group"`
	Physical     yamlMinMax     `yaml:"physical"`
	Normal       yamlMinMax     `yaml:"normal"`
	Thresholds   yamlThresholds `yaml:"thresholds"`
	HealthWeight float32        `yaml:"health_weight"`
	EmaAlpha     float32        `yaml:"ema_alpha"`
	DisplayOpts  map[string]any `yaml:"display_opts"`
}

type yamlMinMax struct {
	Min float32 `yaml:"min"`
	Max float32 `yaml:"max"`
}

type yamlThresholds struct {
	WarnAbove     *float32 `yaml:"warn_above"`
	WarnBelow     *float32 `yaml:"warn_below"`
	CriticalAbove *float32 `yaml:"critical_above"`
	CriticalBelow *float32 `yaml:"critical_below"`
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func ptr32(v float32) *float32 { return &v }

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
