package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/mongodb"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/postgres"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/service"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
	appMiddleware "vcs.technonext.com/carrybee/ride_engine/pkg/middleware"

	_ "vcs.technonext.com/carrybee/ride_engine/docs"
)

// ApiServer holds all the dependencies for the API server
type ApiServer struct {
	config   *config.Config
	postgres *database.PostgresDB
	mongo    *database.MongoDB
	redis    *database.RedisDB
}

// NewServer creates a new API server with the provided dependencies
func NewServer(cfg *config.Config, postgresDB *database.PostgresDB, mongoDB *database.MongoDB, redisDB *database.RedisDB) *ApiServer {
	return &ApiServer{
		config:   cfg,
		postgres: postgresDB,
		mongo:    mongoDB,
		redis:    redisDB,
	}
}

// SetupRoutes initializes all repositories, services, handlers and sets up routes
func (s *ApiServer) SetupRoutes() *echo.Echo {
	// Initialize repositories
	customerRepo := postgres.NewCustomerPostgresRepository(s.postgres)
	driverRepo := postgres.NewDriverPostgresRepository(s.postgres)
	rideRepoMongo := mongodb.NewRideMongoRepository(s.mongo.Database) // MongoDB for rides with geospatial queries
	otpRepo := postgres.NewOTPPostgresRepository(s.postgres)
	onlineStatusRepo := postgres.NewOnlineStatusPostgresRepository(s.postgres.DB)
	locationRepo := mongodb.NewLocationMongoRepository(s.mongo.Database)

	// Initialize services
	otpService := service.NewOTPService(s.redis.Client, otpRepo)
	locationService := service.NewLocationService(locationRepo)
	customerService := service.NewCustomerService(customerRepo, s.config.JWT.Secret, s.config.JWT.Expiration, s.redis.Client)
	driverService := service.NewDriverService(driverRepo, onlineStatusRepo, otpService, locationService, s.config.JWT.Secret, s.config.JWT.Expiration, s.redis.Client)
	rideService := service.NewRideService(rideRepoMongo, locationService, driverService, customerRepo)

	// Initialize handlers
	customerHandler := handler.NewCustomerHandler(customerService)
	driverHandler := handler.NewDriverHandler(driverService)
	rideHandler := handler.NewRideHandler(rideService)

	// Setup Echo router
	e := echo.New()

	// Enable CORS to allow Swagger UI and other clients
	e.Use(middleware.CORS())

	authMiddleware := appMiddleware.NewAuthMiddleware(s.redis.Client, s.config.JWT.Secret)

	// Register routes
	s.registerRoutes(e, authMiddleware, customerHandler, driverHandler, rideHandler)

	return e
}

// registerRoutes registers all the API routes using route groups
func (s *ApiServer) registerRoutes(e *echo.Echo, authMiddleware *appMiddleware.AuthMiddleware, customerHandler *handler.CustomerHandler, driverHandler *handler.DriverHandler, rideHandler *handler.RideHandler) {
	// Register route groups
	api := e.Group("/api/v1")

	s.registerCustomerRoutes(api, customerHandler)
	s.registerDriverRoutes(api, authMiddleware, driverHandler)
	s.registerRideRoutes(api, authMiddleware, rideHandler)

	// Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
}
