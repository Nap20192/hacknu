package services

import (
	"context"
	"time"

	"github.com/Nap20192/hacknu/gen/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Background struct {
	queries sqlc.Queries
}

func NewEmaAlphaService(queries sqlc.Queries) *Background {
	return &Background{queries: queries}
}

func (s *Background) StartEmaAlphaRecalculation(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.queries.RecalculateEmaAlpha(ctx)
		}
	}
}
func (s *Background) PurgeOldTelemetry(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if _, err := s.queries.PurgeTelemetryOlderThan(ctx, pgtype.Interval{Microseconds: int64(interval / time.Microsecond),Valid: true,}); err != nil {
				return err
			}
		}
	}
}
