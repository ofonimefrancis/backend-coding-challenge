package rating

type CreateRatingRequest struct {
	UserID  string `json:"user_id"`
	MovieID string `json:"movie_id"`
	Score   int    `json:"score"`
	Review  string `json:"review,omitempty"`
}

type UpdateRatingRequest struct {
	Score  *int    `json:"score,omitempty"`
	Review *string `json:"review,omitempty"`
}
