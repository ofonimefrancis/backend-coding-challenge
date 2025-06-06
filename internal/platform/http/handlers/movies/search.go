package movies

import "net/http"

func (h *Handler) SearchMovies(w http.ResponseWriter, r *http.Request) {
	searchParams, err := h.parseSearchParams(r)
	if err != nil {
		h.logger.Error("[search_movies_handler] Failed to parse search params", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	moviesList, total, err := h.movieService.SearchMovies(r.Context(), *searchParams)
	if err != nil {
		h.logger.Error("[search_movies_handler] Failed to search movies", "error", err)
		h.handleServiceError(w, err)
		return
	}

	response := &SearchMoviesResponse{
		Movies:  h.moviesToResponse(moviesList),
		Total:   total,
		Limit:   searchParams.Limit,
		Offset:  searchParams.Offset,
		HasMore: searchParams.Offset+searchParams.Limit < int(total),
		Query:   searchParams.Query,
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)

}
