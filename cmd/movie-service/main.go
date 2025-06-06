package main

import (
	"context"
	"log/slog"
	"os"
	"thermondo/config"
	"thermondo/internal/domain/shared"
	"thermondo/internal/pkg/postgres"
	"thermondo/internal/pkg/server"
	movieHandlers "thermondo/internal/platform/http/handlers/movies"
	userHandlers "thermondo/internal/platform/http/handlers/users"
	"thermondo/internal/platform/http/rest"
	"thermondo/internal/platform/repository"
	movieService "thermondo/internal/platform/service/movies"
	userService "thermondo/internal/platform/service/user"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Database
	db, err := postgres.NewConnection(cfg.Database.DSN, cfg.Database.HealthCheck)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	movieRepo := repository.NewMovieRepository(db)
	idGenerator := shared.NewULIDsGenerator()
	timeProvider := shared.NewTimeProvider()

	// Services
	userService := userService.NewUserService(userRepo, idGenerator, timeProvider)
	movieService := movieService.NewMovieService(movieRepo, idGenerator, timeProvider, logger)

	// Handlers
	userHandler := userHandlers.NewHandler(userService, logger)
	movieHandler := movieHandlers.NewHandler(movieService, logger)

	// Router with all handlers
	appRouter := rest.NewRouter(
		logger,
		rest.WithCORS(rest.DefaultCORSOptions()),
		rest.WithHandlers(
			userHandler,
			movieHandler,
		),
	)

	// Server
	srv, err := server.NewServer(
		cfg,
		logger,
		server.WithRouter(appRouter),
	)
	if err != nil {
		logger.Error("Failed to create server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err := srv.Run(context.Background()); err != nil {
		logger.Error("Server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
