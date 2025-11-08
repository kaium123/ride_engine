package repository

import (
	"context"
	"time"
)

// DriverLocation represents a driver's location in the system
type DriverLocation struct {
	DriverID  int64     `bson:"driver_id"`
	Location  GeoJSON   `bson:"location"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// GeoJSON represents a GeoJSON Point
type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"` // [longitude, latitude]
}

type LocationRepository interface {
	UpdateDriverLocation(ctx context.Context, driverID int64, lat, lng float64) error
	FindNearestDrivers(ctx context.Context, lat, lng float64, maxDistance float64, limit int) ([]int64, error)
}
