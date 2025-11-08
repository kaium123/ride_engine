package service

import (
	"context"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/postgres"
)

type RideService struct {
	rideRepo        *postgres.RidePostgresRepository
	locationService *LocationService
}

func NewRideService(
	rideRepo *postgres.RidePostgresRepository,
	locationService *LocationService,
) *RideService {
	return &RideService{
		rideRepo:        rideRepo,
		locationService: locationService,
	}
}

// RequestRide creates a new ride request
func (s *RideService) RequestRide(ctx context.Context, customerID int64, pickupLat, pickupLng, dropoffLat, dropoffLng float64) (*domain.Ride, error) {
	ride := &domain.Ride{
		CustomerID:  customerID,
		PickupLat:   pickupLat,
		PickupLng:   pickupLng,
		DropoffLat:  dropoffLat,
		DropoffLng:  dropoffLng,
		Status:      domain.RideStatusRequested,
		RequestedAt: time.Now(),
	}

	if err := s.rideRepo.Create(ctx, ride); err != nil {
		logger.Error(ctx, "Failed to create ride: %v", err)
		return nil, err
	}

	return ride, nil
}

// GetNearbyRides finds available rides near driver's location
func (s *RideService) GetNearbyRides(ctx context.Context, driverID int64, driverLat, driverLng, maxDistance float64) ([]*domain.Ride, error) {
	rides, err := s.rideRepo.GetRequestedRides(ctx)
	if err != nil {
		logger.Error(ctx, "Failed to get requested rides: %v", err)
		return nil, err
	}

	var nearbyRides []*domain.Ride
	driverLocation := domain.Location{Latitude: driverLat, Longitude: driverLng}

	for _, ride := range rides {
		pickupLocation := domain.Location{Latitude: ride.PickupLat, Longitude: ride.PickupLng}
		distance := driverLocation.DistanceTo(pickupLocation)

		if distance <= maxDistance {
			nearbyRides = append(nearbyRides, ride)
		}
	}

	return nearbyRides, nil
}

// AcceptRide allows driver to accept a ride
func (s *RideService) AcceptRide(ctx context.Context, rideID, driverID int64) error {
	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if err := ride.Accept(driverID); err != nil {
		logger.Error(ctx, "Failed to accept ride: %v", err)
		return err
	}

	return s.rideRepo.Update(ctx, ride)
}

// StartRide starts the ride
func (s *RideService) StartRide(ctx context.Context, rideID int64) error {
	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if err := ride.Start(); err != nil {
		logger.Error(ctx, "Failed to start ride: %v", err)
		return err
	}

	return s.rideRepo.Update(ctx, ride)
}

// CompleteRide completes the ride
func (s *RideService) CompleteRide(ctx context.Context, rideID int64) error {
	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if err := ride.Complete(); err != nil {
		logger.Error(ctx, "Failed to complete ride: %v", err)
		return err
	}

	return s.rideRepo.Update(ctx, ride)
}

// CancelRide cancels the ride
func (s *RideService) CancelRide(ctx context.Context, rideID int64) error {
	ride, err := s.rideRepo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if err := ride.Cancel(); err != nil {
		logger.Error(ctx, "Failed to cancel ride: %v", err)
		return err
	}

	return s.rideRepo.Update(ctx, ride)
}

// GetRideByID retrieves a ride by ID
func (s *RideService) GetRideByID(ctx context.Context, rideID int64) (*domain.Ride, error) {
	return s.rideRepo.GetByID(ctx, rideID)
}
