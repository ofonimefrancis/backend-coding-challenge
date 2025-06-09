package user

import (
	"context"
	"database/sql"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
	"time"

	"github.com/jmoiron/sqlx"
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

// MockRatingRepository is a mock implementation of the rating.Repository interface
type MockRatingRepository struct {
	mock.Mock
}

func (m *MockRatingRepository) GetByUser(ctx context.Context, userID users.UserID, opts ...rating.SearchOption) ([]*rating.Rating, error) {
	args := m.Called(ctx, userID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rating.Rating), args.Error(1)
}

func (m *MockRatingRepository) GetByUserAndMovie(ctx context.Context, userID users.UserID, movieID movies.MovieID) (*rating.Rating, error) {
	args := m.Called(ctx, userID, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingRepository) GetByID(ctx context.Context, id rating.RatingID) (*rating.Rating, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingRepository) GetByMovie(ctx context.Context, movieID movies.MovieID, opts ...rating.SearchOption) ([]*rating.Rating, error) {
	args := m.Called(ctx, movieID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*rating.Rating), args.Error(1)
}

func (m *MockRatingRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRatingRepository) Delete(ctx context.Context, id rating.RatingID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRatingRepository) Exists(ctx context.Context, id rating.RatingID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockRatingRepository) GetGlobalAverageRating(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRatingRepository) GetMovieStats(ctx context.Context, movieID movies.MovieID) (*rating.MovieRatingStats, error) {
	args := m.Called(ctx, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.MovieRatingStats), args.Error(1)
}

func (m *MockRatingRepository) Save(ctx context.Context, r *rating.Rating) (*rating.Rating, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

func (m *MockRatingRepository) Update(ctx context.Context, r *rating.Rating) (*rating.Rating, error) {
	args := m.Called(ctx, r)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rating.Rating), args.Error(1)
}

// MockMovieRepository is a mock implementation of the movies.Repository interface
type MockMovieRepository struct {
	mock.Mock
}

func (m *MockMovieRepository) GetByID(ctx context.Context, id movies.MovieID) (*movies.Movie, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) Count(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMovieRepository) Exists(ctx context.Context, id movies.MovieID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *MockMovieRepository) GetAll(ctx context.Context, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByDirector(ctx context.Context, director string, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, director, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByGenre(ctx context.Context, genre string, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, genre, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetByYearRange(ctx context.Context, startYear, endYear int, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, startYear, endYear, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) GetDB() *sqlx.DB {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*sqlx.DB)
}

func (m *MockMovieRepository) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	callArgs := m.Called(ctx, query, args)
	if callArgs.Get(0) == nil {
		return nil, callArgs.Error(1)
	}
	return callArgs.Get(0).(*sql.Rows), callArgs.Error(1)
}

func (m *MockMovieRepository) Save(ctx context.Context, movie *movies.Movie) (*movies.Movie, error) {
	args := m.Called(ctx, movie)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) ScanMovies(rows *sql.Rows) ([]*movies.Movie, error) {
	args := m.Called(rows)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
}

func (m *MockMovieRepository) SearchByTitle(ctx context.Context, title string, opts ...movies.SearchOption) ([]*movies.Movie, error) {
	args := m.Called(ctx, title, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*movies.Movie), args.Error(1)
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

func (m *MockUserService) GetUserProfile(ctx context.Context, req UserProfileRequest) ([]*UserRatingWithMovie, *UserProfileStats, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).([]*UserRatingWithMovie), args.Get(1).(*UserProfileStats), args.Error(2)
}

func (m *MockUserService) GetUserStats(ctx context.Context, userID string) (*UserProfileStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserProfileStats), args.Error(1)
}

func (m *MockUserService) InvalidateUserCache(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserService) WarmUserCache(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}
