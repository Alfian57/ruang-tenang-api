package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	gomysql "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
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
	loc, err := time.LoadLocation(tz)
	if err != nil {
		tz = "Asia/Jakarta"
		loc, err = time.LoadLocation(tz)
		if err != nil {
			loc = time.FixedZone(tz, 7*60*60)
		}
	}

	dsnCfg, err := buildMySQLConfig(cfg.DatabaseURL, loc)
	if err != nil {
		return nil, err
	}

	connector, err := gomysql.NewConnector(dsnCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build database connector: %w", err)
	}
	sqlDB := sql.OpenDB(connector)

	logLevel := logger.Silent
	if cfg.AppEnv == "development" {
		logLevel = logger.Info
	}

	db, err := openDBFn(gormmysql.New(gormmysql.Config{Conn: sqlDB}), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	return db, nil
}

// buildMySQLConfig turns a mysql:// URL (or a raw go-sql-driver DSN) into a
// parsed driver config. User and password are QueryUnescaped to match the
// behaviour expected by golang-migrate (which unescapes DATABASE_URL
// credentials). The app timezone is applied as the connection location so
// DATETIME values are read/written in APP_TIMEZONE.
func buildMySQLConfig(rawURL string, loc *time.Location) (*gomysql.Config, error) {
	raw := strings.TrimSpace(rawURL)
	raw = strings.TrimPrefix(raw, "mysql://")

	dsnCfg, err := gomysql.ParseDSN(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid database DSN: %w", err)
	}

	if user, err := url.QueryUnescape(dsnCfg.User); err == nil {
		dsnCfg.User = user
	}
	if pass, err := url.QueryUnescape(dsnCfg.Passwd); err == nil {
		dsnCfg.Passwd = pass
	}

	dsnCfg.ParseTime = true
	if loc != nil {
		dsnCfg.Loc = loc
	}
	if dsnCfg.Params == nil {
		dsnCfg.Params = map[string]string{}
	}
	if _, ok := dsnCfg.Params["charset"]; !ok {
		dsnCfg.Params["charset"] = "utf8mb4"
	}
	// Align the MySQL session time zone with APP_TIMEZONE so DB-generated
	// timestamps (DEFAULT CURRENT_TIMESTAMP / NOW()) match timestamps written
	// by GORM. A numeric offset avoids depending on server time zone tables.
	if loc != nil {
		_, offset := time.Now().In(loc).Zone()
		sign := "+"
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		dsnCfg.Params["time_zone"] = fmt.Sprintf("'%s%02d:%02d'", sign, offset/3600, (offset%3600)/60)
	}

	return dsnCfg, nil
}

func GetDB() *gorm.DB {
	return DB
}
