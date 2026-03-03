package database

import (
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

var openDBFn = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(dialector, cfg)
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	dsn := cfg.DatabaseURL

	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := openDBFn(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return db, nil
}

func GetDB() *gorm.DB {
	return DB
}
