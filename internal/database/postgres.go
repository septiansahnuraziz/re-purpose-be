package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// InitializePostgresConnection opens a GORM PostgreSQL connection with retry + pooling.
func InitializePostgresConnection() (*gorm.DB, error) {
	logrus.WithFields(logrus.Fields{
		"host": config.DatabaseHost(),
		"port": config.DatabasePort(),
		"name": config.DatabaseName(),
	}).Info("connecting to PostgreSQL 🛢️")

	db, err := gorm.Open(postgres.Open(config.DatabaseDSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(config.DatabaseMaxOpenConns())
	sqlDB.SetMaxIdleConns(config.DatabaseMaxIdleConns())
	sqlDB.SetConnMaxLifetime(config.DatabaseConnMaxLifetime())
	sqlDB.SetConnMaxIdleTime(config.DatabaseConnMaxIdleTime())

	if err := pingWithRetry(sqlDB); err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"host": config.DatabaseHost(),
		"port": config.DatabasePort(),
	}).Info("connected to PostgreSQL ✅")
	return db, nil
}

func pingWithRetry(db *sql.DB) error {
	attempts := config.DatabaseRetryAttempts()
	if attempts <= 0 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := db.Ping(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < attempts {
			logrus.WithError(lastErr).WithFields(logrus.Fields{
				"attempt":  attempt,
				"attempts": attempts,
			}).Warn("postgres ping failed, retrying")
			time.Sleep(config.DatabasePingInterval())
		}
	}

	return fmt.Errorf("ping postgres after %d attempts: %w", attempts, lastErr)
}
