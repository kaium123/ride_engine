package api

import (
	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/handler"
)

// registerCustomerRoutes registers all customer-related routes
func (s *ApiServer) registerCustomerRoutes(e *echo.Echo, customerHandler *handler.CustomerHandler) {
	e.POST("/api/v1/customers/register", customerHandler.Register)
	e.POST("/api/v1/customers/login", customerHandler.Login)
}
