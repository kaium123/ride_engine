package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/spf13/cobra"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/postgres"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/service"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
	"vcs.technonext.com/carrybee/ride_engine/pkg/middleware"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Ride Engine API server",
	Long:  `Starts the monolithic Ride Engine API server with all routes, database connections, and background workers.`,
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func startServer() {
	// Load configuration
	cfg := config.Load()

	// Initialize PostgreSQL
	postgresDB, err := database.NewPostgresDB(cfg.Postgres)
	if err != nil {
		logger.Fatal("Failed to connect to PostgresSQL : ", err)
	}
	defer postgresDB.Close()

	logger.Info(context.Background(), "Running database migrations...")
	if err := postgres.AutoMigrate(postgresDB.DB); err != nil {
		logger.Fatal("Failed to migrate postgres schema : ", err)
	}
	logger.Info(context.Background(), "Migrations completed successfully")

	// Initialize MongoDB
	mongoDB, err := database.NewMongoDB(cfg.MongoDB)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB : ", err)
	}
	defer mongoDB.Close()

	// Initialize Redis
	redisDB, err := database.NewRedisDB(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis : ", err)
	}
	defer redisDB.Close()

	// Initialize repositories
	customerRepo := postgres.NewCustomerPostgresRepository(postgresDB)
	driverRepo := postgres.NewDriverPostgresRepository(postgresDB)
	rideRepo := postgres.NewRidePostgresRepository(postgresDB)
	otpRepo := postgres.NewOTPPostgresRepository(postgresDB)

	// Initialize services
	otpService := service.NewOTPService(redisDB.Client, otpRepo)
	locationService := service.NewLocationService(mongoDB.Database)
	customerService := service.NewCustomerService(customerRepo, cfg.JWT.Secret, cfg.JWT.Expiration, redisDB.Client)
	driverService := service.NewDriverService(driverRepo, otpService, locationService, cfg.JWT.Secret, cfg.JWT.Expiration, redisDB.Client)
	rideService := service.NewRideService(rideRepo, locationService)

	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(customerService)
	driverHandler := handler.NewDriverHandler(driverService)
	rideHandler := handler.NewRideHandler(rideService)

	mux := http.NewServeMux()
	authMiddleware := middleware.NewAuthMiddleware(redisDB.Client, cfg.JWT.Secret)

	mux.HandleFunc("/api/v1/customers/register", customerHandler.Register)
	mux.HandleFunc("/api/v1/customers/login", customerHandler.Login)
	mux.HandleFunc("/api/v1/drivers/register", driverHandler.Register)
	mux.HandleFunc("/api/v1/drivers/login/request-otp", driverHandler.RequestOTP)
	mux.HandleFunc("/api/v1/drivers/login/verify-otp", driverHandler.VerifyOTP)
	mux.Handle("/api/v1/drivers/location", authMiddleware.Auth(http.HandlerFunc(driverHandler.UpdateLocation)))
	mux.HandleFunc("/api/v1/drivers/status", driverHandler.SetOnlineStatus)
	mux.Handle("/api/v1/rides/nearby", authMiddleware.Auth(http.HandlerFunc(driverHandler.FindNearestDrivers)))
	mux.Handle("/api/v1/rides", authMiddleware.Auth(http.HandlerFunc(rideHandler.RequestRide)))
	mux.Handle("/api/v1/rides/accept", authMiddleware.Auth(http.HandlerFunc(rideHandler.AcceptRide)))
	mux.Handle("/api/v1/rides/start", authMiddleware.Auth(http.HandlerFunc(rideHandler.StartRide)))
	mux.Handle("/api/v1/rides/complete", authMiddleware.Auth(http.HandlerFunc(rideHandler.CompleteRide)))
	mux.Handle("/api/v1/rides/cancel", authMiddleware.Auth(http.HandlerFunc(rideHandler.CancelRide)))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Print all routes
	printRoutes(cfg.Server.Port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed: ", err)
		}
	}()

	logger.Info(context.Background(), "Listening on "+cfg.Server.Port)

	// Wait for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(context.Background(), "Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", err)
	}

	logger.Info(ctx, "Server stopped gracefully")
}

func printRoutes(port string) {
	fmt.Printf("\nRide Engine API Server (CLI)\n")
	fmt.Println("============================")
	fmt.Println("\nCustomer Endpoints:")
	fmt.Println("  POST   /api/v1/customers/register")
	fmt.Println("  POST   /api/v1/customers/login")
	fmt.Println("\nDriver Endpoints:")
	fmt.Println("  POST   /api/v1/drivers/register")
	fmt.Println("  POST   /api/v1/drivers/login/request-otp")
	fmt.Println("  POST   /api/v1/drivers/login/verify-otp")
	fmt.Println("  POST   /api/v1/drivers/location")
	fmt.Println("  POST   /api/v1/drivers/status")
	fmt.Println("\nRide Endpoints:")
	fmt.Println("  POST   /api/v1/rides")
	fmt.Println("  GET    /api/v1/rides/nearby")
	fmt.Println("  POST   /api/v1/rides/accept")
	fmt.Println("  POST   /api/v1/rides/start")
	fmt.Println("  POST   /api/v1/rides/complete")
	fmt.Println("  POST   /api/v1/rides/cancel")
	fmt.Println("\nHealth:")
	fmt.Println("  GET    /health")
	fmt.Printf("\n✅ Server running on http://localhost:%s\n\n", port)
}
