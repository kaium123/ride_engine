package migration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"
	"vcs.technonext.com/carrybee/ride_engine/pkg/utils"

	"github.com/spf13/cobra"
)

// newCmd represents root migration command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new migration",
	Long:  `Create new migration`,
	Run:   createNewMigration,
}
var entityName string

func init() {
	MigrationCmd.AddCommand(newCmd)
	newCmd.Flags().StringVarP(&entityName, "name", "n", "", "Name of the migration")
	newCmd.MarkFlagRequired("name")
}

func createNewMigration(cmd *cobra.Command, args []string) {
	if entityName != "" {
		cwd, err := os.Getwd()
		if err != nil {
			logger.Fatal(err)
			return
		}

		migrationOutputDirectory := "/" + NormalizePath(fmt.Sprintf("%s/pkg/migrations/migrations", cwd))
		processedEntityName := utils.ProcessString(entityName)
		cmd := exec.Command("migrate", "create", "-ext", "sql", "-dir", migrationOutputDirectory, "-format", fmt.Sprintf("%d", time.Now().Unix()), processedEntityName.SnakeCaseLower)
		// Run the command and get the output and error if any
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Fatal("Error running migrate command: ", err)
			return
		}

		// Print the output from the command
		fmt.Printf("Output:\n%s\n", string(output))

		os.Exit(1)

	} else {
		logger.Info(context.Background(), "Please provide a name for the migration with --name flag")
		os.Exit(1)
	}
}

func NormalizePath(input string) string {
	// Use path.Clean to normalize the slashes
	cleanedPath := path.Clean(input)

	// If cleanedPath is "." or "..", it should be returned as is, without trimming the leading slash
	if cleanedPath == "." || cleanedPath == ".." {
		return cleanedPath
	}

	// Remove leading slash if it exists
	return strings.TrimPrefix(cleanedPath, "/")
}
