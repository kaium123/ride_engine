package database

import (
	"context"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
	log "vcs.technonext.com/carrybee/ride_engine/pkg/logger"
)

type PostgresDB struct {
	*gorm.DB
}

func NewPostgresDB(cfg config.PostgresConfig) (*PostgresDB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		log.Error(context.Background(), err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Error(context.Background(), err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error(context.Background(), err)
		return nil, err
	}

	log.Info(context.Background(), "PostgreSQL connected successfully with GORM")
	return &PostgresDB{db}, nil
}

func (db *PostgresDB) Close() error {
	log.Info(context.Background(), "Closing PostgreSQL DB...")
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
