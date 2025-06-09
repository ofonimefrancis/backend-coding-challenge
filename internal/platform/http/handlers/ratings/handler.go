package ratings

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"thermondo/internal/domain/rating"
	"thermondo/internal/pkg/http/response"
	ratingService "thermondo/internal/platform/service/rating"
	"time"

	appErrors "thermondo/internal/pkg/errors"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	ratingService  ratingService.Service
	responseWriter *response.Writer
	logger         *slog.Logger
}

func NewHandler(ratingService ratingService.Service, logger *slog.Logger) *Handler {
	return &Handler{
		ratingService:  ratingService,
		responseWriter: response.NewWriter(logger),
		logger:         logger,
	}
}

func (h *Handler) CreateRating(w http.ResponseWriter, r *http.Request) {
	var req ratingService.CreateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.responseWriter.WriteError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	rating, err := h.ratingService.CreateRating(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create rating", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := h.ratingToCreateResponse(rating)
	h.responseWriter.WriteSuccess(w, response, http.StatusCreated)
}

// GetRatingByID handles GET /ratings/{id}
func (h *Handler) GetRatingByID(w http.ResponseWriter, r *http.Request) {
	ratingID := chi.URLParam(r, "id")
	if ratingID == "" {
		h.logger.Error("Rating ID is required", "error", errors.New("rating ID is required"))
		h.responseWriter.WriteError(w, "Rating ID is required", http.StatusBadRequest)
		return
	}

	rating, err := h.ratingService.GetRatingByID(r.Context(), ratingID)
	if err != nil {
		h.logger.Error("Failed to get rating by ID", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := h.ratingToResponse(rating)
	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

// UpdateRating handles PUT /ratings/{id}
func (h *Handler) UpdateRating(w http.ResponseWriter, r *http.Request) {
	ratingID := chi.URLParam(r, "id")
	if ratingID == "" {
		h.logger.Error("Rating ID is required", "error", errors.New("rating ID is required"))
		h.responseWriter.WriteError(w, "Rating ID is required", http.StatusBadRequest)
		return
	}

	var req ratingService.UpdateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", "error", err)
		h.responseWriter.WriteError(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	rating, err := h.ratingService.UpdateRating(r.Context(), ratingID, req)
	if err != nil {
		h.logger.Error("Failed to update rating", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := h.ratingToResponse(rating)
	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

// DeleteRating handles DELETE /ratings/{id}
func (h *Handler) DeleteRating(w http.ResponseWriter, r *http.Request) {
	ratingID := chi.URLParam(r, "id")
	if ratingID == "" {
		h.logger.Error("Rating ID is required", "error", errors.New("rating ID is required"))
		h.responseWriter.WriteError(w, "Rating ID is required", http.StatusBadRequest)
		return
	}

	err := h.ratingService.DeleteRating(r.Context(), ratingID)
	if err != nil {
		h.logger.Error("Failed to delete rating", "error", err)
		h.handleServiceError(w, err)
		return
	}

	type successResponse struct {
		Message string `json:"message"`
	}

	h.responseWriter.WriteSuccess(w, successResponse{Message: "Rating deleted successfully"}, http.StatusOK)
}

// GetMovieRatings handles GET /movies/{movieId}/ratings
func (h *Handler) GetMovieRatings(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "movieId")
	if movieID == "" {
		h.logger.Error("Movie ID is required", "error", errors.New("movie ID is required"))
		h.responseWriter.WriteError(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	params, err := h.parseListParams(r)
	if err != nil {
		h.logger.Error("Failed to parse list params", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	ratingsList, total, err := h.ratingService.GetMovieRatings(
		r.Context(), movieID, params.Limit, params.Offset, params.SortBy, params.Order,
	)
	if err != nil {
		h.logger.Error("Failed to get movie ratings", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := &RatingsListResponse{
		Ratings: h.ratingsToResponse(ratingsList),
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: params.Offset+params.Limit < int(total),
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

// GetMovieStats handles GET /movies/{movieId}/stats
func (h *Handler) GetMovieStats(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "movieId")
	if movieID == "" {
		h.logger.Error("Movie ID is required", "error", errors.New("movie ID is required"))
		h.responseWriter.WriteError(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	stats, err := h.ratingService.GetMovieStats(r.Context(), movieID)
	if err != nil {
		h.logger.Error("Failed to get movie stats", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := h.statsToResponse(stats)
	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

func (h *Handler) GetUserRating(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	movieID := chi.URLParam(r, "movieId")

	if userID == "" {
		h.logger.Error("User ID is required", "error", errors.New("user ID is required"))
		h.responseWriter.WriteError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	if movieID == "" {
		h.logger.Error("Movie ID is required", "error", errors.New("movie ID is required"))
		h.responseWriter.WriteError(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	rating, err := h.ratingService.GetUserRating(r.Context(), userID, movieID)
	if err != nil {
		h.logger.Error("Failed to get user rating", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := h.ratingToResponse(rating)
	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

// Parameter parsing with validation
type listParams struct {
	Limit  int
	Offset int
	SortBy string
	Order  string
}

func (h *Handler) parseListParams(r *http.Request) (*listParams, error) {
	params := &listParams{
		Limit:  20,           // Default limit
		Offset: 0,            // Default offset
		SortBy: "created_at", // Default sort
		Order:  "desc",       // Default order
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 || limit > 100 {
			h.logger.Error("Invalid limit", "error", err)
			return nil, errors.New("limit must be between 1 and 100")
		}
		params.Limit = limit
	}

	// Parse offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			h.logger.Error("Invalid offset", "error", err)
			return nil, errors.New("offset must be non-negative")
		}
		params.Offset = offset
	}

	// Parse sort_by
	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		if !h.isValidSortField(sortBy) {
			h.logger.Error("Invalid sort_by field", "error", errors.New("invalid sort_by field"))
			return nil, errors.New("invalid sort_by field")
		}
		params.SortBy = sortBy
	}

	// Parse order
	if order := r.URL.Query().Get("order"); order != "" {
		if order != "asc" && order != "desc" {
			h.logger.Error("Invalid order", "error", errors.New("order must be 'asc' or 'desc'"))
			return nil, errors.New("order must be 'asc' or 'desc'")
		}
		params.Order = order
	}

	return params, nil
}

func (h *Handler) isValidSortField(field string) bool {
	validFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"score":      true,
	}
	return validFields[field]
}

// Response transformation methods
func (h *Handler) ratingsToResponse(ratingsList []*rating.Rating) []RatingResponse {
	responses := make([]RatingResponse, len(ratingsList))
	for i, rating := range ratingsList {
		responses[i] = h.ratingToResponse(rating)
	}
	return responses
}

func (h *Handler) ratingToResponse(rating *rating.Rating) RatingResponse {
	return RatingResponse{
		ID:        string(rating.ID),
		UserID:    string(rating.UserID),
		MovieID:   string(rating.MovieID),
		Score:     rating.Score,
		Review:    rating.Review,
		CreatedAt: rating.CreatedAt.Format(time.RFC3339),
		UpdatedAt: rating.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) ratingToCreateResponse(rating *rating.Rating) CreateRatingResponse {
	return CreateRatingResponse{
		ID:        string(rating.ID),
		UserID:    string(rating.UserID),
		MovieID:   string(rating.MovieID),
		Score:     rating.Score,
		Review:    rating.Review,
		CreatedAt: rating.CreatedAt.Format(time.RFC3339),
		UpdatedAt: rating.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) statsToResponse(stats *rating.MovieRatingStats) MovieStatsResponse {
	// Convert map[int]int64 to map[string]int64 for JSON
	scoreCount := make(map[string]int64)
	for score, count := range stats.ScoreCount {
		scoreCount[strconv.Itoa(score)] = count
	}

	return MovieStatsResponse{
		MovieID:      string(stats.MovieID),
		AverageScore: stats.AverageScore,
		TotalRatings: stats.TotalRatings,
		ScoreCount:   scoreCount,
	}
}

func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		h.logger.Error("Service error", "error", appErr)
		h.responseWriter.WriteError(w, appErr.Message, appErr.StatusCode)
		return
	}

	// Handle specific error types
	switch {
	case errors.Is(err, rating.ErrInvalidScore):
		h.logger.Error("Invalid score", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, rating.ErrEmptyUserID), errors.Is(err, rating.ErrEmptyMovieID):
		h.logger.Error("Missing required field", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
	default:
		h.logger.Error("Unexpected service error", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {

	router.Route("/ratings", func(r chi.Router) {
		r.Post("/", h.CreateRating)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetRatingByID)
			r.Put("/", h.UpdateRating)
			r.Delete("/", h.DeleteRating)
		})
	})

	// User-centric rating routes
	router.Route("/users/{userId}/ratings", func(r chi.Router) {
		r.Get("/", h.GetUserRating)
		r.Get("/{movieId}", h.GetUserRating)
	})

	// Movie-centric rating routes
	router.Route("/movies/{movieId}", func(r chi.Router) {
		r.Get("/ratings", h.GetMovieRatings)
		r.Get("/stats", h.GetMovieStats)
	})
}
