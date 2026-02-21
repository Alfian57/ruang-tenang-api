package utils

import (
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
)

func TestGenerateAndValidateToken(t *testing.T) {
	original := config.AppConfig
	t.Cleanup(func() { config.AppConfig = original })

	config.AppConfig = &config.Config{JWTSecret: "test-secret"}

	token, err := GenerateToken(10, "user@example.com", "member", time.Hour)
	if err != nil {
		t.Fatalf("unexpected generate token error: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}

	if claims.UserID != 10 || claims.Email != "user@example.com" || claims.Role != "member" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestValidateTokenWrongSecret(t *testing.T) {
	original := config.AppConfig
	t.Cleanup(func() { config.AppConfig = original })

	config.AppConfig = &config.Config{JWTSecret: "secret-a"}
	token, err := GenerateToken(1, "a@b.com", "member", time.Hour)
	if err != nil {
		t.Fatalf("unexpected generate token error: %v", err)
	}

	config.AppConfig = &config.Config{JWTSecret: "secret-b"}
	if _, err := ValidateToken(token); err == nil {
		t.Fatal("expected validation error for wrong secret")
	}
}
