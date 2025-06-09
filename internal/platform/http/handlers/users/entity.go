package users

type UserProfileResponse struct {
	User    UserResponse                  `json:"user"`
	Stats   UserProfileStatsResponse      `json:"stats"`
	Ratings []UserRatingWithMovieResponse `json:"ratings"`
	HasMore bool                          `json:"has_more"`
	Total   int64                         `json:"total"`
}

type UserProfileStatsResponse struct {
	TotalRatings      int64            `json:"total_ratings"`
	AverageScore      float64          `json:"average_score"`
	ScoreDistribution map[string]int64 `json:"score_distribution"` // String keys for JSON
	FavoriteGenre     string           `json:"favorite_genre"`
	GenreBreakdown    map[string]int64 `json:"genre_breakdown"`
}

type UserRatingWithMovieResponse struct {
	RatingID string `json:"rating_id"`
	Score    int    `json:"score"`
	Review   string `json:"review"`
	RatedAt  string `json:"rated_at"`

	// Movie details
	MovieID     string  `json:"movie_id"`
	Title       string  `json:"title"`
	ReleaseYear int     `json:"release_year"`
	Genre       string  `json:"genre"`
	Director    string  `json:"director"`
	PosterURL   *string `json:"poster_url,omitempty"`

	// Comparison data
	MovieAverage  float64 `json:"movie_average"`
	TotalRatings  int64   `json:"total_ratings"`
	UserVsAverage string  `json:"user_vs_average"`
}
