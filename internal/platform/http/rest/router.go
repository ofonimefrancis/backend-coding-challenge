package rest

import (
	"context"
	"fmt"
	"log/slog"

	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// HealthStatusProvider defines the interface for health status
type HealthStatusProvider interface {
	HealthStatus(ctx context.Context) error
	IsReady() bool
}

// HandlerProvider defines the interface for route handlers
type HandlerProvider interface {
	RegisterRoutes(r chi.Router)
}

// Router wraps chi router with middleware and CORS configuration
type Router struct {
	mux           *chi.Mux
	logger        *slog.Logger
	corsOptions   *cors.Options
	healthChecker HealthStatusProvider
	handlers      []HandlerProvider
}

// RouterOption defines functional options for router configuration
type RouterOption func(*Router)

// WithCORS sets CORS configuration
func WithCORS(options *cors.Options) RouterOption {
	return func(r *Router) {
		r.corsOptions = options
	}
}

// WithHealthProvider sets the health status provider
func WithHealthProvider(provider HealthStatusProvider) RouterOption {
	return func(r *Router) {
		r.healthChecker = provider
	}
}

// WithHandlers registers multiple handler providers
func WithHandlers(handlers ...HandlerProvider) RouterOption {
	return func(r *Router) {
		r.handlers = append(r.handlers, handlers...)
	}
}

// NewRouter creates a new router with middleware and routes
func NewRouter(logger *slog.Logger, opts ...RouterOption) *Router {
	if logger == nil {
		logger = slog.Default()
	}

	router := &Router{
		mux:    chi.NewRouter(),
		logger: logger,
	}

	for _, opt := range opts {
		opt(router)
	}
	router.setupMiddleware()
	router.setupRoutes()

	return router
}

func (r *Router) Handler() http.Handler {
	return r.mux
}

// setupMiddleware configures standard middleware stack
func (r *Router) setupMiddleware() {
	r.mux.Use(middleware.RequestID)
	r.mux.Use(middleware.RealIP)
	r.mux.Use(middleware.Recoverer)
	r.mux.Use(middleware.Timeout(30 * time.Second))

	if r.corsOptions != nil {
		r.mux.Use(cors.Handler(*r.corsOptions))
	}

	r.mux.Use(r.loggingMiddleware)
}

// loggingMiddleware provides structured request logging
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, req.ProtoMajor)

		defer func() {
			r.logger.Info("HTTP request",
				slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", req.RemoteAddr),
				slog.String("user_agent", req.UserAgent()),
				slog.String("request_id", middleware.GetReqID(req.Context())),
			)
		}()

		next.ServeHTTP(ww, req)
	})
}

// setupRoutes configures application routes
func (r *Router) setupRoutes() {
	// Health endpoints
	r.mux.Get("/health", r.handleHealth)
	r.mux.Get("/ready", r.handleReadiness)

	r.mux.Handle("/swagger/*", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./docs/swagger-ui"))))

	// Serve the OpenAPI YAML
	r.mux.Get("/swagger/openapi.yml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/openapi.yml")
	})

	// Swagger UI
	r.mux.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// API versioning
	r.mux.Route("/api/v1", func(v1 chi.Router) {
		// Register all handler providers
		for _, handler := range r.handlers {
			handler.RegisterRoutes(v1)
		}
	})
}

// handleHealth provides a health check endpoint
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()

	if r.healthChecker != nil {
		if err := r.healthChecker.HealthStatus(ctx); err != nil {
			r.logger.Error("Health check failed", slog.String("error", err.Error()))
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// handleReadiness provides a readiness check endpoint
func (r *Router) handleReadiness(w http.ResponseWriter, req *http.Request) {
	if r.healthChecker == nil || !r.healthChecker.IsReady() {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ready","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// DefaultCORSOptions returns a sensible default CORS configuration
func DefaultCORSOptions() *cors.Options {
	return &cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure based on environment
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}
}

// ProductionCORSOptions returns CORS configuration for production
func ProductionCORSOptions(allowedOrigins []string) *cors.Options {
	return &cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}
