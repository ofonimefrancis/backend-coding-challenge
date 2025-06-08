package rating

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
)

// Test helpers
func createTestRating() *rating.Rating {
	return &rating.Rating{
		ID:        rating.RatingID("test-rating-123"),
		UserID:    users.UserID("user-123"),
		MovieID:   movies.MovieID("movie-123"),
		Score:     4,
		Review:    "Great movie!",
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func createTestMovieStats() *rating.MovieRatingStats {
	return &rating.MovieRatingStats{
		MovieID:      movies.MovieID("movie-123"),
		AverageScore: 4.2,
		TotalRatings: 10,
		ScoreCount: map[int]int64{
			1: 0,
			2: 1,
			3: 2,
			4: 3,
			5: 4,
		},
	}
}

func setupTestService() (Service, *mockRatingRepository, *mockIDGenerator, *mockTimeProvider) {
	mockRepo := new(mockRatingRepository)
	mockIDGen := &mockIDGenerator{id: "test-rating-123"}
	mockTimeProvider := &mockTimeProvider{now: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	service := NewTestRatingService(mockRepo, mockIDGen, mockTimeProvider, logger)
	return service, mockRepo, mockIDGen, mockTimeProvider
}

func TestCreateRating(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateRatingRequest
		setupMocks     func(*mockRatingRepository, *mockIDGenerator, *mockTimeProvider)
		expectedError  string
		expectSuccess  bool
		validateResult func(*testing.T, *rating.Rating)
	}{
		{
			name: "successful rating creation",
			request: CreateRatingRequest{
				UserID:  "user-123",
				MovieID: "movie-123",
				Score:   4,
				Review:  "Great movie!",
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockIDGen *mockIDGenerator, mockTimeProvider *mockTimeProvider) {
				mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-123"), movies.MovieID("movie-123")).
					Return(nil, errors.New("not found"))

				expectedRating := createTestRating()
				mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(expectedRating, nil)

				// Global average update - expect one call for deterministic unit test
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.2, nil).Once()
			},
			expectedError: "",
			expectSuccess: true,
			validateResult: func(t *testing.T, result *rating.Rating) {
				assert.Equal(t, rating.RatingID("test-rating-123"), result.ID)
				assert.Equal(t, users.UserID("user-123"), result.UserID)
				assert.Equal(t, movies.MovieID("movie-123"), result.MovieID)
				assert.Equal(t, 4, result.Score)
				assert.Equal(t, "Great movie!", result.Review)
			},
		},
		{
			name: "user already rated movie",
			request: CreateRatingRequest{
				UserID:  "user-123",
				MovieID: "movie-123",
				Score:   4,
				Review:  "Great movie!",
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockIDGen *mockIDGenerator, mockTimeProvider *mockTimeProvider) {
				existingRating := createTestRating()
				mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-123"), movies.MovieID("movie-123")).
					Return(existingRating, nil)
			},
			expectedError: "User has already rated this movie",
			expectSuccess: false,
		},
		{
			name: "invalid score",
			request: CreateRatingRequest{
				UserID:  "user-123",
				MovieID: "movie-123",
				Score:   6, // Invalid score > 5
				Review:  "Great movie!",
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockIDGen *mockIDGenerator, mockTimeProvider *mockTimeProvider) {
				// User hasn't rated this movie before
				mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-123"), movies.MovieID("movie-123")).
					Return(nil, errors.New("not found"))
			},
			expectedError: "score must be between 1 and 5",
			expectSuccess: false,
		},
		{
			name: "repository save error",
			request: CreateRatingRequest{
				UserID:  "user-123",
				MovieID: "movie-123",
				Score:   4,
				Review:  "Great movie!",
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockIDGen *mockIDGenerator, mockTimeProvider *mockTimeProvider) {
				// User hasn't rated this movie before
				mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-123"), movies.MovieID("movie-123")).
					Return(nil, errors.New("not found"))

				// Repository save fails
				mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(nil, errors.New("database error"))
			},
			expectedError: "Failed to create rating",
			expectSuccess: false,
		},
		{
			name: "conflict error from repository",
			request: CreateRatingRequest{
				UserID:  "user-123",
				MovieID: "movie-123",
				Score:   4,
				Review:  "Great movie!",
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockIDGen *mockIDGenerator, mockTimeProvider *mockTimeProvider) {
				// User hasn't rated this movie before (initial check)
				mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-123"), movies.MovieID("movie-123")).
					Return(nil, errors.New("not found"))

				// But repository returns conflict error (race condition)
				mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(nil, errors.New("user has already rated this movie"))
			},
			expectedError: "User has already rated this movie",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, mockIDGen, mockTimeProvider := setupTestService()
			tt.setupMocks(mockRepo, mockIDGen, mockTimeProvider)

			result, err := service.CreateRating(context.Background(), tt.request)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetRatingByID(t *testing.T) {
	tests := []struct {
		name           string
		ratingID       string
		setupMocks     func(*mockRatingRepository)
		expectedError  string
		expectSuccess  bool
		validateResult func(*testing.T, *rating.Rating)
	}{
		{
			name:     "successful get rating by ID",
			ratingID: "test-rating-123",
			setupMocks: func(mockRepo *mockRatingRepository) {
				expectedRating := createTestRating()
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
					Return(expectedRating, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *rating.Rating) {
				assert.Equal(t, rating.RatingID("test-rating-123"), result.ID)
				assert.Equal(t, 4, result.Score)
			},
		},
		{
			name:     "rating not found",
			ratingID: "nonexistent-rating",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("nonexistent-rating")).
					Return(nil, errors.New("not found"))
			},
			expectedError: "Rating not found",
			expectSuccess: false,
		},
		{
			name:     "repository error",
			ratingID: "test-rating-123",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
					Return(nil, errors.New("database error"))
			},
			expectedError: "Failed to get rating",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()
			tt.setupMocks(mockRepo)

			result, err := service.GetRatingByID(context.Background(), tt.ratingID)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateRating(t *testing.T) {
	tests := []struct {
		name           string
		ratingID       string
		request        UpdateRatingRequest
		setupMocks     func(*mockRatingRepository, *mockTimeProvider)
		expectedError  string
		expectSuccess  bool
		validateResult func(*testing.T, *rating.Rating)
	}{
		{
			name:     "successful score update",
			ratingID: "test-rating-123",
			request: UpdateRatingRequest{
				Score: intPtr(5),
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockTimeProvider *mockTimeProvider) {
				existingRating := createTestRating()
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
					Return(existingRating, nil)

				updatedRating := *existingRating
				updatedRating.Score = 5
				updatedRating.UpdatedAt = mockTimeProvider.Now()
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(&updatedRating, nil)

				// Global average update
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.2, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *rating.Rating) {
				assert.Equal(t, 5, result.Score)
				assert.Equal(t, "Great movie!", result.Review) // Review unchanged
			},
		},
		{
			name:     "successful review update",
			ratingID: "test-rating-123",
			request: UpdateRatingRequest{
				Review: stringPtr("Updated review"),
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockTimeProvider *mockTimeProvider) {
				existingRating := createTestRating()
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
					Return(existingRating, nil)

				updatedRating := *existingRating
				updatedRating.Review = "Updated review"
				updatedRating.UpdatedAt = mockTimeProvider.Now()
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(&updatedRating, nil)

				// Global average update
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.2, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *rating.Rating) {
				assert.Equal(t, 4, result.Score) // Score unchanged
				assert.Equal(t, "Updated review", result.Review)
			},
		},
		{
			name:     "update both score and review",
			ratingID: "test-rating-123",
			request: UpdateRatingRequest{
				Score:  intPtr(2),
				Review: stringPtr("Changed my mind"),
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockTimeProvider *mockTimeProvider) {
				existingRating := createTestRating()
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
					Return(existingRating, nil)

				updatedRating := *existingRating
				updatedRating.Score = 2
				updatedRating.Review = "Changed my mind"
				updatedRating.UpdatedAt = mockTimeProvider.Now()
				mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(&updatedRating, nil)

				// Global average update
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.2, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *rating.Rating) {
				assert.Equal(t, 2, result.Score)
				assert.Equal(t, "Changed my mind", result.Review)
			},
		},
		{
			name:     "rating not found",
			ratingID: "nonexistent-rating",
			request: UpdateRatingRequest{
				Score: intPtr(5),
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockTimeProvider *mockTimeProvider) {
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("nonexistent-rating")).
					Return(nil, errors.New("not found"))
			},
			expectedError: "Rating not found",
			expectSuccess: false,
		},
		{
			name:     "invalid score update",
			ratingID: "test-rating-123",
			request: UpdateRatingRequest{
				Score: intPtr(6), // Invalid score
			},
			setupMocks: func(mockRepo *mockRatingRepository, mockTimeProvider *mockTimeProvider) {
				existingRating := createTestRating()
				mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
					Return(existingRating, nil)
			},
			expectedError: "score must be between 1 and 5",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, mockTimeProvider := setupTestService()
			tt.setupMocks(mockRepo, mockTimeProvider)

			result, err := service.UpdateRating(context.Background(), tt.ratingID, tt.request)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteRating(t *testing.T) {
	tests := []struct {
		name          string
		ratingID      string
		setupMocks    func(*mockRatingRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:     "successful deletion",
			ratingID: "test-rating-123",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("Delete", mock.Anything, rating.RatingID("test-rating-123")).
					Return(nil)

				// Global average update
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.2, nil)
			},
			expectSuccess: true,
		},
		{
			name:     "rating not found",
			ratingID: "nonexistent-rating",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("Delete", mock.Anything, rating.RatingID("nonexistent-rating")).
					Return(errors.New("not found"))
			},
			expectedError: "Rating not found",
			expectSuccess: false,
		},
		{
			name:     "repository error",
			ratingID: "test-rating-123",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("Delete", mock.Anything, rating.RatingID("test-rating-123")).
					Return(errors.New("database error"))
			},
			expectedError: "Failed to delete rating",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()
			tt.setupMocks(mockRepo)

			err := service.DeleteRating(context.Background(), tt.ratingID)

			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestGetEnhancedMovieStats(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		setupMocks     func(*mockRatingRepository)
		expectedError  string
		expectSuccess  bool
		validateResult func(*testing.T, *EnhancedMovieStats)
	}{
		{
			name:    "successful enhanced stats calculation",
			movieID: "movie-123",
			setupMocks: func(mockRepo *mockRatingRepository) {
				stats := createTestMovieStats()
				mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("movie-123")).
					Return(stats, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *EnhancedMovieStats) {
				assert.Equal(t, movies.MovieID("movie-123"), result.MovieID)
				assert.Equal(t, 4.2, result.AverageScore)
				assert.Equal(t, int64(10), result.TotalRatings)

				// Bayesian average should be lower than simple average for small sample
				assert.Less(t, result.BayesianAverage, result.AverageScore)

				// Confidence should be 1.0 since we have 10 ratings (= MinVotes)
				assert.Equal(t, 1.0, result.Confidence)

				// Should have explanation
				assert.NotEmpty(t, result.Explanation)
				assert.Contains(t, result.Explanation, "10 user ratings")
			},
		},
		{
			name:    "movie with no ratings",
			movieID: "movie-no-ratings",
			setupMocks: func(mockRepo *mockRatingRepository) {
				stats := &rating.MovieRatingStats{
					MovieID:      movies.MovieID("movie-no-ratings"),
					AverageScore: 0,
					TotalRatings: 0,
					ScoreCount:   map[int]int64{},
				}
				mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("movie-no-ratings")).
					Return(stats, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *EnhancedMovieStats) {
				assert.Equal(t, int64(0), result.TotalRatings)
				assert.Equal(t, 3.0, result.BayesianAverage) // Should equal global average
				assert.Equal(t, 0.0, result.Confidence)
				assert.Contains(t, result.Explanation, "No ratings yet")
			},
		},
		{
			name:    "movie with few ratings",
			movieID: "movie-few-ratings",
			setupMocks: func(mockRepo *mockRatingRepository) {
				stats := &rating.MovieRatingStats{
					MovieID:      movies.MovieID("movie-few-ratings"),
					AverageScore: 5.0, // Perfect rating but only few votes
					TotalRatings: 3,
					ScoreCount: map[int]int64{
						5: 3,
					},
				}
				mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("movie-few-ratings")).
					Return(stats, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, result *EnhancedMovieStats) {
				assert.Equal(t, 5.0, result.AverageScore)
				assert.Equal(t, int64(3), result.TotalRatings)

				// Bayesian should pull toward global average (3.0)
				assert.Greater(t, result.BayesianAverage, 3.0)
				assert.Less(t, result.BayesianAverage, 5.0)

				// Low confidence due to few ratings
				assert.Equal(t, 0.3, result.Confidence) // 3/10
				assert.Contains(t, result.Explanation, "small sample size")
			},
		},
		{
			name:    "repository error",
			movieID: "movie-error",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("movie-error")).
					Return(nil, errors.New("database error"))
			},
			expectedError: "Failed to get movie stats",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()
			tt.setupMocks(mockRepo)

			result, err := service.GetEnhancedMovieStats(context.Background(), tt.movieID)

			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateGlobalAverage(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*mockRatingRepository)
		expectedError  string
		expectSuccess  bool
		validateResult func(*testing.T, Service)
	}{
		{
			name: "successful global average update",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.47, nil)
			},
			expectSuccess: true,
			validateResult: func(t *testing.T, service Service) {
				config := service.GetBayesianConfig()
				assert.Equal(t, 3.47, config.GlobalAverage)
			},
		},
		{
			name: "repository error",
			setupMocks: func(mockRepo *mockRatingRepository) {
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(0.0, errors.New("database error"))
			},
			expectedError: "failed to update global average",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()
			tt.setupMocks(mockRepo)

			err := service.UpdateGlobalAverage(context.Background())

			if tt.expectSuccess {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, service)
				}
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestBayesianConfiguration(t *testing.T) {
	service, _, _, _ := setupTestService()

	// Test default configuration
	defaultConfig := service.GetBayesianConfig()
	assert.Equal(t, int64(10), defaultConfig.MinVotes)
	assert.Equal(t, 3.0, defaultConfig.GlobalAverage)
	assert.Equal(t, 25.0, defaultConfig.ConfidenceK)

	// Test configuration update
	newConfig := BayesianConfig{
		MinVotes:      15,
		GlobalAverage: 3.5,
		ConfidenceK:   30.0,
	}

	service.SetBayesianConfig(newConfig)
	updatedConfig := service.GetBayesianConfig()
	assert.Equal(t, newConfig, updatedConfig)
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

// Benchmark tests
func BenchmarkCreateRating(b *testing.B) {
	service, mockRepo, _, _ := setupTestService()

	// Setup mocks for benchmark
	mockRepo.On("GetByUserAndMovie", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("not found"))
	mockRepo.On("Save", mock.Anything, mock.Anything).
		Return(createTestRating(), nil)
	mockRepo.On("GetGlobalAverageRating", mock.Anything).
		Return(3.2, nil)

	request := CreateRatingRequest{
		UserID:  "user-123",
		MovieID: "movie-123",
		Score:   4,
		Review:  "Benchmark test",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateRating(context.Background(), request)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetEnhancedMovieStats(b *testing.B) {
	service, mockRepo, _, _ := setupTestService()

	// Setup mocks for benchmark
	mockRepo.On("GetMovieStats", mock.Anything, mock.Anything).
		Return(createTestMovieStats(), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetEnhancedMovieStats(context.Background(), "movie-123")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Integration-style tests with multiple operations
func TestRatingWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		operations     []func(*testing.T, Service, *mockRatingRepository)
		finalAssertion func(*testing.T, Service)
	}{
		{
			name: "complete rating lifecycle",
			operations: []func(*testing.T, Service, *mockRatingRepository){
				// 1. Create rating
				func(t *testing.T, service Service, mockRepo *mockRatingRepository) {
					mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-123"), movies.MovieID("movie-123")).
						Return(nil, errors.New("not found")).Once()
					mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*rating.Rating")).
						Return(createTestRating(), nil).Once()
					mockRepo.On("GetGlobalAverageRating", mock.Anything).
						Return(3.2, nil).Once()

					req := CreateRatingRequest{
						UserID:  "user-123",
						MovieID: "movie-123",
						Score:   4,
						Review:  "Great movie!",
					}
					result, err := service.CreateRating(context.Background(), req)
					require.NoError(t, err)
					assert.Equal(t, 4, result.Score)
				},
				// 2. Update rating
				func(t *testing.T, service Service, mockRepo *mockRatingRepository) {
					existingRating := createTestRating()
					mockRepo.On("GetByID", mock.Anything, rating.RatingID("test-rating-123")).
						Return(existingRating, nil).Once()

					updatedRating := *existingRating
					updatedRating.Score = 5
					mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*rating.Rating")).
						Return(&updatedRating, nil).Once()
					mockRepo.On("GetGlobalAverageRating", mock.Anything).
						Return(3.3, nil).Once()

					updateReq := UpdateRatingRequest{Score: intPtr(5)}
					result, err := service.UpdateRating(context.Background(), "test-rating-123", updateReq)
					require.NoError(t, err)
					assert.Equal(t, 5, result.Score)
				},
				// 3. Get enhanced stats
				func(t *testing.T, service Service, mockRepo *mockRatingRepository) {
					mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("movie-123")).
						Return(createTestMovieStats(), nil).Once()

					stats, err := service.GetEnhancedMovieStats(context.Background(), "movie-123")
					require.NoError(t, err)
					assert.Greater(t, stats.BayesianAverage, 0.0)
					assert.Greater(t, stats.Confidence, 0.0)
				},
			},
			finalAssertion: func(t *testing.T, service Service) {
				// Verify final state
				config := service.GetBayesianConfig()
				assert.Equal(t, 3.3, config.GlobalAverage) // Should be updated from operations
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()

			// Execute all operations
			for i, operation := range tt.operations {
				t.Logf("Executing operation %d", i+1)
				operation(t, service, mockRepo)
			}

			// Final assertion
			if tt.finalAssertion != nil {
				tt.finalAssertion(t, service)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// Edge case and error scenario tests
func TestRatingServiceEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupTest func(*testing.T, Service, *mockRatingRepository)
		assertion func(*testing.T)
	}{
		{
			name: "bayesian calculation with zero ratings",
			setupTest: func(t *testing.T, service Service, mockRepo *mockRatingRepository) {
				stats := &rating.MovieRatingStats{
					MovieID:      movies.MovieID("empty-movie"),
					AverageScore: 0,
					TotalRatings: 0,
					ScoreCount:   map[int]int64{},
				}
				mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("empty-movie")).
					Return(stats, nil)

				result, err := service.GetEnhancedMovieStats(context.Background(), "empty-movie")
				require.NoError(t, err)

				// Should return global average when no ratings
				assert.Equal(t, 3.0, result.BayesianAverage)
				assert.Equal(t, 0.0, result.Confidence)
				assert.Contains(t, result.Explanation, "No ratings yet")
			},
		},
		{
			name: "bayesian calculation with exactly minimum votes",
			setupTest: func(t *testing.T, service Service, mockRepo *mockRatingRepository) {
				stats := &rating.MovieRatingStats{
					MovieID:      movies.MovieID("exact-min-votes"),
					AverageScore: 4.5,
					TotalRatings: 10, // Exactly minVotes
					ScoreCount: map[int]int64{
						4: 5,
						5: 5,
					},
				}
				mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("exact-min-votes")).
					Return(stats, nil)

				result, err := service.GetEnhancedMovieStats(context.Background(), "exact-min-votes")
				require.NoError(t, err)

				// Should have full confidence with exactly minVotes
				assert.Equal(t, 1.0, result.Confidence)
				assert.Contains(t, result.Explanation, "confidence")
			},
		},
		{
			name: "concurrent rating creation attempts",
			setupTest: func(t *testing.T, service Service, mockRepo *mockRatingRepository) {
				// Simulate race condition where initial check passes but save fails due to constraint
				mockRepo.On("GetByUserAndMovie", mock.Anything, users.UserID("user-race"), movies.MovieID("movie-race")).
					Return(nil, errors.New("not found")).Times(2)

				// First call succeeds
				mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(createTestRating(), nil).Once()
				mockRepo.On("GetGlobalAverageRating", mock.Anything).
					Return(3.2, nil).Once()

				// Second call fails due to race condition
				mockRepo.On("Save", mock.Anything, mock.AnythingOfType("*rating.Rating")).
					Return(nil, errors.New("user has already rated this movie")).Once()

				req := CreateRatingRequest{
					UserID:  "user-race",
					MovieID: "movie-race",
					Score:   4,
					Review:  "Race condition test",
				}

				// First creation should succeed
				result1, err1 := service.CreateRating(context.Background(), req)
				assert.NoError(t, err1)
				assert.NotNil(t, result1)

				// Second creation should fail with conflict error
				result2, err2 := service.CreateRating(context.Background(), req)
				assert.Error(t, err2)
				assert.Nil(t, result2)
				assert.Contains(t, err2.Error(), "already rated")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()
			tt.setupTest(t, service, mockRepo)
			mockRepo.AssertExpectations(t)
		})
	}
}

// Test configuration scenarios
func TestBayesianConfigurationScenarios(t *testing.T) {
	tests := []struct {
		name               string
		config             BayesianConfig
		movieStats         *rating.MovieRatingStats
		expectedBayesian   float64
		expectedConfidence float64
	}{
		{
			name: "conservative configuration",
			config: BayesianConfig{
				MinVotes:      50, // High minimum
				GlobalAverage: 3.0,
				ConfidenceK:   100, // Very conservative
			},
			movieStats: &rating.MovieRatingStats{
				MovieID:      movies.MovieID("conservative-test"),
				AverageScore: 4.8,
				TotalRatings: 5,
				ScoreCount:   map[int]int64{5: 5},
			},
			expectedBayesian:   3.09, // Should be very close to global average due to high K
			expectedConfidence: 0.1,  // 5/50
		},
		{
			name: "liberal configuration",
			config: BayesianConfig{
				MinVotes:      5, // Low minimum
				GlobalAverage: 3.0,
				ConfidenceK:   5, // Not very conservative
			},
			movieStats: &rating.MovieRatingStats{
				MovieID:      movies.MovieID("liberal-test"),
				AverageScore: 4.8,
				TotalRatings: 5,
				ScoreCount:   map[int]int64{5: 5},
			},
			expectedBayesian:   3.9, // Should be closer to movie average due to low K
			expectedConfidence: 1.0, // 5/5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo, _, _ := setupTestService()

			// Set custom configuration
			service.SetBayesianConfig(tt.config)

			// Setup mock
			mockRepo.On("GetMovieStats", mock.Anything, tt.movieStats.MovieID).
				Return(tt.movieStats, nil)

			// Get enhanced stats
			result, err := service.GetEnhancedMovieStats(context.Background(), string(tt.movieStats.MovieID))
			require.NoError(t, err)

			// Validate Bayesian calculation with tolerance
			assert.InDelta(t, tt.expectedBayesian, result.BayesianAverage, 0.1)
			assert.Equal(t, tt.expectedConfidence, result.Confidence)

			mockRepo.AssertExpectations(t)
		})
	}
}

// Test service with custom configuration
func TestServiceWithCustomConfig(t *testing.T) {
	mockRepo := new(mockRatingRepository)
	mockIDGen := &mockIDGenerator{id: "custom-test-123"}
	mockTimeProvider := &mockTimeProvider{now: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	customConfig := BayesianConfig{
		MinVotes:      15,
		GlobalAverage: 3.5,
		ConfidenceK:   20.0,
	}

	service := NewRatingServiceWithConfig(mockRepo, mockIDGen, mockTimeProvider, logger, customConfig)

	// Verify custom configuration is applied
	config := service.GetBayesianConfig()
	assert.Equal(t, customConfig, config)

	// Test that custom configuration affects calculations
	stats := &rating.MovieRatingStats{
		MovieID:      movies.MovieID("custom-config-test"),
		AverageScore: 4.0,
		TotalRatings: 10,
		ScoreCount:   map[int]int64{4: 10},
	}

	mockRepo.On("GetMovieStats", mock.Anything, movies.MovieID("custom-config-test")).
		Return(stats, nil)

	result, err := service.GetEnhancedMovieStats(context.Background(), "custom-config-test")
	require.NoError(t, err)

	// With custom config: confidence should be 10/15 = 0.667
	assert.InDelta(t, 0.667, result.Confidence, 0.01)

	// Bayesian average should be influenced by custom global average (3.5)
	assert.Greater(t, result.BayesianAverage, 3.5)
	assert.Less(t, result.BayesianAverage, 4.0)

	mockRepo.AssertExpectations(t)
}
