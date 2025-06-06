package user

import (
	"context"
	"thermondo/internal/domain/users"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *users.User) (*users.User, error) {
	args := m.Called(ctx, user)
	var u *users.User
	if args.Get(0) != nil {
		u = args.Get(0).(*users.User)
	}
	return u, args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id users.UserID) (*users.User, error) {
	args := m.Called(ctx, id)
	var u *users.User
	if args.Get(0) != nil {
		u = args.Get(0).(*users.User)
	}
	return u, args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	args := m.Called(ctx, email)
	var u *users.User
	if args.Get(0) != nil {
		u = args.Get(0).(*users.User)
	}
	return u, args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context, page, limit int) ([]*users.User, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]*users.User), args.Error(1)
}

func (m *MockUserRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// MockIDGenerator is a mock implementation of the IDGenerator interface
type MockIDGenerator struct {
	mock.Mock
}

func (m *MockIDGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}

// MockTimeProvider is a mock implementation of the TimeProvider interface
type MockTimeProvider struct {
	mock.Mock
}

func (m *MockTimeProvider) Now() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}
