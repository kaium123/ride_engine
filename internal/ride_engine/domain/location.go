package domain

import (
	"errors"
	"math"
)

// Location represents a geographical location
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Validation errors
var (
	ErrInvalidLatitude  = errors.New("invalid latitude")
	ErrInvalidLongitude = errors.New("invalid longitude")
)

// Validate validates the location
func (l *Location) Validate() error {
	if l.Latitude < -90 || l.Latitude > 90 {
		return ErrInvalidLatitude
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return ErrInvalidLongitude
	}
	return nil
}

// DistanceTo calculates the distance between two locations in kilometers using Haversine formula
func (l *Location) DistanceTo(other Location) float64 {
	const earthRadius = 6371 // Earth's radius in kilometers

	lat1 := l.Latitude * math.Pi / 180
	lat2 := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - l.Latitude) * math.Pi / 180
	deltaLng := (other.Longitude - l.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
