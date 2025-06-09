package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"thermondo/internal/domain/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name              string
		req               users.CreateUserRequest
		setupMocks        func(*MockUserRepository, *MockRatingRepository, *MockMovieRepository, *MockIDGenerator, *MockTimeProvider)
		expectedUser      *users.User
		expectedError     error
		expectedErrorFunc func(*MockIDGenerator, *MockTimeProvider) error
	}{
		{
			name: "successful user creation",
			req: users.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "password123",
				Role:      "user",
			},
			setupMocks: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id")
				timeProv.On("Now").Return(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
				repo.On("FindByEmail", mock.Anything, "john.doe@example.com").Return(nil, nil)
				expectedUser := &users.User{
					ID:        "test-id",
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
					Password:  "password123",
					Role:      users.Role("user"),
					IsActive:  true,
					CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				repo.On("Create", mock.Anything, mock.Anything).Return(expectedUser, nil)
			},
			expectedUser: &users.User{
				ID:        "test-id",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "password123",
				Role:      users.Role("user"),
				IsActive:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: nil,
		},
		{
			name: "email already exists",
			req: users.CreateUserRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "existing@example.com",
				Password:  "password123",
				Role:      "user",
			},
			setupMocks: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				existingUser := &users.User{
					ID:        "existing-id",
					Email:     "existing@example.com",
					FirstName: "Existing",
					LastName:  "User",
				}
				repo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedUser:  nil,
			expectedError: errors.New("user already exists"),
		},
		{
			name: "invalid role",
			req: users.CreateUserRequest{
				FirstName: "Invalid",
				LastName:  "Role",
				Email:     "invalid.role@example.com",
				Password:  "password123",
				Role:      "invalid_role",
			},
			setupMocks: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.On("FindByEmail", mock.Anything, "invalid.role@example.com").Return(nil, nil)
				timeProv.On("Now").Return(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
				idGen.On("Generate").Return("test-id")
			},
			expectedUser: nil,
			expectedErrorFunc: func(idGen *MockIDGenerator, timeProv *MockTimeProvider) error {
				_, err := users.NewUser(
					"Invalid",
					"Role",
					"invalid.role@example.com",
					"password123",
					idGen,
					timeProv,
					users.WithRole(users.Role("invalid_role")),
				)
				return err
			},
		},
		{
			name: "repository error",
			req: users.CreateUserRequest{
				FirstName: "Error",
				LastName:  "Case",
				Email:     "error@example.com",
				Password:  "password123",
				Role:      "user",
			},
			setupMocks: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				idGen.On("Generate").Return("test-id")
				timeProv.On("Now").Return(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC))
				repo.On("FindByEmail", mock.Anything, "error@example.com").Return(nil, nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(MockUserRepository)
			mockRatingRepo := new(MockRatingRepository)
			mockMovieRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProvider := new(MockTimeProvider)
			tt.setupMocks(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProvider)

			service := NewUserService(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProvider, nil)
			user, err := service.CreateUser(context.Background(), tt.req)

			if tt.expectedErrorFunc != nil {
				expectedErr := tt.expectedErrorFunc(mockIDGen, mockTimeProvider)
				assert.Error(t, err)
				assert.Equal(t, expectedErr.Error(), err.Error())
				assert.Nil(t, user)
			} else if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.FirstName, user.FirstName)
				assert.Equal(t, tt.expectedUser.LastName, user.LastName)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.Equal(t, tt.expectedUser.Role, user.Role)
				assert.Equal(t, tt.expectedUser.IsActive, user.IsActive)
				assert.Equal(t, tt.expectedUser.CreatedAt, user.CreatedAt)
				assert.Equal(t, tt.expectedUser.UpdatedAt, user.UpdatedAt)
			}

			mockRepo.AssertExpectations(t)
			mockRatingRepo.AssertExpectations(t)
			mockMovieRepo.AssertExpectations(t)
			mockIDGen.AssertExpectations(t)
			mockTimeProvider.AssertExpectations(t)
		})
	}
}
