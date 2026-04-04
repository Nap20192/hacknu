package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nap20192/hacknu/internal/services"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	locoIDStr := envOrDefault("LOCO_ID", "")
	metricsConfig := envOrDefault("METRICS_CONFIG", "metrics.yaml")
	intervalStr := envOrDefault("TICK_INTERVAL", "1s")
	wsURL := envOrDefault("WS_URL", "ws://localhost:8081/ws/telemetry")

	var locoID uuid.UUID
	if locoIDStr == "" {
		locoID = uuid.New()
		slog.Info("LOCO_ID not set, generated new UUID", "loco_id", locoID)
	} else {
		var err error
		locoID, err = uuid.Parse(locoIDStr)
		if err != nil {
			slog.Error("invalid LOCO_ID: must be a UUID", "value", locoIDStr)
			os.Exit(1)
		}
	}

	interval, err := time.ParseDuration(intervalStr)
	if err != nil {
		slog.Error("invalid TICK_INTERVAL", "value", intervalStr, "err", err)
		os.Exit(1)
	}

	defs, err := loadMetricDefs(metricsConfig)
	if err != nil {
		slog.Error("load metrics config", "err", err)
		os.Exit(1)
	}
	slog.Info("loaded metric definitions", "count", len(defs))

	svc := services.NewSimulationService(wsURL, locoID, interval, defs)
	slog.Info("starting simulation", "loco_id", locoID, "interval", interval, "ws_url", wsURL)
	if err = svc.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("simulation stopped with error", "err", err)
		os.Exit(1)
	}
	slog.Info("simulation stopped")
}

// loadMetricDefs reads metrics.yaml and converts it to simulation MetricDef slice.
func loadMetricDefs(path string) ([]services.MetricDef, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	var cfg metricsConfig
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %q: %w", path, err)
	}
	defs := make([]services.MetricDef, 0, len(cfg.Metrics))
	for _, m := range cfg.Metrics {
		defs = append(defs, services.MetricDef{
			Name:        m.Name,
			PhysicalMin: m.Physical.Min,
			PhysicalMax: m.Physical.Max,
			NormalMin:   m.Normal.Min,
			NormalMax:   m.Normal.Max,
			WarnAbove:   m.Thresholds.WarnAbove,
			WarnBelow:   m.Thresholds.WarnBelow,
			CritAbove:   m.Thresholds.CriticalAbove,
			CritBelow:   m.Thresholds.CriticalBelow,
		})
	}
	return defs, nil
}

// --------------------------------------------------------------------------
// YAML structs
// --------------------------------------------------------------------------

type metricsConfig struct {
	Metrics []yamlMetric `yaml:"metrics"`
}

type yamlMetric struct {
	Name       string         `yaml:"name"`
	Physical   yamlMinMax     `yaml:"physical"`
	Normal     yamlMinMax     `yaml:"normal"`
	Thresholds yamlThresholds `yaml:"thresholds"`
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

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
