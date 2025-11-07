package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"vcs.technonext.com/carrybee/ride_engine/cmd/migration"
)

var rootCmd = &cobra.Command{
	Use:   "ride_engine",
	Short: "Ride Engine API CLI",
	Long:  `A monolithic Ride Engine API server with PostgreSQL, MongoDB, Redis, and HTTP handlers.`,
}

func init() {
	rootCmd.AddCommand(migration.MigrationCmd)
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yml", "config file")
	//rootCmd.PersistentFlags().BoolVarP(&prettyPrintLog, "pretty", "p", false, "pretty print verbose/log")
	//rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// set the value to viper config
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
