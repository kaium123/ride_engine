package api

import (
	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
)

// registerCustomerRoutes registers all customer-related routes
func (s *ApiServer) registerCustomerRoutes(e *echo.Group, customerHandler *handler.CustomerHandler) {
	customers := e.Group("/customers")
	customers.POST("/register", customerHandler.Register)
	customers.POST("/login", customerHandler.Login)
}
