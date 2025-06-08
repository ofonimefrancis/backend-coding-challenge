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

func TestFindUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockUserRepository)
		expectedUser  *users.User
		expectedError error
	}{
		{
			name:   "successful user retrieval",
			userID: "test-id",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("FindByID", mock.Anything, users.UserID("test-id")).Return(&users.User{
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
			expectedUser: &users.User{
				ID:        "test-id",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      users.RoleUser,
				IsActive:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userID: "non-existent-id",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("FindByID", mock.Anything, users.UserID("non-existent-id")).Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: "error-id",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("FindByID", mock.Anything, users.UserID("error-id")).Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil, nil)
			user, err := service.FindUserByID(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectedUser == nil {
					assert.Nil(t, user)
				} else {
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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestFindUserByEmail(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		mockSetup     func(*MockUserRepository)
		expectedUser  *users.User
		expectedError error
	}{
		{
			name:  "successful user retrieval",
			email: "john.doe@example.com",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("FindByEmail", mock.Anything, "john.doe@example.com").Return(&users.User{
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
			expectedUser: &users.User{
				ID:        "test-id",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Role:      users.RoleUser,
				IsActive:  true,
				CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedError: nil,
		},
		{
			name:  "user not found",
			email: "non-existent@example.com",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("FindByEmail", mock.Anything, "non-existent@example.com").Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: nil,
		},
		{
			name:  "repository error",
			email: "error@example.com",
			mockSetup: func(repo *MockUserRepository) {
				repo.On("FindByEmail", mock.Anything, "error@example.com").Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil, nil)
			user, err := service.FindUserByEmail(context.Background(), tt.email)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tt.expectedUser == nil {
					assert.Nil(t, user)
				} else {
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
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name          string
		page          int
		limit         int
		mockSetup     func(*MockUserRepository)
		expectedUsers []*users.User
		expectedTotal int
		expectedError error
	}{
		{
			name:  "successful user list retrieval",
			page:  1,
			limit: 10,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 1, 10).Return([]*users.User{
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
						CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}, nil)
				repo.On("Count", mock.Anything).Return(2, nil)
			},
			expectedUsers: []*users.User{
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
					CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			expectedTotal: 2,
			expectedError: nil,
		},
		{
			name:  "empty user list",
			page:  1,
			limit: 10,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 1, 10).Return([]*users.User{}, nil)
				repo.On("Count", mock.Anything).Return(0, nil)
			},
			expectedUsers: []*users.User{},
			expectedTotal: 0,
			expectedError: nil,
		},
		{
			name:  "repository error on list",
			page:  1,
			limit: 10,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Count", mock.Anything).Return(0, nil)
				repo.On("List", mock.Anything, 1, 10).Return(nil, errors.New("database error"))
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errors.New("database error"),
		},
		{
			name:  "repository error on count",
			page:  1,
			limit: 10,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("Count", mock.Anything).Return(0, errors.New("database error"))
			},
			expectedUsers: nil,
			expectedTotal: 0,
			expectedError: errors.New("database error"),
		},
		{
			name:  "invalid page number",
			page:  0,
			limit: 10,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 1, 10).Return([]*users.User{}, nil)
				repo.On("Count", mock.Anything).Return(0, nil)
			},
			expectedUsers: []*users.User{},
			expectedTotal: 0,
			expectedError: nil,
		},
		{
			name:  "invalid limit",
			page:  1,
			limit: 0,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 1, 20).Return([]*users.User{}, nil)
				repo.On("Count", mock.Anything).Return(0, nil)
			},
			expectedUsers: []*users.User{},
			expectedTotal: 0,
			expectedError: nil,
		},
		{
			name:  "limit exceeds maximum",
			page:  1,
			limit: 101,
			mockSetup: func(repo *MockUserRepository) {
				repo.On("List", mock.Anything, 1, 20).Return([]*users.User{}, nil)
				repo.On("Count", mock.Anything).Return(0, nil)
			},
			expectedUsers: []*users.User{},
			expectedTotal: 0,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil, nil)
			users, total, err := service.ListUsers(context.Background(), tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, users)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Equal(t, len(tt.expectedUsers), len(users))
				assert.Equal(t, tt.expectedTotal, total)

				for i, expectedUser := range tt.expectedUsers {
					user := users[i]
					assert.Equal(t, expectedUser.ID, user.ID)
					assert.Equal(t, expectedUser.FirstName, user.FirstName)
					assert.Equal(t, expectedUser.LastName, user.LastName)
					assert.Equal(t, expectedUser.Email, user.Email)
					assert.Equal(t, expectedUser.Role, user.Role)
					assert.Equal(t, expectedUser.IsActive, user.IsActive)
					assert.Equal(t, expectedUser.CreatedAt, user.CreatedAt)
					assert.Equal(t, expectedUser.UpdatedAt, user.UpdatedAt)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
