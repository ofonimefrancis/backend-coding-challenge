package movies

import (
	"context"
	"database/sql"
	"thermondo/internal/domain/movies"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

// Mock Repository
type MockMovieRepository struct {
	mock.Mock
}

func (m *MockMovieRepository) Save(ctx context.Context, movie *movies.Movie) (*movies.Movie, error) {
	args := m.Called(ctx, movie)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByID(ctx context.Context, id movies.MovieID) (*movies.Movie, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetAll(ctx context.Context, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMovieRepository) SearchByTitle(ctx context.Context, title string, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, title, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByGenre(ctx context.Context, genre string, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, genre, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByDirector(ctx context.Context, director string, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, director, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByYearRange(ctx context.Context, minYear, maxYear int, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, minYear, maxYear, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) Exists(ctx context.Context, id movies.MovieID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockMovieRepository) GetDB() *sqlx.DB {
	return nil
}

func (m *MockMovieRepository) ScanMovies(rows *sql.Rows) ([]*movies.Movie, error) {
	return nil, nil
}

func (m *MockMovieRepository) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}

// Mock ID Generator
type MockIDGenerator struct {
	mock.Mock
}

func (m *MockIDGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

// Mock Time Provider
type MockTimeProvider struct {
	mock.Mock
}

func (m *MockTimeProvider) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}
