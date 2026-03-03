package database

import (
	"errors"
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
	originalOpen := openDBFn
	t.Cleanup(func() { openDBFn = originalOpen })
	openDBFn = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
		return originalOpen(dialector, cfg)
	}

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

func TestConnect_SetsGlobalDBOnSuccess(t *testing.T) {
	originalDB := DB
	originalOpen := openDBFn
	t.Cleanup(func() {
		DB = originalDB
		openDBFn = originalOpen
	})

	fake := &gorm.DB{}
	openDBFn = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
		return fake, nil
	}

	got, err := Connect(&config.Config{AppEnv: "development", DatabaseURL: "postgres://example"})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if got != fake {
		t.Fatalf("expected returned db to match fake")
	}
	if DB != fake {
		t.Fatalf("expected global DB to be set")
	}
}

func TestConnect_WrapsOpenError(t *testing.T) {
	originalOpen := openDBFn
	t.Cleanup(func() { openDBFn = originalOpen })

	openDBFn = func(dialector gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
		return nil, errors.New("boom")
	}

	_, err := Connect(&config.Config{AppEnv: "production", DatabaseURL: "postgres://example"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to connect to database") {
		t.Fatalf("unexpected error: %v", err)
	}
}
