package ratings

import (
	"context"
	"thermondo/internal/domain/rating"

	ratingService "thermondo/internal/platform/service/rating"

	"github.com/stretchr/testify/mock"
)

// MockRatingService is a mock implementation of the rating.Service interface
type MockRatingService struct {
	mock.Mock
}

func (m *MockRatingService) CreateRating(ctx context.Context, req ratingService.CreateRatingRequest) (*rating.Rating, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingService) GetRatingByID(ctx context.Context, id string) (*rating.Rating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingService) UpdateRating(ctx context.Context, id string, req ratingService.UpdateRatingRequest) (*rating.Rating, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingService) DeleteRating(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRatingService) GetMovieRatings(ctx context.Context, movieID string, limit, offset int, sortBy, order string) ([]*rating.Rating, int64, error) {
	args := m.Called(ctx, movieID, limit, offset, sortBy, order)
	return args.Get(0).([]*rating.Rating), args.Get(1).(int64), args.Error(2)
}

func (m *MockRatingService) GetMovieStats(ctx context.Context, movieID string) (*rating.MovieRatingStats, error) {
	args := m.Called(ctx, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.MovieRatingStats), args.Error(1)
}

func (m *MockRatingService) GetUserRating(ctx context.Context, userID, movieID string) (*rating.Rating, error) {
	args := m.Called(ctx, userID, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingService) GetBayesianConfig() ratingService.BayesianConfig {
	args := m.Called()
	return args.Get(0).(ratingService.BayesianConfig)
}

func (m *MockRatingService) SetBayesianConfig(config ratingService.BayesianConfig) {
	m.Called(config)
}

func (m *MockRatingService) GetEnhancedMovieStats(ctx context.Context, movieID string) (*ratingService.EnhancedMovieStats, error) {
	args := m.Called(ctx, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ratingService.EnhancedMovieStats), args.Error(1)
}

func (m *MockRatingService) UpdateGlobalAverage(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRatingService) GetUserRatings(ctx context.Context, userID string, limit, offset int, sortBy, order string) ([]*rating.Rating, int64, error) {
	args := m.Called(ctx, userID, limit, offset, sortBy, order)
	return args.Get(0).([]*rating.Rating), args.Get(1).(int64), args.Error(2)
}
