package users

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
	"thermondo/internal/pkg/http/response"
	userService "thermondo/internal/platform/service/user"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		queryParams    string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful profile retrieval",
			userID:      "test-user-id",
			queryParams: "limit=10&offset=0&sort_by=created_at&order=desc",
			mockSetup: func(service *MockUserService) {
				// Mock user retrieval
				service.On("FindUserByID", mock.Anything, "test-user-id").Return(&users.User{
					ID:        "test-user-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)

				// Mock profile retrieval
				ratings := []*userService.UserRatingWithMovie{
					{
						Rating: &rating.Rating{
							ID:        "rating-1",
							UserID:    "test-user-id",
							MovieID:   "movie-1",
							Score:     5,
							Review:    "Great movie!",
							CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						},
						Movie: &movies.Movie{
							ID:          "movie-1",
							Title:       "The Matrix",
							ReleaseYear: 1999,
							Genre:       "Sci-Fi",
							Director:    "Wachowski Sisters",
							PosterURL:   stringPtr("https://example.com/matrix.jpg"),
						},
						MovieAverage: 4.5,
						TotalRatings: 1000,
						UserVsAvg:    "above",
					},
				}

				stats := &userService.UserProfileStats{
					TotalRatings:      1,
					AverageScore:      5.0,
					ScoreDistribution: map[int]int64{5: 1},
					FavoriteGenre:     "Sci-Fi",
					GenreBreakdown:    map[string]int64{"Sci-Fi": 1},
				}

				service.On("GetUserProfile", mock.Anything, mock.MatchedBy(func(req userService.UserProfileRequest) bool {
					return req.UserID == "test-user-id" &&
						req.Limit == 10 &&
						req.Offset == 0 &&
						req.SortBy == "created_at" &&
						req.Order == "desc"
				})).Return(ratings, stats, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: UserProfileResponse{
				User: UserResponse{
					ID:        "test-user-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Role:      "user",
					IsActive:  true,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
				Stats: UserProfileStatsResponse{
					TotalRatings:      1,
					AverageScore:      5.0,
					ScoreDistribution: map[string]int64{"5": 1},
					FavoriteGenre:     "Sci-Fi",
					GenreBreakdown:    map[string]int64{"Sci-Fi": 1},
				},
				Ratings: []UserRatingWithMovieResponse{
					{
						RatingID:      "rating-1",
						Score:         5,
						Review:        "Great movie!",
						RatedAt:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
						MovieID:       "movie-1",
						Title:         "The Matrix",
						ReleaseYear:   1999,
						Genre:         "Sci-Fi",
						Director:      "Wachowski Sisters",
						PosterURL:     stringPtr("https://example.com/matrix.jpg"),
						MovieAverage:  4.5,
						TotalRatings:  1000,
						UserVsAverage: "above",
					},
				},
				HasMore: false,
				Total:   1,
			},
		},
		{
			name:           "missing user ID",
			userID:         "",
			queryParams:    "",
			mockSetup:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "User ID is required"},
		},
		{
			name:        "invalid limit",
			userID:      "test-user-id",
			queryParams: "limit=101",
			mockSetup: func(service *MockUserService) {
				// No mock setup needed since we validate limit before calling service
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Limit must be between 1 and 100"},
		},
		{
			name:        "invalid sort field",
			userID:      "test-user-id",
			queryParams: "sort_by=invalid",
			mockSetup: func(service *MockUserService) {
				// No mock setup needed since we validate sort field before calling service
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Invalid sort field"},
		},
		{
			name:        "user not found",
			userID:      "non-existent-id",
			queryParams: "",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByID", mock.Anything, "non-existent-id").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   response.ErrorResponse{Error: "user not found"},
		},
		{
			name:        "profile retrieval error",
			userID:      "test-user-id",
			queryParams: "",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByID", mock.Anything, "test-user-id").Return(&users.User{
					ID: "test-user-id",
				}, nil)
				service.On("GetUserProfile", mock.Anything, mock.Anything).Return(nil, nil, errors.New("failed to get profile"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.ErrorResponse{Error: "Internal server error"},
		},
		{
			name:        "user service error",
			userID:      "test-user-id",
			queryParams: "",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByID", mock.Anything, "test-user-id").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.ErrorResponse{Error: "Internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockService := new(MockUserService)
			tt.mockSetup(mockService)

			// Create handler with logger
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewProfileHandler(mockService, logger)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/profile?"+tt.queryParams, nil)
			req = req.WithContext(context.Background())

			// Create response recorder
			recorder := httptest.NewRecorder()

			// Create router and register handler
			router := chi.NewRouter()
			router.Get("/users/{userId}/profile", handler.GetUserProfile)

			// Serve request
			router.ServeHTTP(recorder, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Parse response body
			var responseBody interface{}
			if tt.expectedStatus == http.StatusOK {
				var successResponse UserProfileResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &successResponse)
				assert.NoError(t, err)
				responseBody = successResponse
			} else {
				var errorResponse response.ErrorResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				responseBody = errorResponse
			}

			// Compare response body
			assert.Equal(t, tt.expectedBody, responseBody)

			// Verify mock expectations
			mockService.AssertExpectations(t)
		})
	}
}

// Helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// Add WarmUserCache method to MockUserService
func (m *MockUserService) WarmUserCache(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
