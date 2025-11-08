package postgres

import (
	"context"
	"errors"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"gorm.io/gorm"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
)

var (
	ErrDriverNotFound      = errors.New("driver not found")
	ErrDriverAlreadyExists = errors.New("driver already exists")
)

type DriverPostgresRepository struct {
	db *database.PostgresDB
}

func NewDriverPostgresRepository(db *database.PostgresDB) *DriverPostgresRepository {
	return &DriverPostgresRepository{db: db}
}

func toDriverModel(driver *domain.Driver) *DriverModel {
	return &DriverModel{
		ID:            driver.ID,
		Name:          driver.Name,
		Phone:         driver.Phone,
		VehicleNo:     driver.VehicleNo,
		IsOnline:      driver.IsOnline,
		CurrentLat:    driver.CurrentLat,
		CurrentLng:    driver.CurrentLng,
		LastPingAt:    driver.LastPingAt,
		LastUpdatedAt: driver.LastUpdatedAt,
		CreatedAt:     driver.CreatedAt,
	}
}

func toDriverDomain(model *DriverModel) *domain.Driver {
	return &domain.Driver{
		ID:            model.ID,
		Name:          model.Name,
		Phone:         model.Phone,
		VehicleNo:     model.VehicleNo,
		IsOnline:      model.IsOnline,
		CurrentLat:    model.CurrentLat,
		CurrentLng:    model.CurrentLng,
		LastPingAt:    model.LastPingAt,
		LastUpdatedAt: model.LastUpdatedAt,
		CreatedAt:     model.CreatedAt,
	}
}

func (r *DriverPostgresRepository) Create(ctx context.Context, driver *domain.Driver) error {
	model := toDriverModel(driver)

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		logger.Error(ctx, "Failed to create driver model", result.Error)
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrDriverAlreadyExists
		}
		return result.Error
	}

	driver.ID = model.ID // Set the auto-generated ID
	return nil
}

func (r *DriverPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Driver, error) {
	var model DriverModel

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		logger.Error(ctx, "Failed to get driver model", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrDriverNotFound
		}
		return nil, result.Error
	}

	return toDriverDomain(&model), nil
}

func (r *DriverPostgresRepository) GetByPhone(ctx context.Context, phone string) (*domain.Driver, error) {
	var model DriverModel

	result := r.db.WithContext(ctx).Where("phone = ?", phone).First(&model)
	if result.Error != nil {
		logger.Error(ctx, "Failed to get driver model", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrDriverNotFound
		}
		return nil, result.Error
	}

	return toDriverDomain(&model), nil
}

func (r *DriverPostgresRepository) UpdatePing(ctx context.Context, driverID int64, lat, lng float64, pingTime time.Time) error {
	return r.db.WithContext(ctx).Model(&DriverModel{}).
		Where("id = ?", driverID).
		Updates(map[string]interface{}{
			"current_lat":     lat,
			"current_lng":     lng,
			"last_ping_at":    pingTime,
			"is_online":       true,
			"last_updated_at": pingTime,
		}).Error
}

func (r *DriverPostgresRepository) SetOnlineStatus(ctx context.Context, driverID int64, isOnline bool) error {
	return r.db.WithContext(ctx).Model(&DriverModel{}).
		Where("id = ?", driverID).
		Update("is_online", isOnline).Error
}

func (r *DriverPostgresRepository) GetOnlineDrivers(ctx context.Context) ([]*domain.Driver, error) {
	var models []DriverModel

	result := r.db.WithContext(ctx).Where("is_online = ?", true).Find(&models)
	if result.Error != nil {
		logger.Error(ctx, "Failed to get online drivers", result.Error)
		return nil, result.Error
	}

	drivers := make([]*domain.Driver, len(models))
	for i, model := range models {
		drivers[i] = toDriverDomain(&model)
	}

	return drivers, nil
}

func (r *DriverPostgresRepository) MarkOfflineIfInactive(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).Model(&DriverModel{}).
		Where("last_ping_at < ? AND is_online = ?", cutoff, true).
		Update("is_online", false).Error
}
