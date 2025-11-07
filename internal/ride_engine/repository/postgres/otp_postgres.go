package postgres

import (
	"context"
	"time"

	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
)

type OTPPostgresRepository struct {
	db *database.PostgresDB
}

func NewOTPPostgresRepository(db *database.PostgresDB) *OTPPostgresRepository {
	return &OTPPostgresRepository{db: db}
}

// SaveOTP saves OTP to database for audit trail
func (r *OTPPostgresRepository) SaveOTP(ctx context.Context, phone, otp, purpose string, expiresAt time.Time) error {
	model := &OTPModel{
		Phone:      phone,
		OTP:        otp,
		Purpose:    purpose,
		IsVerified: false,
		IsExpired:  false,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// VerifyOTP marks OTP as verified and returns true if valid
func (r *OTPPostgresRepository) VerifyOTP(ctx context.Context, phone, otp string) (bool, error) {
	var model OTPModel

	// Find the most recent non-expired, non-verified OTP for this phone
	err := r.db.WithContext(ctx).
		Where("phone = ? AND otp = ? AND is_verified = ? AND is_expired = ? AND expires_at > ?",
			phone, otp, false, false, time.Now()).
		Order("created_at DESC").
		First(&model).Error

	if err != nil {
		return false, nil // OTP not found or expired
	}

	// Mark as verified
	now := time.Now()
	model.IsVerified = true
	model.VerifiedAt = &now

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return false, err
	}

	return true, nil
}

// MarkExpired marks all non-verified OTPs for a phone as expired
func (r *OTPPostgresRepository) MarkExpired(ctx context.Context, phone string) error {
	return r.db.WithContext(ctx).
		Model(&OTPModel{}).
		Where("phone = ? AND is_verified = ? AND is_expired = ?", phone, false, false).
		Update("is_expired", true).Error
}

// GetOTPHistory retrieves OTP history for a phone number
func (r *OTPPostgresRepository) GetOTPHistory(ctx context.Context, phone string, limit int) ([]OTPModel, error) {
	var otps []OTPModel

	err := r.db.WithContext(ctx).
		Where("phone = ?", phone).
		Order("created_at DESC").
		Limit(limit).
		Find(&otps).Error

	return otps, err
}

// CleanupExpiredOTPs removes expired OTPs older than specified duration (for maintenance)
func (r *OTPPostgresRepository) CleanupExpiredOTPs(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", olderThan).
		Delete(&OTPModel{}).Error
}
