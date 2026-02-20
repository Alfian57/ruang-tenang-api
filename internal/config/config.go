package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv             string   `mapstructure:"APP_ENV"`
	AppPort            string   `mapstructure:"APP_PORT"`
	AppTimezone        string   `mapstructure:"APP_TIMEZONE"`
	DatabaseURL        string   `mapstructure:"DATABASE_URL"`
	DBHost             string   `mapstructure:"DB_HOST"`
	DBPort             string   `mapstructure:"DB_PORT"`
	DBUser             string   `mapstructure:"DB_USER"`
	DBPassword         string   `mapstructure:"DB_PASSWORD"`
	DBName             string   `mapstructure:"DB_NAME"`
	JWTSecret          string   `mapstructure:"JWT_SECRET"`
	JWTExpiryHours     int      `mapstructure:"JWT_EXPIRY_HOURS"`
	CORSAllowedOrigins []string // parsed from CORS_ALLOWED_ORIGINS (comma-separated)
	GeminiAPIKey       string   `mapstructure:"GEMINI_API_KEY"`
}

var AppConfig *Config

func LoadConfig() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	viper.AutomaticEnv()

	// Set defaults for non-critical vars
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_TIMEZONE", "Asia/Jakarta")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if .env doesn't exist, we can read from env vars
	}

	databaseURL := strings.TrimSpace(viper.GetString("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = buildDatabaseURLFromParts(
			viper.GetString("DB_HOST"),
			viper.GetString("DB_PORT"),
			viper.GetString("DB_USER"),
			viper.GetString("DB_PASSWORD"),
			viper.GetString("DB_NAME"),
		)
	}

	// Fail fast: validate required env vars
	required := []string{"APP_ENV", "PORT", "JWT_SECRET", "CORS_ALLOWED_ORIGINS"}
	var missing []string
	for _, key := range required {
		if viper.GetString(key) == "" {
			missing = append(missing, key)
		}
	}
	if databaseURL == "" {
		missing = append(missing, "DATABASE_URL (or DB_HOST, DB_PORT, DB_USER, DB_NAME)")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %s", strings.Join(missing, ", "))
	}

	// Parse CORS origins (comma-separated)
	rawOrigins := viper.GetString("CORS_ALLOWED_ORIGINS")
	var origins []string
	for _, o := range strings.Split(rawOrigins, ",") {
		trimmed := strings.TrimSpace(o)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	// Prefer PORT as source of truth, keep APP_PORT as fallback for backward compatibility.
	appPort := viper.GetString("PORT")
	if appPort == "" {
		appPort = viper.GetString("APP_PORT")
	}

	config := &Config{
		AppEnv:             viper.GetString("APP_ENV"),
		AppPort:            appPort,
		AppTimezone:        viper.GetString("APP_TIMEZONE"),
		DatabaseURL:        databaseURL,
		DBHost:             viper.GetString("DB_HOST"),
		DBPort:             viper.GetString("DB_PORT"),
		DBUser:             viper.GetString("DB_USER"),
		DBPassword:         viper.GetString("DB_PASSWORD"),
		DBName:             viper.GetString("DB_NAME"),
		JWTSecret:          viper.GetString("JWT_SECRET"),
		JWTExpiryHours:     viper.GetInt("JWT_EXPIRY_HOURS"),
		CORSAllowedOrigins: origins,
		GeminiAPIKey:       viper.GetString("GEMINI_API_KEY"),
	}

	AppConfig = config
	return config, nil
}

func buildDatabaseURLFromParts(host, port, user, password, dbName string) string {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	user = strings.TrimSpace(user)
	password = strings.TrimSpace(password)
	dbName = strings.TrimSpace(dbName)

	if host == "" || port == "" || user == "" || dbName == "" {
		return ""
	}

	u := &url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + dbName,
	}
	if password == "" {
		u.User = url.User(user)
	} else {
		u.User = url.UserPassword(user, password)
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String()
}
