package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv         string `mapstructure:"APP_ENV"`
	AppPort        string `mapstructure:"APP_PORT"`
	DBHost         string `mapstructure:"DB_HOST"`
	DBPort         string `mapstructure:"DB_PORT"`
	DBUser         string `mapstructure:"DB_USER"`
	DBPassword     string `mapstructure:"DB_PASSWORD"`
	DBName         string `mapstructure:"DB_NAME"`
	JWTSecret      string `mapstructure:"JWT_SECRET"`
	JWTExpiryHours int    `mapstructure:"JWT_EXPIRY_HOURS"`
	ClientOrigin   string `mapstructure:"CLIENT_ORIGIN"`
	GeminiAPIKey   string `mapstructure:"GEMINI_API_KEY"`
}

var AppConfig *Config

func LoadConfig() (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "ruang_tenang")
	viper.SetDefault("JWT_SECRET", "your-super-secret-jwt-key")
	viper.SetDefault("JWT_EXPIRY_HOURS", 24)
	viper.SetDefault("CLIENT_ORIGIN", "http://localhost:3000")

	if err := viper.ReadInConfig(); err != nil {
		// It's okay if .env doesn't exist, we can read from env vars
	}

	config := &Config{
		AppEnv:         viper.GetString("APP_ENV"),
		AppPort:        viper.GetString("APP_PORT"),
		DBHost:         viper.GetString("DB_HOST"),
		DBPort:         viper.GetString("DB_PORT"),
		DBUser:         viper.GetString("DB_USER"),
		DBPassword:     viper.GetString("DB_PASSWORD"),
		DBName:         viper.GetString("DB_NAME"),
		JWTSecret:      viper.GetString("JWT_SECRET"),
		JWTExpiryHours: viper.GetInt("JWT_EXPIRY_HOURS"),
		ClientOrigin:   viper.GetString("CLIENT_ORIGIN"),
		GeminiAPIKey:   viper.GetString("GEMINI_API_KEY"),
	}

	AppConfig = config
	return config, nil
}
