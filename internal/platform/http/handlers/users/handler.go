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
	router.Route("/api/v1/users", func(r chi.Router) {
		// Collection routes (operate on multiple users)
		r.Get("/", h.ListUsers)   // GET /api/v1/users - List all users
		r.Post("/", h.CreateUser) // POST /api/v1/users - Create new user

		//Resource routes (operate on single user by ID)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetUser) // GET /api/v1/users/{id} - Get user by ID
		})
	})
}
