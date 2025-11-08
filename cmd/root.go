package cmd

import (
	"github.com/spf13/cobra"
	"vcs.technonext.com/carrybee/ride_engine/cmd/migration"
)

var rootCmd = &cobra.Command{
	Use:   "ride_engine",
	Short: "Ride Engine API CLI",
	Long:  `A monolithic Ride Engine API server with PostgreSQL, MongoDB, Redis, and HTTP handlers.`,
}

func init() {
	rootCmd.AddCommand(migration.MigrationCmd)
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
