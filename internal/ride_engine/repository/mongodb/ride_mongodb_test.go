package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
)

// setupTestDB creates a test MongoDB connection
func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()

	// Connect to test MongoDB instance
	clientOptions := options.Client().ApplyURI("mongodb://root:secret@localhost:27016/?authSource=admin")
	client, err := mongo.Connect(ctx, clientOptions)
	require.NoError(t, err)

	// Use a test database
	db := client.Database("ride_engine_test")

	// Cleanup function
	cleanup := func() {
		// Drop test database
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return db, cleanup
}

func TestRideMongoRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	ride := &domain.Ride{
		CustomerID:  123,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	err := repo.Create(ctx, ride)
	assert.NoError(t, err)
	assert.NotZero(t, ride.ID, "Ride ID should be generated")
}

func TestRideMongoRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Create a ride first
	ride := &domain.Ride{
		CustomerID:  123,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	err := repo.Create(ctx, ride)
	require.NoError(t, err)

	// Get the ride by ID
	retrieved, err := repo.GetByID(ctx, ride.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, ride.ID, retrieved.ID)
	assert.Equal(t, ride.CustomerID, retrieved.CustomerID)
	assert.Equal(t, ride.Status, retrieved.Status)
}

func TestRideMongoRepository_GetByID_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Try to get non-existent ride
	retrieved, err := repo.GetByID(ctx, 99999)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Equal(t, ErrRideNotFound, err)
}

func TestRideMongoRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Create a ride
	ride := &domain.Ride{
		CustomerID:  123,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	err := repo.Create(ctx, ride)
	require.NoError(t, err)

	// Accept the ride
	driverID := int64(456)
	err = ride.Accept(driverID)
	require.NoError(t, err)

	// Update the ride
	err = repo.Update(ctx, ride)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, ride.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.RideStatusAccepted, updated.Status)
	assert.NotNil(t, updated.DriverID)
	assert.Equal(t, driverID, *updated.DriverID)
	assert.NotNil(t, updated.AcceptedAt)
}

func TestRideMongoRepository_GetRequestedRides(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Create multiple rides with different statuses
	rides := []*domain.Ride{
		{
			CustomerID:  1,
			PickupLat:   23.8100,
			PickupLng:   90.4120,
			DropoffLat:  23.7509,
			DropoffLng:  90.3761,
			Status:      domain.RideStatusRequested,
			RequestedAt: time.Now(),
		},
		{
			CustomerID:  2,
			PickupLat:   23.8200,
			PickupLng:   90.4220,
			DropoffLat:  23.7609,
			DropoffLng:  90.3861,
			Status:      domain.RideStatusRequested,
			RequestedAt: time.Now(),
		},
		{
			CustomerID:  3,
			PickupLat:   23.8300,
			PickupLng:   90.4320,
			DropoffLat:  23.7709,
			DropoffLng:  90.3961,
			Status:      domain.RideStatusAccepted,
			RequestedAt: time.Now(),
		},
	}

	for _, ride := range rides {
		err := repo.Create(ctx, ride)
		require.NoError(t, err)
	}

	// Get requested rides
	requested, err := repo.GetRequestedRides(ctx)
	assert.NoError(t, err)
	assert.Len(t, requested, 2, "Should return only requested rides")
}

func TestRideMongoRepository_GetNearbyRequestedRides(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Create rides at different locations
	now := time.Now()

	// Ride near driver (within 5km)
	nearRide := &domain.Ride{
		CustomerID:  1,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: now,
	}
	err := repo.Create(ctx, nearRide)
	require.NoError(t, err)

	// Ride far from driver (> 5km)
	farRide := &domain.Ride{
		CustomerID:  2,
		PickupLat:   23.9000, // ~10km away
		PickupLng:   90.5000,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: now,
	}
	err = repo.Create(ctx, farRide)
	require.NoError(t, err)

	// Driver location
	driverLat := 23.8103
	driverLng := 90.4125
	maxDistance := 5000.0 // 5km

	// Get nearby rides
	nearby, err := repo.GetNearbyRequestedRides(ctx, driverLat, driverLng, maxDistance, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, nearby, "Should find at least one nearby ride")

	// Verify the near ride is in results
	found := false
	for _, ride := range nearby {
		if ride.ID == nearRide.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Near ride should be in results")
}

func TestRideMongoRepository_GetNearbyRequestedRides_TimeFilter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Create a fresh ride (within 5 minutes)
	freshRide := &domain.Ride{
		CustomerID:  1,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}
	err := repo.Create(ctx, freshRide)
	require.NoError(t, err)

	// Driver location
	driverLat := 23.8103
	driverLng := 90.4125
	maxDistance := 10000.0

	// Get nearby rides
	nearby, err := repo.GetNearbyRequestedRides(ctx, driverLat, driverLng, maxDistance, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, nearby, "Should find fresh ride")
}

func TestRideMongoRepository_GetNearbyRequestedRides_WithLimit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	// Create 10 rides at same location
	for i := 0; i < 10; i++ {
		ride := &domain.Ride{
			CustomerID:  int64(i + 1),
			PickupLat:   23.8100,
			PickupLng:   90.4120,
			DropoffLat:  23.7509,
			DropoffLng:  90.3761,
			Status:      domain.RideStatusRequested,
			RequestedAt: time.Now(),
		}
		err := repo.Create(ctx, ride)
		require.NoError(t, err)
	}

	// Get nearby rides with limit of 5
	nearby, err := repo.GetNearbyRequestedRides(ctx, 23.8103, 90.4125, 10000.0, 5)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len(nearby), 5, "Should respect limit")
}

func TestRideMongoRepository_GetByCustomerID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	customerID := int64(123)

	// Create rides for different customers
	for i := 0; i < 3; i++ {
		ride := &domain.Ride{
			CustomerID:  customerID,
			PickupLat:   23.8100,
			PickupLng:   90.4120,
			DropoffLat:  23.7509,
			DropoffLng:  90.3761,
			Status:      domain.RideStatusRequested,
			RequestedAt: time.Now(),
		}
		err := repo.Create(ctx, ride)
		require.NoError(t, err)
	}

	// Create ride for different customer
	otherRide := &domain.Ride{
		CustomerID:  456,
		PickupLat:   23.8100,
		PickupLng:   90.4120,
		DropoffLat:  23.7509,
		DropoffLng:  90.3761,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}
	err := repo.Create(ctx, otherRide)
	require.NoError(t, err)

	// Get rides by customer ID
	rides, err := repo.GetByCustomerID(ctx, customerID)
	assert.NoError(t, err)
	assert.Len(t, rides, 3, "Should return only customer's rides")
}

func TestRideMongoRepository_GetByDriverID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRideMongoRepository(db)
	ctx := context.Background()

	driverID := int64(456)

	// Create and accept rides
	for i := 0; i < 2; i++ {
		ride := &domain.Ride{
			CustomerID:  int64(i + 1),
			PickupLat:   23.8100,
			PickupLng:   90.4120,
			DropoffLat:  23.7509,
			DropoffLng:  90.3761,
			Status:      domain.RideStatusRequested,
			RequestedAt: time.Now(),
		}
		err := repo.Create(ctx, ride)
		require.NoError(t, err)

		// Accept ride
		err = ride.Accept(driverID)
		require.NoError(t, err)
		err = repo.Update(ctx, ride)
		require.NoError(t, err)
	}

	// Get rides by driver ID
	rides, err := repo.GetByDriverID(ctx, driverID)
	assert.NoError(t, err)
	assert.Len(t, rides, 2, "Should return driver's rides")
}
