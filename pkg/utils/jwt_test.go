package utils

import (
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/golang-jwt/jwt/v5"
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

func TestValidateToken_UnexpectedSigningMethod(t *testing.T) {
	original := config.AppConfig
	t.Cleanup(func() { config.AppConfig = original })

	config.AppConfig = &config.Config{JWTSecret: "test-secret"}

	tok := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"user_id": 1})
	badToken, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("unexpected signing error: %v", err)
	}

	_, err = ValidateToken(badToken)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "unexpected signing method") {
		t.Fatalf("unexpected error: %v", err)
	}
}
