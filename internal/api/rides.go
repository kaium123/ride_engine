package api

import (
	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
	"vcs.technonext.com/carrybee/ride_engine/pkg/middleware"
)

// registerRideRoutes registers all ride-related routes
func (s *ApiServer) registerRideRoutes(e *echo.Echo, authMiddleware *middleware.AuthMiddleware, rideHandler *handler.RideHandler) {
	e.POST("/api/v1/rides", rideHandler.RequestRide, authMiddleware.AuthEcho)
	e.POST("/api/v1/rides/accept", rideHandler.AcceptRide, authMiddleware.AuthEcho)
	e.POST("/api/v1/rides/start", rideHandler.StartRide, authMiddleware.AuthEcho)
	e.POST("/api/v1/rides/complete", rideHandler.CompleteRide, authMiddleware.AuthEcho)
	e.POST("/api/v1/rides/cancel", rideHandler.CancelRide, authMiddleware.AuthEcho)
}
