package user

import (
	"context"
	"errors"
	"fmt"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
	"thermondo/internal/pkg/cache"
	pkgerrors "thermondo/internal/pkg/errors"
	"thermondo/internal/pkg/interfaces"
)

type UserProfileRequest struct {
	UserID string `json:"user_id"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	SortBy string `json:"sort_by"` // "created_at", "score", "title"
	Order  string `json:"order"`   // "asc", "desc"
}

type UserRatingWithMovie struct {
	Rating       *rating.Rating `json:"rating"`
	Movie        *movies.Movie  `json:"movie"`
	MovieAverage float64        `json:"movie_average"`   // How this movie is rated by others
	TotalRatings int64          `json:"total_ratings"`   // How many people rated this movie
	UserVsAvg    string         `json:"user_vs_average"` // "above", "below", "same"
}

type UserProfileStats struct {
	TotalRatings      int64            `json:"total_ratings"`
	AverageScore      float64          `json:"average_score"`
	ScoreDistribution map[int]int64    `json:"score_distribution"` // User's rating distribution
	FavoriteGenre     string           `json:"favorite_genre"`
	GenreBreakdown    map[string]int64 `json:"genre_breakdown"`
}

type userService struct {
	userRepository users.UserRepository
	ratingRepo     rating.Repository
	movieRepo      movies.Repository
	idGenerator    interfaces.IDGenerator
	timeProvider   interfaces.TimeProvider
	cache          cache.Cache
}

func NewUserService(
	userRepository users.UserRepository,
	ratingRepo rating.Repository,
	movieRepo movies.Repository,
	idGenerator interfaces.IDGenerator,
	timeProvider interfaces.TimeProvider,
	cache cache.Cache,
) UserService {
	return &userService{
		userRepository: userRepository,
		ratingRepo:     ratingRepo,
		movieRepo:      movieRepo,
		idGenerator:    idGenerator,
		timeProvider:   timeProvider,
		cache:          cache,
	}
}

func (s *userService) GetUserProfile(ctx context.Context, req UserProfileRequest) ([]*UserRatingWithMovie, *UserProfileStats, error) {
	// Check if user exists
	user, err := s.userRepository.FindByID(ctx, users.UserID(req.UserID))
	if err != nil {
		return nil, nil, pkgerrors.NewInternalError("Failed to check user existence")
	}
	if user == nil {
		return nil, nil, errors.New("user not found")
	}

	// Try to get profile from cache
	cacheKey := cache.UserProfileKeyFunc(req.UserID, req.Limit, req.Offset, req.SortBy)
	var cachedProfile []*UserRatingWithMovie
	if err := s.cache.Get(ctx, cacheKey, &cachedProfile); err == nil {
		// If we have cached profile, get stats from cache as well
		statsKey := cache.UserStatsKeyFunc(req.UserID)
		var cachedStats *UserProfileStats
		if err := s.cache.Get(ctx, statsKey, &cachedStats); err == nil {
			return cachedProfile, cachedStats, nil
		}
	}

	// Get user's ratings with pagination
	searchOptions := []rating.SearchOption{
		rating.WithLimit(req.Limit),
		rating.WithOffset(req.Offset),
		rating.WithSort(req.SortBy, req.Order),
	}

	userRatings, err := s.ratingRepo.GetByUser(ctx, users.UserID(req.UserID), searchOptions...)
	if err != nil {
		return nil, nil, pkgerrors.NewInternalError("Failed to get user ratings")
	}

	// Get movie details and stats for each rating
	userRatingsWithMovies := make([]*UserRatingWithMovie, len(userRatings))
	for i, userRating := range userRatings {
		movie, err := s.movieRepo.GetByID(ctx, userRating.MovieID)
		if err != nil {
			continue
		}

		// Get movie rating stats
		movieStats, err := s.ratingRepo.GetMovieStats(ctx, userRating.MovieID)
		if err != nil {
			movieStats = &rating.MovieRatingStats{
				AverageScore: 0,
				TotalRatings: 0,
			}
		}

		// Compare user rating vs movie average
		userVsAvg := s.compareUserRatingToAverage(userRating.Score, movieStats.AverageScore)

		userRatingsWithMovies[i] = &UserRatingWithMovie{
			Rating:       userRating,
			Movie:        movie,
			MovieAverage: movieStats.AverageScore,
			TotalRatings: movieStats.TotalRatings,
			UserVsAvg:    userVsAvg,
		}
	}

	// Get user stats
	userStats, err := s.GetUserStats(ctx, req.UserID)
	if err != nil {
		return nil, nil, pkgerrors.NewInternalError("Failed to get user statistics")
	}

	// Cache the results
	if err := s.cache.Set(ctx, cacheKey, userRatingsWithMovies, cache.UserProfileTTL); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache user profile: %v\n", err)
	}

	statsKey := cache.UserStatsKeyFunc(req.UserID)
	if err := s.cache.Set(ctx, statsKey, userStats, cache.UserStatsTTL); err != nil {
		// Log cache error but don't fail the request
		fmt.Printf("Failed to cache user stats: %v\n", err)
	}

	return userRatingsWithMovies, userStats, nil
}

func (s *userService) GetUserStats(ctx context.Context, userID string) (*UserProfileStats, error) {
	cacheKey := cache.UserStatsKeyFunc(userID)
	var cachedStats *UserProfileStats
	if err := s.cache.Get(ctx, cacheKey, &cachedStats); err == nil {
		return cachedStats, nil
	}

	user, err := s.userRepository.FindByID(ctx, users.UserID(userID))
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to check user existence")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	allRatings, err := s.ratingRepo.GetByUser(ctx, users.UserID(userID))
	if err != nil {
		return nil, pkgerrors.NewInternalError("Failed to get user ratings for stats")
	}

	if len(allRatings) == 0 {
		emptyStats := &UserProfileStats{
			TotalRatings:      0,
			AverageScore:      0,
			ScoreDistribution: make(map[int]int64),
			FavoriteGenre:     "",
			GenreBreakdown:    make(map[string]int64),
		}

		if err := s.cache.Set(ctx, cacheKey, emptyStats, cache.UserStatsTTL); err != nil {
			fmt.Printf("Failed to cache empty user stats: %v\n", err)
		}

		return emptyStats, nil
	}

	// Calculate basic stats
	totalScore := 0
	scoreDistribution := make(map[int]int64)
	genreBreakdown := make(map[string]int64)

	for _, rating := range allRatings {
		totalScore += rating.Score
		scoreDistribution[rating.Score]++

		// Get movie for genre info
		if movie, err := s.movieRepo.GetByID(ctx, rating.MovieID); err == nil {
			genreBreakdown[movie.Genre]++
		}
	}

	averageScore := float64(totalScore) / float64(len(allRatings))

	// Find favorite genre
	favoriteGenre := ""
	maxGenreCount := int64(0)
	for genre, count := range genreBreakdown {
		if count > maxGenreCount {
			maxGenreCount = count
			favoriteGenre = genre
		}
	}

	stats := &UserProfileStats{
		TotalRatings:      int64(len(allRatings)),
		AverageScore:      averageScore,
		ScoreDistribution: scoreDistribution,
		FavoriteGenre:     favoriteGenre,
		GenreBreakdown:    genreBreakdown,
	}

	// Cache the stats
	if err := s.cache.Set(ctx, cacheKey, stats, cache.UserStatsTTL); err != nil {
		fmt.Printf("Failed to cache user stats: %v\n", err)
	}

	return stats, nil
}

func (s *userService) compareUserRatingToAverage(userScore int, movieAverage float64) string {
	if movieAverage == 0 {
		return "only_rating" // User is the only one who rated
	}

	userFloat := float64(userScore)
	diff := userFloat - movieAverage

	switch {
	case diff > 0.5:
		return "much_above"
	case diff > 0.1:
		return "above"
	case diff < -0.5:
		return "much_below"
	case diff < -0.1:
		return "below"
	default:
		return "same"
	}
}

// InvalidateUserCache invalidates all cached data for a user
func (s *userService) InvalidateUserCache(ctx context.Context, userID string) error {
	// Delete all user profile cache entries
	profilePattern := "user_profile:" + userID + ":*:*:*"
	if err := s.cache.DeletePattern(ctx, profilePattern); err != nil {
		return fmt.Errorf("failed to delete profile cache: %w", err)
	}

	// Delete user stats cache
	statsKey := cache.UserStatsKeyFunc(userID)
	if err := s.cache.Delete(ctx, statsKey); err != nil {
		return fmt.Errorf("failed to delete stats cache: %w", err)
	}

	return nil
}
