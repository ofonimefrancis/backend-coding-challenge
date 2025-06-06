package movies

import (
	"encoding/json"
	"net/http"
	"thermondo/internal/domain/movies"
	"time"
)

func (h *Handler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	var req movies.CreateMovieRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.responseWriter.WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	movie, err := h.movieService.CreateMovie(r.Context(), req)
	if err != nil {
		h.logger.Error("[create_movie_handler] Failed to create movie", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := CreateMovieResponse{
		ID:           string(movie.ID),
		Title:        movie.Title,
		Description:  movie.Description,
		ReleaseYear:  movie.ReleaseYear,
		Genre:        movie.Genre,
		Director:     movie.Director,
		DurationMins: movie.DurationMins,
		Rating:       string(movie.Rating),
		Language:     movie.Language,
		Country:      movie.Country,
		Budget:       movie.Budget,
		Revenue:      movie.Revenue,
		IMDbID:       movie.IMDbID,
		PosterURL:    movie.PosterURL,
		CreatedAt:    movie.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    movie.UpdatedAt.Format(time.RFC3339),
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusCreated)

}
