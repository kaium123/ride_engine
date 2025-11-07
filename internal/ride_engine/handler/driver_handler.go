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
func (h *DriverHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req RegisterDriverRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	driver, err := h.service.Register(r.Context(), req.Name, req.Phone, req.VehicleNo)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendJSON(w, http.StatusCreated, driver)
}

// RequestOTP handles OTP generation and sending
func (h *DriverHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req RequestOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.RequestOTP(r.Context(), req.Phone)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendMessage(w, http.StatusOK, "OTP sent successfully")
}

// VerifyOTP handles OTP verification and login
func (h *DriverHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	driver, token, err := h.service.VerifyOTP(r.Context(), req.Phone, req.OTP)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err)
		return
	}

	SendJSON(w, http.StatusOK, AuthResponse{
		Customer: driver,
		Token:    token,
	})
}

// UpdateLocation handles driver location updates
func (h *DriverHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
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

	var req UpdateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.UpdateLocation(r.Context(), driverID, req.Latitude, req.Longitude)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}

	SendMessage(w, http.StatusOK, "Location updated successfully")
}

// SetOnlineStatus handles driver online/offline status
func (h *DriverHandler) SetOnlineStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
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

	var req SetOnlineStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	err := h.service.SetOnlineStatus(r.Context(), driverID, req.IsOnline)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err)
		return
	}

	status := "offline"
	if req.IsOnline {
		status = "online"
	}
	SendMessage(w, http.StatusOK, "Driver is now "+status)
}

func (h *DriverHandler) FindNearestDrivers(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	radiusStr := r.URL.Query().Get("radius")
	limitStr := r.URL.Query().Get("limit")

	if latStr == "" || lngStr == "" {
		http.Error(w, "lat and lng are required", http.StatusBadRequest)
		return
	}

	lat, err1 := strconv.ParseFloat(latStr, 64)
	lng, err2 := strconv.ParseFloat(lngStr, 64)
	if err1 != nil || err2 != nil {
		http.Error(w, "invalid coordinates", http.StatusBadRequest)
		return
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

	driverIDs, err := h.service.GetNearestDrivers(r.Context(), lat, lng, radius, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"drivers": driverIDs,
		"count":   len(driverIDs),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
