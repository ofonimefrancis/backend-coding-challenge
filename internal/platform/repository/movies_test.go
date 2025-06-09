package repository

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"thermondo/internal/domain/movies"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMovieTestDB(t *testing.T) *sqlx.DB {
	// Replace with your actual test database connection string
	connStr := "postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable"
	db, err := sqlx.Connect("postgres", connStr)
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

func TestMovieRepository_Save(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	rating := string(movies.RatingPG13)
	movie, err := movies.NewMovie(
		"Test Movie",
		"Test Description",
		2024,
		"Action",
		"Test Director",
		120,
		"English",
		"USA",
		&mockIDGenerator{id: "test-id-save"},
		&mockTimeProvider{now: time.Now()},
		movies.WithRating(rating),
	)
	require.NoError(t, err)

	savedMovie, err := repo.Save(context.Background(), movie)
	require.NoError(t, err)
	// Compare only the relevant fields, ignoring timestamps
	assert.Equal(t, movie.ID, savedMovie.ID)
	assert.Equal(t, movie.Title, savedMovie.Title)
	assert.Equal(t, movie.Description, savedMovie.Description)
	assert.Equal(t, movie.ReleaseYear, savedMovie.ReleaseYear)
	assert.Equal(t, movie.Genre, savedMovie.Genre)
	assert.Equal(t, movie.Director, savedMovie.Director)
	assert.Equal(t, movie.DurationMins, savedMovie.DurationMins)
	assert.Equal(t, movie.Rating, savedMovie.Rating)
	assert.Equal(t, movie.Language, savedMovie.Language)
	assert.Equal(t, movie.Country, savedMovie.Country)
}

// Add mock implementations for IDGenerator and TimeProvider
type mockIDGenerator struct {
	id string
}

func (m *mockIDGenerator) Generate() string {
	return m.id
}

type mockTimeProvider struct {
	now time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	return m.now
}

func TestMovieRepository_GetByID(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	expectedMovie := &movies.Movie{
		ID:           "test-id-get",
		Title:        "Test Movie",
		Description:  "Test Description",
		ReleaseYear:  2024,
		Genre:        "Action",
		Director:     "Test Director",
		DurationMins: 120,
		Rating:       movies.RatingPG13,
		Language:     "English",
		Country:      "USA",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert the movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, expectedMovie.ID, expectedMovie.Title, expectedMovie.Description, expectedMovie.ReleaseYear, expectedMovie.Genre, expectedMovie.Director, expectedMovie.DurationMins, expectedMovie.Rating, expectedMovie.Language, expectedMovie.Country, expectedMovie.CreatedAt, expectedMovie.UpdatedAt)
	require.NoError(t, err)

	movie, err := repo.GetByID(context.Background(), "test-id-get")
	require.NoError(t, err)
	// Compare only the relevant fields, ignoring timestamps
	assert.Equal(t, expectedMovie.ID, movie.ID)
	assert.Equal(t, expectedMovie.Title, movie.Title)
	assert.Equal(t, expectedMovie.Description, movie.Description)
	assert.Equal(t, expectedMovie.ReleaseYear, movie.ReleaseYear)
	assert.Equal(t, expectedMovie.Genre, movie.Genre)
	assert.Equal(t, expectedMovie.Director, movie.Director)
	assert.Equal(t, expectedMovie.DurationMins, movie.DurationMins)
	assert.Equal(t, expectedMovie.Rating, movie.Rating)
	assert.Equal(t, expectedMovie.Language, movie.Language)
	assert.Equal(t, expectedMovie.Country, movie.Country)
}

func TestMovieRepository_GetAll(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	expectedMovies := []*movies.Movie{
		{
			ID:           "test-id-all-1",
			Title:        "Test Movie 1",
			Description:  "Test Description 1",
			ReleaseYear:  2024,
			Genre:        "Action",
			Director:     "Test Director 1",
			DurationMins: 120,
			Rating:       movies.RatingPG13,
			Language:     "English",
			Country:      "USA",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "test-id-all-2",
			Title:        "Test Movie 2",
			Description:  "Test Description 2",
			ReleaseYear:  2024,
			Genre:        "Comedy",
			Director:     "Test Director 2",
			DurationMins: 90,
			Rating:       movies.RatingPG,
			Language:     "English",
			Country:      "UK",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Insert the movies into the database
	for _, movie := range expectedMovies {
		_, err := db.Exec(`
			INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`, movie.ID, movie.Title, movie.Description, movie.ReleaseYear, movie.Genre, movie.Director, movie.DurationMins, movie.Rating, movie.Language, movie.Country, movie.CreatedAt, movie.UpdatedAt)
		require.NoError(t, err)
	}

	movies, err := repo.GetAll(context.Background(), movies.WithLimit(10), movies.WithOffset(0))
	require.NoError(t, err)
	// Sort both slices by ID for comparison
	sort.Slice(movies, func(i, j int) bool { return movies[i].ID < movies[j].ID })
	sort.Slice(expectedMovies, func(i, j int) bool { return expectedMovies[i].ID < expectedMovies[j].ID })
	// Compare only the relevant fields, ignoring timestamps
	for i, movie := range movies {
		assert.Equal(t, expectedMovies[i].ID, movie.ID)
		assert.Equal(t, expectedMovies[i].Title, movie.Title)
		assert.Equal(t, expectedMovies[i].Description, movie.Description)
		assert.Equal(t, expectedMovies[i].ReleaseYear, movie.ReleaseYear)
		assert.Equal(t, expectedMovies[i].Genre, movie.Genre)
		assert.Equal(t, expectedMovies[i].Director, movie.Director)
		assert.Equal(t, expectedMovies[i].DurationMins, movie.DurationMins)
		assert.Equal(t, expectedMovies[i].Rating, movie.Rating)
		assert.Equal(t, expectedMovies[i].Language, movie.Language)
		assert.Equal(t, expectedMovies[i].Country, movie.Country)
	}
}

func TestMovieRepository_SearchByTitle(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	expectedMovies := []*movies.Movie{
		{
			ID:           "test-id-search",
			Title:        "Test Movie 1",
			Description:  "Test Description 1",
			ReleaseYear:  2024,
			Genre:        "Action",
			Director:     "Test Director 1",
			DurationMins: 120,
			Rating:       movies.RatingPG13,
			Language:     "English",
			Country:      "USA",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Insert the movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, expectedMovies[0].ID, expectedMovies[0].Title, expectedMovies[0].Description, expectedMovies[0].ReleaseYear, expectedMovies[0].Genre, expectedMovies[0].Director, expectedMovies[0].DurationMins, expectedMovies[0].Rating, expectedMovies[0].Language, expectedMovies[0].Country, expectedMovies[0].CreatedAt, expectedMovies[0].UpdatedAt)
	require.NoError(t, err)

	movies, err := repo.SearchByTitle(context.Background(), "test", movies.WithLimit(10), movies.WithOffset(0))
	require.NoError(t, err)
	// Compare only the relevant fields, ignoring timestamps
	for i, movie := range movies {
		assert.Equal(t, expectedMovies[i].ID, movie.ID)
		assert.Equal(t, expectedMovies[i].Title, movie.Title)
		assert.Equal(t, expectedMovies[i].Description, movie.Description)
		assert.Equal(t, expectedMovies[i].ReleaseYear, movie.ReleaseYear)
		assert.Equal(t, expectedMovies[i].Genre, movie.Genre)
		assert.Equal(t, expectedMovies[i].Director, movie.Director)
		assert.Equal(t, expectedMovies[i].DurationMins, movie.DurationMins)
		assert.Equal(t, expectedMovies[i].Rating, movie.Rating)
		assert.Equal(t, expectedMovies[i].Language, movie.Language)
		assert.Equal(t, expectedMovies[i].Country, movie.Country)
	}
}

func TestMovieRepository_GetByGenre(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	expectedMovies := []*movies.Movie{
		{
			ID:           "test-id-genre",
			Title:        "Test Movie 1",
			Description:  "Test Description 1",
			ReleaseYear:  2024,
			Genre:        "Action",
			Director:     "Test Director 1",
			DurationMins: 120,
			Rating:       movies.RatingPG13,
			Language:     "English",
			Country:      "USA",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Insert the movie into the database with unpadded ID
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, expectedMovies[0].ID, expectedMovies[0].Title, expectedMovies[0].Description, expectedMovies[0].ReleaseYear, expectedMovies[0].Genre, expectedMovies[0].Director, expectedMovies[0].DurationMins, expectedMovies[0].Rating, expectedMovies[0].Language, expectedMovies[0].Country, expectedMovies[0].CreatedAt, expectedMovies[0].UpdatedAt)
	require.NoError(t, err)

	movies, err := repo.GetByGenre(context.Background(), "action", movies.WithLimit(10), movies.WithOffset(0))
	require.NoError(t, err)
	// Compare only the relevant fields, ignoring timestamps
	for i, movie := range movies {
		assert.Equal(t, expectedMovies[i].ID, movie.ID)
		assert.Equal(t, expectedMovies[i].Title, movie.Title)
		assert.Equal(t, expectedMovies[i].Description, movie.Description)
		assert.Equal(t, expectedMovies[i].ReleaseYear, movie.ReleaseYear)
		assert.Equal(t, expectedMovies[i].Genre, movie.Genre)
		assert.Equal(t, expectedMovies[i].Director, movie.Director)
		assert.Equal(t, expectedMovies[i].DurationMins, movie.DurationMins)
		assert.Equal(t, expectedMovies[i].Rating, movie.Rating)
		assert.Equal(t, expectedMovies[i].Language, movie.Language)
		assert.Equal(t, expectedMovies[i].Country, movie.Country)
	}
}

func TestMovieRepository_GetByDirector(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	expectedMovies := []*movies.Movie{
		{
			ID:           "test-id-director",
			Title:        "Test Movie 1",
			Description:  "Test Description 1",
			ReleaseYear:  2024,
			Genre:        "Action",
			Director:     "Test Director",
			DurationMins: 120,
			Rating:       movies.RatingPG13,
			Language:     "English",
			Country:      "USA",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Pad the ID to 26 characters
	paddedID := fmt.Sprintf("%-26s", expectedMovies[0].ID)

	// Insert the movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, paddedID, expectedMovies[0].Title, expectedMovies[0].Description, expectedMovies[0].ReleaseYear, expectedMovies[0].Genre, expectedMovies[0].Director, expectedMovies[0].DurationMins, expectedMovies[0].Rating, expectedMovies[0].Language, expectedMovies[0].Country, expectedMovies[0].CreatedAt, expectedMovies[0].UpdatedAt)
	require.NoError(t, err)

	movies, err := repo.GetByDirector(context.Background(), "test director", movies.WithLimit(10), movies.WithOffset(0))
	require.NoError(t, err)
	assert.Equal(t, expectedMovies[0].ID, movies[0].ID)
	assert.Equal(t, expectedMovies[0].Title, movies[0].Title)
	assert.Equal(t, expectedMovies[0].Description, movies[0].Description)
	assert.Equal(t, expectedMovies[0].ReleaseYear, movies[0].ReleaseYear)
	assert.Equal(t, expectedMovies[0].Genre, movies[0].Genre)
	assert.Equal(t, expectedMovies[0].Director, movies[0].Director)
	assert.Equal(t, expectedMovies[0].DurationMins, movies[0].DurationMins)
	assert.Equal(t, expectedMovies[0].Rating, movies[0].Rating)
	assert.Equal(t, expectedMovies[0].Language, movies[0].Language)
	assert.Equal(t, expectedMovies[0].Country, movies[0].Country)
}

func TestMovieRepository_GetByYearRange(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	expectedMovies := []*movies.Movie{
		{
			ID:           "test-id-year",
			Title:        "Test Movie 1",
			Description:  "Test Description 1",
			ReleaseYear:  2024,
			Genre:        "Action",
			Director:     "Test Director",
			DurationMins: 120,
			Rating:       movies.RatingPG13,
			Language:     "English",
			Country:      "USA",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
	}

	// Pad the ID to 26 characters
	paddedID := fmt.Sprintf("%-26s", expectedMovies[0].ID)

	// Insert the movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, paddedID, expectedMovies[0].Title, expectedMovies[0].Description, expectedMovies[0].ReleaseYear, expectedMovies[0].Genre, expectedMovies[0].Director, expectedMovies[0].DurationMins, expectedMovies[0].Rating, expectedMovies[0].Language, expectedMovies[0].Country, expectedMovies[0].CreatedAt, expectedMovies[0].UpdatedAt)
	require.NoError(t, err)

	movies, err := repo.GetByYearRange(context.Background(), 2020, 2025, movies.WithLimit(10), movies.WithOffset(0))
	require.NoError(t, err)
	fmt.Printf("[GetByYearRange] expected ID: '%s' (len=%d), actual ID: '%s' (len=%d)\n", expectedMovies[0].ID, len(expectedMovies[0].ID), movies[0].ID, len(movies[0].ID))
	assert.Equal(t, expectedMovies[0].ID, movies[0].ID)
	assert.Equal(t, expectedMovies[0].Title, movies[0].Title)
	assert.Equal(t, expectedMovies[0].Description, movies[0].Description)
	assert.Equal(t, expectedMovies[0].ReleaseYear, movies[0].ReleaseYear)
	assert.Equal(t, expectedMovies[0].Genre, movies[0].Genre)
	assert.Equal(t, expectedMovies[0].Director, movies[0].Director)
	assert.Equal(t, expectedMovies[0].DurationMins, movies[0].DurationMins)
	assert.Equal(t, expectedMovies[0].Rating, movies[0].Rating)
	assert.Equal(t, expectedMovies[0].Language, movies[0].Language)
	assert.Equal(t, expectedMovies[0].Country, movies[0].Country)
}

func TestMovieRepository_Count(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	// Insert a movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "test-id-count", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, movies.RatingPG13, "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	count, err := repo.Count(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMovieRepository_Exists(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	// Insert a movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, "test-id-exists", "Test Movie", "Test Description", 2024, "Action", "Test Director", 120, movies.RatingPG13, "English", "USA", time.Now(), time.Now())
	require.NoError(t, err)

	exists, err := repo.Exists(context.Background(), "test-id-exists")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMovieRepository_GetByID_NotFound(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	movie, err := repo.GetByID(context.Background(), "non-existent-id")
	require.Error(t, err)
	assert.Nil(t, movie)
	assert.Contains(t, err.Error(), "not found")
}

func TestMovieRepository_Save_Error(t *testing.T) {
	db := setupMovieTestDB(t)
	defer db.Close()

	repo := NewMovieRepository(db)

	movie := &movies.Movie{
		ID:           "test-id-error",
		Title:        "Test Movie",
		Description:  "Test Description",
		ReleaseYear:  2024,
		Genre:        "Action",
		Director:     "Test Director",
		DurationMins: 120,
		Rating:       movies.RatingPG13,
		Language:     "English",
		Country:      "USA",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Insert the movie into the database
	_, err := db.Exec(`
		INSERT INTO movies (id, title, description, release_year, genre, director, duration_mins, rating, language, country, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, movie.ID, movie.Title, movie.Description, movie.ReleaseYear, movie.Genre, movie.Director, movie.DurationMins, movie.Rating, movie.Language, movie.Country, movie.CreatedAt, movie.UpdatedAt)
	require.NoError(t, err)

	// Try to save the same movie again
	_, err = repo.Save(context.Background(), movie)
	require.Error(t, err)
}
