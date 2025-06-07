package movies

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"log/slog"
	"thermondo/internal/domain/movies"
	appErrors "thermondo/internal/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test Helpers
func createTestMovie() *movies.Movie {
	return &movies.Movie{
		ID:           movies.MovieID("test-id-123"),
		Title:        "Test Movie",
		Description:  "Test Description",
		ReleaseYear:  2023,
		Genre:        "Action",
		Director:     "Test Director",
		DurationMins: 120,
		Language:     "English",
		Country:      "USA",
		Rating:       movies.Rating("PG-13"),
		Budget:       int64Ptr(100000000),
		Revenue:      int64Ptr(500000000),
		IMDbID:       stringPtr("tt1234567"),
		PosterURL:    stringPtr("https://example.com/poster.jpg"),
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func createMockRows(movies []*movies.Movie) *sql.Rows {
	// Create a mock sql.Rows that will be used by the repository
	rows := &sql.Rows{}
	return rows
}

// Tests
func TestCreateMovie(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		req           movies.CreateMovieRequest
		mockSetup     func(*MockMovieRepository, *MockIDGenerator, *MockTimeProvider)
		expectedError error
		expectedMovie *movies.Movie
		expectedCalls func(*MockMovieRepository, *MockIDGenerator, *MockTimeProvider)
	}{
		{
			name: "success - all fields provided",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
				Rating:       stringPtr(movies.RatingPG13.String()),
				Budget:       int64Ptr(100000000),
				Revenue:      int64Ptr(500000000),
				IMDbID:       stringPtr("tt1234567"),
				PosterURL:    stringPtr("https://example.com/poster.jpg"),
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
				repo.On("Save", ctx, mock.AnythingOfType("*movies.Movie")).Return(createTestMovie(), nil)
			},
			expectedMovie: createTestMovie(),
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertExpectations(t)
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "success - minimal required fields",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
				repo.On("Save", ctx, mock.AnythingOfType("*movies.Movie")).Return(createTestMovie(), nil)
			},
			expectedMovie: createTestMovie(),
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertExpectations(t)
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if empty title",
			req: movies.CreateMovieRequest{
				Title:        "",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "error - invalid release year (too old)",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  movies.FirstMovieYear - 1,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now).Maybe()
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "error - invalid release year (too far in future)",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  now.Year() + movies.MaxFutureYears + 1,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if empty genre",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if empty director",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if invalid duration",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 0,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if empty language",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if empty country",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if negative budget",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
				Budget:       int64Ptr(-1000000),
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if negative revenue",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
				Revenue:      int64Ptr(-500000),
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertNotCalled(t, "Save")
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if invalid rating",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
				Rating:       stringPtr("INVALID"),
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
				repo.On("Save", ctx, mock.AnythingOfType("*movies.Movie")).Return(createTestMovie(), nil)
			},
			expectedMovie: createTestMovie(),
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertExpectations(t)
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should fail if movie with ID already exists",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
				repo.On("Save", ctx, mock.AnythingOfType("*movies.Movie")).Return(nil, errors.New("movie with ID already exists"))
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertExpectations(t)
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
		{
			name: "should return error if repository internal error",
			req: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "Test Description",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			mockSetup: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id-123")
				timeProv.On("Now").Return(now)
				repo.On("Save", ctx, mock.AnythingOfType("*movies.Movie")).Return(nil, errors.New("database error"))
			},
			expectedError: &appErrors.AppError{},
			expectedCalls: func(repo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.AssertExpectations(t)
				idGen.AssertExpectations(t)
				timeProv.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProvider := new(MockTimeProvider)
			logger := slog.Default()

			service := NewMovieService(mockRepo, mockIDGen, mockTimeProvider, logger)
			tt.mockSetup(mockRepo, mockIDGen, mockTimeProvider)

			result, err := service.CreateMovie(ctx, tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedMovie.ID, result.ID)
				assert.Equal(t, tt.expectedMovie.Title, result.Title)
			}

			tt.expectedCalls(mockRepo, mockIDGen, mockTimeProvider)
		})
	}
}

func TestGetAllMovies(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		limit          int
		offset         int
		sortBy         string
		order          string
		mockSetup      func(*MockMovieRepository)
		expectedMovies []*movies.Movie
		expectedCount  int64
		expectedError  error
	}{
		{
			name:   "should return all movies based on limit and offset",
			limit:  10,
			offset: 0,
			sortBy: "title",
			order:  "asc",
			mockSetup: func(repo *MockMovieRepository) {
				expectedMovies := []*movies.Movie{
					createTestMovie(),
					createTestMovie(),
				}
				repo.On("GetAll", ctx, mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(2), nil)
			},
			expectedMovies: []*movies.Movie{
				createTestMovie(),
				createTestMovie(),
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:   "should return error if repository error on get all",
			limit:  10,
			offset: 0,
			sortBy: "title",
			order:  "asc",
			mockSetup: func(repo *MockMovieRepository) {
				repo.On("GetAll", ctx, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedMovies: nil,
			expectedCount:  0,
			expectedError:  &appErrors.AppError{},
		},
		{
			name:   "should return error if repository error on count",
			limit:  10,
			offset: 0,
			sortBy: "title",
			order:  "asc",
			mockSetup: func(repo *MockMovieRepository) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				repo.On("GetAll", ctx, mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(0), errors.New("count error"))
			},
			expectedMovies: nil,
			expectedCount:  0,
			expectedError:  &appErrors.AppError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProvider := new(MockTimeProvider)
			logger := slog.Default()

			service := NewMovieService(mockRepo, mockIDGen, mockTimeProvider, logger)
			tt.mockSetup(mockRepo)

			result, count, err := service.GetAllMovies(ctx, tt.limit, tt.offset, tt.sortBy, tt.order)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, result)
				assert.Equal(t, int64(0), count)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedMovies, result)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetMovieByID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		movieID       string
		mockSetup     func(*MockMovieRepository)
		expectedMovie *movies.Movie
		expectedError error
	}{
		{
			name:    "should return movie by id",
			movieID: "test-id-123",
			mockSetup: func(repo *MockMovieRepository) {
				expectedMovie := createTestMovie()
				repo.On("GetByID", ctx, movies.MovieID("test-id-123")).Return(expectedMovie, nil)
			},
			expectedMovie: createTestMovie(),
			expectedError: nil,
		},
		{
			name:    "should return error if movie not found",
			movieID: "non-existent-id",
			mockSetup: func(repo *MockMovieRepository) {
				repo.On("GetByID", ctx, movies.MovieID("non-existent-id")).Return(nil, errors.New("not found"))
			},
			expectedMovie: nil,
			expectedError: &appErrors.AppError{},
		},
		{
			name:    "error - internal error",
			movieID: "test-id-123",
			mockSetup: func(repo *MockMovieRepository) {
				repo.On("GetByID", ctx, movies.MovieID("test-id-123")).Return(nil, errors.New("database error"))
			},
			expectedMovie: nil,
			expectedError: &appErrors.AppError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProvider := new(MockTimeProvider)
			logger := slog.Default()

			service := NewMovieService(mockRepo, mockIDGen, mockTimeProvider, logger)
			tt.mockSetup(mockRepo)

			result, err := service.GetMovieByID(ctx, tt.movieID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedMovie, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestSearchMovies(t *testing.T) {
	ctx := context.Background()
	currentTime := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		req            movies.SearchMoviesRequest
		mockSetup      func(*MockMovieRepository, *MockTimeProvider)
		expectedMovies []*movies.Movie
		expectedCount  int64
		expectedError  error
	}{
		{
			name: "should search movies by title",
			req: movies.SearchMoviesRequest{
				Query:  "Test",
				Limit:  10,
				Offset: 0,
				SortBy: "title",
				Order:  "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				repo.On("SearchByTitle", ctx, "Test", mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(1), nil)
			},
			expectedMovies: []*movies.Movie{createTestMovie()},
			expectedCount:  1,
			expectedError:  nil,
		},
		{
			name: "should search movies by genre",
			req: movies.SearchMoviesRequest{
				Genre:  "Action",
				Limit:  10,
				Offset: 0,
				SortBy: "title",
				Order:  "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				repo.On("GetByGenre", ctx, "Action", mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(1), nil)
			},
			expectedMovies: []*movies.Movie{createTestMovie()},
			expectedCount:  1,
			expectedError:  nil,
		},
		{
			name: "should search movies by director",
			req: movies.SearchMoviesRequest{
				Director: "Test Director",
				Limit:    10,
				Offset:   0,
				SortBy:   "title",
				Order:    "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				repo.On("GetByDirector", ctx, "Test Director", mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(1), nil)
			},
			expectedMovies: []*movies.Movie{createTestMovie()},
			expectedCount:  1,
			expectedError:  nil,
		},
		{
			name: "should search movies by year range",
			req: movies.SearchMoviesRequest{
				MinYear: intPtr(2020),
				MaxYear: intPtr(2024),
				Limit:   10,
				Offset:  0,
				SortBy:  "title",
				Order:   "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				timeProv.On("Now").Return(currentTime)
				repo.On("GetByYearRange", ctx, 2020, 2024, mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(1), nil)
			},
			expectedMovies: []*movies.Movie{createTestMovie()},
			expectedCount:  1,
			expectedError:  nil,
		},
		{
			name: "should search movies by year range with only min year",
			req: movies.SearchMoviesRequest{
				MinYear: intPtr(2020),
				Limit:   10,
				Offset:  0,
				SortBy:  "title",
				Order:   "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				timeProv.On("Now").Return(currentTime)
				expectedMaxYear := 2024 + movies.MaxFutureYears
				repo.On("GetByYearRange", ctx, 2020, expectedMaxYear, mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(1), nil)
			},
			expectedMovies: []*movies.Movie{createTestMovie()},
			expectedCount:  1,
			expectedError:  nil,
		},
		{
			name: "should return all movies when no search criteria provided",
			req: movies.SearchMoviesRequest{
				Limit:  10,
				Offset: 0,
				SortBy: "title",
				Order:  "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				expectedMovies := []*movies.Movie{createTestMovie()}
				repo.On("GetAll", ctx, mock.Anything).Return(expectedMovies, nil)
				repo.On("Count", ctx).Return(int64(1), nil)
			},
			expectedMovies: []*movies.Movie{createTestMovie()},
			expectedCount:  1,
			expectedError:  nil,
		},
		{
			name: "should return error if search fails",
			req: movies.SearchMoviesRequest{
				Query:  "Test",
				Limit:  10,
				Offset: 0,
				SortBy: "title",
				Order:  "asc",
			},
			mockSetup: func(repo *MockMovieRepository, timeProv *MockTimeProvider) {
				repo.On("SearchByTitle", ctx, "Test", mock.Anything).Return(nil, errors.New("search error"))
			},
			expectedMovies: nil,
			expectedCount:  0,
			expectedError:  &appErrors.AppError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProvider := new(MockTimeProvider)
			logger := slog.Default()

			service := NewMovieService(mockRepo, mockIDGen, mockTimeProvider, logger)
			tt.mockSetup(mockRepo, mockTimeProvider)

			result, count, err := service.SearchMovies(ctx, tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedError, err)
				assert.Nil(t, result)
				assert.Equal(t, int64(0), count)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedMovies, result)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockRepo.AssertExpectations(t)
			mockTimeProvider.AssertExpectations(t)
		})
	}
}
