package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/postgres"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"
	"vcs.technonext.com/carrybee/ride_engine/pkg/utils"
)

type DriverService struct {
	driverRepo       *postgres.DriverPostgresRepository
	onlineStatusRepo repository.OnlineStatusRepository
	otpService       *OTPService
	locationService  *LocationService
	jwtSecret        string
	jwtExpiry        int
	redis            *redis.Client
}

func NewDriverService(
	driverRepo *postgres.DriverPostgresRepository,
	onlineStatusRepo repository.OnlineStatusRepository,
	otpService *OTPService,
	locationService *LocationService,
	jwtSecret string,
	jwtExpiry int,
	redis *redis.Client,
) *DriverService {
	return &DriverService{
		driverRepo:       driverRepo,
		onlineStatusRepo: onlineStatusRepo,
		otpService:       otpService,
		locationService:  locationService,
		jwtSecret:        jwtSecret,
		jwtExpiry:        jwtExpiry,
		redis:            redis,
	}
}

// Register creates a new driver account
func (s *DriverService) Register(ctx context.Context, name, phone, vehicleNo string) (*domain.Driver, error) {

	existingDriver, err := s.driverRepo.GetByPhone(ctx, phone)
	if err == nil && existingDriver != nil {
		logger.Error(ctx, fmt.Sprintf("driver with phone %s already exists", phone))
		return nil, errors.New("driver with this phone already exists")
	}

	driver := &domain.Driver{
		Name:      name,
		Phone:     phone,
		VehicleNo: vehicleNo,
		IsOnline:  false,
		CreatedAt: time.Now(),
	}

	if err := domain.ValidateDriver(driver); err != nil {
		logger.Error(ctx, fmt.Sprintf("invalid driver: %v", err))
		return nil, err
	}

	if err := s.driverRepo.Create(ctx, driver); err != nil {
		logger.Error(ctx, fmt.Sprintf("error creating driver: %v", err))
		return nil, err
	}

	return driver, nil
}

// RequestOTP generates and sends OTP to driver's phone
func (s *DriverService) RequestOTP(ctx context.Context, phone string) error {
	if phone == "" {
		logger.Error(ctx, "phone is required")
		return errors.New("phone is required")
	}

	_, err := s.driverRepo.GetByPhone(ctx, phone)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("driver with phone %s not found", phone))
		return errors.New("driver not found")
	}

	otp := s.otpService.GenerateOTP()
	if config.GetConfig().Environment == "development" {
		otp = "123456"
	}

	if err := s.otpService.SaveOTP(ctx, phone, otp, "driver_login"); err != nil {
		logger.Error(ctx, fmt.Sprintf("error saving otp: %v", err))
		return err
	}

	fmt.Printf("OTP for driver %s: %s\n", phone, otp)

	return nil
}

// VerifyOTP verifies OTP and logs in the driver
func (s *DriverService) VerifyOTP(ctx context.Context, phone, otp string) (*domain.Driver, string, error) {
	if phone == "" || otp == "" {
		logger.Error(ctx, "phone and OTP are required")
		return nil, "", errors.New("phone and OTP are required")
	}

	valid, err := s.otpService.VerifyOTP(ctx, phone, otp)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error verifying otp: %v", err))
		return nil, "", err
	}

	if !valid {
		logger.Error(ctx, fmt.Sprintf("invalid otp: %s", otp))
		return nil, "", errors.New("invalid or expired OTP")
	}

	driver, err := s.driverRepo.GetByPhone(ctx, phone)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("driver with phone %s not found", phone))
		return nil, "", err
	}

	token, err := utils.GenerateJWT(driver.ID, "driver", s.jwtSecret, s.jwtExpiry)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error generating token: %v", err))
		return nil, "", err
	}

	key := fmt.Sprintf("jwt:driver:%d", driver.ID)
	err = s.redis.Set(ctx, key, token, time.Duration(s.jwtExpiry)*time.Hour).Err()
	if err != nil {
		logger.Error(ctx, fmt.Sprintf("error saving token: %v", err))
		return nil, "", fmt.Errorf("failed to store JWT in Redis: %v", err)
	}

	return driver, token, nil
}

// UpdateLocation updates driver's location in both PostgreSQL and MongoDB
func (s *DriverService) UpdateLocation(ctx context.Context, driverID int64, lat, lng float64) error {

	if err := s.locationService.UpdateDriverLocation(ctx, driverID, lat, lng); err != nil {
		logger.Error(ctx, fmt.Sprintf("error updating driver location: %v", err))
		return err
	}

	return nil
}

// GetByID retrieves a driver by ID
func (s *DriverService) GetByID(ctx context.Context, id int64) (*domain.Driver, error) {
	return s.driverRepo.GetByID(ctx, id)
}

func (s *DriverService) GetNearestDrivers(ctx context.Context, lat, lng, radius float64, limit int) ([]int64, error) {
	if radius <= 0 {
		radius = 3000 // default 3 km
	}
	if limit <= 0 {
		limit = 5
	}

	nearestDrivers, err := s.locationService.FindNearestDrivers(ctx, lat, lng, radius, limit)
	if err != nil {
		return nil, err
	}

	return nearestDrivers, nil
}
