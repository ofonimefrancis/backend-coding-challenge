package migrate

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

// RunMigrations runs all pending migrations
func RunMigrations(db *sql.DB, migrationsDir string, logger *slog.Logger) error {
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Applied migrations", "count", n)
	return nil
}

// RollbackMigrations rolls back the last migration
func RollbackMigrations(db *sql.DB, migrationsDir string, logger *slog.Logger) error {
	migrations := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Down)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	logger.Info("Rolled back migrations", "count", n)
	return nil
}
