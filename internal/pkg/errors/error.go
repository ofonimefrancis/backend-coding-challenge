package errors

import "net/http"

type AppError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
	Code       string `json:"code,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusConflict,
		Code:       "CONFLICT",
	}
}

func NewInternalError(message string) *AppError {
	return &AppError{
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_ERROR",
	}
}
