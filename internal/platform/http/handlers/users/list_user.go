package users

import (
	"encoding/json"
	"net/http"
	"strconv"
	"thermondo/internal/domain/users"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	if email != "" {
		user, err := h.userService.FindUserByEmail(r.Context(), email)
		if err != nil {
			h.logger.Error("Failed to get user by email", "error", err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	// Parse pagination parameters
	//TODO: limit and page should be constants set in the config
	page := 1
	limit := 20 // Default page size

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed // Cap at 100 to prevent abuse
		}
	}

	// Get paginated users
	var users []*users.User
	users, total, err := h.userService.ListUsers(r.Context(), page, limit)
	if err != nil {
		h.logger.Error("Failed to get users", "error", err)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	// Build response with pagination info
	response := map[string]interface{}{
		"users": users,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + limit - 1) / limit, // Ceiling division
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
