package users

import (
	"encoding/json"
	"errors"
	"net/http"
	domainUser "thermondo/internal/domain/users"
	appErrors "thermondo/internal/pkg/errors"
	"time"
)

var (
	ErrInvalidRequest    = errors.New("invalid request")
	ErrFirstNameRequired = errors.New("first name is required")
	ErrLastNameRequired  = errors.New("last name is required")
	ErrEmailRequired     = errors.New("email is required")
	ErrPasswordRequired  = errors.New("password is required")
	ErrRoleRequired      = errors.New("role is required")
)

type createUserResponse struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  *bool     `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req domainUser.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("[create_user_handler] Invalid JSON", "error", err)
		h.responseWriter.WriteError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	savedUser, err := h.userService.CreateUser(r.Context(), req)
	if err != nil {
		h.logger.Error("[create_user_handler] Failed to create user", "error", err)
		h.handleCreateUserServiceError(w, err)
		return
	}

	response := createUserResponse{
		ID:        savedUser.ID.String(),
		FirstName: savedUser.FirstName,
		LastName:  savedUser.LastName,
		Email:     savedUser.Email,
		Role:      string(savedUser.Role),
		IsActive:  &savedUser.IsActive,
		CreatedAt: savedUser.CreatedAt,
	}

	h.responseWriter.WriteSuccess(w, response, http.StatusCreated)
}

func (h *Handler) handleCreateUserServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, domainUser.ErrUserAlreadyExists) {
		h.responseWriter.WriteError(w, "user already exists", http.StatusConflict)
		return
	}

	var appErr *appErrors.AppError
	if errors.As(err, &appErr) {
		h.responseWriter.WriteError(w, appErr.Message, appErr.StatusCode)
		return
	}

	h.logger.Error("[create_user_handler] Internal server error", "error", err)
	h.responseWriter.WriteError(w, "Internal server error", http.StatusInternalServerError)
}
