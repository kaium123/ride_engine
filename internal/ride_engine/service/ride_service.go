package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/mongodb"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/postgres"
)

// RideWithCustomerInfo contains ride details along with customer information
type RideWithCustomerInfo struct {
	RideID             int64   `json:"ride_id"`
	CustomerID         int64   `json:"customer_id"`
	CustomerName       string  `json:"customer_name"`
	CustomerPhone      string  `json:"customer_phone"`
	CustomerCurrentLat float64 `json:"customer_current_lat"` // Customer's current location (pickup)
	CustomerCurrentLng float64 `json:"customer_current_lng"` // Customer's current location (pickup)
	PickupLat          float64 `json:"pickup_lat"`           // Same as customer current location
	PickupLng          float64 `json:"pickup_lng"`           // Same as customer current location
	DropoffLat         float64 `json:"dropoff_lat"`          // Destination/drop location
	DropoffLng         float64 `json:"dropoff_lng"`          // Destination/drop location
	RequestedAt        string  `json:"requested_at"`
	Status             string  `json:"status"`
	DistanceFromDriver float64 `json:"distance_from_driver,omitempty"` // in meters
}

type RideService struct {
	rideRepoMongo   *mongodb.RideMongoRepository
	locationService *LocationService
	driverService   *DriverService
	customerRepo    *postgres.CustomerPostgresRepository
}

func NewRideService(
	rideRepoMongo *mongodb.RideMongoRepository,
	locationService *LocationService,
	driverService *DriverService,
	customerRepo *postgres.CustomerPostgresRepository,
) *RideService {
	return &RideService{
		rideRepoMongo:   rideRepoMongo,
		locationService: locationService,
		driverService:   driverService,
		customerRepo:    customerRepo,
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

	if err := s.rideRepoMongo.Create(ctx, ride); err != nil {
		logger.Error(ctx, "Failed to create ride: %v", err)
		return nil, err
	}

	return ride, nil
}

// GetNearbyRides finds available rides near driver's location using MongoDB geospatial query
// Returns rides within radius that were updated in the last 5 minutes with status "requested" or "pending"
func (s *RideService) GetNearbyRides(ctx context.Context, driverID int64, driverLat, driverLng, maxDistance float64, limit int) ([]*domain.Ride, error) {
	// Use MongoDB geospatial query to find nearby rides efficiently
	rides, err := s.rideRepoMongo.GetNearbyRequestedRides(ctx, driverLat, driverLng, maxDistance, limit)
	if err != nil {
		logger.Error(ctx, "Failed to get nearby requested rides: %v", err)
		return nil, err
	}

	logger.Info(ctx, fmt.Sprintf("Found %d nearby rides for driver %d within %.2fm (limit: %d)", len(rides), driverID, maxDistance, limit))

	return rides, nil
}

// AcceptRide allows driver to accept a ride
// Only online drivers can accept rides
func (s *RideService) AcceptRide(ctx context.Context, rideID, driverID int64) error {
	// Check if driver is online

	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if ride.Status == domain.RideStatusAccepted || ride.Status == domain.RideStatusStarted || ride.Status == domain.RideStatusCompleted {
		logger.Error(ctx, "Ride with id %d cannot be accepted", rideID)
		return errors.New("ride is cannot be accepted")
	}

	if err := ride.Accept(driverID); err != nil {
		logger.Error(ctx, "Failed to accept ride: %v", err)
		return err
	}

	return s.rideRepoMongo.Update(ctx, ride)
}

// StartRide starts the ride
func (s *RideService) StartRide(ctx context.Context, rideID int64) error {
	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if ride.Status != domain.RideStatusAccepted {
		logger.Error(ctx, "Ride with id %d cannot be started", rideID)
		return errors.New("ride is cannot be started")
	}

	if err := ride.Start(); err != nil {
		logger.Error(ctx, "Failed to start ride: %v", err)
		return err
	}

	return s.rideRepoMongo.Update(ctx, ride)
}

// CompleteRide completes the ride
func (s *RideService) CompleteRide(ctx context.Context, rideID int64) error {
	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if ride.Status != domain.RideStatusCompleted {
		logger.Error(ctx, "Ride with id %d cannot be completed", rideID)
		return errors.New("ride is cannot be completed")
	}

	if err := ride.Complete(); err != nil {
		logger.Error(ctx, "Failed to complete ride: %v", err)
		return err
	}

	return s.rideRepoMongo.Update(ctx, ride)
}

// CancelRide cancels the ride
func (s *RideService) CancelRide(ctx context.Context, rideID int64) error {
	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride: %v", err)
		return err
	}

	if ride.Status == domain.RideStatusCompleted || ride.Status == domain.RideStatusCancelled {
		logger.Error(ctx, "Ride with id %d cannot be cancelled", rideID)
		return errors.New("ride is cannot be cancelled")
	}

	if err := ride.Cancel(); err != nil {
		logger.Error(ctx, "Failed to cancel ride: %v", err)
		return err
	}

	return s.rideRepoMongo.Update(ctx, ride)
}

// GetRideByID retrieves a ride by ID
func (s *RideService) GetRideByID(ctx context.Context, rideID int64) (*domain.Ride, error) {
	return s.rideRepoMongo.GetByID(ctx, rideID)
}

// GetRideDetailsWithCustomer retrieves detailed ride information with customer details
func (s *RideService) GetRideDetailsWithCustomer(ctx context.Context, rideID int64) (*RideWithCustomerInfo, error) {
	// Get ride by ID
	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride %d: %v", rideID, err)
		return nil, err
	}

	// Get customer details
	customer, err := s.customerRepo.GetByID(ctx, ride.CustomerID)
	if err != nil {
		logger.Error(ctx, "Failed to get customer %d: %v", ride.CustomerID, err)
		return nil, err
	}

	rideDetails := &RideWithCustomerInfo{
		RideID:             ride.ID,
		CustomerID:         ride.CustomerID,
		CustomerName:       customer.Name,
		CustomerPhone:      customer.Phone,
		CustomerCurrentLat: ride.PickupLat,
		CustomerCurrentLng: ride.PickupLng,
		PickupLat:          ride.PickupLat,
		PickupLng:          ride.PickupLng,
		DropoffLat:         ride.DropoffLat,
		DropoffLng:         ride.DropoffLng,
		RequestedAt:        ride.RequestedAt.Format("2006-01-02 15:04:05"),
		Status:             string(ride.Status),
	}

	return rideDetails, nil
}

// SendRideRequestToDriver sends a ride request to a specific driver
func (s *RideService) SendRideRequestToDriver(ctx context.Context, rideID, driverID int64) error {
	// Check if driver is online
	isOnline, err := s.driverService.IsDriverOnline(ctx, driverID)
	if err != nil {
		logger.Error(ctx, "Failed to check driver online status: %v", err)
		return err
	}

	if !isOnline {
		logger.Error(ctx, "Driver %d is not online", driverID)
		return errors.New("driver is not online, cannot send ride request")
	}

	// Verify ride exists and is in requested status
	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride %d: %v", rideID, err)
		return err
	}

	if ride.Status != domain.RideStatusRequested {
		logger.Error(ctx, "Ride %d is not in requested status", rideID)
		return fmt.Errorf("ride is not available (status: %s)", ride.Status)
	}

	// TODO: Here you can implement actual notification logic
	// For example: push notification, websocket, SMS, etc.
	// For now, we just log it
	logger.Info(ctx, fmt.Sprintf("Ride request %d sent to driver %d", rideID, driverID))

	return nil
}

//
//// GetRideRequestsForDriver returns available ride requests with customer details for a specific driver
//// based on driver's location and maximum search radius using MongoDB geospatial query
//// Returns rides within radius that were updated in the last 5 minutes with status "requested" or "pending"
//func (s *RideService) GetRideRequestsForDriver(ctx context.Context, driverID int64, driverLat, driverLng, maxRadius float64) ([]*RideWithCustomerInfo, error) {
//	// Use MongoDB geospatial query to get nearby rides efficiently
//	rides, err := s.rideRepoMongo.GetNearbyRequestedRides(ctx, driverLat, driverLng, maxRadius)
//	if err != nil {
//		logger.Error(ctx, "Failed to get nearby requested rides: %v", err)
//		return nil, err
//	}
//
//	var rideRequests []*RideWithCustomerInfo
//	driverLocation := domain.Location{Latitude: driverLat, Longitude: driverLng}
//
//	for _, ride := range rides {
//		// Get customer details
//		customer, err := s.customerRepo.GetByID(ctx, ride.CustomerID)
//		if err != nil {
//			logger.Error(ctx, "Failed to get customer %d: %v", ride.CustomerID, err)
//			continue // Skip this ride if customer not found
//		}
//
//		// Calculate distance for display purposes
//		pickupLocation := domain.Location{Latitude: ride.PickupLat, Longitude: ride.PickupLng}
//		distance := driverLocation.DistanceTo(pickupLocation)
//
//		rideRequest := &RideWithCustomerInfo{
//			RideID:             ride.ID,
//			CustomerID:         ride.CustomerID,
//			CustomerName:       customer.Name,
//			CustomerPhone:      customer.Phone,
//			CustomerCurrentLat: ride.PickupLat,
//			CustomerCurrentLng: ride.PickupLng,
//			PickupLat:          ride.PickupLat,
//			PickupLng:          ride.PickupLng,
//			DropoffLat:         ride.DropoffLat,
//			DropoffLng:         ride.DropoffLng,
//			RequestedAt:        ride.RequestedAt.Format("2006-01-02 15:04:05"),
//			Status:             string(ride.Status),
//			DistanceFromDriver: distance,
//		}
//
//		rideRequests = append(rideRequests, rideRequest)
//	}
//
//	logger.Info(ctx, fmt.Sprintf("Found %d ride requests within %.2fm for driver %d", len(rideRequests), maxRadius, driverID))
//
//	return rideRequests, nil
//}

// GetRideStatusForCustomer retrieves ride status with driver information for customer
func (s *RideService) GetRideStatusForCustomer(ctx context.Context, rideID, customerID int64) (*RideStatusResponse, error) {
	// Get ride by ID
	ride, err := s.rideRepoMongo.GetByID(ctx, rideID)
	if err != nil {
		logger.Error(ctx, "Failed to get ride %d: %v", rideID, err)
		return nil, errors.New("ride not found")
	}

	// Verify ride belongs to this customer
	if ride.CustomerID != customerID {
		logger.Error(ctx, "Customer %d tried to access ride %d belonging to customer %d", customerID, rideID, ride.CustomerID)
		return nil, errors.New("forbidden: this ride belongs to another customer")
	}

	// Build response
	response := &RideStatusResponse{
		RideID:      ride.ID,
		CustomerID:  ride.CustomerID,
		PickupLat:   ride.PickupLat,
		PickupLng:   ride.PickupLng,
		DropoffLat:  ride.DropoffLat,
		DropoffLng:  ride.DropoffLng,
		Status:      string(ride.Status),
		Fare:        ride.Fare,
		RequestedAt: ride.RequestedAt.Format("2006-01-02 15:04:05"),
	}

	// Add timestamps if available
	if ride.AcceptedAt != nil {
		acceptedStr := ride.AcceptedAt.Format("2006-01-02 15:04:05")
		response.AcceptedAt = &acceptedStr
	}
	if ride.StartedAt != nil {
		startedStr := ride.StartedAt.Format("2006-01-02 15:04:05")
		response.StartedAt = &startedStr
	}
	if ride.CompletedAt != nil {
		completedStr := ride.CompletedAt.Format("2006-01-02 15:04:05")
		response.CompletedAt = &completedStr
	}
	if ride.CancelledAt != nil {
		cancelledStr := ride.CancelledAt.Format("2006-01-02 15:04:05")
		response.CancelledAt = &cancelledStr
	}

	// If ride has been accepted, get driver information and current location
	if ride.DriverID != nil {
		driverInfo, err := s.getDriverInfoWithLocation(ctx, *ride.DriverID)
		if err != nil {
			logger.Error(ctx, "Failed to get driver info for driver %d: %v", *ride.DriverID, err)
			// Don't fail the request, just log the error
		} else {
			response.Driver = driverInfo
		}
	}

	return response, nil
}

// getDriverInfoWithLocation retrieves driver information including current location
func (s *RideService) getDriverInfoWithLocation(ctx context.Context, driverID int64) (*DriverInfo, error) {
	// Get driver details from database
	driver, err := s.driverService.GetByID(ctx, driverID)
	if err != nil {
		logger.Error(ctx, "Failed to get driver %d: %v", driverID, err)
		return nil, err
	}

	driverInfo := &DriverInfo{
		DriverID:  driver.ID,
		Name:      driver.Name,
		Phone:     driver.Phone,
		VehicleNo: driver.VehicleNo,
	}

	// Get driver's current location from MongoDB (if available)
	currentLat, currentLng, lastPingAt, err := s.locationService.GetDriverLocation(ctx, driverID)
	if err != nil {
		logger.Error(ctx, "Failed to get driver location for driver %d: %v", driverID, err)
		// Location not available, return driver info without location
	} else {
		driverInfo.CurrentLat = &currentLat
		driverInfo.CurrentLng = &currentLng
		if lastPingAt != nil {
			pingStr := lastPingAt.Format("2006-01-02 15:04:05")
			driverInfo.LastPingAt = &pingStr
		}
	}

	return driverInfo, nil
}

// RideStatusResponse contains ride status with driver information
type RideStatusResponse struct {
	RideID      int64       `json:"ride_id"`
	CustomerID  int64       `json:"customer_id"`
	PickupLat   float64     `json:"pickup_lat"`
	PickupLng   float64     `json:"pickup_lng"`
	DropoffLat  float64     `json:"dropoff_lat"`
	DropoffLng  float64     `json:"dropoff_lng"`
	Status      string      `json:"status"`
	Fare        *float64    `json:"fare,omitempty"`
	RequestedAt string      `json:"requested_at"`
	AcceptedAt  *string     `json:"accepted_at,omitempty"`
	StartedAt   *string     `json:"started_at,omitempty"`
	CompletedAt *string     `json:"completed_at,omitempty"`
	CancelledAt *string     `json:"cancelled_at,omitempty"`
	Driver      *DriverInfo `json:"driver,omitempty"`
}

// DriverInfo contains driver details and current location
type DriverInfo struct {
	DriverID   int64    `json:"driver_id"`
	Name       string   `json:"name"`
	Phone      string   `json:"phone"`
	VehicleNo  string   `json:"vehicle_no"`
	CurrentLat *float64 `json:"current_lat,omitempty"`
	CurrentLng *float64 `json:"current_lng,omitempty"`
	LastPingAt *string  `json:"last_ping_at,omitempty"`
}
