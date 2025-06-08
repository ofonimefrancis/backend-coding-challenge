package rating

import (
	"context"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
	"time"

	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type mockRatingRepository struct {
	mock.Mock
}

func (m *mockRatingRepository) Save(ctx context.Context, r *rating.Rating) (*rating.Rating, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *mockRatingRepository) GetByID(ctx context.Context, id rating.RatingID) (*rating.Rating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *mockRatingRepository) GetByUserAndMovie(ctx context.Context, userID users.UserID, movieID movies.MovieID) (*rating.Rating, error) {
	args := m.Called(ctx, userID, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *mockRatingRepository) GetByUser(ctx context.Context, userID users.UserID, options ...rating.SearchOption) ([]*rating.Rating, error) {
	args := m.Called(ctx, userID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rating.Rating), args.Error(1)
}

func (m *mockRatingRepository) GetByMovie(ctx context.Context, movieID movies.MovieID, options ...rating.SearchOption) ([]*rating.Rating, error) {
	args := m.Called(ctx, movieID, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rating.Rating), args.Error(1)
}

func (m *mockRatingRepository) Update(ctx context.Context, r *rating.Rating) (*rating.Rating, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *mockRatingRepository) Delete(ctx context.Context, id rating.RatingID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRatingRepository) GetMovieStats(ctx context.Context, movieID movies.MovieID) (*rating.MovieRatingStats, error) {
	args := m.Called(ctx, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.MovieRatingStats), args.Error(1)
}

func (m *mockRatingRepository) Exists(ctx context.Context, id rating.RatingID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *mockRatingRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRatingRepository) GetGlobalAverageRating(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}

type mockIDGenerator struct {
	id string
}

func (m *mockIDGenerator) Generate() string {
	return m.id
}

type mockTimeProvider struct {
	now time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	return m.now
}
