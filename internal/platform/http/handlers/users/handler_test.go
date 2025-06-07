package users

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"thermondo/internal/domain/users"
	"thermondo/internal/pkg/http/response"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of the UserService interface
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, user users.CreateUserRequest) (*users.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *MockUserService) FindUserByID(ctx context.Context, id string) (*users.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *MockUserService) FindUserByEmail(ctx context.Context, email string) (*users.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*users.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, page, limit int) ([]*users.User, int, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*users.User), args.Int(1), args.Error(2)
}

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
				service.On("CreateUser", mock.Anything, mock.Anything).Return(&users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Role:      users.RoleUser,
					IsActive:  true,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: createUserResponse{
				ID:        "test-id",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      "user",
				IsActive:  boolPtr(true),
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(service *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   response.ErrorResponse{Error: "Invalid JSON"},
		},
		{
			name: "user already exists",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "existing@example.com",
				Password:  "password123",
				Role:      "user",
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.Anything).Return(nil, users.ErrUserAlreadyExists)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   response.ErrorResponse{Error: "user already exists"},
		},
		{
			name: "internal server error",
			requestBody: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "error@example.com",
				Password:  "password123",
				Role:      "user",
			},
			mockSetup: func(service *MockUserService) {
				service.On("CreateUser", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
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
			handler := NewHandler(mockService, nil)

			// Create request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			// Execute
			handler.CreateUser(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			if resp, ok := tt.expectedBody.(createUserResponse); ok {
				var actual createUserResponse
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
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
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
			handler := NewHandler(mockService, nil)

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
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        "test-id-2",
						FirstName: "Jane",
						LastName:  "Smith",
						Email:     "jane.smith@example.com",
						Role:      "admin",
						IsActive:  true,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
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
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
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
			handler := NewHandler(mockService, nil)
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
