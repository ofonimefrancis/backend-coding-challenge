package movies

import (
	"strings"
	"thermondo/internal/domain/shared"
	"time"
)

const (
	FirstMovieYear = 1888
	MaxFutureYears = 5 // How far ahead we allow movie announcements
)

type MovieID string

type Movie struct {
	ID           MovieID   `db:"id"`
	Title        string    `db:"title"`
	Description  string    `db:"description"`
	ReleaseYear  int       `db:"release_year"`
	Genre        string    `db:"genre"`
	Director     string    `db:"director"`
	DurationMins int       `db:"duration_mins"`
	Rating       Rating    `db:"rating"` // G, PG, PG13, Restricted, NC17, etc.
	Language     string    `db:"language"`
	Country      string    `db:"country"`
	Budget       *int64    `db:"budget"`     // Optional, int64(avoid floating point issues)
	Revenue      *int64    `db:"revenue"`    // Optional, int64(avoid floating point issues)
	IMDbID       *string   `db:"imdb_id"`    // Optional external reference
	PosterURL    *string   `db:"poster_url"` // Optional poster image
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type CreateMovieRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ReleaseYear  int     `json:"release_year"`
	Genre        string  `json:"genre"`
	Director     string  `json:"director"`
	DurationMins int     `json:"duration_mins"`
	Rating       *string `json:"rating,omitempty"`
	Language     string  `json:"language"`
	Country      string  `json:"country"`
	Budget       *int64  `json:"budget,omitempty"`
	Revenue      *int64  `json:"revenue,omitempty"`
	IMDbID       *string `json:"imdb_id,omitempty"`
	PosterURL    *string `json:"poster_url,omitempty"`
}

type SearchMoviesRequest struct {
	Query    string `json:"query,omitempty"`
	Genre    string `json:"genre,omitempty"`
	Director string `json:"director,omitempty"`
	MinYear  *int   `json:"min_year,omitempty"`
	MaxYear  *int   `json:"max_year,omitempty"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	SortBy   string `json:"sort_by"`
	Order    string `json:"order"`
}

func NewMovie(
	title, description string,
	releaseYear int,
	genre, director string,
	durationMins int,
	language, country string,
	idGenerator shared.IDGenerator,
	timeProvider shared.TimeProvider,
	options ...MovieOption,
) (*Movie, error) {
	movie := &Movie{
		ID:           MovieID(idGenerator.Generate()),
		Title:        strings.TrimSpace(title),
		Description:  strings.TrimSpace(description),
		ReleaseYear:  releaseYear,
		Genre:        strings.TrimSpace(genre),
		Director:     strings.TrimSpace(director),
		DurationMins: durationMins,
		Language:     strings.TrimSpace(language),
		Country:      strings.TrimSpace(country),
		CreatedAt:    timeProvider.Now(),
		UpdatedAt:    timeProvider.Now(),
	}

	for _, option := range options {
		option(movie)
	}

	if err := movie.Validate(); err != nil {
		return nil, err
	}

	return movie, nil
}

func (m *Movie) Validate() error {
	if m.Title == "" {
		return ErrEmptyTitle
	}

	currentYear := time.Now().Year()
	if m.ReleaseYear < FirstMovieYear || m.ReleaseYear > currentYear+MaxFutureYears {
		return ErrInvalidYear
	}

	if m.Genre == "" {
		return ErrEmptyGenre
	}

	if m.Director == "" {
		return ErrEmptyDirector
	}

	if m.DurationMins <= 0 {
		return ErrInvalidDuration
	}

	if m.Language == "" {
		return ErrEmptyLanguage
	}

	if m.Country == "" {
		return ErrEmptyCountry
	}

	if m.Budget != nil && *m.Budget < 0 {
		return ErrInvalidBudget
	}

	if m.Revenue != nil && *m.Revenue < 0 {
		return ErrInvalidRevenue
	}

	return nil
}
