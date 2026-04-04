package services

import (
	"context"
	"log/slog"
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
)

type healthService struct {
	queries sqlc.Querier
}

func NewHealthService(queries sqlc.Querier) *healthService {
	return &healthService{queries: queries}
}

func (s *healthService) StartHealthCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.SnapState(ctx); err != nil {
				slog.Warn("health check failed", "err", err)
			}
			slog.Info("health check completed")
		}
	}
}

func (s *healthService) SnapState(ctx context.Context) error {
	return nil
}
