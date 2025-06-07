package movies

import (
	"context"
	"thermondo/internal/domain/movies"

	"github.com/stretchr/testify/mock"
)

// Mock service for testing
type mockMovieService struct {
	mock.Mock
}

func (m *mockMovieService) CreateMovie(ctx context.Context, req movies.CreateMovieRequest) (*movies.Movie, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*movies.Movie), args.Error(1)
}

func (m *mockMovieService) GetMovieByID(ctx context.Context, id string) (*movies.Movie, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*movies.Movie), args.Error(1)
}

func (m *mockMovieService) SearchMovies(ctx context.Context, req movies.SearchMoviesRequest) ([]*movies.Movie, int64, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*movies.Movie), args.Get(1).(int64), args.Error(2)
}

func (m *mockMovieService) GetAllMovies(ctx context.Context, limit, offset int, sortBy, order string) ([]*movies.Movie, int64, error) {
	args := m.Called(ctx, limit, offset, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*movies.Movie), args.Get(1).(int64), args.Error(2)
}
