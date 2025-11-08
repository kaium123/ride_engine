package service

import (
	"context"

	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository"
)

type LocationService struct {
	repo repository.LocationRepository
}

func NewLocationService(repo repository.LocationRepository) *LocationService {
	return &LocationService{repo: repo}
}

// UpdateDriverLocation updates driver's current location
func (s *LocationService) UpdateDriverLocation(ctx context.Context, driverID int64, lat, lng float64) error {
	return s.repo.UpdateDriverLocation(ctx, driverID, lat, lng)
}

// FindNearestDrivers finds drivers within maxDistance (in meters)
func (s *LocationService) FindNearestDrivers(ctx context.Context, lat, lng float64, maxDistance float64, limit int) ([]int64, error) {
	return s.repo.FindNearestDrivers(ctx, lat, lng, maxDistance, limit)
}
