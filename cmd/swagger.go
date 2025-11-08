package cmd

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/internal/api"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	_ "vcs.technonext.com/carrybee/ride_engine/docs"

	"github.com/spf13/cobra"
)

var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "Start the Swagger documentation server",
	Long:  `Starts a standalone Swagger documentation server for the Ride Engine API.`,
	Run: func(cmd *cobra.Command, args []string) {
		startSwaggerServer()
	},
}

func init() {
	rootCmd.AddCommand(swaggerCmd)
}

func startSwaggerServer() {
	// Load configuration
	cfg := config.Load()

	// Convert port string to int
	port, err := strconv.Atoi(cfg.Swagger.Port)
	if err != nil {
		logger.Fatal("Invalid swagger port: ", err)
	}

	// Create swagger server
	swaggerServer := api.NewSwagger(api.SwaggerServerOpts{
		ListenPort: port,
	})

	// Start server in a goroutine
	go func() {
		if err := swaggerServer.Run(); err != nil {
			logger.Fatal("Swagger server failed: ", err)
		}
	}()

	logger.Info(context.Background(), "Swagger server running on port "+cfg.Swagger.Port)
	logger.Info(context.Background(), "Access Swagger UI at: http://localhost:"+cfg.Swagger.Port+"/swagger/index.html")

	// Wait for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(context.Background(), "Swagger server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := swaggerServer.Shutdown(ctx); err != nil {
		logger.Fatal("Swagger server forced to shutdown: ", err)
	}

	logger.Info(ctx, "Swagger server stopped gracefully")
}
