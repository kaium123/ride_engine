package postgres

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"
)

var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrCustomerAlreadyExists = errors.New("customer already exists")
)

type CustomerPostgresRepository struct {
	db *database.PostgresDB
}

func NewCustomerPostgresRepository(db *database.PostgresDB) *CustomerPostgresRepository {
	return &CustomerPostgresRepository{db: db}
}

func toCustomerModel(customer *domain.Customer, password string) *CustomerModel {
	return &CustomerModel{
		ID:        customer.ID,
		Name:      customer.Name,
		Email:     customer.Email,
		Phone:     customer.Phone,
		Password:  password,
		CreatedAt: customer.CreatedAt,
	}
}

func toCustomerDomain(model *CustomerModel) *domain.Customer {
	return &domain.Customer{
		ID:        model.ID,
		Name:      model.Name,
		Email:     model.Email,
		Phone:     model.Phone,
		CreatedAt: model.CreatedAt,
	}
}

func (r *CustomerPostgresRepository) Create(ctx context.Context, customer *domain.Customer, password string) error {
	model := toCustomerModel(customer, password)

	result := r.db.WithContext(ctx).Create(model)
	if result.Error != nil {
		logger.Error(ctx, "error creating customer", result.Error)
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrCustomerAlreadyExists
		}
		return result.Error
	}

	customer.ID = model.ID // Set the auto-generated ID
	return nil
}

func (r *CustomerPostgresRepository) GetByID(ctx context.Context, id int64) (*domain.Customer, error) {
	var model CustomerModel

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
	if result.Error != nil {
		logger.Error(ctx, "error getting customer", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, result.Error
	}

	return toCustomerDomain(&model), nil
}

func (r *CustomerPostgresRepository) GetByEmail(ctx context.Context, email string) (*domain.Customer, string, error) {
	var model CustomerModel

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&model)
	if result.Error != nil {
		logger.Error(ctx, "error getting customer", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, "", ErrCustomerNotFound
		}
		return nil, "", result.Error
	}

	return toCustomerDomain(&model), model.Password, nil
}

func (r *CustomerPostgresRepository) GetByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	var model CustomerModel

	result := r.db.WithContext(ctx).Where("phone = ?", phone).First(&model)
	if result.Error != nil {
		logger.Error(ctx, "error getting customer", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, result.Error
	}

	return toCustomerDomain(&model), nil
}

func (r *CustomerPostgresRepository) Update(ctx context.Context, customer *domain.Customer) error {
	result := r.db.WithContext(ctx).Model(&CustomerModel{}).
		Where("id = ?", customer.ID).
		Updates(map[string]interface{}{
			"name":  customer.Name,
			"email": customer.Email,
			"phone": customer.Phone,
		})

	if result.Error != nil {
		logger.Error(ctx, "error updating customer", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Error(ctx, "error updating customer", ErrCustomerNotFound)
		return ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerPostgresRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&CustomerModel{})

	if result.Error != nil {
		logger.Error(ctx, "error deleting customer", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		logger.Error(ctx, "error deleting customer", ErrCustomerNotFound)
		return ErrCustomerNotFound
	}

	return nil
}
