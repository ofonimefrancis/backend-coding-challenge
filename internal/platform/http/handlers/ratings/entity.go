package ratings

type CreateRatingResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	MovieID   string `json:"movie_id"`
	Score     int    `json:"score"`
	Review    string `json:"review"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type RatingResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	MovieID   string `json:"movie_id"`
	Score     int    `json:"score"`
	Review    string `json:"review"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type RatingsListResponse struct {
	Ratings []RatingResponse `json:"ratings"`
	Total   int64            `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	HasMore bool             `json:"has_more"`
}

type MovieStatsResponse struct {
	MovieID      string           `json:"movie_id"`
	AverageScore float64          `json:"average_score"`
	TotalRatings int64            `json:"total_ratings"`
	ScoreCount   map[string]int64 `json:"score_count"` // String keys for JSON
}
