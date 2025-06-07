package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Writer struct {
	logger *slog.Logger
}

func NewWriter(logger *slog.Logger) *Writer {
	return &Writer{logger: logger}
}

func (w *Writer) WriteSuccess(resp http.ResponseWriter, data any, statusCode int) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)

	if err := json.NewEncoder(resp).Encode(data); err != nil {
		w.logger.Error("Failed to encode success response", "error", err)
	}
}

func (w *Writer) WriteError(resp http.ResponseWriter, message string, statusCode int) {
	errorResponse := ErrorResponse{
		Error: message,
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)

	if err := json.NewEncoder(resp).Encode(errorResponse); err != nil {
		w.logger.Error("Failed to encode error response", "error", err)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type APIResponse struct {
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
	Status  int       `json:"status"`
	Message string    `json:"message,omitempty"`
}

type APIError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (w *Writer) WriteAPIResponse(resp http.ResponseWriter, data interface{}, err *APIError, statusCode int, message string) {
	response := APIResponse{
		Data:    data,
		Error:   err,
		Status:  statusCode,
		Message: message,
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)

	if encodeErr := json.NewEncoder(resp).Encode(response); encodeErr != nil {
		w.logger.Error("Failed to encode API response", "error", encodeErr)
	}
}
