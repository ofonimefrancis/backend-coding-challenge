package users

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userService.FindUserByID(r.Context(), id)
	if err != nil {
		h.responseWriter.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO: convert to response DTO
	h.responseWriter.WriteSuccess(w, user, http.StatusOK)
}
