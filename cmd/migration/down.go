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

// downCmd represents root migration command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Down drop table and indices",
	Long:  `Down drop table and indices`,
	PreRun: func(cmd *cobra.Command, args []string) {

	},
	Run: down,
}

func init() {
	MigrationCmd.AddCommand(downCmd)
}

func down(cmd *cobra.Command, args []string) {
	//ctx := context.Background()
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

	if dbConfig.Options != nil {
		val := url.Values(dbConfig.Options)
		uri.RawQuery = val.Encode()
	}

	migrateDB, err := SQLFromUrl(uri.String())
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
		return
	}
	defer migrateDB.Close()

	if err := migrateFromFS(migrateDB, "down", dbConfig.Database, migrationFiles); err != nil {
		logger.Fatal("Failed to migrate:", err)
		return
	}

	logger.Info(context.Background(), "Migration down successful!")
}
