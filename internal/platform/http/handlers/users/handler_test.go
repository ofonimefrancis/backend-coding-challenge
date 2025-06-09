package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"thermondo/internal/domain/users"
	"thermondo/internal/pkg/http/response"
	"time"

	"thermondo/internal/pkg/password"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful user creation",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "password123",
				Role:      "user",
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("users.CreateUserRequest")).Return(&users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: UserResponse{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      "user",
				IsActive:  true,
			},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"firstName": 123, // Invalid type for firstName
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("users.CreateUserRequest")).Return(nil, errors.New("invalid input"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Invalid request body"},
		},
		{
			name: "missing required fields",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				// Missing lastName, email, and password
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("users.CreateUserRequest")).Return(nil, errors.New("invalid input"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Invalid request body"},
		},
		{
			name: "invalid email format",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "invalid-email",
				Password:  "password123",
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("users.CreateUserRequest")).Return(nil, errors.New("invalid input"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Invalid request body"},
		},
		{
			name: "invalid role",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "password123",
				Role:      "invalid-role",
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("users.CreateUserRequest")).Return(nil, errors.New("invalid input"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Invalid request body"},
		},
		{
			name: "service error",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "password123",
				Role:      "user",
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.AnythingOfType("users.CreateUserRequest")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.ErrorResponse{Error: "Internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			handler := NewHandler(mockService, logger, "test-secret")

			// Create request
			var reqBody []byte
			var err error
			if tt.requestBody != nil {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			handler.CreateUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if resp, ok := tt.expectedBody.(UserResponse); ok {
				var actual UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				// Don't compare ID, CreatedAt, and UpdatedAt as they are generated
				assert.Equal(t, resp.FirstName, actual.FirstName)
				assert.Equal(t, resp.LastName, actual.LastName)
				assert.Equal(t, resp.Email, actual.Email)
				assert.Equal(t, resp.Role, actual.Role)
				assert.Equal(t, resp.IsActive, actual.IsActive)
			} else if resp, ok := tt.expectedBody.(response.ErrorResponse); ok {
				var actual response.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, resp, actual)
			} else {
				t.Fatalf("unexpected expectedBody type: %T", tt.expectedBody)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	mockService := new(MockUserService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	handler := NewHandler(mockService, logger, "test-secret")

	// Create a properly hashed password for testing
	hashedPassword, _ := password.HashPassword("password123")

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful login",
			requestBody: loginRequest{
				Email:    "john@example.com",
				Password: "password123",
			},
			mockSetup: func() {
				mockService.On("FindUserByEmail", mock.Anything, "john@example.com").Return(&users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
					Password:  hashedPassword,
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: loginResponse{
				Token:     "mock-token",
				ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			},
		},
		{
			name: "invalid credentials",
			requestBody: loginRequest{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			mockSetup: func() {
				mockService.On("FindUserByEmail", mock.Anything, "john@example.com").Return(&users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john@example.com",
					Password:  hashedPassword,
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid credentials",
		},
		{
			name:        "invalid request body",
			requestBody: "not a json",
			mockSetup: func() {
				// No mock setup needed for invalid JSON
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			tt.mockSetup()

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			// Create router and register handler
			router := chi.NewRouter()
			router.Post("/login", handler.Login)

			// Serve request
			router.ServeHTTP(w, req)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response loginResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Greater(t, response.ExpiresAt, time.Now().Unix())
			} else {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody.(string), response["error"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:   "successful user retrieval",
			userID: "test-id",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByID", mock.Anything, "test-id").Return(&users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: UserResponse{
				ID:        "test-id",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      "user",
				IsActive:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
		},
		{
			name:   "user not found",
			userID: "non-existent",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByID", mock.Anything, "non-existent").Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.ErrorResponse{Error: "user not found"},
		},
		{
			name:   "internal server error",
			userID: "error-id",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByID", mock.Anything, "error-id").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.ErrorResponse{Error: "database error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			handler := NewHandler(mockService, logger, "test-secret")

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			w := httptest.NewRecorder()

			// Setup chi router context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Execute
			handler.GetUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if resp, ok := tt.expectedBody.(UserResponse); ok {
				var actual UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, resp, actual)
			} else if resp, ok := tt.expectedBody.(response.ErrorResponse); ok {
				var actual response.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, resp, actual)
			} else {
				t.Fatalf("unexpected expectedBody type: %T", tt.expectedBody)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful user list retrieval",
			queryParams: "page=1&limit=10",
			mockSetup: func(service *MockUserService) {
				users := []*users.User{
					{
						ID:        "test-id-1",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "john.doe@example.com",
						Role:      users.RoleUser,
						IsActive:  true,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        "test-id-2",
						FirstName: "Jane",
						LastName:  "Smith",
						Email:     "jane.smith@example.com",
						Role:      users.RoleAdmin,
						IsActive:  true,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				service.On("ListUsers", mock.Anything, 1, 10).Return(users, 2, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: ListUsersResponse{
				Users: []UserResponse{
					{
						ID:        "test-id-1",
						FirstName: "John",
						LastName:  "Doe",
						Email:     "john.doe@example.com",
						Role:      "user",
						IsActive:  true,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
						UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					},
					{
						ID:        "test-id-2",
						FirstName: "Jane",
						LastName:  "Smith",
						Email:     "jane.smith@example.com",
						Role:      "admin",
						IsActive:  true,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
						UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
					},
				},
				Pagination: &Pagination{
					Page:       1,
					Limit:      10,
					Total:      2,
					TotalPages: 1,
				},
			},
		},
		{
			name:        "search by email",
			queryParams: "email=john.doe@example.com",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByEmail", mock.Anything, "john.doe@example.com").Return(&users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: UserResponse{
				ID:        "test-id",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      "user",
				IsActive:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			},
		},
		{
			name:        "user not found by email",
			queryParams: "email=nonexistent@example.com",
			mockSetup: func(service *MockUserService) {
				service.On("FindUserByEmail", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   response.ErrorResponse{Error: "User not found"},
		},
		{
			name:        "internal server error",
			queryParams: "page=1&limit=10",
			mockSetup: func(service *MockUserService) {
				service.On("ListUsers", mock.Anything, 1, 10).Return([]*users.User{}, 0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   response.ErrorResponse{Error: "Failed to get users"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.mockSetup(mockService)
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
			handler := NewHandler(mockService, logger, "test-secret")
			req := httptest.NewRequest(http.MethodGet, "/users?"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			handler.ListUsers(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if resp, ok := tt.expectedBody.(ListUsersResponse); ok {
				var actual ListUsersResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, resp, actual)
			} else if resp, ok := tt.expectedBody.(UserResponse); ok {
				var actual UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, resp, actual)
			} else if resp, ok := tt.expectedBody.(response.ErrorResponse); ok {
				var actual response.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &actual)
				assert.NoError(t, err)
				assert.Equal(t, resp, actual)
			} else {
				t.Fatalf("unexpected expectedBody type: %T", tt.expectedBody)
			}
		})
	}
}

// Helper function to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
