package movies

import "errors"

var (
	ErrEmptyTitle      = errors.New("title cannot be empty")
	ErrInvalidYear     = errors.New("release year must be between 1888 and current year + 5")
	ErrEmptyGenre      = errors.New("genre cannot be empty")
	ErrEmptyDirector   = errors.New("director cannot be empty")
	ErrInvalidDuration = errors.New("duration must be greater than 0")
	ErrEmptyLanguage   = errors.New("language cannot be empty")
	ErrEmptyCountry    = errors.New("country cannot be empty")
	ErrInvalidBudget   = errors.New("budget must be non-negative")
	ErrInvalidRevenue  = errors.New("revenue must be non-negative")
)
