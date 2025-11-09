package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository"
	"vcs.technonext.com/carrybee/ride_engine/pkg/utils"
)

type CustomerService struct {
	repo      repository.CustomerRepository
	jwtSecret string
	jwtExpiry int
	redis     *redis.Client
}

func NewCustomerService(repo repository.CustomerRepository, jwtSecret string, jwtExpiry int, redis *redis.Client) *CustomerService {
	return &CustomerService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
		redis:     redis,
	}
}

// Register creates a new customer account
func (s *CustomerService) Register(ctx context.Context, name, email, phone, password string) (*domain.Customer, string, error) {
	if name == "" || email == "" || phone == "" || password == "" {
		logger.Error(ctx, "all fields are required")
		return nil, "", errors.New("all fields are required")
	}

	existingCustomer, _, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingCustomer != nil {
		logger.Error(ctx, "Customer with email already exists")
		return nil, "", errors.New("customer with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	customer := &domain.Customer{
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: time.Now(),
	}

	if err := domain.ValidateCustomer(customer); err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	if err := s.repo.Create(ctx, customer, hashedPassword); err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	token, err := utils.GenerateJWT(customer.ID, "customer", s.jwtSecret, s.jwtExpiry)
	if err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	key := fmt.Sprintf("jwt:user:%d", customer.ID)
	err = s.redis.Set(ctx, key, token, time.Duration(s.jwtExpiry)*time.Second).Err()
	if err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	return customer, token, nil
}

// Login authenticates a customer
func (s *CustomerService) Login(ctx context.Context, email, password string) (*domain.Customer, string, error) {
	if email == "" || password == "" {
		logger.Error(ctx, "email and password are required")
		return nil, "", errors.New("invalid email or password")
	}

	customer, hashedPassword, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		logger.Error(ctx, err)
		return nil, "", errors.New("invalid email or password")
	}

	if !utils.CheckPassword(password, hashedPassword) {
		logger.Error(ctx, "invalid password")
		return nil, "", errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWT(customer.ID, "customer", s.jwtSecret, s.jwtExpiry)
	if err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	key := fmt.Sprintf("jwt:user:%d", customer.ID)
	err = s.redis.Set(ctx, key, token, time.Duration(s.jwtExpiry)*time.Second).Err()
	if err != nil {
		logger.Error(ctx, err)
		return nil, "", err
	}

	return customer, token, nil
}

// GetByID retrieves a customer by ID
func (s *CustomerService) GetByID(ctx context.Context, id int64) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, id)
}
