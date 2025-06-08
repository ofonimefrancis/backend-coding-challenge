package rating

import (
	"context"
	"log/slog"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/shared"
	"time"
)

const (
	DefaultGlobalAverage = 3.0
)

func NewRatingServiceWithStartup(
	ratingRepo rating.Repository,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	logger *slog.Logger,
) (Service, error) {
	service := &ratingService{
		ratingRepo:     ratingRepo,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		logger:         logger,
		bayesianConfig: DefaultBayesianConfig(),
		globalAverage:  DefaultGlobalAverage,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := service.UpdateGlobalAverage(ctx); err != nil {
		logger.Warn("Failed to initialize global average on startup, using default", "error", err)
	}

	return service, nil
}

func (s *ratingService) StartGlobalAverageUpdater(ctx context.Context, updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	s.logger.Info("Starting global average updater", "interval", updateInterval)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping global average updater")
			return
		case <-ticker.C:
			if err := s.UpdateGlobalAverage(ctx); err != nil {
				s.logger.Error("Periodic global average update failed", "error", err)
			}
		}
	}
}
