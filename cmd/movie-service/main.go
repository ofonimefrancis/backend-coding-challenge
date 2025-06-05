package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"thermondo/config"
	"thermondo/internal/pkg/postgres"
	"thermondo/internal/pkg/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("cfg", cfg)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

	httpServer, err := server.NewServer(
		cfg,
		logger,
	)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	db, err := postgres.NewConnection(cfg.Database.DSN, cfg.Database.HealthCheck)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)

	// setup individual services and repository

	// Run the server
	ctx := context.Background()
	if err := httpServer.Run(ctx); err != nil {
		logger.Error("Server execution failed",
			slog.String("error", err.Error()),
			slog.String("addr", httpServer.Addr()),
		)
		os.Exit(1)
	}
}
