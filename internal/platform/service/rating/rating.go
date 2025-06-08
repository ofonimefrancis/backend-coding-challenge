package rating

import (
	"context"
	"fmt"
	"log/slog"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/shared"
	"thermondo/internal/domain/users"
	"thermondo/internal/pkg/errors"
)

// Bayesian rating configuration
type BayesianConfig struct {
	MinVotes      int64   // Minimum votes to consider reliable (e.g., 10)
	GlobalAverage float64 // Global average rating across all movies
	ConfidenceK   float64 // Confidence parameter - higher = more conservative
}

// Default Bayesian configuration
func DefaultBayesianConfig() BayesianConfig {
	return BayesianConfig{
		MinVotes:      10,   // Require at least 10 votes for full confidence
		GlobalAverage: 3.0,  // Assume global average of 3.0 (middle of 1-5 scale)
		ConfidenceK:   25.0, // Confidence factor (similar to IMDb)
	}
}

// Enhanced movie stats with Bayesian calculation
type EnhancedMovieStats struct {
	*rating.MovieRatingStats
	BayesianAverage float64 `json:"bayesian_average"`
	Confidence      float64 `json:"confidence"`  // How confident we are (0-1)
	Percentile      float64 `json:"percentile"`  // Percentile rank among all movies
	Explanation     string  `json:"explanation"` // Human-readable explanation
}

type Service interface {
	CreateRating(ctx context.Context, req CreateRatingRequest) (*rating.Rating, error)
	GetRatingByID(ctx context.Context, id string) (*rating.Rating, error)
	GetUserRating(ctx context.Context, userID, movieID string) (*rating.Rating, error)
	UpdateRating(ctx context.Context, id string, req UpdateRatingRequest) (*rating.Rating, error)
	DeleteRating(ctx context.Context, id string) error
	GetUserRatings(ctx context.Context, userID string, limit, offset int, sortBy, order string) ([]*rating.Rating, int64, error)
	GetMovieRatings(ctx context.Context, movieID string, limit, offset int, sortBy, order string) ([]*rating.Rating, int64, error)
	GetMovieStats(ctx context.Context, movieID string) (*rating.MovieRatingStats, error)

	// Enhanced methods with Bayesian calculation
	GetEnhancedMovieStats(ctx context.Context, movieID string) (*EnhancedMovieStats, error)
	UpdateGlobalAverage(ctx context.Context) error
	GetBayesianConfig() BayesianConfig
	SetBayesianConfig(config BayesianConfig)
}

type ratingService struct {
	ratingRepo     rating.Repository
	idGenerator    shared.IDGenerator
	timeProvider   shared.TimeProvider
	logger         *slog.Logger
	bayesianConfig BayesianConfig
	globalAverage  float64 // Cached global average
	testMode       bool    // If true, run background updates synchronously (for tests)
}

func NewRatingService(
	ratingRepo rating.Repository,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	logger *slog.Logger,
) Service {
	return &ratingService{
		ratingRepo:     ratingRepo,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		logger:         logger,
		bayesianConfig: DefaultBayesianConfig(),
		globalAverage:  3.0, // Default until first calculation
		testMode:       false,
	}
}

// NewTestRatingService is used for tests to enable synchronous background updates
func NewTestRatingService(
	ratingRepo rating.Repository,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	logger *slog.Logger,
) Service {
	return &ratingService{
		ratingRepo:     ratingRepo,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		logger:         logger,
		bayesianConfig: DefaultBayesianConfig(),
		globalAverage:  3.0,
		testMode:       true,
	}
}

func NewRatingServiceWithConfig(
	ratingRepo rating.Repository,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	logger *slog.Logger,
	config BayesianConfig,
) Service {
	return &ratingService{
		ratingRepo:     ratingRepo,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		logger:         logger,
		bayesianConfig: config,
		globalAverage:  config.GlobalAverage,
		testMode:       false,
	}
}

// Bayesian Average Calculation
// Formula: (C * m + R * v) / (C + v)
// Where:
// - C = confidence parameter (minimum votes needed for reliability)
// - m = global average rating
// - R = average rating for this movie
// - v = number of votes for this movie
func (s *ratingService) calculateBayesianAverage(movieAverage float64, movieVotes int64) float64 {
	C := s.bayesianConfig.ConfidenceK
	m := s.globalAverage
	R := movieAverage
	v := float64(movieVotes)

	if v == 0 {
		return m // Return global average if no votes
	}

	bayesianAvg := (C*m + R*v) / (C + v)
	s.logger.Debug("Calculated Bayesian average",
		"movie_avg", R,
		"movie_votes", v,
		"global_avg", m,
		"confidence_k", C,
		"bayesian_avg", bayesianAvg,
	)

	return bayesianAvg
}

// Calculate confidence score (0-1) based on number of ratings
func (s *ratingService) calculateConfidence(totalRatings int64) float64 {
	if totalRatings >= s.bayesianConfig.MinVotes {
		return 1.0
	}
	confidence := float64(totalRatings) / float64(s.bayesianConfig.MinVotes)
	return confidence
}

// Estimate percentile rank based on Bayesian average
func (s *ratingService) estimatePercentile(bayesianAverage float64) float64 {
	// TODO: Simplified percentile estimation - in production you might calculate this from actual data
	switch {
	case bayesianAverage >= 4.5:
		return 95.0
	case bayesianAverage >= 4.2:
		return 90.0
	case bayesianAverage >= 4.0:
		return 80.0
	case bayesianAverage >= 3.8:
		return 70.0
	case bayesianAverage >= 3.5:
		return 60.0
	case bayesianAverage >= 3.2:
		return 50.0
	case bayesianAverage >= 3.0:
		return 40.0
	case bayesianAverage >= 2.5:
		return 20.0
	case bayesianAverage >= 2.0:
		return 10.0
	default:
		return 5.0
	}
}

// Generate human-readable explanation
func (s *ratingService) generateExplanation(stats *EnhancedMovieStats) string {
	totalRatings := stats.TotalRatings
	confidence := stats.Confidence

	if totalRatings == 0 {
		return "No ratings yet. Score shows global average."
	}

	if totalRatings < s.bayesianConfig.MinVotes {
		return fmt.Sprintf("Rating adjusted for small sample size (%d ratings). Bayesian average considers global trends.", totalRatings)
	}

	if confidence >= 0.95 {
		return fmt.Sprintf("High confidence rating based on %d user ratings.", totalRatings)
	}

	if confidence >= 0.8 {
		return fmt.Sprintf("Reliable rating based on %d user ratings.", totalRatings)
	}

	return fmt.Sprintf("Rating based on %d user ratings with %.0f%% confidence.", totalRatings, confidence*100)
}

func (s *ratingService) CreateRating(ctx context.Context, req CreateRatingRequest) (*rating.Rating, error) {
	existingRating, err := s.ratingRepo.GetByUserAndMovie(ctx, users.UserID(req.UserID), movies.MovieID(req.MovieID))
	if err == nil && existingRating != nil {
		s.logger.Warn("User attempted to rate movie twice",
			"user_id", req.UserID,
			"movie_id", req.MovieID)
		return nil, errors.NewConflictError("User has already rated this movie")
	}

	newRating, err := rating.NewRating(
		users.UserID(req.UserID),
		movies.MovieID(req.MovieID),
		req.Score,
		req.Review,
		s.idGenerator,
		s.timeProvider,
	)
	if err != nil {
		s.logger.Error("Failed to create rating domain object", "error", err)
		return nil, errors.NewBadRequestError(err.Error())
	}

	s.logger.Info("Creating rating",
		"user_id", req.UserID,
		"movie_id", req.MovieID,
		"score", req.Score)

	savedRating, err := s.ratingRepo.Save(ctx, newRating)
	if err != nil {
		if isConflictError(err) {
			s.logger.Error("Conflict when saving rating", "error", err)
			return nil, errors.NewConflictError("User has already rated this movie")
		}
		s.logger.Error("Failed to save rating to repository", "error", err)
		return nil, errors.NewInternalError("Failed to create rating")
	}

	// Update global average in background after new rating
	if s.testMode {
		_ = s.UpdateGlobalAverage(context.Background())
	} else {
		go func() {
			if err := s.UpdateGlobalAverage(context.Background()); err != nil {
				s.logger.Error("Failed to update global average", "error", err)
			}
		}()
	}

	return savedRating, nil
}

func (s *ratingService) GetRatingByID(ctx context.Context, id string) (*rating.Rating, error) {
	ratingObj, err := s.ratingRepo.GetByID(ctx, rating.RatingID(id))
	if err != nil {
		if isNotFoundError(err) {
			s.logger.Debug("Rating not found", "rating_id", id)
			return nil, errors.NewNotFoundError("Rating not found")
		}
		s.logger.Error("Failed to get rating by ID", "error", err, "rating_id", id)
		return nil, errors.NewInternalError("Failed to get rating")
	}

	return ratingObj, nil
}

func (s *ratingService) GetUserRating(ctx context.Context, userID, movieID string) (*rating.Rating, error) {
	ratingObj, err := s.ratingRepo.GetByUserAndMovie(ctx, users.UserID(userID), movies.MovieID(movieID))
	if err != nil {
		if isNotFoundError(err) {
			s.logger.Debug("User rating not found", "user_id", userID, "movie_id", movieID)
			return nil, errors.NewNotFoundError("Rating not found")
		}
		s.logger.Error("Failed to get user rating", "error", err, "user_id", userID, "movie_id", movieID)
		return nil, errors.NewInternalError("Failed to get rating")
	}

	return ratingObj, nil
}

func (s *ratingService) UpdateRating(ctx context.Context, id string, req UpdateRatingRequest) (*rating.Rating, error) {
	existingRating, err := s.ratingRepo.GetByID(ctx, rating.RatingID(id))
	if err != nil {
		if isNotFoundError(err) {
			s.logger.Debug("Rating not found for update", "rating_id", id)
			return nil, errors.NewNotFoundError("Rating not found")
		}
		s.logger.Error("Failed to get rating for update", "error", err, "rating_id", id)
		return nil, errors.NewInternalError("Failed to get rating for update")
	}

	updatedRating := *existingRating

	if req.Score != nil {
		if err := updatedRating.UpdateScore(*req.Score, s.timeProvider); err != nil {
			s.logger.Error("Failed to update score", "error", err, "new_score", *req.Score)
			return nil, errors.NewBadRequestError(err.Error())
		}
		s.logger.Info("Updated rating score", "rating_id", id, "new_score", *req.Score)
	}

	if req.Review != nil {
		if err := updatedRating.UpdateReview(*req.Review, s.timeProvider); err != nil {
			s.logger.Error("Failed to update review", "error", err)
			return nil, errors.NewBadRequestError(err.Error())
		}
		s.logger.Info("Updated rating review", "rating_id", id)
	}

	savedRating, err := s.ratingRepo.Update(ctx, &updatedRating)
	if err != nil {
		s.logger.Error("Failed to save updated rating", "error", err, "rating_id", id)
		return nil, errors.NewInternalError("Failed to update rating")
	}

	// Update global average in background after rating update
	if s.testMode {
		_ = s.UpdateGlobalAverage(context.Background())
	} else {
		go func() {
			if err := s.UpdateGlobalAverage(context.Background()); err != nil {
				s.logger.Error("Failed to update global average after rating update", "error", err)
			}
		}()
	}

	return savedRating, nil
}

func (s *ratingService) DeleteRating(ctx context.Context, id string) error {
	err := s.ratingRepo.Delete(ctx, rating.RatingID(id))
	if err != nil {
		if isNotFoundError(err) {
			s.logger.Debug("Rating not found for deletion", "rating_id", id)
			return errors.NewNotFoundError("Rating not found")
		}
		s.logger.Error("Failed to delete rating", "error", err, "rating_id", id)
		return errors.NewInternalError("Failed to delete rating")
	}

	s.logger.Info("Deleted rating", "rating_id", id)

	// Update global average in background after deletion
	if s.testMode {
		_ = s.UpdateGlobalAverage(context.Background())
	} else {
		go func() {
			if err := s.UpdateGlobalAverage(context.Background()); err != nil {
				s.logger.Error("Failed to update global average after rating deletion", "error", err)
			}
		}()
	}

	return nil
}

func (s *ratingService) GetUserRatings(ctx context.Context, userID string, limit, offset int, sortBy, order string) ([]*rating.Rating, int64, error) {
	searchOptions := []rating.SearchOption{
		rating.WithLimit(limit),
		rating.WithOffset(offset),
		rating.WithSort(sortBy, order),
	}

	ratingsList, err := s.ratingRepo.GetByUser(ctx, users.UserID(userID), searchOptions...)
	if err != nil {
		s.logger.Error("Failed to get user ratings", "error", err, "user_id", userID)
		return nil, 0, errors.NewInternalError("Failed to get user ratings")
	}

	totalCount := int64(len(ratingsList))
	s.logger.Debug("Retrieved user ratings", "user_id", userID, "count", totalCount)

	return ratingsList, totalCount, nil
}

func (s *ratingService) GetMovieRatings(ctx context.Context, movieID string, limit, offset int, sortBy, order string) ([]*rating.Rating, int64, error) {
	searchOptions := []rating.SearchOption{
		rating.WithLimit(limit),
		rating.WithOffset(offset),
		rating.WithSort(sortBy, order),
	}

	ratingsList, err := s.ratingRepo.GetByMovie(ctx, movies.MovieID(movieID), searchOptions...)
	if err != nil {
		s.logger.Error("Failed to get movie ratings", "error", err, "movie_id", movieID)
		return nil, 0, errors.NewInternalError("Failed to get movie ratings")
	}

	totalCount := int64(len(ratingsList))
	s.logger.Debug("Retrieved movie ratings", "movie_id", movieID, "count", totalCount)

	return ratingsList, totalCount, nil
}

func (s *ratingService) GetMovieStats(ctx context.Context, movieID string) (*rating.MovieRatingStats, error) {
	stats, err := s.ratingRepo.GetMovieStats(ctx, movies.MovieID(movieID))
	if err != nil {
		s.logger.Error("Failed to get movie stats", "error", err, "movie_id", movieID)
		return nil, errors.NewInternalError("Failed to get movie stats")
	}

	return stats, nil
}

// Enhanced method with Bayesian calculation
func (s *ratingService) GetEnhancedMovieStats(ctx context.Context, movieID string) (*EnhancedMovieStats, error) {
	// Get basic stats
	stats, err := s.ratingRepo.GetMovieStats(ctx, movies.MovieID(movieID))
	if err != nil {
		s.logger.Error("Failed to get movie stats for Bayesian calculation", "error", err, "movie_id", movieID)
		return nil, errors.NewInternalError("Failed to get movie stats")
	}

	// Calculate Bayesian metrics
	bayesianAvg := s.calculateBayesianAverage(stats.AverageScore, stats.TotalRatings)
	confidence := s.calculateConfidence(stats.TotalRatings)
	percentile := s.estimatePercentile(bayesianAvg)

	enhancedStats := &EnhancedMovieStats{
		MovieRatingStats: stats,
		BayesianAverage:  bayesianAvg,
		Confidence:       confidence,
		Percentile:       percentile,
	}

	// Add explanation
	enhancedStats.Explanation = s.generateExplanation(enhancedStats)

	s.logger.Debug("Calculated enhanced movie stats",
		"movie_id", movieID,
		"simple_avg", stats.AverageScore,
		"bayesian_avg", bayesianAvg,
		"confidence", confidence,
		"total_ratings", stats.TotalRatings)

	return enhancedStats, nil
}

// Update global average rating across all movies
func (s *ratingService) UpdateGlobalAverage(ctx context.Context) error {
	s.logger.Info("Updating global average rating")

	newGlobalAverage, err := s.ratingRepo.GetGlobalAverageRating(ctx)
	if err != nil {
		s.logger.Error("Failed to calculate global average from repository", "error", err)
		return fmt.Errorf("failed to update global average: %w", err)
	}

	oldAverage := s.globalAverage
	s.globalAverage = newGlobalAverage

	s.bayesianConfig.GlobalAverage = newGlobalAverage

	s.logger.Info("Successfully updated global average",
		"old_average", oldAverage,
		"new_average", newGlobalAverage,
		"change", newGlobalAverage-oldAverage)

	return nil
}

// Configuration methods
func (s *ratingService) GetBayesianConfig() BayesianConfig {
	return s.bayesianConfig
}

func (s *ratingService) SetBayesianConfig(config BayesianConfig) {
	s.logger.Info("Updating Bayesian configuration",
		"old_min_votes", s.bayesianConfig.MinVotes,
		"new_min_votes", config.MinVotes,
		"old_global_avg", s.bayesianConfig.GlobalAverage,
		"new_global_avg", config.GlobalAverage,
		"old_confidence_k", s.bayesianConfig.ConfidenceK,
		"new_confidence_k", config.ConfidenceK)

	s.bayesianConfig = config
	s.globalAverage = config.GlobalAverage
}

// Helper functions
func isConflictError(err error) bool {
	return fmt.Sprintf("%v", err) == "user has already rated this movie"
}

func isNotFoundError(err error) bool {
	return fmt.Sprintf("%v", err) == "not found"
}
