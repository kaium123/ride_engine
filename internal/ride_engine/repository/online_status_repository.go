package repository

import (
	"context"
	"time"
)

// OnlineDriver represents an online driver record
type OnlineDriver struct {
	DriverID     int64     `json:"driver_id"`
	IsOnline     bool      `json:"is_online"`
	LastPingAt   time.Time `json:"last_ping_at"`
	WentOnlineAt time.Time `json:"went_online_at"`
	CurrentLat   *float64  `json:"current_lat,omitempty"`
	CurrentLng   *float64  `json:"current_lng,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OnlineStatusRepository interface {
	UpsertOnlineDriver(ctx context.Context, driverID int64, lat, lng float64) error
	SetDriverOffline(ctx context.Context, driverID int64) error
	IsDriverOnline(ctx context.Context, driverID int64) (bool, error)
	GetOnlineDrivers(ctx context.Context) ([]int64, error)
	RemoveInactiveDrivers(ctx context.Context, cutoffTime time.Time) error
	GetOnlineDriversByIDs(ctx context.Context, driverIDs []int64) ([]int64, error)
}
