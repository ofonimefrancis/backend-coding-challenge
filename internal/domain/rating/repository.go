package rating

import (
	"context"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/users"
)

type SearchOptions struct {
	Limit  int
	Offset int
	SortBy string // "created_at", "score"
	Order  string // "asc", "desc"
}

func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Limit:  20,
		Offset: 0,
		SortBy: "created_at",
		Order:  "desc",
	}
}

type SearchOption func(*SearchOptions)

func WithLimit(limit int) SearchOption {
	return func(opts *SearchOptions) {
		opts.Limit = limit
	}
}

func WithOffset(offset int) SearchOption {
	return func(opts *SearchOptions) {
		opts.Offset = offset
	}
}

func WithSort(sortBy, order string) SearchOption {
	return func(opts *SearchOptions) {
		opts.SortBy = sortBy
		opts.Order = order
	}
}

type Repository interface {
	Save(ctx context.Context, rating *Rating) (*Rating, error)
	GetByID(ctx context.Context, id RatingID) (*Rating, error)
	GetByUserAndMovie(ctx context.Context, userID users.UserID, movieID movies.MovieID) (*Rating, error)
	GetByUser(ctx context.Context, userID users.UserID, options ...SearchOption) ([]*Rating, error)
	GetByMovie(ctx context.Context, movieID movies.MovieID, options ...SearchOption) ([]*Rating, error)
	Update(ctx context.Context, rating *Rating) (*Rating, error)
	Delete(ctx context.Context, id RatingID) error
	GetMovieStats(ctx context.Context, movieID movies.MovieID) (*MovieRatingStats, error)
	Exists(ctx context.Context, id RatingID) (bool, error)
	Count(ctx context.Context) (int64, error)

	GetGlobalAverageRating(ctx context.Context) (float64, error)
}

type MovieRatingStats struct {
	MovieID      movies.MovieID `json:"movie_id"`
	AverageScore float64        `json:"average_score"`
	TotalRatings int64          `json:"total_ratings"`
	ScoreCount   map[int]int64  `json:"score_count"` // Score (1-5) -> Count
}
