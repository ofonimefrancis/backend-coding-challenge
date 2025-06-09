package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Add a mock cache struct
type mockCache struct{ mock.Mock }

func (m *mockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
func (m *mockCache) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}
func (m *mockCache) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}
func (m *mockCache) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}
func (m *mockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}
func (m *mockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}
func (m *mockCache) MGet(ctx context.Context, keys []string, dest interface{}) error {
	args := m.Called(ctx, keys, dest)
	return args.Error(0)
}
func (m *mockCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	args := m.Called(ctx, items, ttl)
	return args.Error(0)
}
func (m *mockCache) Ping(ctx context.Context) error { args := m.Called(ctx); return args.Error(0) }
func (m *mockCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(time.Duration), args.Error(1)
}

func TestFindUserByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockUserRepository, *MockRatingRepository, *MockMovieRepository, *MockIDGenerator, *MockTimeProvider)
		expectedUser  *users.User
		expectedError error
	}{
		{
			name:   "successful user retrieval",
			userID: "test-id",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.On("FindByID", mock.Anything, users.UserID("non-existent-id")).Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: nil,
		},
		{
			name:   "repository error",
			userID: "error-id",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.On("FindByID", mock.Anything, users.UserID("error-id")).Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRatingRepo := new(MockRatingRepository)
			mockMovieRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProv := new(MockTimeProvider)
			tt.mockSetup(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv)

			mockCache := new(mockCache)
			service := NewUserService(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)
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
		mockSetup     func(*MockUserRepository, *MockRatingRepository, *MockMovieRepository, *MockIDGenerator, *MockTimeProvider)
		expectedUser  *users.User
		expectedError error
	}{
		{
			name:  "successful user retrieval",
			email: "john.doe@example.com",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.On("FindByEmail", mock.Anything, "non-existent@example.com").Return(nil, nil)
			},
			expectedUser:  nil,
			expectedError: nil,
		},
		{
			name:  "repository error",
			email: "error@example.com",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
				repo.On("FindByEmail", mock.Anything, "error@example.com").Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRatingRepo := new(MockRatingRepository)
			mockMovieRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProv := new(MockTimeProvider)
			tt.mockSetup(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv)

			mockCache := new(mockCache)
			service := NewUserService(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)
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
		mockSetup     func(*MockUserRepository, *MockRatingRepository, *MockMovieRepository, *MockIDGenerator, *MockTimeProvider)
		expectedUsers []*users.User
		expectedTotal int
		expectedError error
	}{
		{
			name:  "successful user list retrieval",
			page:  1,
			limit: 10,
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider) {
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
			mockRatingRepo := new(MockRatingRepository)
			mockMovieRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProv := new(MockTimeProvider)
			tt.mockSetup(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv)

			mockCache := new(mockCache)
			service := NewUserService(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)
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

func TestGetUserProfile(t *testing.T) {
	tests := []struct {
		name            string
		req             UserProfileRequest
		mockSetup       func(*MockUserRepository, *MockRatingRepository, *MockMovieRepository, *MockIDGenerator, *MockTimeProvider, *mockCache)
		expectedRatings []*UserRatingWithMovie
		expectedStats   *UserProfileStats
		expectedError   error
	}{
		{
			name: "successful profile retrieval",
			req: UserProfileRequest{
				UserID: "test-id",
				Limit:  10,
				Offset: 0,
				SortBy: "created_at",
				Order:  "desc",
			},
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				// Mock cache miss
				cache.On("Get", mock.Anything, "user_profile:test-id:10:0:created_at", mock.Anything).Return(errors.New("cache miss"))
				cache.On("Get", mock.Anything, "user_stats:test-id", mock.Anything).Return(errors.New("cache miss"))

				// Mock user existence check
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

				// Mock ratings
				ratings := []*rating.Rating{
					{
						ID:        "rating-1",
						UserID:    "test-id",
						MovieID:   "movie-1",
						Score:     4,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				ratingRepo.On("GetByUser", mock.Anything, users.UserID("test-id"), mock.Anything).Return(ratings, nil)

				// Mock movie
				movie := &movies.Movie{
					ID:          "movie-1",
					Title:       "Test Movie",
					Genre:       "Action",
					ReleaseYear: 2023,
				}
				movieRepo.On("GetByID", mock.Anything, movies.MovieID("movie-1")).Return(movie, nil)

				// Mock movie stats
				movieStats := &rating.MovieRatingStats{
					AverageScore: 4.5,
					TotalRatings: 100,
				}
				ratingRepo.On("GetMovieStats", mock.Anything, movies.MovieID("movie-1")).Return(movieStats, nil)

				// Mock cache set
				cache.On("Set", mock.Anything, "user_profile:test-id:10:0:created_at", mock.Anything, mock.Anything).Return(nil)
				cache.On("Set", mock.Anything, "user_stats:test-id", mock.Anything, mock.Anything).Return(nil)
			},
			expectedRatings: []*UserRatingWithMovie{
				{
					Rating: &rating.Rating{
						ID:        "rating-1",
						UserID:    "test-id",
						MovieID:   "movie-1",
						Score:     4,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					Movie: &movies.Movie{
						ID:          "movie-1",
						Title:       "Test Movie",
						Genre:       "Action",
						ReleaseYear: 2023,
					},
					MovieAverage: 4.5,
					TotalRatings: 100,
					UserVsAvg:    "below",
				},
			},
			expectedStats: &UserProfileStats{
				TotalRatings: 1,
				AverageScore: 4.0,
				ScoreDistribution: map[int]int64{
					4: 1,
				},
				FavoriteGenre: "Action",
				GenreBreakdown: map[string]int64{
					"Action": 1,
				},
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			req: UserProfileRequest{
				UserID: "non-existent-id",
				Limit:  10,
				Offset: 0,
			},
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				repo.On("FindByID", mock.Anything, users.UserID("non-existent-id")).Return(nil, nil)
			},
			expectedRatings: nil,
			expectedStats:   nil,
			expectedError:   errors.New("user not found"),
		},
		{
			name: "error getting ratings",
			req: UserProfileRequest{
				UserID: "test-id",
				Limit:  10,
				Offset: 0,
				SortBy: "created_at",
			},
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				// Mock cache miss
				cache.On("Get", mock.Anything, "user_profile:test-id:10:0:created_at", mock.Anything).Return(errors.New("cache miss"))
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
				ratingRepo.On("GetByUser", mock.Anything, users.UserID("test-id"), mock.Anything).Return(nil, errors.New("Failed to get user ratings"))
			},
			expectedRatings: nil,
			expectedStats:   nil,
			expectedError:   errors.New("Failed to get user ratings"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRatingRepo := new(MockRatingRepository)
			mockMovieRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProv := new(MockTimeProvider)
			mockCache := new(mockCache)
			tt.mockSetup(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)

			service := NewUserService(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)
			ratings, stats, err := service.GetUserProfile(context.Background(), tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, ratings)
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ratings)
				assert.NotNil(t, stats)
				assert.Equal(t, len(tt.expectedRatings), len(ratings))
				assert.Equal(t, tt.expectedStats.TotalRatings, stats.TotalRatings)
				assert.Equal(t, tt.expectedStats.AverageScore, stats.AverageScore)
				assert.Equal(t, tt.expectedStats.FavoriteGenre, stats.FavoriteGenre)
				assert.Equal(t, tt.expectedStats.ScoreDistribution, stats.ScoreDistribution)
				assert.Equal(t, tt.expectedStats.GenreBreakdown, stats.GenreBreakdown)

				for i, expectedRating := range tt.expectedRatings {
					rating := ratings[i]
					assert.Equal(t, expectedRating.Rating.ID, rating.Rating.ID)
					assert.Equal(t, expectedRating.Rating.UserID, rating.Rating.UserID)
					assert.Equal(t, expectedRating.Rating.MovieID, rating.Rating.MovieID)
					assert.Equal(t, expectedRating.Rating.Score, rating.Rating.Score)
					assert.Equal(t, expectedRating.Rating.CreatedAt, rating.Rating.CreatedAt)
					assert.Equal(t, expectedRating.Movie.ID, rating.Movie.ID)
					assert.Equal(t, expectedRating.Movie.Title, rating.Movie.Title)
					assert.Equal(t, expectedRating.Movie.Genre, rating.Movie.Genre)
					assert.Equal(t, expectedRating.Movie.ReleaseYear, rating.Movie.ReleaseYear)
					assert.Equal(t, expectedRating.MovieAverage, rating.MovieAverage)
					assert.Equal(t, expectedRating.TotalRatings, rating.TotalRatings)
					assert.Equal(t, expectedRating.UserVsAvg, rating.UserVsAvg)
				}
			}

			mockRepo.AssertExpectations(t)
			mockRatingRepo.AssertExpectations(t)
			mockMovieRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestGetUserStats(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		mockSetup     func(*MockUserRepository, *MockRatingRepository, *MockMovieRepository, *MockIDGenerator, *MockTimeProvider, *mockCache)
		expectedStats *UserProfileStats
		expectedError error
	}{
		{
			name:   "successful stats retrieval",
			userID: "test-id",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				// Mock cache miss
				cache.On("Get", mock.Anything, "user_stats:test-id", mock.Anything).Return(errors.New("cache miss"))
				// Mock cache set
				cache.On("Set", mock.Anything, "user_stats:test-id", mock.Anything, mock.Anything).Return(nil)
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

				// Mock ratings
				ratings := []*rating.Rating{
					{
						ID:        "rating-1",
						UserID:    "test-id",
						MovieID:   "movie-1",
						Score:     4,
						CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					},
					{
						ID:        "rating-2",
						UserID:    "test-id",
						MovieID:   "movie-2",
						Score:     5,
						CreatedAt: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					},
				}
				ratingRepo.On("GetByUser", mock.Anything, users.UserID("test-id"), mock.Anything).Return(ratings, nil)

				// Mock movies
				movie1 := &movies.Movie{
					ID:          "movie-1",
					Title:       "Test Movie 1",
					Genre:       "Action",
					ReleaseYear: 2023,
				}
				movie2 := &movies.Movie{
					ID:          "movie-2",
					Title:       "Test Movie 2",
					Genre:       "Drama",
					ReleaseYear: 2023,
				}
				movieRepo.On("GetByID", mock.Anything, movies.MovieID("movie-1")).Return(movie1, nil)
				movieRepo.On("GetByID", mock.Anything, movies.MovieID("movie-2")).Return(movie2, nil)
			},
			expectedStats: &UserProfileStats{
				TotalRatings: 2,
				AverageScore: 4.5,
				ScoreDistribution: map[int]int64{
					4: 1,
					5: 1,
				},
				FavoriteGenre: "Action",
				GenreBreakdown: map[string]int64{
					"Action": 1,
					"Drama":  1,
				},
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userID: "non-existent-id",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				// Mock cache miss
				cache.On("Get", mock.Anything, "user_stats:non-existent-id", mock.Anything).Return(errors.New("cache miss"))
				repo.On("FindByID", mock.Anything, users.UserID("non-existent-id")).Return(nil, nil)
			},
			expectedStats: nil,
			expectedError: errors.New("user not found"),
		},
		{
			name:   "error getting ratings",
			userID: "test-id",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				// Mock cache miss
				cache.On("Get", mock.Anything, "user_stats:test-id", mock.Anything).Return(errors.New("cache miss"))
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
				ratingRepo.On("GetByUser", mock.Anything, users.UserID("test-id"), mock.Anything).Return(nil, errors.New("Failed to get user ratings for stats"))
			},
			expectedStats: nil,
			expectedError: errors.New("Failed to get user ratings for stats"),
		},
		{
			name:   "no ratings",
			userID: "test-id",
			mockSetup: func(repo *MockUserRepository, ratingRepo *MockRatingRepository, movieRepo *MockMovieRepository, idGen *MockIDGenerator, timeProv *MockTimeProvider, cache *mockCache) {
				// Mock cache miss
				cache.On("Get", mock.Anything, "user_stats:test-id", mock.Anything).Return(errors.New("cache miss"))
				// Mock cache set
				cache.On("Set", mock.Anything, "user_stats:test-id", mock.Anything, mock.Anything).Return(nil)
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
				ratingRepo.On("GetByUser", mock.Anything, users.UserID("test-id"), mock.Anything).Return([]*rating.Rating{}, nil)
			},
			expectedStats: &UserProfileStats{
				TotalRatings:      0,
				AverageScore:      0,
				ScoreDistribution: make(map[int]int64),
				FavoriteGenre:     "",
				GenreBreakdown:    make(map[string]int64),
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockRatingRepo := new(MockRatingRepository)
			mockMovieRepo := new(MockMovieRepository)
			mockIDGen := new(MockIDGenerator)
			mockTimeProv := new(MockTimeProvider)
			mockCache := new(mockCache)
			tt.mockSetup(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)

			service := NewUserService(mockRepo, mockRatingRepo, mockMovieRepo, mockIDGen, mockTimeProv, mockCache)
			stats, err := service.GetUserStats(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, stats)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, stats)
				assert.Equal(t, tt.expectedStats.TotalRatings, stats.TotalRatings)
				assert.Equal(t, tt.expectedStats.AverageScore, stats.AverageScore)
				// Accept either genre as FavoriteGenre if there is a tie
				if len(tt.expectedStats.GenreBreakdown) > 1 {
					_, ok := tt.expectedStats.GenreBreakdown[stats.FavoriteGenre]
					assert.True(t, ok, "FavoriteGenre should be one of the genres in GenreBreakdown")
				} else {
					assert.Equal(t, tt.expectedStats.FavoriteGenre, stats.FavoriteGenre)
				}
				assert.Equal(t, tt.expectedStats.ScoreDistribution, stats.ScoreDistribution)
				assert.Equal(t, tt.expectedStats.GenreBreakdown, stats.GenreBreakdown)
			}

			mockRepo.AssertExpectations(t)
			mockRatingRepo.AssertExpectations(t)
			mockMovieRepo.AssertExpectations(t)
		})
	}
}
