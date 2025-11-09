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

type RideHandler struct {
	service *service.RideService
}

func NewRideHandler(service *service.RideService) *RideHandler {
	return &RideHandler{service: service}
}

type RequestRideRequest struct {
	PickupLat  float64 `json:"pickup_lat"`
	PickupLng  float64 `json:"pickup_lng"`
	DropoffLat float64 `json:"dropoff_lat"`
	DropoffLng float64 `json:"dropoff_lng"`
}

// RequestRide handles customer ride requests
// @Summary Request a new ride
// @Description Create a new ride request with pickup and dropoff locations
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RequestRideRequest true "Ride request details"
// @Success 201 {object} map[string]interface{} "Ride created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /rides [post]
func (h *RideHandler) RequestRide(c echo.Context) error {
	ctx := c.Request().Context()
	customerID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("no user id from context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing customer ID in context"})
	}
	fmt.Println("customer ID from context:", customerID)

	role, ok := middleware.GetUserRoleFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("no user role from context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
	}

	if role != "customer" {
		logger.Error(ctx, errors.New("invalid role"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid role"})
	}

	var req RequestRideRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	ride, err := h.service.RequestRide(ctx, customerID, req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, ride)
}

// GetNearbyRides handles getting nearby rides for drivers
func (h *RideHandler) GetNearbyRides(c echo.Context) error {
	ctx := c.Request().Context()
	driverID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing customer ID in context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing driver ID in context"})
	}
	fmt.Println("Driver ID from context:", driverID)

	role, ok := middleware.GetUserRoleFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing role in context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
	}
	if role != "driver" {
		logger.Error(ctx, errors.New("role is not driver"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid role in context"})
	}

	latStr := c.QueryParam("lat")
	lngStr := c.QueryParam("lng")
	maxDistStr := c.QueryParam("max_distance")

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)
	maxDistance, _ := strconv.ParseFloat(maxDistStr, 64)

	if maxDistance == 0 {
		maxDistance = 10000 // default 10km in meters
	}

	rides, err := h.service.GetNearbyRides(ctx, driverID, lat, lng, maxDistance)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, rides)
}

// AcceptRide handles driver accepting a ride
// @Summary Accept a ride request
// @Description Driver accepts a ride request
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ride_id query integer true "Ride ID to accept"
// @Success 200 {object} MessageResponse "Ride accepted successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Router /rides/accept [post]
func (h *RideHandler) AcceptRide(c echo.Context) error {
	ctx := c.Request().Context()
	rideIDStr := c.QueryParam("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	driverID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing customer ID in context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing driver ID in context"})
	}
	fmt.Println("Driver ID from context:", driverID)

	role, ok := middleware.GetUserRoleFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing role in context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
	}
	if role != "driver" {
		logger.Error(ctx, errors.New("role is not driver"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid role in context"})
	}

	err = h.service.AcceptRide(ctx, rideID, driverID)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "Ride accepted successfully"})
}

// StartRide handles starting a ride
// @Summary Start a ride
// @Description Mark a ride as started
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ride_id query integer true "Ride ID to start"
// @Success 200 {object} MessageResponse "Ride started successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /rides/start [post]
func (h *RideHandler) StartRide(c echo.Context) error {
	ctx := c.Request().Context()
	rideIDStr := c.QueryParam("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err = h.service.StartRide(c.Request().Context(), rideID)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "Ride started successfully"})
}

// CompleteRide handles completing a ride
// @Summary Complete a ride
// @Description Mark a ride as completed
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ride_id query integer true "Ride ID to complete"
// @Success 200 {object} MessageResponse "Ride completed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /rides/complete [post]
func (h *RideHandler) CompleteRide(c echo.Context) error {
	ctx := c.Request().Context()
	rideIDStr := c.QueryParam("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err = h.service.CompleteRide(ctx, rideID)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "Ride completed successfully"})
}

// CancelRide handles cancelling a ride
// @Summary Cancel a ride
// @Description Cancel an active or pending ride
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ride_id query integer true "Ride ID to cancel"
// @Success 200 {object} MessageResponse "Ride cancelled successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Router /rides/cancel [post]
func (h *RideHandler) CancelRide(c echo.Context) error {
	ctx := c.Request().Context()
	rideIDStr := c.QueryParam("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	err = h.service.CancelRide(c.Request().Context(), rideID)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, MessageResponse{Message: "Ride cancelled successfully"})
}
