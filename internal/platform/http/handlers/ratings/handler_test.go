package ratings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"thermondo/internal/domain/rating"
	ratingService "thermondo/internal/platform/service/rating"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createTestRating() *rating.Rating {
	return &rating.Rating{
		ID:        rating.RatingID("test-rating-123"),
		UserID:    "test-user-123",
		MovieID:   "test-movie-123",
		Score:     5,
		Review:    "Great movie!",
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func createRequestBody(v interface{}) io.Reader {
	jsonData, _ := json.Marshal(v)
	return bytes.NewReader(jsonData)
}

func TestCreateRating(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockRatingService)
		expectedStatus int
		expectedBody   func(*testing.T, string)
		expectError    bool
	}{
		{
			name: "successful rating creation",
			requestBody: ratingService.CreateRatingRequest{
				UserID:  "test-user-123",
				MovieID: "test-movie-123",
				Score:   5,
				Review:  "Great movie!",
			},
			setupMock: func(m *MockRatingService) {
				expectedReq := ratingService.CreateRatingRequest{
					UserID:  "test-user-123",
					MovieID: "test-movie-123",
					Score:   5,
					Review:  "Great movie!",
				}
				m.On("CreateRating", mock.Anything, expectedReq).Return(createTestRating(), nil)
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body string) {
				var response CreateRatingResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Equal(t, "test-rating-123", response.ID)
				assert.Equal(t, "test-user-123", response.UserID)
				assert.Equal(t, "test-movie-123", response.MovieID)
				assert.Equal(t, 5, response.Score)
				assert.Equal(t, "Great movie!", response.Review)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.CreatedAt)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.UpdatedAt)
			},
			expectError: false,
		},
		{
			name: "invalid score",
			requestBody: ratingService.CreateRatingRequest{
				UserID:  "test-user-123",
				MovieID: "test-movie-123",
				Score:   6, // Invalid score (should be 1-5)
				Review:  "Great movie!",
			},
			setupMock: func(m *MockRatingService) {
				m.On("CreateRating", mock.Anything, mock.Anything).Return(nil, errors.New("invalid score"))
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "invalid score")
			},
			expectError: true,
		},
		{
			name:        "malformed JSON",
			requestBody: `{"user_id": "test-user-123", "score": "not_a_number"}`,
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid JSON payload")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRatingService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger)

			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = bytes.NewReader([]byte(str))
			} else {
				body = createRequestBody(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/ratings", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler.CreateRating(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(t, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetRatingByID(t *testing.T) {
	tests := []struct {
		name           string
		ratingID       string
		setupMock      func(*MockRatingService)
		expectedStatus int
		expectedBody   func(*testing.T, string)
		expectError    bool
	}{
		{
			name:     "successful rating retrieval",
			ratingID: "test-rating-123",
			setupMock: func(m *MockRatingService) {
				m.On("GetRatingByID", mock.Anything, "test-rating-123").Return(createTestRating(), nil)
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body string) {
				var response RatingResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Equal(t, "test-rating-123", response.ID)
				assert.Equal(t, "test-user-123", response.UserID)
				assert.Equal(t, "test-movie-123", response.MovieID)
				assert.Equal(t, 5, response.Score)
				assert.Equal(t, "Great movie!", response.Review)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.CreatedAt)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.UpdatedAt)
			},
			expectError: false,
		},
		{
			name:     "rating not found",
			ratingID: "non-existent",
			setupMock: func(m *MockRatingService) {
				m.On("GetRatingByID", mock.Anything, "non-existent").Return(nil, errors.New("rating not found"))
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "rating not found")
			},
			expectError: true,
		},
		{
			name:     "missing rating ID",
			ratingID: "",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Rating ID is required")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRatingService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest(http.MethodGet, "/ratings/"+tt.ratingID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.ratingID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetRatingByID(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(t, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetMovieRatings(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		queryParams    string
		setupMock      func(*MockRatingService)
		expectedStatus int
		expectedBody   func(*testing.T, string)
		expectError    bool
	}{
		{
			name:        "successful movie ratings retrieval",
			movieID:     "test-movie-123",
			queryParams: "limit=10&offset=0&sort_by=created_at&order=desc",
			setupMock: func(m *MockRatingService) {
				ratings := []*rating.Rating{createTestRating()}
				m.On("GetMovieRatings", mock.Anything, "test-movie-123", 10, 0, "created_at", "desc").Return(ratings, int64(1), nil)
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body string) {
				var response RatingsListResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Len(t, response.Ratings, 1)
				assert.Equal(t, int64(1), response.Total)
				assert.Equal(t, 10, response.Limit)
				assert.Equal(t, 0, response.Offset)
				assert.False(t, response.HasMore)

				rating := response.Ratings[0]
				assert.Equal(t, "test-rating-123", rating.ID)
				assert.Equal(t, "test-user-123", rating.UserID)
				assert.Equal(t, "test-movie-123", rating.MovieID)
				assert.Equal(t, 5, rating.Score)
				assert.Equal(t, "Great movie!", rating.Review)
			},
			expectError: false,
		},
		{
			name:        "invalid limit",
			movieID:     "test-movie-123",
			queryParams: "limit=101",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "limit must be between 1 and 100")
			},
			expectError: true,
		},
		{
			name:        "invalid sort field",
			movieID:     "test-movie-123",
			queryParams: "sort_by=invalid_field",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "invalid sort_by field")
			},
			expectError: true,
		},
		{
			name:        "missing movie ID",
			movieID:     "",
			queryParams: "",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Movie ID is required")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRatingService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest(http.MethodGet, "/movies/"+tt.movieID+"/ratings?"+tt.queryParams, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("movieId", tt.movieID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetMovieRatings(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(t, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetMovieStats(t *testing.T) {
	tests := []struct {
		name           string
		movieID        string
		setupMock      func(*MockRatingService)
		expectedStatus int
		expectedBody   func(*testing.T, string)
		expectError    bool
	}{
		{
			name:    "successful movie stats retrieval",
			movieID: "test-movie-123",
			setupMock: func(m *MockRatingService) {
				stats := &rating.MovieRatingStats{
					MovieID:      "test-movie-123",
					AverageScore: 4.5,
					TotalRatings: 100,
					ScoreCount: map[int]int64{
						5: 50,
						4: 30,
						3: 20,
					},
				}
				m.On("GetMovieStats", mock.Anything, "test-movie-123").Return(stats, nil)
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body string) {
				var response MovieStatsResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Equal(t, "test-movie-123", response.MovieID)
				assert.Equal(t, 4.5, response.AverageScore)
				assert.Equal(t, int64(100), response.TotalRatings)
				assert.Equal(t, int64(50), response.ScoreCount["5"])
				assert.Equal(t, int64(30), response.ScoreCount["4"])
				assert.Equal(t, int64(20), response.ScoreCount["3"])
			},
			expectError: false,
		},
		{
			name:    "movie not found",
			movieID: "non-existent",
			setupMock: func(m *MockRatingService) {
				m.On("GetMovieStats", mock.Anything, "non-existent").Return(nil, errors.New("movie not found"))
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "movie not found")
			},
			expectError: true,
		},
		{
			name:    "missing movie ID",
			movieID: "",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Movie ID is required")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRatingService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest(http.MethodGet, "/movies/"+tt.movieID+"/stats", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("movieId", tt.movieID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetMovieStats(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(t, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}

func TestGetUserRating(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		movieID        string
		setupMock      func(*MockRatingService)
		expectedStatus int
		expectedBody   func(*testing.T, string)
		expectError    bool
	}{
		{
			name:    "successful user rating retrieval",
			userID:  "test-user-123",
			movieID: "test-movie-123",
			setupMock: func(m *MockRatingService) {
				m.On("GetUserRating", mock.Anything, "test-user-123", "test-movie-123").Return(createTestRating(), nil)
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body string) {
				var response RatingResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Equal(t, "test-rating-123", response.ID)
				assert.Equal(t, "test-user-123", response.UserID)
				assert.Equal(t, "test-movie-123", response.MovieID)
				assert.Equal(t, 5, response.Score)
				assert.Equal(t, "Great movie!", response.Review)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.CreatedAt)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.UpdatedAt)
			},
			expectError: false,
		},
		{
			name:    "rating not found",
			userID:  "test-user-123",
			movieID: "test-movie-123",
			setupMock: func(m *MockRatingService) {
				m.On("GetUserRating", mock.Anything, "test-user-123", "test-movie-123").Return(nil, errors.New("rating not found"))
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "rating not found")
			},
			expectError: true,
		},
		{
			name:    "missing user ID",
			userID:  "",
			movieID: "test-movie-123",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "User ID is required")
			},
			expectError: true,
		},
		{
			name:    "missing movie ID",
			userID:  "test-user-123",
			movieID: "",
			setupMock: func(m *MockRatingService) {
				m.On("GetBayesianConfig").Return(ratingService.DefaultBayesianConfig()).Maybe()
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Movie ID is required")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRatingService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger)

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/ratings/"+tt.movieID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("userId", tt.userID)
			rctx.URLParams.Add("movieId", tt.movieID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.GetUserRating(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			tt.expectedBody(t, rr.Body.String())
			mockService.AssertExpectations(t)
		})
	}
}
