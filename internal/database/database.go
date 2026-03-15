package database

import (
	"fmt"
	"net/url"
	"strings"
	"time"

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

	tz := strings.TrimSpace(cfg.AppTimezone)
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	if _, err := time.LoadLocation(tz); err != nil {
		tz = "Asia/Jakarta"
	}

	dsn := withPostgresTimezone(cfg.DatabaseURL, tz)

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

	if err := db.Exec("SELECT set_config('TIMEZONE', ?, false)", tz).Error; err != nil {
		return nil, fmt.Errorf("failed to set database timezone: %w", err)
	}

	DB = db
	return db, nil
}

func withPostgresTimezone(dsn, timezone string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return dsn
	}

	q := u.Query()
	q.Set("TimeZone", timezone)
	u.RawQuery = q.Encode()
	return u.String()
}

func GetDB() *gorm.DB {
	return DB
}
