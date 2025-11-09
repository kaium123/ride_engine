package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
)

// Note: These tests are simplified unit tests that test the domain logic
// For full integration tests with MongoDB, see the repository layer tests
// The RideService has concrete dependencies (not interfaces), so we focus on
// testing the business logic and domain model behavior here

func TestRide_Accept(t *testing.T) {
	ride := &domain.Ride{
		ID:          1,
		CustomerID:  123,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	driverID := int64(456)
	err := ride.Accept(driverID)

	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusAccepted, ride.Status)
	assert.NotNil(t, ride.DriverID)
	assert.Equal(t, driverID, *ride.DriverID)
	assert.NotNil(t, ride.AcceptedAt)
}

func TestRide_Accept_Pending(t *testing.T) {
	ride := &domain.Ride{
		ID:          1,
		CustomerID:  123,
		Status:      domain.RideStatusPending,
		RequestedAt: time.Now(),
	}

	driverID := int64(456)
	err := ride.Accept(driverID)

	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusAccepted, ride.Status)
	assert.NotNil(t, ride.DriverID)
	assert.Equal(t, driverID, *ride.DriverID)
}

func TestRide_Accept_AlreadyAccepted(t *testing.T) {
	existingDriverID := int64(789)
	ride := &domain.Ride{
		ID:         1,
		CustomerID: 123,
		Status:     domain.RideStatusAccepted,
		DriverID:   &existingDriverID,
	}

	driverID := int64(456)
	err := ride.Accept(driverID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in requested or pending status")
	assert.Equal(t, domain.RideStatusAccepted, ride.Status)
	assert.Equal(t, existingDriverID, *ride.DriverID)
}

func TestRide_Start(t *testing.T) {
	driverID := int64(456)
	ride := &domain.Ride{
		ID:         1,
		CustomerID: 123,
		Status:     domain.RideStatusAccepted,
		DriverID:   &driverID,
	}

	err := ride.Start()

	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusStarted, ride.Status)
	assert.NotNil(t, ride.StartedAt)
}

func TestRide_Start_NotAccepted(t *testing.T) {
	ride := &domain.Ride{
		ID:         1,
		CustomerID: 123,
		Status:     domain.RideStatusRequested,
	}

	err := ride.Start()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ride must be accepted before starting")
	assert.Equal(t, domain.RideStatusRequested, ride.Status)
}

func TestRide_Complete(t *testing.T) {
	driverID := int64(456)
	now := time.Now()
	ride := &domain.Ride{
		ID:         1,
		CustomerID: 123,
		Status:     domain.RideStatusStarted,
		DriverID:   &driverID,
		StartedAt:  &now,
	}

	err := ride.Complete()

	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusCompleted, ride.Status)
	assert.NotNil(t, ride.CompletedAt)
}

func TestRide_Complete_NotStarted(t *testing.T) {
	driverID := int64(456)
	ride := &domain.Ride{
		ID:         1,
		CustomerID: 123,
		Status:     domain.RideStatusAccepted,
		DriverID:   &driverID,
	}

	err := ride.Complete()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ride must be started before completing")
	assert.Equal(t, domain.RideStatusAccepted, ride.Status)
}

func TestRide_Cancel_Requested(t *testing.T) {
	ride := &domain.Ride{
		ID:          1,
		CustomerID:  123,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	err := ride.Cancel()

	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusCancelled, ride.Status)
	assert.NotNil(t, ride.CancelledAt)
}

func TestRide_Cancel_Accepted(t *testing.T) {
	driverID := int64(456)
	now := time.Now()
	ride := &domain.Ride{
		ID:         1,
		CustomerID: 123,
		Status:     domain.RideStatusAccepted,
		DriverID:   &driverID,
		AcceptedAt: &now,
	}

	err := ride.Cancel()

	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusCancelled, ride.Status)
	assert.NotNil(t, ride.CancelledAt)
}

func TestRide_Cancel_AlreadyCompleted(t *testing.T) {
	driverID := int64(456)
	now := time.Now()
	ride := &domain.Ride{
		ID:          1,
		CustomerID:  123,
		Status:      domain.RideStatusCompleted,
		DriverID:    &driverID,
		CompletedAt: &now,
	}

	err := ride.Cancel()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel completed ride")
	assert.Equal(t, domain.RideStatusCompleted, ride.Status)
}

func TestValidateDriver(t *testing.T) {
	tests := []struct {
		name      string
		driver    *domain.Driver
		shouldErr bool
	}{
		{
			name: "Valid driver",
			driver: &domain.Driver{
				Name:      "John Doe",
				Phone:     "1234567890",
				VehicleNo: "ABC-123",
			},
			shouldErr: false,
		},
		{
			name: "Missing phone",
			driver: &domain.Driver{
				Name:      "John Doe",
				Phone:     "",
				VehicleNo: "ABC-123",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateDriver(tt.driver)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCustomer(t *testing.T) {
	tests := []struct {
		name      string
		customer  *domain.Customer
		shouldErr bool
	}{
		{
			name: "Valid customer",
			customer: &domain.Customer{
				Name:  "Jane Doe",
				Phone: "9876543210",
				Email: "jane@example.com",
			},
			shouldErr: false,
		},
		{
			name: "Missing phone",
			customer: &domain.Customer{
				Name:  "Jane Doe",
				Phone: "",
				Email: "jane@example.com",
			},
			shouldErr: true,
		},
		{
			name: "Missing email",
			customer: &domain.Customer{
				Name:  "Jane Doe",
				Phone: "9876543210",
				Email: "",
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateCustomer(tt.customer)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRideStatusTransitions(t *testing.T) {
	// Test complete ride lifecycle
	ride := &domain.Ride{
		ID:          1,
		CustomerID:  123,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	// Step 1: Accept ride
	driverID := int64(456)
	err := ride.Accept(driverID)
	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusAccepted, ride.Status)

	// Step 2: Start ride
	err = ride.Start()
	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusStarted, ride.Status)

	// Step 3: Complete ride
	err = ride.Complete()
	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusCompleted, ride.Status)

	// Verify all timestamps are set
	assert.NotNil(t, ride.AcceptedAt)
	assert.NotNil(t, ride.StartedAt)
	assert.NotNil(t, ride.CompletedAt)
	assert.Nil(t, ride.CancelledAt)
}

func TestRideStatusTransitions_CancelAfterAccept(t *testing.T) {
	// Test cancellation after acceptance
	ride := &domain.Ride{
		ID:          1,
		CustomerID:  123,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	// Accept ride
	driverID := int64(456)
	err := ride.Accept(driverID)
	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusAccepted, ride.Status)

	// Cancel ride
	err = ride.Cancel()
	assert.NoError(t, err)
	assert.Equal(t, domain.RideStatusCancelled, ride.Status)
	assert.NotNil(t, ride.CancelledAt)
}

func TestRideStatusTransitions_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus domain.RideStatus
		action        string
		shouldErr     bool
	}{
		{
			name:          "Cannot start requested ride",
			initialStatus: domain.RideStatusRequested,
			action:        "start",
			shouldErr:     true,
		},
		{
			name:          "Cannot complete accepted ride",
			initialStatus: domain.RideStatusAccepted,
			action:        "complete",
			shouldErr:     true,
		},
		{
			name:          "Cannot cancel completed ride",
			initialStatus: domain.RideStatusCompleted,
			action:        "cancel",
			shouldErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ride := &domain.Ride{
				ID:          1,
				CustomerID:  123,
				Status:      tt.initialStatus,
				RequestedAt: time.Now(),
			}

			var err error
			switch tt.action {
			case "start":
				err = ride.Start()
			case "complete":
				err = ride.Complete()
			case "cancel":
				err = ride.Cancel()
			}

			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
