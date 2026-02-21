package database

import (
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"gorm.io/gorm"
)

func TestConnect_RequiresDatabaseURL(t *testing.T) {
	_, err := Connect(&config.Config{DatabaseURL: ""})
	if err == nil {
		t.Fatalf("expected error when database url is empty")
	}
}

func TestGetDB_ReturnsCurrentGlobalDB(t *testing.T) {
	original := DB
	defer func() { DB = original }()

	fake := &gorm.DB{}
	DB = fake

	if got := GetDB(); got != fake {
		t.Fatalf("expected same DB pointer")
	}
}

func TestConnect_InvalidDSN(t *testing.T) {
	t.Run("development env", func(t *testing.T) {
		_, err := Connect(&config.Config{
			AppEnv:      "development",
			DatabaseURL: "postgres://user:pass@127.0.0.1:1/db?sslmode=disable&connect_timeout=1",
		})
		if err == nil {
			t.Fatal("expected connection error")
		}
		if !strings.Contains(err.Error(), "failed to connect to database") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("production env", func(t *testing.T) {
		_, err := Connect(&config.Config{
			AppEnv:      "production",
			DatabaseURL: "postgres://user:pass@127.0.0.1:1/db?sslmode=disable&connect_timeout=1",
		})
		if err == nil {
			t.Fatal("expected connection error")
		}
		if !strings.Contains(err.Error(), "failed to connect to database") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}