package users

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userService.FindUserByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Service error", "error", err)
		h.responseWriter.WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Role:      string(user.Role),
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}
