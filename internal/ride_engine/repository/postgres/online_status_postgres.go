package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository"
)

// OnlineDriverModel represents the online_drivers table
type OnlineDriverModel struct {
	DriverID     int64     `gorm:"column:driver_id;primaryKey"`
	IsOnline     bool      `gorm:"column:is_online;not null;default:true"`
	LastPingAt   time.Time `gorm:"column:last_ping_at;not null;default:CURRENT_TIMESTAMP"`
	WentOnlineAt time.Time `gorm:"column:went_online_at;not null;default:CURRENT_TIMESTAMP"`
	CurrentLat   *float64  `gorm:"column:current_lat"`
	CurrentLng   *float64  `gorm:"column:current_lng"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP"`
}

func (OnlineDriverModel) TableName() string {
	return "online_drivers"
}

type OnlineStatusPostgresRepository struct {
	db *gorm.DB
}

func NewOnlineStatusPostgresRepository(db *gorm.DB) repository.OnlineStatusRepository {
	return &OnlineStatusPostgresRepository{db: db}
}

// UpsertOnlineDriver creates or updates online driver record with location ping
func (r *OnlineStatusPostgresRepository) UpsertOnlineDriver(ctx context.Context, driverID int64, lat, lng float64) error {
	now := time.Now()

	var existing OnlineDriverModel
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		newDriver := OnlineDriverModel{
			DriverID:     driverID,
			IsOnline:     true,
			LastPingAt:   now,
			WentOnlineAt: now,
			CurrentLat:   &lat,
			CurrentLng:   &lng,
			UpdatedAt:    now,
		}
		return r.db.WithContext(ctx).Create(&newDriver).Error
	} else if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"is_online":    true,
		"last_ping_at": now,
		"current_lat":  lat,
		"current_lng":  lng,
		"updated_at":   now,
	}

	return r.db.WithContext(ctx).
		Model(&OnlineDriverModel{}).
		Where("driver_id = ?", driverID).
		Updates(updates).Error
}

// SetDriverOffline removes driver from online drivers table
func (r *OnlineStatusPostgresRepository) SetDriverOffline(ctx context.Context, driverID int64) error {
	return r.db.WithContext(ctx).
		Where("driver_id = ?", driverID).
		Delete(&OnlineDriverModel{}).Error
}

// IsDriverOnline A driver is considered online if they exist in online_drivers table AND last ping was within 2 minutes
func (r *OnlineStatusPostgresRepository) IsDriverOnline(ctx context.Context, driverID int64) (bool, error) {
	// Calculate cutoff time (2 minutes ago)
	cutoffTime := time.Now().Add(-2 * time.Minute)

	var count int64
	err := r.db.WithContext(ctx).
		Model(&OnlineDriverModel{}).
		Where("driver_id = ? AND is_online = ? AND last_ping_at > ?", driverID, true, cutoffTime).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetOnlineDrivers returns list of all online driver IDs
func (r *OnlineStatusPostgresRepository) GetOnlineDrivers(ctx context.Context) ([]int64, error) {

	cutoffTime := time.Now().Add(-2 * time.Minute) // Calculate cutoff time (2 minutes ago)

	var driverIDs []int64
	err := r.db.WithContext(ctx).
		Model(&OnlineDriverModel{}).
		Where("is_online = ? AND last_ping_at > ?", true, cutoffTime).
		Pluck("driver_id", &driverIDs).Error

	if err != nil {
		return nil, err
	}

	return driverIDs, nil
}

// RemoveInactiveDrivers removes drivers who haven't pinged since cutoffTime
func (r *OnlineStatusPostgresRepository) RemoveInactiveDrivers(ctx context.Context, cutoffTime time.Time) error {
	return r.db.WithContext(ctx).
		Where("last_ping_at < ?", cutoffTime).
		Delete(&OnlineDriverModel{}).Error
}

// GetOnlineDriversByIDs filters a list of driver IDs to only those currently online
func (r *OnlineStatusPostgresRepository) GetOnlineDriversByIDs(ctx context.Context, driverIDs []int64) ([]int64, error) {
	if len(driverIDs) == 0 {
		return []int64{}, nil
	}

	cutoffTime := time.Now().Add(-2 * time.Minute) // Calculate cutoff time (2 minutes ago)

	var onlineDriverIDs []int64
	err := r.db.WithContext(ctx).
		Model(&OnlineDriverModel{}).
		Where("driver_id IN ? AND is_online = ? AND last_ping_at > ?", driverIDs, true, cutoffTime).
		Pluck("driver_id", &onlineDriverIDs).Error

	if err != nil {
		return nil, err
	}

	return onlineDriverIDs, nil
}
