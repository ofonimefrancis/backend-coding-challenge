package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/rating"
	"thermondo/internal/domain/users"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ratingRepository struct {
	db *sqlx.DB
}

func NewRatingRepository(db *sqlx.DB) rating.Repository {
	return &ratingRepository{db: db}
}

func (r *ratingRepository) Save(ctx context.Context, rating *rating.Rating) (*rating.Rating, error) {
	query := `
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	var savedRating = rating
	err := r.db.QueryRowContext(
		ctx, query,
		rating.ID, rating.UserID, rating.MovieID, rating.Score,
		rating.Review, rating.CreatedAt, rating.UpdatedAt,
	).Scan(&savedRating.ID, &savedRating.CreatedAt, &savedRating.UpdatedAt)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("user has already rated this movie")
		}
		return nil, fmt.Errorf("failed to save rating: %w", err)
	}

	return savedRating, nil
}

func (r *ratingRepository) GetByID(ctx context.Context, id rating.RatingID) (*rating.Rating, error) {
	query := `
		SELECT id, user_id, movie_id, score, review, created_at, updated_at
		FROM ratings WHERE id = $1`

	rating := &rating.Rating{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rating.ID, &rating.UserID, &rating.MovieID, &rating.Score,
		&rating.Review, &rating.CreatedAt, &rating.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rating with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return rating, nil
}

func (r *ratingRepository) GetByUserAndMovie(ctx context.Context, userID users.UserID, movieID movies.MovieID) (*rating.Rating, error) {
	query := `
		SELECT id, user_id, movie_id, score, review, created_at, updated_at
		FROM ratings WHERE user_id = $1 AND movie_id = $2`

	rating := &rating.Rating{}
	err := r.db.QueryRowContext(ctx, query, userID, movieID).Scan(
		&rating.ID, &rating.UserID, &rating.MovieID, &rating.Score,
		&rating.Review, &rating.CreatedAt, &rating.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rating not found for user %s and movie %s", userID, movieID)
		}
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return rating, nil
}

func (r *ratingRepository) GetByUser(ctx context.Context, userID users.UserID, options ...rating.SearchOption) ([]*rating.Rating, error) {
	opts := rating.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, movie_id, score, review, created_at, updated_at
		FROM ratings 
		WHERE user_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, r.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	return r.queryRatings(ctx, query, userID, opts.Limit, opts.Offset)
}

func (r *ratingRepository) GetByMovie(ctx context.Context, movieID movies.MovieID, options ...rating.SearchOption) ([]*rating.Rating, error) {
	opts := rating.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, movie_id, score, review, created_at, updated_at
		FROM ratings 
		WHERE movie_id = $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, r.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	return r.queryRatings(ctx, query, movieID, opts.Limit, opts.Offset)
}

func (r *ratingRepository) Update(ctx context.Context, rating *rating.Rating) (*rating.Rating, error) {
	query := `
		UPDATE ratings SET
			score = $2, review = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, created_at, updated_at`

	rating.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx, query,
		rating.ID, rating.Score, rating.Review, rating.UpdatedAt,
	).Scan(&rating.ID, &rating.CreatedAt, &rating.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rating with ID %s not found", rating.ID)
		}
		return nil, fmt.Errorf("failed to update rating: %w", err)
	}

	return rating, nil
}

func (r *ratingRepository) Delete(ctx context.Context, id rating.RatingID) error {
	query := `DELETE FROM ratings WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete rating: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("rating with ID %s not found", id)
	}

	return nil
}

func (r *ratingRepository) GetMovieStats(ctx context.Context, movieID movies.MovieID) (*rating.MovieRatingStats, error) {
	query := `
		SELECT 
			ROUND(AVG(score::decimal), 2) as average_score,
			COUNT(*) as total_ratings,
			score,
			COUNT(*) as score_count
		FROM ratings 
		WHERE movie_id = $1
		GROUP BY ROLLUP(score)
		ORDER BY score`

	rows, err := r.db.QueryContext(ctx, query, movieID)
	if err != nil {
		return nil, fmt.Errorf("failed to get movie stats: %w", err)
	}
	defer rows.Close()

	stats := &rating.MovieRatingStats{
		MovieID:    movieID,
		ScoreCount: make(map[int]int64),
	}

	for rows.Next() {
		var avgScore sql.NullFloat64
		var totalRatings int64
		var score sql.NullInt64
		var scoreCount int64

		err := rows.Scan(&avgScore, &totalRatings, &score, &scoreCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie stats: %w", err)
		}

		if !score.Valid {
			stats.AverageScore = avgScore.Float64
			stats.TotalRatings = totalRatings
		} else {
			stats.ScoreCount[int(score.Int64)] = scoreCount
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating movie stats: %w", err)
	}

	return stats, nil
}

func (r *ratingRepository) Exists(ctx context.Context, id rating.RatingID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM ratings WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check rating existence: %w", err)
	}

	return exists, nil
}

func (r *ratingRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM ratings`

	var count int64
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count ratings: %w", err)
	}

	return count, nil
}

func (r *ratingRepository) GetGlobalAverageRating(ctx context.Context) (float64, error) {
	query := `
		SELECT ROUND(AVG(score::decimal), 2) as global_average
		FROM ratings`

	var globalAvg sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query).Scan(&globalAvg)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate global average rating: %w", err)
	}

	// Return default if no ratings exist yet
	if !globalAvg.Valid {
		return 3.0, nil // Default middle value for 1-5 scale
	}

	return globalAvg.Float64, nil
}

// Helper methods
func (r *ratingRepository) getSortColumn(sortBy string) string {
	switch sortBy {
	case "score":
		return "score"
	case "created_at":
		return "created_at"
	case "updated_at":
		return "updated_at"
	default:
		return "created_at"
	}
}

func (r *ratingRepository) queryRatings(ctx context.Context, query string, args ...interface{}) ([]*rating.Rating, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query ratings: %w", err)
	}
	defer rows.Close()

	var ratingsList []*rating.Rating
	for rows.Next() {
		rating := &rating.Rating{}
		err := rows.Scan(
			&rating.ID, &rating.UserID, &rating.MovieID, &rating.Score,
			&rating.Review, &rating.CreatedAt, &rating.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		ratingsList = append(ratingsList, rating)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ratings: %w", err)
	}

	return ratingsList, nil
}
