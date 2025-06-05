package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewConnection(dsn string, hasHealthCheck bool) (*sqlx.DB, error) {
	fmt.Println("dsn", dsn)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Ping the database to check if the connection is healthy
	if hasHealthCheck {
		if err := db.Ping(); err != nil {
			return nil, err
		}
	}

	return db, nil
}
