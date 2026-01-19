package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/registryx/registryx/backend/pkg/config"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}
