package users

import (
	"context"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
	userService "thermondo/internal/platform/service/user"
	"time"

	"github.com/stretchr/testify/mock"
)

// UserProfileRequest represents a request to get a user's profile
type UserProfileRequest struct {
	UserID string
	Page   int
	Limit  int
}

// UserRatingWithMovie represents a user's rating with movie details
type UserRatingWithMovie struct {
	Rating    *rating.Rating
	Movie     *movies.Movie
	CreatedAt time.Time
}

// UserProfileStats represents statistics about a user's ratings
type UserProfileStats struct {
	TotalRatings      int
	AverageRating     float64
	RatingCounts      map[int]int
	FavoriteGenres    []string
	FavoriteActors    []string
	FavoriteDirectors []string
}

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

func (m *MockUserService) GetUserProfile(ctx context.Context, req userService.UserProfileRequest) ([]*userService.UserRatingWithMovie, *userService.UserProfileStats, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*userService.UserRatingWithMovie), args.Get(1).(*userService.UserProfileStats), args.Error(2)
}

func (m *MockUserService) GetUserStats(ctx context.Context, userID string) (*userService.UserProfileStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userService.UserProfileStats), args.Error(1)
}

func (m *MockUserService) InvalidateUserCache(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
