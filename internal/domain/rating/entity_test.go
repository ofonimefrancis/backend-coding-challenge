package rating

import (
	"testing"
	"time"

	"thermondo/internal/domain/movies"
	"thermondo/internal/domain/users"

	"github.com/stretchr/testify/assert"
)

type mockIDGenerator struct{}

func (m *mockIDGenerator) Generate() string { return "mock-id" }

type mockTimeProvider struct{ now time.Time }

func (m *mockTimeProvider) Now() time.Time { return m.now }

func TestNewRating_Success(t *testing.T) {
	idGen := &mockIDGenerator{}
	timeNow := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeProv := &mockTimeProvider{now: timeNow}

	r, err := NewRating(users.UserID("user-1"), movies.MovieID("movie-1"), 5, "  Great!  ", idGen, timeProv)
	assert.NoError(t, err)
	assert.Equal(t, RatingID("mock-id"), r.ID)
	assert.Equal(t, users.UserID("user-1"), r.UserID)
	assert.Equal(t, movies.MovieID("movie-1"), r.MovieID)
	assert.Equal(t, 5, r.Score)
	assert.Equal(t, "Great!", r.Review)
	assert.Equal(t, timeNow, r.CreatedAt)
	assert.Equal(t, timeNow, r.UpdatedAt)
}

func TestNewRating_ValidationErrors(t *testing.T) {
	idGen := &mockIDGenerator{}
	timeProv := &mockTimeProvider{now: time.Now()}

	_, err := NewRating("", "movie-1", 5, "Review", idGen, timeProv)
	assert.ErrorIs(t, err, ErrEmptyUserID)

	_, err = NewRating("user-1", "", 5, "Review", idGen, timeProv)
	assert.ErrorIs(t, err, ErrEmptyMovieID)

	_, err = NewRating("user-1", "movie-1", 0, "Review", idGen, timeProv)
	assert.ErrorIs(t, err, ErrInvalidScore)

	_, err = NewRating("user-1", "movie-1", 6, "Review", idGen, timeProv)
	assert.ErrorIs(t, err, ErrInvalidScore)
}

func TestRating_Validate(t *testing.T) {
	timeNow := time.Now()
	r := &Rating{
		ID:        "id",
		UserID:    "user-1",
		MovieID:   "movie-1",
		Score:     5,
		Review:    "Review",
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	assert.NoError(t, r.Validate())

	r.UserID = ""
	assert.ErrorIs(t, r.Validate(), ErrEmptyUserID)
	r.UserID = "user-1"
	r.MovieID = ""
	assert.ErrorIs(t, r.Validate(), ErrEmptyMovieID)
	r.MovieID = "movie-1"
	r.Score = 0
	assert.ErrorIs(t, r.Validate(), ErrInvalidScore)
}

func TestRating_UpdateScore(t *testing.T) {
	timeNow := time.Now()
	timeProv := &mockTimeProvider{now: timeNow}
	r := &Rating{Score: 3, UpdatedAt: time.Time{}}

	err := r.UpdateScore(5, timeProv)
	assert.NoError(t, err)
	assert.Equal(t, 5, r.Score)
	assert.Equal(t, timeNow, r.UpdatedAt)

	err = r.UpdateScore(0, timeProv)
	assert.ErrorIs(t, err, ErrInvalidScore)
}

func TestRating_UpdateReview(t *testing.T) {
	timeNow := time.Now()
	timeProv := &mockTimeProvider{now: timeNow}
	r := &Rating{Review: "Old", UpdatedAt: time.Time{}}

	err := r.UpdateReview("  New Review  ", timeProv)
	assert.NoError(t, err)
	assert.Equal(t, "New Review", r.Review)
	assert.Equal(t, timeNow, r.UpdatedAt)
}

func TestRating_UpdateScore_MinMax(t *testing.T) {
	timeNow := time.Now()
	timeProv := &mockTimeProvider{now: timeNow}
	r := &Rating{Score: 3, UpdatedAt: time.Time{}}

	err := r.UpdateScore(1, timeProv)
	assert.NoError(t, err)
	assert.Equal(t, 1, r.Score)

	err = r.UpdateScore(5, timeProv)
	assert.NoError(t, err)
	assert.Equal(t, 5, r.Score)
}

func TestRating_UpdateReview_EmptyWhitespace(t *testing.T) {
	timeNow := time.Now()
	timeProv := &mockTimeProvider{now: timeNow}
	r := &Rating{Review: "Old", UpdatedAt: time.Time{}}

	err := r.UpdateReview("", timeProv)
	assert.NoError(t, err)
	assert.Equal(t, "", r.Review)

	err = r.UpdateReview("   ", timeProv)
	assert.NoError(t, err)
	assert.Equal(t, "", r.Review)
}

func TestNewRating_WhitespaceReview(t *testing.T) {
	idGen := &mockIDGenerator{}
	timeProv := &mockTimeProvider{now: time.Now()}

	r, err := NewRating("user-1", "movie-1", 3, "   ", idGen, timeProv)
	assert.NoError(t, err)
	assert.Equal(t, "", r.Review)
}

func TestRating_Validate_NegativeScore(t *testing.T) {
	r := &Rating{UserID: "user-1", MovieID: "movie-1", Score: -1}
	assert.ErrorIs(t, r.Validate(), ErrInvalidScore)
}

func TestRating_UpdateScore_NegativeAndOverMax(t *testing.T) {
	timeNow := time.Now()
	timeProv := &mockTimeProvider{now: timeNow}
	r := &Rating{Score: 3, UpdatedAt: time.Time{}}

	err := r.UpdateScore(-1, timeProv)
	assert.ErrorIs(t, err, ErrInvalidScore)

	err = r.UpdateScore(6, timeProv)
	assert.ErrorIs(t, err, ErrInvalidScore)
}

func TestRating_UpdateReview_NoChange(t *testing.T) {
	timeNow := time.Now()
	timeProv := &mockTimeProvider{now: timeNow}
	r := &Rating{Review: "Same", UpdatedAt: time.Time{}}

	err := r.UpdateReview("Same", timeProv)
	assert.NoError(t, err)
	assert.Equal(t, "Same", r.Review)
}

func TestNewRating_LongReview(t *testing.T) {
	idGen := &mockIDGenerator{}
	timeProv := &mockTimeProvider{now: time.Now()}
	longReview := make([]byte, 1000)
	for i := range longReview {
		longReview[i] = 'a'
	}
	r, err := NewRating("user-1", "movie-1", 4, string(longReview), idGen, timeProv)
	assert.NoError(t, err)
	assert.Equal(t, string(longReview), r.Review)
}
