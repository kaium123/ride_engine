package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/labstack/echo/v4"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/service"
	"vcs.technonext.com/carrybee/ride_engine/pkg/middleware"
)

type DriverHandler struct {
	service *service.DriverService
}

func NewDriverHandler(service *service.DriverService) *DriverHandler {
	return &DriverHandler{service: service}
}

type RegisterDriverRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	VehicleNo string `json:"vehicle_no"`
}

type RequestOTPRequest struct {
	Phone string `json:"phone"`
}

type VerifyOTPRequest struct {
	Phone string `json:"phone"`
	OTP   string `json:"otp"`
}

type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type SetOnlineStatusRequest struct {
	IsOnline bool `json:"is_online"`
}

// Register handles driver registration
// @Summary Register a new driver
// @Description Register a new driver with name, phone, and vehicle number
// @Tags Drivers
// @Accept json
// @Produce json
// @Param request body RegisterDriverRequest true "Driver registration details"
// @Success 201 {object} map[string]interface{} "Driver registered successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /drivers/register [post]
func (h *DriverHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	var req RegisterDriverRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	driver, err := h.service.Register(ctx, req.Name, req.Phone, req.VehicleNo)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, driver)
}

// RequestOTP handles OTP generation and sending
// @Summary Request OTP for driver login
// @Description Send an OTP to the driver's phone number for authentication
// @Tags Drivers
// @Accept json
// @Produce json
// @Param request body RequestOTPRequest true "Phone number to send OTP"
// @Success 200 {object} MessageResponse "OTP sent successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /drivers/login/request-otp [post]
func (h *DriverHandler) RequestOTP(c echo.Context) error {
	ctx := c.Request().Context()
	var req RequestOTPRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err := h.service.RequestOTP(ctx, req.Phone)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "OTP sent successfully"})
}

// VerifyOTP handles OTP verification and login
// @Summary Verify OTP and login driver
// @Description Verify the OTP sent to driver's phone and authenticate
// @Tags Drivers
// @Accept json
// @Produce json
// @Param request body VerifyOTPRequest true "Phone and OTP for verification"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /drivers/login/verify-otp [post]
func (h *DriverHandler) VerifyOTP(c echo.Context) error {
	ctx := c.Request().Context()
	var req VerifyOTPRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	driver, token, err := h.service.VerifyOTP(ctx, req.Phone, req.OTP)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Customer: driver,
		Token:    token,
	})
}

// UpdateLocation handles driver location updates
// @Summary Update driver location
// @Description Update the current location of the authenticated driver
// @Tags Drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateLocationRequest true "Driver's current location"
// @Success 200 {object} MessageResponse "Location updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers/location [post]
func (h *DriverHandler) UpdateLocation(c echo.Context) error {
	ctx := c.Request().Context()
	driverID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing user id"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing driver ID in context"})
	}
	fmt.Println("Driver ID from context:", driverID)

	role, ok := middleware.GetUserRoleFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing user role"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
	}
	if role != "driver" {
		logger.Error(ctx, errors.New("invalid role"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid role in context"})
	}

	var req UpdateLocationRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err := h.service.UpdateLocation(ctx, driverID, req.Latitude, req.Longitude)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "Location updated successfully"})
}

// SetOnlineStatus handles driver online/offline status
// @Summary Set driver online/offline status
// @Description Update whether the driver is available to accept rides
// @Tags Drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SetOnlineStatusRequest true "Driver's online status"
// @Success 200 {object} MessageResponse "Status updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers/status [post]
func (h *DriverHandler) SetOnlineStatus(c echo.Context) error {
	ctx := c.Request().Context()
	driverID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing user id"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing driver ID in context"})
	}
	fmt.Println("Driver ID from context:", driverID)

	role, ok := middleware.GetUserRoleFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing user role"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
	}
	if role != "driver" {
		logger.Error(ctx, errors.New("invalid role"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid role in context"})
	}

	var req SetOnlineStatusRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err := h.service.SetOnlineStatus(ctx, driverID, req.IsOnline)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	status := "offline"
	if req.IsOnline {
		status = "online"
	}
	return c.JSON(http.StatusOK, MessageResponse{Message: "Driver is now " + status})
}

// FindNearestDrivers finds nearest available drivers
// @Summary Find nearest drivers
// @Description Find nearest available drivers within a specified radius
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param lat query number true "Latitude of the location"
// @Param lng query number true "Longitude of the location"
// @Param radius query number false "Search radius in meters (default: 3000)"
// @Param limit query integer false "Maximum number of drivers to return (default: 5)"
// @Success 200 {object} map[string]interface{} "List of nearest drivers"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /rides/nearby [get]
func (h *DriverHandler) FindNearestDrivers(c echo.Context) error {
	ctx := c.Request().Context()
	latStr := c.QueryParam("lat")
	lngStr := c.QueryParam("lng")
	radiusStr := c.QueryParam("radius")
	limitStr := c.QueryParam("limit")

	if latStr == "" || lngStr == "" {
		logger.Error(ctx, errors.New("missing latitude, lng"))
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "lat and lng are required"})
	}

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		logger.Error(ctx, errors.New("invalid latitude, lng"))
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid coordinates"})
	}

	radius := 3000.0
	if radiusStr != "" {
		if v, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			radius = v
		}
	}

	limit := 5
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}

	driverIDs, err := h.service.GetNearestDrivers(ctx, lat, lng, radius, limit)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	resp := map[string]interface{}{
		"drivers": driverIDs,
		"count":   len(driverIDs),
	}

	return c.JSON(http.StatusOK, resp)
}
