package handler

import (
	"net/http"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/service"
)

type CustomerHandler struct {
	service *service.CustomerService
}

func NewCustomerHandler(service *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{service: service}
}

type RegisterCustomerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginCustomerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Customer interface{} `json:"customer"`
	Token    string      `json:"token"`
}

// Register handles customer registration
// @Summary Register a new customer
// @Description Register a new customer with name, email, phone, and password
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body RegisterCustomerRequest true "Customer registration details"
// @Success 201 {object} AuthResponse "Customer registered successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /customers/register [post]
func (h *CustomerHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	var req RegisterCustomerRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	customer, token, err := h.service.Register(ctx, req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, AuthResponse{
		Customer: customer,
		Token:    token,
	})
}

// Login handles customer login
// @Summary Login a customer
// @Description Authenticate a customer with email and password
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body LoginCustomerRequest true "Customer login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /customers/login [post]
func (h *CustomerHandler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	var req LoginCustomerRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	customer, token, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Customer: customer,
		Token:    token,
	})
}
