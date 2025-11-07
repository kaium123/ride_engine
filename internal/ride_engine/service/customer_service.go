package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

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
	existingCustomer, _, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existingCustomer != nil {
		return nil, "", errors.New("customer with this email already exists")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	customer := &domain.Customer{
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: time.Now(),
	}

	if err := domain.ValidateCustomer(customer); err != nil {
		return nil, "", err
	}

	if err := s.repo.Create(ctx, customer, hashedPassword); err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateJWT(customer.ID, "customer", s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, "", err
	}

	key := fmt.Sprintf("jwt:user:%d", customer.ID)
	err = s.redis.Set(ctx, key, token, time.Duration(s.jwtExpiry)*time.Second).Err()
	if err != nil {
		return nil, "", fmt.Errorf("failed to store JWT in Redis: %v", err)
	}

	return customer, token, nil
}

// Login authenticates a customer
func (s *CustomerService) Login(ctx context.Context, email, password string) (*domain.Customer, string, error) {
	customer, hashedPassword, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", errors.New("invalid email or password")
	}

	if !utils.CheckPassword(password, hashedPassword) {
		return nil, "", errors.New("invalid email or password")
	}

	token, err := utils.GenerateJWT(customer.ID, "customer", s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, "", err
	}

	key := fmt.Sprintf("jwt:user:%d", customer.ID)
	fmt.Println(key)
	err = s.redis.Set(ctx, key, token, time.Duration(s.jwtExpiry)*time.Second).Err()
	if err != nil {
		return nil, "", fmt.Errorf("failed to store JWT in Redis: %v", err)
	}

	return customer, token, nil
}

// GetByID retrieves a customer by ID
func (s *CustomerService) GetByID(ctx context.Context, id int64) (*domain.Customer, error) {
	return s.repo.GetByID(ctx, id)
}
