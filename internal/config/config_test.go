package config

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func resetConfigState() {
	viper.Reset()
	AppConfig = nil
}

func TestBuildDatabaseURLFromParts(t *testing.T) {
	t.Run("with password", func(t *testing.T) {
		got := buildDatabaseURLFromParts("localhost", "5432", "postgres", "secret", "ruang_tenang")
		if !strings.Contains(got, "postgres://postgres:secret@localhost:5432/ruang_tenang") {
			t.Fatalf("unexpected database URL: %s", got)
		}
		if !strings.Contains(got, "sslmode=disable") {
			t.Fatalf("missing sslmode in URL: %s", got)
		}
	})

	t.Run("without password", func(t *testing.T) {
		got := buildDatabaseURLFromParts("localhost", "5432", "postgres", "", "ruang_tenang")
		if strings.Contains(got, ":@") {
			t.Fatalf("unexpected password marker in URL: %s", got)
		}
	})

	t.Run("missing required parts", func(t *testing.T) {
		got := buildDatabaseURLFromParts("", "5432", "postgres", "", "ruang_tenang")
		if got != "" {
			t.Fatalf("expected empty URL for incomplete parts, got %s", got)
		}
	})
}

func TestLoadConfigWithDatabaseURL(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "8090")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000, https://example.com")
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/db?sslmode=disable")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected load config error: %v", err)
	}

	if cfg.AppPort != "8090" {
		t.Fatalf("expected app port 8090, got %s", cfg.AppPort)
	}
	if len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(cfg.CORSAllowedOrigins))
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("expected database URL to be set")
	}
}

func TestLoadConfigBuildsDatabaseURLFromParts(t *testing.T) {
	resetConfigState()
	t.Cleanup(resetConfigState)

	t.Setenv("APP_ENV", "test")
	t.Setenv("PORT", "8090")
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "ruang_tenang")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected load config error: %v", err)
	}

	if !strings.Contains(cfg.DatabaseURL, "postgres://postgres@localhost:5432/ruang_tenang") {
		t.Fatalf("unexpected generated database URL: %s", cfg.DatabaseURL)
	}
}
