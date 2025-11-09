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

type GetNearbyRidesRequest struct {
	Lat         float64 `json:"lat" validate:"required"`
	Lng         float64 `json:"lng" validate:"required"`
	MaxDistance float64 `json:"max_distance"` // in meters, default 10000
	Limit       int     `json:"limit"`        // max number of rides to return, default 50
}

// GetNearbyRides handles getting nearby rides for drivers (Short Polling Endpoint)
// @Summary Get nearby available rides for driver
// @Description Driver polls this endpoint to get available rides within a radius. Returns rides with status "requested" or "pending" updated within last 5 minutes.
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GetNearbyRidesRequest true "Driver location and search parameters"
// @Success 200 {array} domain.Ride "List of nearby available rides"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized - driver must be logged in"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /rides/nearby [post]
func (h *RideHandler) GetNearbyRides(c echo.Context) error {
	ctx := c.Request().Context()
	driverID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing driver ID in context"))
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

	var req GetNearbyRidesRequest
	if err := c.Bind(&req); err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Validate required fields
	if req.Lat == 0 || req.Lng == 0 {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "lat and lng are required"})
	}

	// Set defaults
	if req.MaxDistance == 0 {
		req.MaxDistance = 10000 // default 10km in meters
	}
	if req.Limit == 0 {
		req.Limit = 50 // default 50 rides
	}

	// Validate limits
	if req.Limit > 100 {
		req.Limit = 100 // cap at 100 rides
	}
	if req.Limit < 1 {
		req.Limit = 1 // minimum 1 ride
	}

	rides, err := h.service.GetNearbyRides(ctx, driverID, req.Lat, req.Lng, req.MaxDistance, req.Limit)
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

// GetRideDetails handles getting ride details by ride_id
// @Summary Get ride details
// @Description Get detailed information about a specific ride including customer info
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ride_id query integer true "Ride ID"
// @Success 200 {object} service.RideWithCustomerInfo "Ride details with customer information"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Ride not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /rides/details [get]
func (h *RideHandler) GetRideDetails(c echo.Context) error {
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

	// Parse ride_id from query parameter
	rideIDStr := c.QueryParam("ride_id")
	if rideIDStr == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "ride_id is required"})
	}

	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid ride_id"})
	}

	// Get ride details with customer info
	rideDetails, err := h.service.GetRideDetailsWithCustomer(ctx, rideID)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, rideDetails)
}

type RideStatusResponse struct {
	RideID      int64    `json:"ride_id"`
	CustomerID  int64    `json:"customer_id"`
	PickupLat   float64  `json:"pickup_lat"`
	PickupLng   float64  `json:"pickup_lng"`
	DropoffLat  float64  `json:"dropoff_lat"`
	DropoffLng  float64  `json:"dropoff_lng"`
	Status      string   `json:"status"`
	Fare        *float64 `json:"fare,omitempty"`
	RequestedAt string   `json:"requested_at"`
	AcceptedAt  *string  `json:"accepted_at,omitempty"`
	StartedAt   *string  `json:"started_at,omitempty"`
	CompletedAt *string  `json:"completed_at,omitempty"`
	CancelledAt *string  `json:"cancelled_at,omitempty"`

	// Driver information (only if ride is accepted/started/completed)
	Driver *DriverInfo `json:"driver,omitempty"`
}

type DriverInfo struct {
	DriverID   int64    `json:"driver_id"`
	Name       string   `json:"name"`
	Phone      string   `json:"phone"`
	VehicleNo  string   `json:"vehicle_no"`
	CurrentLat *float64 `json:"current_lat,omitempty"`  // Driver's current location
	CurrentLng *float64 `json:"current_lng,omitempty"`  // Driver's current location
	LastPingAt *string  `json:"last_ping_at,omitempty"` // Last location update time
}

type SendRideRequestToDriverRequest struct {
	RideID   int64 `json:"ride_id" validate:"required"`
	DriverID int64 `json:"driver_id" validate:"required"`
}

//
//// SendRideRequestToDriver handles sending a specific ride request to a specific driver
//// @Summary Send ride request to driver
//// @Description Send a ride request notification to a specific driver
//// @Tags Rides
//// @Accept json
//// @Produce json
//// @Security BearerAuth
//// @Param request body SendRideRequestToDriverRequest true "Ride and Driver IDs"
//// @Success 200 {object} MessageResponse "Ride request sent successfully"
//// @Failure 400 {object} ErrorResponse "Invalid request"
//// @Failure 401 {object} ErrorResponse "Unauthorized"
//// @Failure 500 {object} ErrorResponse "Internal server error"
//// @Router /rides/send-request [post]
//func (h *RideHandler) SendRideRequestToDriver(c echo.Context) error {
//
//	ctx := c.Request().Context()
//	customerID, ok := middleware.GetUserIDFromEcho(c)
//	if !ok {
//		logger.Error(ctx, errors.New("no user id from context"))
//		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing customer ID in context"})
//	}
//	fmt.Println("customer ID from context:", customerID)
//
//	role, ok := middleware.GetUserRoleFromEcho(c)
//	if !ok {
//		logger.Error(ctx, errors.New("no user role from context"))
//		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
//	}
//
//	if role != "customer" {
//		logger.Error(ctx, errors.New("invalid role"))
//		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid role"})
//	}
//
//	// Parse request body
//	var req SendRideRequestToDriverRequest
//	if err := c.Bind(&req); err != nil {
//		logger.Error(ctx, err)
//		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
//	}
//
//	// Validate required fields
//	if req.RideID == 0 || req.DriverID == 0 {
//		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "ride_id and driver_id are required"})
//	}
//
//	// Send ride request to driver
//	err := h.service.SendRideRequestToDriver(ctx, req.RideID, req.DriverID)
//	if err != nil {
//		logger.Error(ctx, err)
//		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
//	}
//
//	return c.JSON(http.StatusOK, MessageResponse{Message: "Ride request sent to driver successfully"})
//}

// GetRideStatus handles getting ride status for customers
// @Summary Get ride status for customer
// @Description Get current status of a ride including driver information and location if driver has accepted
// @Tags Rides
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ride_id query integer true "Ride ID"
// @Success 200 {object} RideStatusResponse "Ride status with driver information"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 403 {object} ErrorResponse "Forbidden - not your ride"
// @Failure 404 {object} ErrorResponse "Ride not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /rides/status [get]
func (h *RideHandler) GetRideStatus(c echo.Context) error {
	ctx := c.Request().Context()

	customerID, ok := middleware.GetUserIDFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing customer ID in context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing customer ID in context"})
	}

	role, ok := middleware.GetUserRoleFromEcho(c)
	if !ok {
		logger.Error(ctx, errors.New("missing role in context"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing role in context"})
	}

	if role != "customer" {
		logger.Error(ctx, errors.New("invalid role"))
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "only customers can check ride status"})
	}

	// Parse ride_id from query parameter
	rideIDStr := c.QueryParam("ride_id")
	if rideIDStr == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "ride_id is required"})
	}

	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		logger.Error(ctx, err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid ride_id"})
	}

	// Get ride status with driver information
	rideStatus, err := h.service.GetRideStatusForCustomer(ctx, rideID, customerID)
	if err != nil {
		logger.Error(ctx, err)
		if err.Error() == "ride not found" {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		}
		if err.Error() == "forbidden: this ride belongs to another customer" {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, rideStatus)
}
