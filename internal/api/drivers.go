package api

import (
	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
	appMiddleware "vcs.technonext.com/carrybee/ride_engine/pkg/middleware"
)

// registerDriverRoutes registers all driver-related routes
func (s *ApiServer) registerDriverRoutes(e *echo.Group, authMiddleware *appMiddleware.AuthMiddleware, driverHandler *handler.DriverHandler) {
	drivers := e.Group("/drivers")
	// Public routes
	drivers.POST("/register", driverHandler.Register)
	drivers.POST("/login/request-otp", driverHandler.RequestOTP)
	drivers.POST("/login/verify-otp", driverHandler.VerifyOTP)

	// Protected routes
	drivers.POST("/location", driverHandler.UpdateLocation, authMiddleware.AuthEcho)
	drivers.POST("/status", driverHandler.SetOnlineStatus)
	e.POST("/rides/nearby", driverHandler.FindNearestDrivers, authMiddleware.AuthEcho)
}
