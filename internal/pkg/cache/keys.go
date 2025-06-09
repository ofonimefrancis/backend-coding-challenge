package cache

import (
	"fmt"
	"time"
)

// Cache key patterns for different domains
const (
	// Movie-related cache keys
	MovieStatsKey   = "movie_stats:%s"        // movie_stats:{movie_id}
	MovieSearchKey  = "movie_search:%s:%d:%d" // movie_search:{query}:{limit}:{offset}
	MovieDetailsKey = "movie_details:%s"      // movie_details:{movie_id}

	// User-related cache keys
	UserProfileKey = "user_profile:%s:%d:%d:%s" // user_profile:{user_id}:{limit}:{offset}:{sort}
	UserStatsKey   = "user_stats:%s"            // user_stats:{user_id}
	UserRatingKey  = "user_rating:%s:%s"        // user_rating:{user_id}:{movie_id}

	// Global cache keys
	GlobalAverageKey = "global_average"
	TopMoviesKey     = "top_movies:%d" // top_movies:{limit}

	// Cache TTL constants
	MovieStatsTTL    = 15 * time.Minute
	UserProfileTTL   = 10 * time.Minute
	UserStatsTTL     = 5 * time.Minute
	GlobalAverageTTL = 1 * time.Hour
	MovieSearchTTL   = 20 * time.Minute
)

// Cache key builders
func MovieStatsKeyFunc(movieID string) string {
	return fmt.Sprintf(MovieStatsKey, movieID)
}

func UserProfileKeyFunc(userID string, limit, offset int, sortBy string) string {
	return fmt.Sprintf(UserProfileKey, userID, limit, offset, sortBy)
}

func UserStatsKeyFunc(userID string) string {
	return fmt.Sprintf(UserStatsKey, userID)
}

func UserRatingKeyFunc(userID, movieID string) string {
	return fmt.Sprintf(UserRatingKey, userID, movieID)
}

func MovieSearchKeyFunc(query string, limit, offset int) string {
	return fmt.Sprintf(MovieSearchKey, query, limit, offset)
}
