package main

import (
	"context"
	"log/slog"
	"os"
	"thermondo/config"
	"thermondo/internal/domain/shared"
	"thermondo/internal/pkg/cache"
	"thermondo/internal/pkg/postgres"
	"thermondo/internal/pkg/server"
	movieHandlers "thermondo/internal/platform/http/handlers/movies"
	ratingHandlers "thermondo/internal/platform/http/handlers/ratings"
	userHandlers "thermondo/internal/platform/http/handlers/users"
	"thermondo/internal/platform/http/rest"
	"thermondo/internal/platform/repository"
	movieService "thermondo/internal/platform/service/movies"
	ratingService "thermondo/internal/platform/service/rating"
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

	// Determine environment
	appEnv := os.Getenv("APP_ENV")
	var c cache.Cache
	if appEnv == "production" {
		redisConfig := cache.RedisConfig{
			Host:     cfg.Redis.Host,
			Port:     cfg.Redis.Port, // Default Redis port
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}
		c, err = cache.NewRedisCache(redisConfig, "thermondo")
		if err != nil {
			logger.Error("Failed to initialize Redis cache", slog.String("error", err.Error()))
			os.Exit(1)
		}
		logger.Info("Using Redis cache")
	} else {
		c = cache.NewNoOpCache()
		logger.Info("Using NoOp (in-memory) cache for non-production environment", slog.String("env", appEnv))
	}
	defer c.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db, c)
	movieRepo := repository.NewMovieRepository(db)
	ratingRepo := repository.NewRatingRepository(db)
	idGenerator := shared.NewULIDsGenerator()
	timeProvider := shared.NewTimeProvider()

	// Services
	userService := userService.NewUserService(userRepo, ratingRepo, movieRepo, idGenerator, timeProvider, c)
	movieService := movieService.NewMovieService(movieRepo, idGenerator, timeProvider, logger)
	ratingService := ratingService.NewRatingService(ratingRepo, idGenerator, timeProvider, logger)

	// Handlers
	userHandler := userHandlers.NewHandler(userService, logger, cfg.JWT.Secret)
	movieHandler := movieHandlers.NewHandler(movieService, logger)
	ratingHandler := ratingHandlers.NewHandler(ratingService, logger)
	userProfileHandler := userHandlers.NewProfileHandler(userService, logger)

	// Router with all handlers
	appRouter := rest.NewRouter(
		logger,
		rest.WithCORS(rest.DefaultCORSOptions()),
		rest.WithHandlers(
			userHandler,
			movieHandler,
			ratingHandler,
			userProfileHandler,
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
