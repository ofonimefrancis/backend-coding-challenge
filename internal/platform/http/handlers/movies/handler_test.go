package movies

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"thermondo/internal/domain/movies"
	"thermondo/internal/pkg/errors"
)

// Test helper to create a test movie
func createTestMovie() *movies.Movie {
	budget := int64(10000000000)  // $100M (in cents)
	revenue := int64(25000000000) // $250M (in cents)
	imdbID := "tt1234567"
	posterURL := "https://example.com/poster.jpg"

	return &movies.Movie{
		ID:           movies.MovieID("test-movie-123"),
		Title:        "Test Movie",
		Description:  "A great test movie",
		ReleaseYear:  2023,
		Genre:        "Action",
		Director:     "Test Director",
		DurationMins: 120,
		Rating:       "PG-13",
		Language:     "English",
		Country:      "USA",
		Budget:       &budget,
		Revenue:      &revenue,
		IMDbID:       &imdbID,
		PosterURL:    &posterURL,
		CreatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// Test helper to create request body
func createRequestBody(data interface{}) io.Reader {
	jsonData, _ := json.Marshal(data)
	return bytes.NewReader(jsonData)
}

func TestCreateMovieHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*mockMovieService)
		expectedStatus int
		expectedBody   func(t *testing.T, body string)
		expectError    bool
	}{
		{
			name: "successful movie creation",
			requestBody: movies.CreateMovieRequest{
				Title:        "Test Movie",
				Description:  "A great test movie",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			setupMock: func(m *mockMovieService) {
				expectedReq := movies.CreateMovieRequest{
					Title:        "Test Movie",
					Description:  "A great test movie",
					ReleaseYear:  2023,
					Genre:        "Action",
					Director:     "Test Director",
					DurationMins: 120,
					Language:     "English",
					Country:      "USA",
				}

				movie := createTestMovie()
				m.On("CreateMovie", mock.Anything, expectedReq).Return(movie, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body string) {
				var response CreateMovieResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Equal(t, "test-movie-123", response.ID)
				assert.Equal(t, "Test Movie", response.Title)
				assert.Equal(t, "A great test movie", response.Description)
				assert.Equal(t, 2023, response.ReleaseYear)
				assert.Equal(t, "Action", response.Genre)
				assert.Equal(t, "Test Director", response.Director)
				assert.Equal(t, 120, response.DurationMins)
				assert.Equal(t, "PG-13", response.Rating)
				assert.Equal(t, "English", response.Language)
				assert.Equal(t, "USA", response.Country)
				assert.NotNil(t, response.Budget)
				assert.Equal(t, int64(10000000000), *response.Budget)
				assert.NotNil(t, response.Revenue)
				assert.Equal(t, int64(25000000000), *response.Revenue)
				assert.NotNil(t, response.IMDbID)
				assert.Equal(t, "tt1234567", *response.IMDbID)
				assert.NotNil(t, response.PosterURL)
				assert.Equal(t, "https://example.com/poster.jpg", *response.PosterURL)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.CreatedAt)
				assert.Equal(t, "2024-01-01T12:00:00Z", response.UpdatedAt)
			},
			expectError: false,
		},
		{
			name: "successful movie creation with minimal fields",
			requestBody: movies.CreateMovieRequest{
				Title:        "Minimal Movie",
				Description:  "A minimal test movie",
				ReleaseYear:  2023,
				Genre:        "Drama",
				Director:     "Minimal Director",
				DurationMins: 90,
				Language:     "English",
				Country:      "UK",
			},
			setupMock: func(m *mockMovieService) {
				expectedReq := movies.CreateMovieRequest{
					Title:        "Minimal Movie",
					Description:  "A minimal test movie",
					ReleaseYear:  2023,
					Genre:        "Drama",
					Director:     "Minimal Director",
					DurationMins: 90,
					Language:     "English",
					Country:      "UK",
				}

				movie := &movies.Movie{
					ID:           movies.MovieID("minimal-movie-123"),
					Title:        "Minimal Movie",
					Description:  "A minimal test movie",
					ReleaseYear:  2023,
					Genre:        "Drama",
					Director:     "Minimal Director",
					DurationMins: 90,
					Rating:       "",
					Language:     "English",
					Country:      "UK",
					Budget:       nil,
					Revenue:      nil,
					IMDbID:       nil,
					PosterURL:    nil,
					CreatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				}
				m.On("CreateMovie", mock.Anything, expectedReq).Return(movie, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body string) {
				var response CreateMovieResponse
				err := json.Unmarshal([]byte(body), &response)
				require.NoError(t, err)

				assert.Equal(t, "minimal-movie-123", response.ID)
				assert.Equal(t, "Minimal Movie", response.Title)
				assert.Nil(t, response.Budget)
				assert.Nil(t, response.Revenue)
				assert.Nil(t, response.IMDbID)
				assert.Nil(t, response.PosterURL)
			},
			expectError: false,
		},
		{
			name:        "invalid JSON payload",
			requestBody: `{"invalid": json}`,
			setupMock: func(m *mockMovieService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid JSON")
			},
			expectError: true,
		},
		{
			name: "service returns bad request error",
			requestBody: movies.CreateMovieRequest{
				Title:        "",
				Description:  "A test movie",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			setupMock: func(m *mockMovieService) {
				m.On("CreateMovie", mock.Anything, mock.AnythingOfType("movies.CreateMovieRequest")).
					Return(nil, errors.NewBadRequestError("title cannot be empty"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "title cannot be empty")
			},
			expectError: true,
		},
		{
			name: "service returns conflict error",
			requestBody: movies.CreateMovieRequest{
				Title:        "Duplicate Movie",
				Description:  "A duplicate test movie",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			setupMock: func(m *mockMovieService) {
				m.On("CreateMovie", mock.Anything, mock.AnythingOfType("movies.CreateMovieRequest")).
					Return(nil, errors.NewConflictError("Movie with this ID already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Movie with this ID already exists")
			},
			expectError: true,
		},
		{
			name: "service returns internal server error",
			requestBody: movies.CreateMovieRequest{
				Title:        "Server Error Movie",
				Description:  "A test movie that causes server error",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			setupMock: func(m *mockMovieService) {
				m.On("CreateMovie", mock.Anything, mock.AnythingOfType("movies.CreateMovieRequest")).
					Return(nil, errors.NewInternalError("Database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Database connection failed")
			},
			expectError: true,
		},
		{
			name: "service returns unexpected error",
			requestBody: movies.CreateMovieRequest{
				Title:        "Unexpected Error Movie",
				Description:  "A test movie that causes unexpected error",
				ReleaseYear:  2023,
				Genre:        "Action",
				Director:     "Test Director",
				DurationMins: 120,
				Language:     "English",
				Country:      "USA",
			},
			setupMock: func(m *mockMovieService) {
				m.On("CreateMovie", mock.Anything, mock.AnythingOfType("movies.CreateMovieRequest")).
					Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Internal server error")
			},
			expectError: true,
		},
		{
			name:           "empty request body",
			requestBody:    "",
			setupMock:      func(m *mockMovieService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid JSON")
			},
			expectError: true,
		},
		{
			name:           "malformed JSON",
			requestBody:    `{"title": "Test Movie", "release_year": "not_a_number"}`,
			setupMock:      func(m *mockMovieService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body string) {
				assert.Contains(t, body, "Invalid JSON")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(mockMovieService)
			tt.setupMock(mockService)

			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			handler := NewHandler(mockService, logger)

			var body io.Reader
			if str, ok := tt.requestBody.(string); ok {
				body = strings.NewReader(str)
			} else {
				body = createRequestBody(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/movies", body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler.CreateMovie(rr, req)
			assert.Equal(t, tt.expectedStatus, rr.Code)

			responseBody := rr.Body.String()
			tt.expectedBody(t, responseBody)
			mockService.AssertExpectations(t)
		})
	}
}

// Test for response structure validation
func TestCreateMovieResponse_Structure(t *testing.T) {
	movie := createTestMovie()

	response := CreateMovieResponse{
		ID:           string(movie.ID),
		Title:        movie.Title,
		Description:  movie.Description,
		ReleaseYear:  movie.ReleaseYear,
		Genre:        movie.Genre,
		Director:     movie.Director,
		DurationMins: movie.DurationMins,
		Rating:       movie.Rating.String(),
		Language:     movie.Language,
		Country:      movie.Country,
		Budget:       movie.Budget,
		Revenue:      movie.Revenue,
		IMDbID:       movie.IMDbID,
		PosterURL:    movie.PosterURL,
		CreatedAt:    movie.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    movie.UpdatedAt.Format(time.RFC3339),
	}

	// Test JSON marshaling/unmarshaling
	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var unmarshaledResponse CreateMovieResponse
	err = json.Unmarshal(jsonData, &unmarshaledResponse)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, response.ID, unmarshaledResponse.ID)
	assert.Equal(t, response.Title, unmarshaledResponse.Title)
	assert.Equal(t, response.Description, unmarshaledResponse.Description)
	assert.Equal(t, response.ReleaseYear, unmarshaledResponse.ReleaseYear)
	assert.Equal(t, response.Genre, unmarshaledResponse.Genre)
	assert.Equal(t, response.Director, unmarshaledResponse.Director)
	assert.Equal(t, response.DurationMins, unmarshaledResponse.DurationMins)
	assert.Equal(t, response.Rating, unmarshaledResponse.Rating)
	assert.Equal(t, response.Language, unmarshaledResponse.Language)
	assert.Equal(t, response.Country, unmarshaledResponse.Country)
	assert.Equal(t, response.Budget, unmarshaledResponse.Budget)
	assert.Equal(t, response.Revenue, unmarshaledResponse.Revenue)
	assert.Equal(t, response.IMDbID, unmarshaledResponse.IMDbID)
	assert.Equal(t, response.PosterURL, unmarshaledResponse.PosterURL)
	assert.Equal(t, response.CreatedAt, unmarshaledResponse.CreatedAt)
	assert.Equal(t, response.UpdatedAt, unmarshaledResponse.UpdatedAt)
}

// Benchmark test for performance
func BenchmarkCreateMovieHandler(b *testing.B) {
	mockService := new(mockMovieService)
	movie := createTestMovie()

	mockService.On("CreateMovie", mock.Anything, mock.AnythingOfType("movies.CreateMovieRequest")).
		Return(movie, nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(mockService, logger)

	requestBody := createRequestBody(movies.CreateMovieRequest{
		Title:        "Benchmark Movie",
		Description:  "A benchmark test movie",
		ReleaseYear:  2023,
		Genre:        "Action",
		Director:     "Benchmark Director",
		DurationMins: 120,
		Language:     "English",
		Country:      "USA",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/movies", requestBody)
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()

		// Reset request body for each iteration
		req.Body = io.NopCloser(createRequestBody(movies.CreateMovieRequest{
			Title:        "Benchmark Movie",
			Description:  "A benchmark test movie",
			ReleaseYear:  2023,
			Genre:        "Action",
			Director:     "Benchmark Director",
			DurationMins: 120,
			Language:     "English",
			Country:      "USA",
		}))

		handler.CreateMovie(rr, req)
	}
}
