package users

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`
	Pagination *Pagination    `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email != "" {
		user, err := h.userService.FindUserByEmail(r.Context(), email)
		if err != nil {
			h.responseWriter.WriteError(w, "User not found", http.StatusNotFound)
			return
		}
		userResponse := UserResponse{
			ID:        strings.TrimSpace(user.ID.String()),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		h.responseWriter.WriteSuccess(w, userResponse, http.StatusOK)
		return
	}

	// Parse pagination parameters
	//TODO: Remove magic numbers
	// pagination defaults can be set in the config
	page := 1
	limit := 20 // Default page size

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get paginated users
	users, total, err := h.userService.ListUsers(r.Context(), page, limit)
	if err != nil {
		h.responseWriter.WriteError(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	var userResponses []UserResponse
	for _, user := range users {
		userResponses = append(userResponses, UserResponse{
			ID:        strings.TrimSpace(user.ID.String()),
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			Role:      string(user.Role),
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	response := ListUsersResponse{
		Users: userResponses,
		Pagination: &Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}
