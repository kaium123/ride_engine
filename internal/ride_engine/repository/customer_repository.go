package repository

import (
	"context"

	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
)

type CustomerRepository interface {
	Create(ctx context.Context, customer *domain.Customer, password string) error
	GetByID(ctx context.Context, id int64) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, string, error) // returns customer and hashed password
	GetByPhone(ctx context.Context, phone string) (*domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, id int64) error
}
