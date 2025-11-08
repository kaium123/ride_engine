package api

import (
	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
	"vcs.technonext.com/carrybee/ride_engine/pkg/middleware"
)

// registerRideRoutes registers all ride-related routes
func (s *ApiServer) registerRideRoutes(e *echo.Group, authMiddleware *middleware.AuthMiddleware, rideHandler *handler.RideHandler) {
	rides := e.Group("/rides")
	rides.POST("/", rideHandler.RequestRide, authMiddleware.AuthEcho)
	rides.POST("/accept", rideHandler.AcceptRide, authMiddleware.AuthEcho)
	rides.POST("/start", rideHandler.StartRide, authMiddleware.AuthEcho)
	rides.POST("/complete", rideHandler.CompleteRide, authMiddleware.AuthEcho)
	rides.POST("/cancel", rideHandler.CancelRide, authMiddleware.AuthEcho)

}
