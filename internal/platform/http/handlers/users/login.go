package users

import (
	"encoding/json"
	"net/http"
	"thermondo/internal/pkg/password"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("[login_handler] Invalid JSON", "error", err)
		h.responseWriter.WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find user by email
	user, err := h.userService.FindUserByEmail(r.Context(), req.Email)
	if err != nil {
		h.logger.Error("[login_handler] Failed to find user", "error", err)
		h.responseWriter.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := password.VerifyPassword(req.Password, user.Password); err != nil {
		h.logger.Error("[login_handler] Invalid password", "error", err)
		h.responseWriter.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    string(user.Role),
		"exp":     expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.Error("[login_handler] Failed to generate token", "error", err)
		h.responseWriter.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := loginResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt.Unix(),
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusOK)
}
