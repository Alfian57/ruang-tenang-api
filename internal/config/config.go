package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	AppEnv                 string   `mapstructure:"APP_ENV"`
	AppPort                string   `mapstructure:"APP_PORT"`
	AppTimezone            string   `mapstructure:"APP_TIMEZONE"`
	DatabaseURL            string   `mapstructure:"DATABASE_URL"`
	DBHost                 string   `mapstructure:"DB_HOST"`
	DBPort                 string   `mapstructure:"DB_PORT"`
	DBUser                 string   `mapstructure:"DB_USER"`
	DBPassword             string   `mapstructure:"DB_PASSWORD"`
	DBName                 string   `mapstructure:"DB_NAME"`
	JWTSecret              string   `mapstructure:"JWT_SECRET"`
	JWTExpiryHours         int      `mapstructure:"JWT_EXPIRY_HOURS"`
	CORSAllowedOrigins     []string // parsed from CORS_ALLOWED_ORIGINS (comma-separated)
	GeminiAPIKey           string   `mapstructure:"GEMINI_API_KEY"`
	AI                     AIConfig // centralized AI/model configuration
	MidtransBaseURL        string   `mapstructure:"MIDTRANS_BASE_URL"`
	MidtransServerKey      string   `mapstructure:"MIDTRANS_SERVER_KEY"`
	ChatDailyMessageLimit  int      `mapstructure:"CHAT_DAILY_MESSAGE_LIMIT"`
	ChatQuotaResetInterval string   `mapstructure:"CHAT_QUOTA_RESET_INTERVAL"`
	VAPIDPublicKey         string   `mapstructure:"VAPID_PUBLIC_KEY"`
	VAPIDPrivateKey        string   `mapstructure:"VAPID_PRIVATE_KEY"`
	VAPIDContact           string   `mapstructure:"VAPID_CONTACT"`
}

var AppConfig *Config

// AIConfig memusatkan konfigurasi model AI (Gemini) agar mudah diganti dari
// satu tempat. Tiap area fitur punya model sendiri sehingga bisa diatur
// terpisah lewat environment variable, namun semuanya berbagi API key yang sama.
type AIConfig struct {
	APIKey string `mapstructure:"GEMINI_API_KEY"`
	// Model per area fitur. Default diisi di LoadConfig.
	ChatModel       string `mapstructure:"AI_CHAT_MODEL"`
	ModerationModel string `mapstructure:"AI_MODERATION_MODEL"`
	JournalModel    string `mapstructure:"AI_JOURNAL_MODEL"`
	WellnessModel   string `mapstructure:"AI_WELLNESS_MODEL"`
}

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
	viper.SetDefault("FRONTEND_URL", "")
	viper.SetDefault("MIDTRANS_BASE_URL", "https://app.sandbox.midtrans.com")
	viper.SetDefault("CHAT_DAILY_MESSAGE_LIMIT", 100)
	viper.SetDefault("CHAT_QUOTA_RESET_INTERVAL", "24h")

	// AI/model defaults — ganti di sini (atau lewat env) untuk mengubah model
	// yang dipakai tiap area fitur tanpa menyentuh kode service.
	viper.SetDefault("AI_CHAT_MODEL", "gemini-flash-latest")
	viper.SetDefault("AI_MODERATION_MODEL", "gemini-flash-latest")
	viper.SetDefault("AI_JOURNAL_MODEL", "gemini-2.0-flash")
	viper.SetDefault("AI_WELLNESS_MODEL", "gemini-flash-latest")

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

	// Parse CORS origins from comma-separated list + optional frontend URL.
	origins := parseAllowedOrigins(
		viper.GetString("CORS_ALLOWED_ORIGINS"),
		viper.GetString("FRONTEND_URL"),
	)

	// Prefer PORT as source of truth, keep APP_PORT as fallback for backward compatibility.
	appPort := viper.GetString("PORT")
	if appPort == "" {
		appPort = viper.GetString("APP_PORT")
	}

	config := &Config{
		AppEnv:                 viper.GetString("APP_ENV"),
		AppPort:                appPort,
		AppTimezone:            viper.GetString("APP_TIMEZONE"),
		DatabaseURL:            databaseURL,
		DBHost:                 viper.GetString("DB_HOST"),
		DBPort:                 viper.GetString("DB_PORT"),
		DBUser:                 viper.GetString("DB_USER"),
		DBPassword:             viper.GetString("DB_PASSWORD"),
		DBName:                 viper.GetString("DB_NAME"),
		JWTSecret:              viper.GetString("JWT_SECRET"),
		JWTExpiryHours:         viper.GetInt("JWT_EXPIRY_HOURS"),
		CORSAllowedOrigins:     origins,
		GeminiAPIKey:           viper.GetString("GEMINI_API_KEY"),
		AI: AIConfig{
			APIKey:          viper.GetString("GEMINI_API_KEY"),
			ChatModel:       viper.GetString("AI_CHAT_MODEL"),
			ModerationModel: viper.GetString("AI_MODERATION_MODEL"),
			JournalModel:    viper.GetString("AI_JOURNAL_MODEL"),
			WellnessModel:   viper.GetString("AI_WELLNESS_MODEL"),
		},
		MidtransBaseURL:        viper.GetString("MIDTRANS_BASE_URL"),
		MidtransServerKey:      viper.GetString("MIDTRANS_SERVER_KEY"),
		ChatDailyMessageLimit:  viper.GetInt("CHAT_DAILY_MESSAGE_LIMIT"),
		ChatQuotaResetInterval: viper.GetString("CHAT_QUOTA_RESET_INTERVAL"),
		VAPIDPublicKey:         viper.GetString("VAPID_PUBLIC_KEY"),
		VAPIDPrivateKey:        viper.GetString("VAPID_PRIVATE_KEY"),
		VAPIDContact:           viper.GetString("VAPID_CONTACT"),
	}

	AppConfig = config
	return config, nil
}

func parseAllowedOrigins(rawOrigins, frontendURL string) []string {
	origins := make([]string, 0, 4)
	seen := make(map[string]struct{})

	addOrigin := func(origin string) {
		normalized := normalizeOrigin(origin)
		if normalized == "" {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}
		seen[normalized] = struct{}{}
		origins = append(origins, normalized)
	}

	for _, o := range strings.Split(rawOrigins, ",") {
		addOrigin(o)
	}
	addOrigin(frontendURL)

	return origins
}

func normalizeOrigin(origin string) string {
	trimmed := strings.TrimSpace(origin)
	return strings.TrimRight(trimmed, "/")
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
