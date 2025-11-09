package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"github.com/redis/go-redis/v9"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository/postgres"
)

type OTPService struct {
	redis   *redis.Client
	otpRepo *postgres.OTPPostgresRepository
}

func NewOTPService(redisClient *redis.Client, otpRepo *postgres.OTPPostgresRepository) *OTPService {
	return &OTPService{
		redis:   redisClient,
		otpRepo: otpRepo,
	}
}

func (s *OTPService) GenerateOTP() string {
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// SaveOTP saves OTP in both Redis (for fast validation) and PostgreSQL (for visualization)
func (s *OTPService) SaveOTP(ctx context.Context, phone, otp, purpose string) error {
	expiresAt := time.Now().Add(2 * time.Minute)

	key := fmt.Sprintf("otp:%s", phone)
	if err := s.redis.Set(ctx, key, otp, 2*time.Minute).Err(); err != nil {
		logger.Error(ctx, fmt.Sprintf("Failed to save OTP to Redis: %v", err))
		return err
	}

	if err := s.otpRepo.SaveOTP(ctx, phone, otp, purpose, expiresAt); err != nil {
		logger.Error(ctx, fmt.Sprintf("save otp error: %v", err))
	}

	return nil
}

// VerifyOTP verifies OTP from both Redis and PostgreSQL
func (s *OTPService) VerifyOTP(ctx context.Context, phone, otp string) (bool, error) {
	key := fmt.Sprintf("otp:%s", phone)
	storedOTP, err := s.redis.Get(ctx, key).Result()

	if err == redis.Nil {
		valid, dbErr := s.otpRepo.VerifyOTP(ctx, phone, otp)
		return valid, dbErr
	}

	if err != nil {
		// Redis error, fallback to database
		return s.otpRepo.VerifyOTP(ctx, phone, otp)
	}

	if storedOTP == otp {
		s.redis.Del(ctx, key)

		if _, err := s.otpRepo.VerifyOTP(ctx, phone, otp); err != nil {
			logger.Error(ctx, fmt.Sprintf("verify otp error: %v", err))
		}

		return true, nil
	}

	return false, nil
}

// InvalidateOTP marks all pending OTPs for a phone as expired
func (s *OTPService) InvalidateOTP(ctx context.Context, phone string) error {
	key := fmt.Sprintf("otp:%s", phone)
	s.redis.Del(ctx, key)

	return s.otpRepo.MarkExpired(ctx, phone)
}
