package postgres

import (
	"time"
)

// CustomerModel represents the customers table
type CustomerModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"type:varchar(255);not null"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Phone     string    `gorm:"type:varchar(20);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (CustomerModel) TableName() string {
	return "customers"
}

// DriverModel represents the drivers table
type DriverModel struct {
	ID            int64      `gorm:"primaryKey;autoIncrement"`
	Name          string     `gorm:"type:varchar(255);not null"`
	Phone         string     `gorm:"type:varchar(20);uniqueIndex;not null"`
	VehicleNo     string     `gorm:"type:varchar(50)"`
	IsOnline      bool       `gorm:"not null;default:false;index"`
	CurrentLat    *float64   `gorm:"type:double precision"`
	CurrentLng    *float64   `gorm:"type:double precision"`
	LastPingAt    *time.Time `gorm:"type:timestamp;index"`
	LastUpdatedAt *time.Time `gorm:"type:timestamp"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (DriverModel) TableName() string {
	return "drivers"
}

// RideModel represents the rides table
type RideModel struct {
	ID          int64         `gorm:"primaryKey;autoIncrement"`
	CustomerID  int64         `gorm:"not null;index"`
	DriverID    *int64        `gorm:"index"`
	PickupLat   float64       `gorm:"type:double precision;not null"`
	PickupLng   float64       `gorm:"type:double precision;not null"`
	DropoffLat  float64       `gorm:"type:double precision;not null"`
	DropoffLng  float64       `gorm:"type:double precision;not null"`
	Status      string        `gorm:"type:varchar(20);not null;index"`
	Fare        *float64      `gorm:"type:decimal(10,2)"`
	RequestedAt time.Time     `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	AcceptedAt  *time.Time    `gorm:"type:timestamp"`
	StartedAt   *time.Time    `gorm:"type:timestamp"`
	CompletedAt *time.Time    `gorm:"type:timestamp"`
	CancelledAt *time.Time    `gorm:"type:timestamp"`
	Customer    CustomerModel `gorm:"foreignKey:CustomerID;references:ID;constraint:OnDelete:CASCADE"`
	Driver      *DriverModel  `gorm:"foreignKey:DriverID;references:ID;constraint:OnDelete:SET NULL"`
}

func (RideModel) TableName() string {
	return "rides"
}

// OTPModel represents the otp_records table for audit trail
type OTPModel struct {
	ID         int64      `gorm:"primaryKey;autoIncrement"`
	Phone      string     `gorm:"type:varchar(20);not null;index"`
	OTP        string     `gorm:"type:varchar(10);not null"`
	Purpose    string     `gorm:"type:varchar(50);not null"` // driver_login, customer_verification, etc
	IsVerified bool       `gorm:"not null;default:false"`
	IsExpired  bool       `gorm:"not null;default:false"`
	ExpiresAt  time.Time  `gorm:"not null;index"`
	VerifiedAt *time.Time `gorm:"type:timestamp"`
	CreatedAt  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (OTPModel) TableName() string {
	return "otp_records"
}
