package repository

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"

	"thermondo/internal/domain/rating"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRatingTestDB creates a test database connection
func setupRatingTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("postgres", "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable")
	require.NoError(t, err)

	// Clean up the database before each test
	_, err = db.Exec(`
		TRUNCATE TABLE ratings CASCADE;
		TRUNCATE TABLE movies CASCADE;
		TRUNCATE TABLE users CASCADE;
	`)
	require.NoError(t, err)

	return db
}

func TestRatingRepository_Save(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create a user and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-save", "test-save@example.com", "password123", "Test", "User", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-save", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	rating := &rating.Rating{
		ID:        "test-id-save",
		UserID:    "user-id-save",
		MovieID:   "movie-id-save",
		Score:     5,
		Review:    "Great movie!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	savedRating, err := repo.Save(context.Background(), rating)
	require.NoError(t, err)
	assert.Equal(t, rating, savedRating)
}

func TestRatingRepository_GetByID(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create a user and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-get", "test-get@example.com", "password123", "Test", "User", "user", true, time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-get", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	expectedCreatedAt := time.Now().UTC().Truncate(time.Second)
	expectedUpdatedAt := expectedCreatedAt

	expectedRating := &rating.Rating{
		ID:        "test-id-get",
		UserID:    "user-id-get",
		MovieID:   "movie-id-get",
		Score:     5,
		Review:    "Great movie!",
		CreatedAt: expectedCreatedAt,
		UpdatedAt: expectedUpdatedAt,
	}

	// Insert the rating into the database
	_, err = db.Exec(`
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, expectedRating.ID, expectedRating.UserID, expectedRating.MovieID, expectedRating.Score, expectedRating.Review, expectedRating.CreatedAt, expectedRating.UpdatedAt)
	require.NoError(t, err)

	result, err := repo.GetByID(context.Background(), "test-id-get")
	require.NoError(t, err)
	assert.Equal(t, expectedRating.ID, result.ID)
	assert.Equal(t, expectedRating.UserID, result.UserID)
	assert.Equal(t, expectedRating.MovieID, result.MovieID)
	assert.Equal(t, expectedRating.Score, result.Score)
	assert.Equal(t, expectedRating.Review, result.Review)
	assert.True(t, expectedRating.CreatedAt.Equal(result.CreatedAt), "CreatedAt times should be equal")
	assert.True(t, expectedRating.UpdatedAt.Equal(result.UpdatedAt), "UpdatedAt times should be equal")
}

func TestRatingRepository_GetByUserAndMovie(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// Create test user and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-get-by-user-movie", "test@example.com", "password123", "Test", "User", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-get-by-user-movie", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, "test-id-get-by-user-movie", "user-id-get-by-user-movie", "movie-id-get-by-user-movie", 5, "Great movie!", time.Now(), time.Now())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	rating, err := repo.GetByUserAndMovie(context.Background(), "user-id-get-by-user-movie", "movie-id-get-by-user-movie")
	require.NoError(t, err)
	require.NotNil(t, rating)

	// Trim any trailing spaces from IDs before comparison
	assert.Equal(t, strings.TrimSpace("test-id-get-by-user-movie"), strings.TrimSpace(string(rating.ID)))
	assert.Equal(t, strings.TrimSpace("user-id-get-by-user-movie"), strings.TrimSpace(string(rating.UserID)))
	assert.Equal(t, "movie-id-get-by-user-movie", string(rating.MovieID))
	assert.Equal(t, 5, rating.Score)
	assert.Equal(t, "Great movie!", rating.Review)
}

func TestRatingRepository_GetByMovie(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create users and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-get-by-movie-1", "test-get-by-movie-1@example.com", "password123", "Test", "User1", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-get-by-movie-2", "test-get-by-movie-2@example.com", "password123", "Test", "User2", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-get-by-movie", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	expectedRatings := []*rating.Rating{
		{
			ID:        "test-id-get-by-movie-1",
			UserID:    "user-id-get-by-movie-1",
			MovieID:   "movie-id-get-by-movie",
			Score:     5,
			Review:    "Great movie!",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "test-id-get-by-movie-2",
			UserID:    "user-id-get-by-movie-2",
			MovieID:   "movie-id-get-by-movie",
			Score:     4,
			Review:    "Good movie",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Insert the ratings into the database
	for _, r := range expectedRatings {
		_, err := db.Exec(`
			INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, r.ID, r.UserID, r.MovieID, r.Score, r.Review, r.CreatedAt, r.UpdatedAt)
		require.NoError(t, err)
	}

	ratings, err := repo.GetByMovie(context.Background(), "movie-id-get-by-movie")
	require.NoError(t, err)

	// Sort both slices by ID for comparison
	sort.Slice(ratings, func(i, j int) bool { return ratings[i].ID < ratings[j].ID })
	sort.Slice(expectedRatings, func(i, j int) bool { return expectedRatings[i].ID < expectedRatings[j].ID })

	// Compare ratings ignoring CreatedAt and UpdatedAt
	for i := range ratings {
		assert.Equal(t, expectedRatings[i].ID, ratings[i].ID)
		assert.Equal(t, expectedRatings[i].UserID, ratings[i].UserID)
		assert.Equal(t, expectedRatings[i].MovieID, ratings[i].MovieID)
		assert.Equal(t, expectedRatings[i].Score, ratings[i].Score)
		assert.Equal(t, expectedRatings[i].Review, ratings[i].Review)
	}
}

func TestRatingRepository_GetByUser(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create user and movies to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-get-by-user", "test-get-by-user@example.com", "password123", "Test", "User", "user", true, time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-get-by-user-1", "Test Movie 1", "Test Description 1", 2024, "Action", "Test Director 1", 120, "PG-13", "English", "USA", time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-get-by-user-2", "Test Movie 2", "Test Description 2", 2024, "Comedy", "Test Director 2", 90, "PG", "English", "UK", time.Now().UTC(), time.Now().UTC())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	expectedRatings := []*rating.Rating{
		{
			ID:        "test-id-get-by-user-1",
			UserID:    "user-id-get-by-user",
			MovieID:   "movie-id-get-by-user-1",
			Score:     5,
			Review:    "Great movie!",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
		{
			ID:        "test-id-get-by-user-2",
			UserID:    "user-id-get-by-user",
			MovieID:   "movie-id-get-by-user-2",
			Score:     4,
			Review:    "Good movie",
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}

	// Insert the ratings into the database
	for _, r := range expectedRatings {
		_, err := db.Exec(`
			INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, r.ID, r.UserID, r.MovieID, r.Score, r.Review, r.CreatedAt, r.UpdatedAt)
		require.NoError(t, err)
	}

	ratings, err := repo.GetByUser(context.Background(), "user-id-get-by-user")
	require.NoError(t, err)

	// Sort both slices by ID for comparison
	sort.Slice(ratings, func(i, j int) bool { return ratings[i].ID < ratings[j].ID })
	sort.Slice(expectedRatings, func(i, j int) bool { return expectedRatings[i].ID < expectedRatings[j].ID })

	assert.Equal(t, expectedRatings, ratings)
}

func TestRatingRepository_Count(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create a user and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-count", "test-count@example.com", "password123", "Test", "User", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-count", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	// Insert a rating into the database
	_, err = db.Exec(`
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, "test-id-count", "user-id-count", "movie-id-count", 5, "Great movie!", time.Now(), time.Now())
	require.NoError(t, err)

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRatingRepository_Exists(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create a user and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-exists", "test-exists@example.com", "password123", "Test", "User", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-exists", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	// Insert a rating into the database
	_, err = db.Exec(`
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, "test-id-exists", "user-id-exists", "movie-id-exists", 5, "Great movie!", time.Now(), time.Now())
	require.NoError(t, err)

	exists, err := repo.Exists(context.Background(), "test-id-exists")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRatingRepository_GetByID_NotFound(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	repo := NewRatingRepository(db)

	rating, err := repo.GetByID(context.Background(), "non-existent")
	require.Error(t, err)
	assert.Nil(t, rating)
	assert.Contains(t, err.Error(), "not found")
}

func TestRatingRepository_GetByUserAndMovie_NotFound(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	repo := NewRatingRepository(db)

	rating, err := repo.GetByUserAndMovie(context.Background(), "non-existent-user", "non-existent-movie")
	require.Error(t, err)
	assert.Nil(t, rating)
	assert.Contains(t, err.Error(), "not found")
}

func TestRatingRepository_Save_Error(t *testing.T) {
	db := setupRatingTestDB(t)
	defer db.Close()

	// First create a user and movie to satisfy foreign key constraints
	_, err := db.Exec(`
		INSERT INTO users (id, email, password, first_name, last_name, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, "user-id-error", "test-error@example.com", "password123", "Test", "User", "user", true, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "movie-id-error", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, "PG-13", "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	repo := NewRatingRepository(db)

	rating := &rating.Rating{
		ID:        "test-id-error",
		UserID:    "user-id-error",
		MovieID:   "movie-id-error",
		Score:     5,
		Review:    "Great movie!",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert the rating into the database
	_, err = db.Exec(`
		INSERT INTO ratings (id, user_id, movie_id, score, review, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, rating.ID, rating.UserID, rating.MovieID, rating.Score, rating.Review, rating.CreatedAt, rating.UpdatedAt)
	require.NoError(t, err)

	// Try to save the same rating again
	_, err = repo.Save(context.Background(), rating)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user has already rated this movie")
}
