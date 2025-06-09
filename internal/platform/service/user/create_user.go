package user

import (
	"context"
	"thermondo/internal/domain/users"
)

type UserFilters struct {
	Name   string
	Email  string
	Status string
	Page   int
	Limit  int
}

type UserService interface {
	CreateUser(ctx context.Context, user users.CreateUserRequest) (*users.User, error)
	FindUserByID(ctx context.Context, id string) (*users.User, error)
	FindUserByEmail(ctx context.Context, email string) (*users.User, error)
	ListUsers(ctx context.Context, page, limit int) ([]*users.User, int, error)

	// User profile
	GetUserProfile(ctx context.Context, req UserProfileRequest) ([]*UserRatingWithMovie, *UserProfileStats, error)
	GetUserStats(ctx context.Context, userID string) (*UserProfileStats, error)
	InvalidateUserCache(ctx context.Context, userID string) error
}

func (s *userService) CreateUser(ctx context.Context, user users.CreateUserRequest) (*users.User, error) {
	existingUser, err := s.userRepository.FindByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, users.ErrUserAlreadyExists
	}

	u, err := users.NewUser(
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		s.idGenerator,
		s.timeProvider,
		users.WithRole(users.Role(user.Role)),
		users.WithTimestamps(s.timeProvider.Now(), s.timeProvider.Now()),
	)
	if err != nil {
		return nil, err
	}

	savedUser, err := s.userRepository.Create(ctx, u)
	if err != nil {
		return nil, err
	}

	return savedUser, nil
}

func (s *userService) FindUserByID(ctx context.Context, id string) (*users.User, error) {
	return s.userRepository.FindByID(ctx, users.UserID(id))
}

func (s *userService) FindUserByEmail(ctx context.Context, email string) (*users.User, error) {
	return s.userRepository.FindByEmail(ctx, email)
}

func (s *userService) ListUsers(ctx context.Context, page, limit int) ([]*users.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get total count for pagination info
	total, err := s.userRepository.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated users
	users, err := s.userRepository.List(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
