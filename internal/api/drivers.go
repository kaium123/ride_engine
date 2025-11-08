package api

import (
	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
	appMiddleware "vcs.technonext.com/carrybee/ride_engine/pkg/middleware"
)

// registerDriverRoutes registers all driver-related routes
func (s *ApiServer) registerDriverRoutes(e *echo.Echo, authMiddleware *appMiddleware.AuthMiddleware, driverHandler *handler.DriverHandler) {
	// Public routes
	e.POST("/api/v1/drivers/register", driverHandler.Register)
	e.POST("/api/v1/drivers/login/request-otp", driverHandler.RequestOTP)
	e.POST("/api/v1/drivers/login/verify-otp", driverHandler.VerifyOTP)

	// Protected routes
	e.POST("/api/v1/drivers/location", driverHandler.UpdateLocation, authMiddleware.AuthEcho)
	e.POST("/api/v1/drivers/status", driverHandler.SetOnlineStatus)
	e.GET("/api/v1/rides/nearby", driverHandler.FindNearestDrivers, authMiddleware.AuthEcho)
}
