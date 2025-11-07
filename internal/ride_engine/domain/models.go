package domain

import (
	"errors"
	"time"
)

// UserType represents the type of user
type UserType string

const (
	UserTypeCustomer UserType = "customer"
	UserTypeDriver   UserType = "driver"
)

// User represents a base user in the system
type User struct {
	ID        int64     `json:"id"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email,omitempty"`
	Type      UserType  `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// Customer represents a customer/rider
type Customer struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

// Driver represents a driver
type Driver struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Phone         string     `json:"phone"`
	VehicleNo     string     `json:"vehicle_no"`
	IsOnline      bool       `json:"is_online"`
	CurrentLat    *float64   `json:"current_lat,omitempty"`
	CurrentLng    *float64   `json:"current_lng,omitempty"`
	LastPingAt    *time.Time `json:"last_ping_at,omitempty"`
	LastUpdatedAt *time.Time `json:"last_updated_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// RideStatus represents the status of a ride
type RideStatus string

const (
	RideStatusRequested RideStatus = "requested"
	RideStatusAccepted  RideStatus = "accepted"
	RideStatusStarted   RideStatus = "started"
	RideStatusCompleted RideStatus = "completed"
	RideStatusCancelled RideStatus = "cancelled"
)

// Ride represents a ride request
type Ride struct {
	ID              int64      `json:"id"`
	CustomerID      int64      `json:"customer_id"`
	DriverID        *int64     `json:"driver_id,omitempty"`
	PickupLat       float64    `json:"pickup_lat"`
	PickupLng       float64    `json:"pickup_lng"`
	DropoffLat      float64    `json:"dropoff_lat"`
	DropoffLng      float64    `json:"dropoff_lng"`
	Status          RideStatus `json:"status"`
	Fare            *float64   `json:"fare,omitempty"`
	RequestedAt     time.Time  `json:"requested_at"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
	PickupLocation  Location   `json:"-"`
	DropoffLocation Location   `json:"-"`
}

// Validation errors
var (
	ErrInvalidPhone      = errors.New("invalid phone number")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidUserType   = errors.New("invalid user type")
	ErrInvalidRideStatus = errors.New("invalid ride status")
)

// ValidateCustomer validates customer data
func ValidateCustomer(c *Customer) error {
	if c.Phone == "" {
		return ErrInvalidPhone
	}
	if c.Email == "" {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateDriver validates driver data
func ValidateDriver(d *Driver) error {
	if d.Phone == "" {
		return ErrInvalidPhone
	}
	//if d.Name == "" {
	//	return errors.New("driver name is required")
	//}
	return nil
}

// Accept marks the ride as accepted by a driver
func (r *Ride) Accept(driverID int64) error {
	if r.Status != RideStatusRequested {
		return errors.New("ride is not in requested status")
	}
	now := time.Now()
	r.DriverID = &driverID
	r.Status = RideStatusAccepted
	r.AcceptedAt = &now
	return nil
}

// Start marks the ride as started
func (r *Ride) Start() error {
	if r.Status != RideStatusAccepted {
		return errors.New("ride must be accepted before starting")
	}
	now := time.Now()
	r.Status = RideStatusStarted
	r.StartedAt = &now
	return nil
}

// Complete marks the ride as completed
func (r *Ride) Complete() error {
	if r.Status != RideStatusStarted {
		return errors.New("ride must be started before completing")
	}
	now := time.Now()
	r.Status = RideStatusCompleted
	r.CompletedAt = &now
	return nil
}

// Cancel marks the ride as cancelled
func (r *Ride) Cancel() error {
	if r.Status == RideStatusCompleted {
		return errors.New("cannot cancel completed ride")
	}
	now := time.Now()
	r.Status = RideStatusCancelled
	r.CancelledAt = &now
	return nil
}
