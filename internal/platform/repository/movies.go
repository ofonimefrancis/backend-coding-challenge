package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"thermondo/internal/domain/movies"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type movieRepository struct {
	db *sqlx.DB
}

func NewMovieRepository(db *sqlx.DB) movies.Repository {
	return &movieRepository{db: db}
}

func (m *movieRepository) GetAll(ctx context.Context, options ...movies.SearchOption) ([]*movies.Movie, error) {
	opts := movies.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, release_year, genre, director,
			   duration_mins, rating, language, country, budget, revenue,
			   imdb_id, poster_url, created_at, updated_at
		FROM movies 
		ORDER BY %s %s
		LIMIT $1 OFFSET $2`, m.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	return m.queryMovies(ctx, query, opts.Limit, opts.Offset)
}

func (m *movieRepository) GetByID(ctx context.Context, id movies.MovieID) (*movies.Movie, error) {
	query := `
		SELECT id, title, description, release_year, genre, director,
			   duration_mins, rating, language, country, budget, revenue,
			   imdb_id, poster_url, created_at, updated_at
		FROM movies WHERE id = $1`

	movie := &movies.Movie{}
	var movieID string
	err := m.db.QueryRowContext(ctx, query, id).Scan(
		&movieID, &movie.Title, &movie.Description, &movie.ReleaseYear,
		&movie.Genre, &movie.Director, &movie.DurationMins, &movie.Rating,
		&movie.Language, &movie.Country, &movie.Budget, &movie.Revenue,
		&movie.IMDbID, &movie.PosterURL, &movie.CreatedAt, &movie.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("movie with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get movie: %w", err)
	}

	movie.ID = movies.MovieID(strings.TrimSpace(movieID))
	return movie, nil
}

func (m *movieRepository) Save(ctx context.Context, movie *movies.Movie) (*movies.Movie, error) {
	query := `
		INSERT INTO movies (
			id, title, description, release_year, genre, director,
			duration_mins, rating, language, country, budget, revenue,
			imdb_id, poster_url, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) RETURNING id, created_at, updated_at`

	var savedMovie movies.Movie = *movie
	var savedID string
	err := m.db.QueryRowContext(
		ctx, query,
		movie.ID, movie.Title, movie.Description, movie.ReleaseYear,
		movie.Genre, movie.Director, movie.DurationMins, movie.Rating,
		movie.Language, movie.Country, movie.Budget, movie.Revenue,
		movie.IMDbID, movie.PosterURL, movie.CreatedAt, movie.UpdatedAt,
	).Scan(&savedID, &savedMovie.CreatedAt, &savedMovie.UpdatedAt)

	if err != nil {
		// Check if the error is a unique constraint violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, fmt.Errorf("movie with ID %s already exists", movie.ID)
		}
		return nil, fmt.Errorf("failed to save movie: %w", err)
	}

	savedMovie.ID = movies.MovieID(strings.TrimSpace(savedID))
	return &savedMovie, nil
}

func (m *movieRepository) SearchByTitle(ctx context.Context, title string, options ...movies.SearchOption) ([]*movies.Movie, error) {
	opts := movies.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, release_year, genre, director,
			   duration_mins, rating, language, country, budget, revenue,
			   imdb_id, poster_url, created_at, updated_at
		FROM movies 
		WHERE title ILIKE $1
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, m.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	searchPattern := "%" + strings.ToLower(title) + "%"
	return m.queryMovies(ctx, query, searchPattern, opts.Limit, opts.Offset)
}

func (m *movieRepository) GetByGenre(ctx context.Context, genre string, options ...movies.SearchOption) ([]*movies.Movie, error) {
	opts := movies.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, release_year, genre, director,
			   duration_mins, rating, language, country, budget, revenue,
			   imdb_id, poster_url, created_at, updated_at
		FROM movies 
		WHERE LOWER(genre) = LOWER($1)
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, m.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	return m.queryMovies(ctx, query, genre, opts.Limit, opts.Offset)
}

func (m *movieRepository) GetByDirector(ctx context.Context, director string, options ...movies.SearchOption) ([]*movies.Movie, error) {
	opts := movies.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, release_year, genre, director,
			   duration_mins, rating, language, country, budget, revenue,
			   imdb_id, poster_url, created_at, updated_at
		FROM movies 
		WHERE LOWER(director) = LOWER($1)
		ORDER BY %s %s
		LIMIT $2 OFFSET $3`, m.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	return m.queryMovies(ctx, query, director, opts.Limit, opts.Offset)
}

func (m *movieRepository) GetByYearRange(ctx context.Context, startYear, endYear int, options ...movies.SearchOption) ([]*movies.Movie, error) {
	opts := movies.DefaultSearchOptions()
	for _, option := range options {
		option(&opts)
	}

	query := fmt.Sprintf(`
		SELECT id, title, description, release_year, genre, director,
			   duration_mins, rating, language, country, budget, revenue,
			   imdb_id, poster_url, created_at, updated_at
		FROM movies 
		WHERE release_year BETWEEN $1 AND $2
		ORDER BY %s %s
		LIMIT $3 OFFSET $4`, m.getSortColumn(opts.SortBy), strings.ToUpper(opts.Order))

	return m.queryMovies(ctx, query, startYear, endYear, opts.Limit, opts.Offset)
}

func (m *movieRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM movies`

	var count int64
	err := m.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count movies: %w", err)
	}

	return count, nil
}

// Exists implements movies.Repository.
func (m *movieRepository) Exists(ctx context.Context, id movies.MovieID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM movies WHERE id = $1)`

	var exists bool
	err := m.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check movie existence: %w", err)
	}
	return exists, nil
}

// GetDB returns the database connection
func (m *movieRepository) GetDB() *sqlx.DB {
	return m.db
}

// ScanMovies scans the rows into a slice of movies
func (m *movieRepository) ScanMovies(rows *sql.Rows) ([]*movies.Movie, error) {
	var moviesList []*movies.Movie
	for rows.Next() {
		movie := &movies.Movie{}
		var id string
		err := rows.Scan(
			&id, &movie.Title, &movie.Description, &movie.ReleaseYear,
			&movie.Genre, &movie.Director, &movie.DurationMins, &movie.Rating,
			&movie.Language, &movie.Country, &movie.Budget, &movie.Revenue,
			&movie.IMDbID, &movie.PosterURL, &movie.CreatedAt, &movie.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan movie: %w", err)
		}
		movie.ID = movies.MovieID(strings.TrimSpace(id))
		moviesList = append(moviesList, movie)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating movies: %w", err)
	}

	return moviesList, nil
}

func (r *movieRepository) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return r.db.QueryContext(ctx, query, args...)
}
