package movies

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetAllMovies(w http.ResponseWriter, r *http.Request) {
	params, err := h.parseListParams(r)
	if err != nil {
		h.logger.Error("[get_all_movies_handler] Failed to parse list params", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	moviesList, total, err := h.movieService.GetAllMovies(
		r.Context(),
		params.Limit,
		params.Offset,
		params.SortBy,
		params.Order,
	)
	if err != nil {
		h.logger.Error("[get_all_movies_handler] Failed to get all movies", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := &MoviesListResponse{
		Movies:  h.moviesToResponse(moviesList),
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: params.Offset+params.Limit < int(total),
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}

func (h *Handler) GetMovie(w http.ResponseWriter, r *http.Request) {
	movieID := chi.URLParam(r, "id")
	if movieID == "" {
		h.logger.Error("[get_movie_handler] Movie ID is required")
		h.responseWriter.WriteError(w, "Movie ID is required", http.StatusBadRequest)
		return
	}

	movie, err := h.movieService.GetMovieByID(r.Context(), movieID)
	if err != nil {
		h.logger.Error("[get_movie_handler] Failed to get movie", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := h.movieToResponse(movie)
	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}
