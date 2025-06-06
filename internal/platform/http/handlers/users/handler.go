package users

import (
	"log/slog"
	"os"
	"thermondo/internal/pkg/http/response"
	userService "thermondo/internal/platform/service/user"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger         *slog.Logger
	userService    userService.UserService
	responseWriter *response.Writer
}

func NewHandler(userService userService.UserService, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return &Handler{
		logger:         logger,
		userService:    userService,
		responseWriter: response.NewWriter(logger),
	}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Route("/users", func(r chi.Router) {
		r.Get("/list", h.ListUsers)
		r.Post("/", h.CreateUser)

		//Resource routes (operate on single user by ID)
		r.Route("/{id}", func(r chi.Router) {
		})
	})
}
