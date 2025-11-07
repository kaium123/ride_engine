package migrations

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	_ "github.com/lib/pq"
)

//go:embed migrations
var migrations embed.FS

type Config struct {
	URL string
}

func New(cfg *Config) (*sql.DB, error) {
	if cfg.URL == "" {
		logger.Error(context.Background(), "no migration url provided")
		return nil, errors.New("no migration url provided")
	}

	db, err := sql.Open("postgres", cfg.URL)
	if err != nil {
		logger.Error(context.Background(), "failed to open migration db")
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error(context.Background(), "failed to open migration db")
		return nil, err
	}

	return db, nil
}

func GetMigrations() fs.FS {
	return migrations
}
