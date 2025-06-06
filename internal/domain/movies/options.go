package movies

import "strings"

type MovieOption func(*Movie)

//=================================== Movie Options ===================================

func WithBudget(budget int64) MovieOption {
	return func(m *Movie) {
		m.Budget = &budget
	}
}

func WithRevenue(revenue int64) MovieOption {
	return func(m *Movie) {
		m.Revenue = &revenue
	}
}

func WithIMDbID(imdbID string) MovieOption {
	return func(m *Movie) {
		imdbID = strings.TrimSpace(imdbID)
		if imdbID != "" {
			m.IMDbID = &imdbID
		}
	}
}

func WithPosterURL(posterURL string) MovieOption {
	return func(m *Movie) {
		posterURL = strings.TrimSpace(posterURL)
		if posterURL != "" {
			m.PosterURL = &posterURL
		}
	}
}
func WithRating(rating string) MovieOption {
	return func(m *Movie) {
		// if the rating is not valid, use the restricted and let admin decide
		if IsValid(rating) {
			m.Rating = Rating(rating)
		}
		m.Rating = RatingRestricted
	}
}

//=================================== Search Options ===================================

type SearchOptions struct {
	Limit  int
	Offset int
	SortBy string // "title", "release_year", "created_at"
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
