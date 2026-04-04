// Package main is the entry point for the Locomotive Digital Twin API server.
//
//	@title			Locomotive Digital Twin API
//	@version		1.0
//	@description	Real-time telemetry ingestion, health scoring, and diagnostic REST API.
//	@contact.name	HackNU Team
//	@host			localhost:8080
//	@BasePath		/
//
//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/Nap20192/hacknu/docs" // swagger generated docs

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/Nap20192/hacknu/internal/alerts/api"
	"github.com/Nap20192/hacknu/internal/config"
	"github.com/Nap20192/hacknu/internal/hub"
	"github.com/Nap20192/hacknu/internal/spec"
	"github.com/Nap20192/hacknu/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()

	// ── Logger ────────────────────────────────────────────────────────────────
	log, err := logger.InitLogger(cfg.LogLevel, cfg.LogPretty, cfg.LogDir)
	if err != nil {
		slog.Error("failed to init logger", "err", err)
		os.Exit(1)
	}
	slog.SetDefault(log)

	// ── Database ──────────────────────────────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		slog.Error("database ping failed", "err", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	queries := sqlc.New(pool)

	// ── Rule Registry: load from DB, refresh every 60 s ──────────────────────
	registry := spec.NewRuleRegistry()
	if err := refreshRegistry(context.Background(), queries, registry); err != nil {
		slog.Warn("initial rule registry load failed", "err", err)
	}

	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := refreshRegistry(context.Background(), queries, registry); err != nil {
				slog.Warn("rule registry refresh failed", "err", err)
			}
		}
	}()

	// ── WebSocket Hub ─────────────────────────────────────────────────────────
	wsHub := hub.NewManager()
	wsHub.StartWrite(context.Background())

	// ── Telemetry Processor ───────────────────────────────────────────────────
	// Reads raw frames from the hub, runs them through the Engine,
	// and broadcasts HealthSnapshot JSON back to all clients.

	// ── Fiber App ─────────────────────────────────────────────────────────────
	app := api.NewApp(queries)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("server error", "err", err)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	wsHub.Shutdown()
	slog.Info("server stopped")
}

func refreshRegistry(ctx context.Context, q *sqlc.Queries, r *spec.RuleRegistry) error {
	defs, err := q.ListMetricDefinitions(ctx)
	if err != nil {
		return err
	}
	r.Refresh(defs)
	slog.Info("rule registry refreshed", "count", len(defs))
	return nil
}
