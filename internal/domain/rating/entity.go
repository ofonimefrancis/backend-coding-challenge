package rating

import (
	"errors"
	"strings"
	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/shared"
	"thermondo/internal/domain/users"
	"time"
)

type RatingID string

type Rating struct {
	ID        RatingID       `db:"id"`
	UserID    users.UserID   `db:"user_id"`
	MovieID   movies.MovieID `db:"movie_id"`
	Score     int            `db:"score"`
	Review    string         `db:"review"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

var (
	ErrInvalidScore = errors.New("score must be between 1 and 5")
	ErrEmptyUserID  = errors.New("user ID cannot be empty")
	ErrEmptyMovieID = errors.New("movie ID cannot be empty")
)

func NewRating(
	userID users.UserID,
	movieID movies.MovieID,
	score int,
	review string,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
) (*Rating, error) {
	rating := &Rating{
		ID:        RatingID(idGenerator.Generate()),
		UserID:    userID,
		MovieID:   movieID,
		Score:     score,
		Review:    strings.TrimSpace(review),
		CreatedAt: timeProvider.Now(),
		UpdatedAt: timeProvider.Now(),
	}

	if err := rating.Validate(); err != nil {
		return nil, err
	}

	return rating, nil
}

func (r *Rating) Validate() error {
	if r.UserID == "" {
		return ErrEmptyUserID
	}

	if r.MovieID == "" {
		return ErrEmptyMovieID
	}

	if r.Score < 1 || r.Score > 5 {
		return ErrInvalidScore
	}

	return nil
}

func (r *Rating) UpdateScore(score int, timeProvider shared.TimeProvider) error {
	if score < 1 || score > 5 {
		return ErrInvalidScore
	}

	r.Score = score
	r.UpdatedAt = timeProvider.Now()
	return nil
}

func (r *Rating) UpdateReview(review string, timeProvider shared.TimeProvider) error {
	r.Review = strings.TrimSpace(review)
	r.UpdatedAt = timeProvider.Now()
	return nil
}
