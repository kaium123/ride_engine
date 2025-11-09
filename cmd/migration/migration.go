package migration

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/spf13/cobra"
	"vcs.technonext.com/carrybee/ride_engine/pkg/migrations"
)

// MigrationCmd represents root migration command
var MigrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Migration create/drop table and indices",
	Long:  `Migration create/drop table and indices`,
}

func migrateFromFS(db *sql.DB, commandStatus, database string, files fs.FS) error {
	src, err := httpfs.New(http.FS(files), "migrations")
	if err != nil {
		return fmt.Errorf("failed to initialize migration source: %w", err)
	}

	return migrateFromSource(db, commandStatus, database, src)
}

func migrateFromSource(db *sql.DB, commandStatus, database string, files source.Driver) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Fatal(err)
	}

	m, err := migrate.NewWithInstance("httpfs", files, database, driver)
	if err != nil {
		logger.Fatal("Failed to create migration instance: ", err)
	}

	logger.Info(context.Background(), "Running migration . . .")
	if commandStatus == "down" {
		err = m.Down()
	} else {
		err = m.Up()
	}

	if err == migrate.ErrNoChange || err == migrate.ErrNilVersion {
		//log.Println("No changes were made during the migration")
		return nil
	}

	if err != nil {
		//log.Println("Migration failed: %v", err)
		return err
	}

	logger.Info(context.Background(), "Migration applied successfully.")
	return nil
}

func SQLFromUrl(url string) (*sql.DB, error) {
	cfg := &migrations.Config{URL: url}
	db, err := migrations.New(cfg)

	return db, err
}
