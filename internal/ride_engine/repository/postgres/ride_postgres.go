package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
)

var (
	ErrRideNotFound = errors.New("ride not found")
)

type RidePostgresRepository struct {
	db *database.PostgresDB
}

func NewRidePostgresRepository(db *database.PostgresDB) *RidePostgresRepository {
	return &RidePostgresRepository{db: db}
}

func toRideModel(ride *domain.Ride) *RideModel {
	return &RideModel{
		ID:          ride.ID,
		CustomerID:  ride.CustomerID,
		DriverID:    ride.DriverID,
		PickupLat:   ride.PickupLat,
		PickupLng:   ride.PickupLng,
		DropoffLat:  ride.DropoffLat,
		DropoffLng:  ride.DropoffLng,
		Status:      string(ride.Status),
		Fare:        ride.Fare,
		RequestedAt: ride.RequestedAt,
		AcceptedAt:  ride.AcceptedAt,
		StartedAt:   ride.StartedAt,
		CompletedAt: ride.CompletedAt,
		CancelledAt: ride.CancelledAt,
	}
}

func toRideDomain(model *RideModel) *domain.Ride {
	return &domain.Ride{
		ID:          model.ID,
		CustomerID:  model.CustomerID,
		DriverID:    model.DriverID,
		PickupLat:   model.PickupLat,
		PickupLng:   model.PickupLng,
		DropoffLat:  model.DropoffLat,
		DropoffLng:  model.DropoffLng,
		Status:      domain.RideStatus(model.Status),
		Fare:        model.Fare,
		RequestedAt: model.RequestedAt,
		AcceptedAt:  model.AcceptedAt,
		StartedAt:   model.StartedAt,
		CompletedAt: model.CompletedAt,
		CancelledAt: model.CancelledAt,
	}
}

func (r *RidePostgresRepository) Create(ctx context.Context, ride *domain.Ride) error {
	model := toRideModel(ride)

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		return result.Error
	}

	ride.ID = model.ID // Set the auto-generated ID
	return nil
}

func (r *RidePostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Ride, error) {
	var model RideModel

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRideNotFound
		}
		return nil, result.Error
	}

	return toRideDomain(&model), nil
}

func (r *RidePostgresRepository) Update(ctx context.Context, ride *domain.Ride) error {
	model := toRideModel(ride)

	result := r.db.WithContext(ctx).Model(&RideModel{}).
		Where("id = ?", ride.ID).
		Updates(model)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrRideNotFound
	}

	return nil
}

func (r *RidePostgresRepository) GetRequestedRides(ctx context.Context) ([]*domain.Ride, error) {
	var models []RideModel

	result := r.db.WithContext(ctx).Where("status = ?", "requested").Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	rides := make([]*domain.Ride, len(models))
	for i, model := range models {
		rides[i] = toRideDomain(&model)
	}

	return rides, nil
}

func (r *RidePostgresRepository) GetByCustomerID(ctx context.Context, customerID int64) ([]*domain.Ride, error) {
	var models []RideModel

	result := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Order("requested_at DESC").Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	rides := make([]*domain.Ride, len(models))
	for i, model := range models {
		rides[i] = toRideDomain(&model)
	}

	return rides, nil
}

func (r *RidePostgresRepository) GetByDriverID(ctx context.Context, driverID int64) ([]*domain.Ride, error) {
	var models []RideModel

	result := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Order("requested_at DESC").Find(&models)
	if result.Error != nil {
		return nil, result.Error
	}

	rides := make([]*domain.Ride, len(models))
	for i, model := range models {
		rides[i] = toRideDomain(&model)
	}

	return rides, nil
}
