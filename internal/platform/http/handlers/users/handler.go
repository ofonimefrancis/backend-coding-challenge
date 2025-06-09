package users

import (
	"log/slog"
	"os"
	"thermondo/internal/pkg/http/response"
	userService "thermondo/internal/platform/service/user"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	userService    userService.UserService
	logger         *slog.Logger
	responseWriter *response.Writer
	jwtSecret      string
}

func NewHandler(userService userService.UserService, logger *slog.Logger, jwtSecret string) *Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return &Handler{
		userService:    userService,
		logger:         logger,
		responseWriter: response.NewWriter(logger),
		jwtSecret:      jwtSecret,
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/users", func(r chi.Router) {
		r.Post("/", h.CreateUser)
		r.Post("/login", h.Login)
		r.Get("/", h.ListUsers)
		r.Get("/{id}", h.GetUser)
	})
}
