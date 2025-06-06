package movies

type CreateMovieResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ReleaseYear  int     `json:"release_year"`
	Genre        string  `json:"genre"`
	Director     string  `json:"director"`
	DurationMins int     `json:"duration_mins"`
	Rating       string  `json:"rating"`
	Language     string  `json:"language"`
	Country      string  `json:"country"`
	Budget       *int64  `json:"budget,omitempty"`
	Revenue      *int64  `json:"revenue,omitempty"`
	IMDbID       *string `json:"imdb_id,omitempty"`
	PosterURL    *string `json:"poster_url,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type MovieResponse struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ReleaseYear  int     `json:"release_year"`
	Genre        string  `json:"genre"`
	Director     string  `json:"director"`
	DurationMins int     `json:"duration_mins"`
	Rating       string  `json:"rating"`
	Language     string  `json:"language"`
	Country      string  `json:"country"`
	Budget       *int64  `json:"budget,omitempty"`
	Revenue      *int64  `json:"revenue,omitempty"`
	IMDbID       *string `json:"imdb_id,omitempty"`
	PosterURL    *string `json:"poster_url,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

type MoviesListResponse struct {
	Movies  []MovieResponse `json:"movies"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
}

type SearchMoviesResponse struct {
	Movies  []MovieResponse `json:"movies"`
	Total   int64           `json:"total"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
	Query   string          `json:"query,omitempty"`
}
