package migration

import (
	"context"
	"fmt"
	"net/url"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/spf13/cobra"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
	"vcs.technonext.com/carrybee/ride_engine/pkg/migrations"
)

// upCmd represents root migration command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Up create table and indices",
	Long:  `Up create table and indices`,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: up,
}

func init() {
	MigrationCmd.AddCommand(upCmd)
}

func up(cmd *cobra.Command, args []string) {
	ctx := context.Background()
	logger.Info(ctx, "Creating tables and indices")
	migrationFiles := migrations.GetMigrations()
	config := config.Load()

	dbConfig := config.Postgres

	if dbConfig.Port != 0 {
		dbConfig.Host = fmt.Sprintf("%s:%d", dbConfig.Host, dbConfig.Port)
	}

	uri := url.URL{
		Scheme: "postgres",
		Host:   dbConfig.Host,
		Path:   dbConfig.Database,
		User:   url.UserPassword(dbConfig.User, dbConfig.Password),
	}

	fmt.Println(uri.String())

	if dbConfig.Options != nil {
		fmt.Println("options:", dbConfig.Options)
		val := url.Values(dbConfig.Options)
		uri.RawQuery = val.Encode()
	}

	fmt.Println(uri.String())

	migrateDB, err := SQLFromUrl(uri.String())
	if err != nil {
		fmt.Println(err)
		logger.Fatal("Failed to connect to database: %v", err)
		return
	}
	defer migrateDB.Close()

	if err := migrateFromFS(migrateDB, "up", dbConfig.Database, migrationFiles); err != nil {
		logger.Fatal("Failed to migrate: %v", err)
		return
	}

	logger.Info(ctx, "Creating tables and indices successful!")
}
