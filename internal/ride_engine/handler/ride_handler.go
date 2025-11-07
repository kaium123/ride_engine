package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"vcs.technonext.com/carrybee/ride_engine/pkg/middleware"

	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/service"
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
func (h *RideHandler) RequestRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	fmt.Println("sdf")
	customerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, errors.New("missing customer ID in context"))
		return
	}
	fmt.Println("customer ID from context:", customerID)

	var req RequestRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	ride, err := h.service.RequestRide(r.Context(), customerID, req.PickupLat, req.PickupLng, req.DropoffLat, req.DropoffLng)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}

	SendJSON(w, http.StatusCreated, ride)
}

// GetNearbyRides handles getting nearby rides for drivers
func (h *RideHandler) GetNearbyRides(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	driverID, ok := middleware.GetUserID(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, errors.New("missing driver ID in context"))
		return
	}
	fmt.Println("Driver ID from context:", driverID)

	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, errors.New("missing role in context"))
		return
	}
	if role != "driver" {
		SendError(w, http.StatusUnauthorized, errors.New("invalid role in context"))
		return
	}

	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	maxDistStr := r.URL.Query().Get("max_distance")

	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)
	maxDistance, _ := strconv.ParseFloat(maxDistStr, 64)

	if maxDistance == 0 {
		maxDistance = 10000 // default 10km in meters
	}

	rides, err := h.service.GetNearbyRides(r.Context(), driverID, lat, lng, maxDistance)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}

	SendJSON(w, http.StatusOK, rides)
}

// AcceptRide handles driver accepting a ride
func (h *RideHandler) AcceptRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	rideIDStr := r.URL.Query().Get("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	driverID, ok := middleware.GetUserID(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, errors.New("missing driver ID in context"))
		return
	}
	fmt.Println("Driver ID from context:", driverID)

	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		SendError(w, http.StatusUnauthorized, errors.New("missing role in context"))
		return
	}
	if role != "driver" {
		SendError(w, http.StatusUnauthorized, errors.New("invalid role in context"))
		return
	}

	err = h.service.AcceptRide(r.Context(), rideID, driverID)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendMessage(w, http.StatusOK, "Ride accepted successfully")
}

// StartRide handles starting a ride
func (h *RideHandler) StartRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	rideIDStr := r.URL.Query().Get("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.StartRide(r.Context(), rideID)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendMessage(w, http.StatusOK, "Ride started successfully")
}

// CompleteRide handles completing a ride
func (h *RideHandler) CompleteRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	rideIDStr := r.URL.Query().Get("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.CompleteRide(r.Context(), rideID)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendMessage(w, http.StatusOK, "Ride completed successfully")
}

// CancelRide handles cancelling a ride
func (h *RideHandler) CancelRide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	rideIDStr := r.URL.Query().Get("ride_id")
	rideID, err := strconv.ParseInt(rideIDStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	err = h.service.CancelRide(r.Context(), rideID)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendMessage(w, http.StatusOK, "Ride cancelled successfully")
}
