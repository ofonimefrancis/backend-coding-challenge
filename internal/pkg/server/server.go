// server/server.go
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"thermondo/config"
	"time"
)

var (
	ErrServerAlreadyRunning = errors.New("server is already running")
	ErrServerNotRunning     = errors.New("server is not running")
	ErrInvalidConfiguration = errors.New("invalid server configuration")
)

// ServerState represents the current state of the server
type ServerState int

const (
	StateStopped ServerState = iota
	StateStarting
	StateRunning
	StateStopping
)

// HealthChecker defines the interface for health check dependencies
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// RouterProvider defines the interface for providing HTTP handlers
type RouterProvider interface {
	Handler() http.Handler
}

// Server represents an HTTP server with graceful shutdown capabilities
type Server struct {
	httpServer    *http.Server
	logger        *slog.Logger
	config        config.Configuration
	state         ServerState
	healthChecker HealthChecker
	router        RouterProvider

	// Channels for coordinating server lifecycle
	shutdownCh chan struct{}
	doneCh     chan error
}

// NewServer creates a new HTTP server instance with the given configuration
func NewServer(conf config.Configuration, logger *slog.Logger, opts ...ServerOption) (*Server, error) {
	if err := validateConfig(conf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfiguration, err)
	}

	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	server := &Server{
		logger:     logger,
		config:     conf,
		state:      StateStopped,
		shutdownCh: make(chan struct{}),
		doneCh:     make(chan error, 1),
	}

	// Apply functional options
	for _, opt := range opts {
		opt(server)
	}

	// Ensure we have a router
	if server.router == nil {
		return nil, fmt.Errorf("%w: router provider is required", ErrInvalidConfiguration)
	}

	// Setup HTTP server with timeouts and secure defaults
	server.httpServer = &http.Server{
		Addr:         net.JoinHostPort(conf.Server.Host, conf.Server.Port),
		Handler:      server.router.Handler(),
		ReadTimeout:  time.Duration(conf.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(conf.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(conf.Server.IdleTimeout) * time.Second,
	}

	return server, nil
}

// validateConfig ensures the server configuration is valid
func validateConfig(conf config.Configuration) error {
	if conf.Server.Port == "" {
		return errors.New("server port is required")
	}
	if conf.Server.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}
	return nil
}

// Start begins listening for HTTP requests
func (s *Server) Start(ctx context.Context) error {
	if s.state != StateStopped {
		return ErrServerAlreadyRunning
	}

	s.state = StateStarting
	s.logger.Info("Starting HTTP server", slog.String("addr", s.httpServer.Addr))

	go func() {
		defer close(s.doneCh)

		s.state = StateRunning
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server failed to start", slog.String("error", err.Error()))
			s.doneCh <- err
			return
		}
		s.doneCh <- nil
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-s.doneCh:
		if err != nil {
			s.state = StateStopped
			return fmt.Errorf("failed to start server: %w", err)
		}
	case <-time.After(5 * time.Second):
		s.logger.Info("Server started successfully", slog.String("addr", s.httpServer.Addr))
	}

	return nil
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.state != StateRunning {
		return ErrServerNotRunning
	}

	s.state = StateStopping
	s.logger.Info("Shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Error during server shutdown", slog.String("error", err.Error()))
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.state = StateStopped
	s.logger.Info("Server shutdown completed")
	return nil
}

// Run starts the server and handles graceful shutdown on OS signals
func (s *Server) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := s.Start(runCtx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		s.logger.Info("Received shutdown signal", slog.String("signal", sig.String()))
	case <-runCtx.Done():
		s.logger.Info("Context cancelled, shutting down")
	case err := <-s.doneCh:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		time.Duration(s.config.Server.ShutdownTimeout)*time.Second,
	)
	defer shutdownCancel()

	return s.Shutdown(shutdownCtx)
}

// State returns the current server state
func (s *Server) State() ServerState {
	return s.state
}

// Addr returns the server's listening address
func (s *Server) Addr() string {
	if s.httpServer == nil {
		return ""
	}
	return s.httpServer.Addr
}

// HealthStatus provides server health information
func (s *Server) HealthStatus(ctx context.Context) error {
	if s.healthChecker != nil {
		return s.healthChecker.HealthCheck(ctx)
	}
	return nil
}

// IsReady returns whether the server is ready to handle requests
func (s *Server) IsReady() bool {
	return s.state == StateRunning
}
