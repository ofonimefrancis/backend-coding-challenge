package users

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"thermondo/internal/domain/users"
	appErrors "thermondo/internal/pkg/errors"
	"thermondo/internal/pkg/http/response"
	userService "thermondo/internal/platform/service/user"
	"time"

	"github.com/go-chi/chi/v5"
)

type ProfileHandler struct {
	userService    userService.UserService
	responseWriter *response.Writer
	logger         *slog.Logger
}

func NewProfileHandler(userService userService.UserService, logger *slog.Logger) *ProfileHandler {
	return &ProfileHandler{
		userService:    userService,
		responseWriter: response.NewWriter(logger),
		logger:         logger,
	}
}

// GetUserProfile handles GET /users/{userId}/profile
func (h *ProfileHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if userID == "" {
		h.logger.Error("User ID is required")
		h.responseWriter.WriteError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	limit := h.getIntParam(r, "limit", 20)
	offset := h.getIntParam(r, "offset", 0)
	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}
	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc"
	}

	// Validate parameters
	if limit < 1 || limit > 100 {
		h.logger.Error("Invalid limit", "limit", limit)
		h.responseWriter.WriteError(w, "Limit must be between 1 and 100", http.StatusBadRequest)
		return
	}

	if !h.isValidSortField(sortBy) {
		h.logger.Error("Invalid sort field", "sort_by", sortBy)
		h.responseWriter.WriteError(w, "Invalid sort field", http.StatusBadRequest)
		return
	}

	req := userService.UserProfileRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
		SortBy: sortBy,
		Order:  order,
	}

	// Get user basic info
	user, err := h.userService.FindUserByID(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to find user", "error", err)
		h.handleServiceError(w, err)
		return
	}

	if user == nil {
		h.logger.Error("User not found", "userID", userID)
		h.responseWriter.WriteError(w, "user not found", http.StatusNotFound)
		return
	}

	// Get user profile with ratings
	ratingsWithMovies, stats, err := h.userService.GetUserProfile(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to get user profile", "error", err)
		h.handleServiceError(w, err)
		return
	}

	// Build response
	response := &UserProfileResponse{
		User:    h.userToResponse(user),
		Stats:   h.statsToResponse(stats),
		Ratings: h.ratingsWithMoviesToResponse(ratingsWithMovies),
		HasMore: offset+limit < int(stats.TotalRatings),
		Total:   stats.TotalRatings,
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

// Helper methods
func (h *ProfileHandler) getIntParam(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func (h *ProfileHandler) isValidSortField(field string) bool {
	validFields := map[string]bool{
		"created_at": true,
		"score":      true,
		"title":      true,
	}
	return validFields[field]
}

func (h *ProfileHandler) userToResponse(user *users.User) UserResponse {
	return UserResponse{
		ID:        string(user.ID),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *ProfileHandler) statsToResponse(stats *userService.UserProfileStats) UserProfileStatsResponse {
	// Convert score distribution to string keys
	scoreDistribution := make(map[string]int64)
	for score, count := range stats.ScoreDistribution {
		scoreDistribution[strconv.Itoa(score)] = count
	}

	return UserProfileStatsResponse{
		TotalRatings:      stats.TotalRatings,
		AverageScore:      stats.AverageScore,
		ScoreDistribution: scoreDistribution,
		FavoriteGenre:     stats.FavoriteGenre,
		GenreBreakdown:    stats.GenreBreakdown,
	}
}

func (h *ProfileHandler) ratingsWithMoviesToResponse(ratingsWithMovies []*userService.UserRatingWithMovie) []UserRatingWithMovieResponse {
	responses := make([]UserRatingWithMovieResponse, len(ratingsWithMovies))
	for i, rwm := range ratingsWithMovies {
		responses[i] = UserRatingWithMovieResponse{
			RatingID:      string(rwm.Rating.ID),
			Score:         rwm.Rating.Score,
			Review:        rwm.Rating.Review,
			RatedAt:       rwm.Rating.CreatedAt.Format(time.RFC3339),
			MovieID:       string(rwm.Movie.ID),
			Title:         rwm.Movie.Title,
			ReleaseYear:   rwm.Movie.ReleaseYear,
			Genre:         rwm.Movie.Genre,
			Director:      rwm.Movie.Director,
			PosterURL:     rwm.Movie.PosterURL,
			MovieAverage:  rwm.MovieAverage,
			TotalRatings:  rwm.TotalRatings,
			UserVsAverage: rwm.UserVsAvg,
		}
	}
	return responses
}

func (h *ProfileHandler) handleServiceError(w http.ResponseWriter, err error) {
	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		h.logger.Error("Service error", "error", appErr)
		h.responseWriter.WriteError(w, appErr.Message, appErr.StatusCode)
		return
	}

	// Handle specific error types
	switch {
	case errors.Is(err, users.ErrInvalidEmail):
		h.logger.Error("Invalid email", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, users.ErrEmptyEmail), errors.Is(err, users.ErrEmptyFirstName), errors.Is(err, users.ErrEmptyLastName):
		h.logger.Error("Missing required field", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
	default:
		h.logger.Error("Unexpected service error", "error", err)
		h.responseWriter.WriteError(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *ProfileHandler) RegisterRoutes(r chi.Router) {
	r.Get("/user/{userId}/profile", h.GetUserProfile)
}
