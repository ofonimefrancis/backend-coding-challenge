package repository

import (
	"context"
	"fmt"
	"strings"
	"thermondo/internal/domain/movies"
)

func (r *movieRepository) getSortColumn(sortBy string) string {
	switch sortBy {
	case "title":
		return "title"
	case "release_year":
		return "release_year"
	case "created_at":
		return "created_at"
	default:
		return "created_at"
	}
}

// Helper method for querying multiple movies
func (r *movieRepository) queryMovies(ctx context.Context, query string, args ...interface{}) ([]*movies.Movie, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}
	defer rows.Close()

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
		// Trim any padding from the ID
		movie.ID = movies.MovieID(strings.TrimSpace(id))
		moviesList = append(moviesList, movie)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating movies: %w", err)
	}

	return moviesList, nil
}
