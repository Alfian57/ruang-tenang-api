package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv             string   `mapstructure:"APP_ENV"`
	AppPort            string   `mapstructure:"APP_PORT"`
	AppTimezone        string   `mapstructure:"APP_TIMEZONE"`
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
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("APP_TIMEZONE", "Asia/Jakarta")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("CORS_ALLOWED_ORIGINS", "http://localhost:3000")

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if .env doesn't exist, we can read from env vars
	}

	// Fail fast: validate required env vars
	required := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "JWT_SECRET"}
	var missing []string
	for _, key := range required {
		if viper.GetString(key) == "" {
			missing = append(missing, key)
		}
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

	config := &Config{
		AppEnv:             viper.GetString("APP_ENV"),
		AppPort:            viper.GetString("APP_PORT"),
		AppTimezone:        viper.GetString("APP_TIMEZONE"),
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
