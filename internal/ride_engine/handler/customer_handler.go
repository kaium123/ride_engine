package handler

import (
	"encoding/json"
	"net/http"

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
func (h *CustomerHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req RegisterCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	customer, token, err := h.service.Register(r.Context(), req.Name, req.Email, req.Phone, req.Password)
	if err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	SendJSON(w, http.StatusCreated, AuthResponse{
		Customer: customer,
		Token:    token,
	})
}

// Login handles customer login
func (h *CustomerHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendError(w, http.StatusMethodNotAllowed, http.ErrNotSupported)
		return
	}

	var req LoginCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, err)
		return
	}

	customer, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err)
		return
	}

	SendJSON(w, http.StatusOK, AuthResponse{
		Customer: customer,
		Token:    token,
	})
}
