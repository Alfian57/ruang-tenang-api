package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// KeyRotationManager handles JWT key rotation
type KeyRotationManager struct {
	CurrentKey  []byte
	PreviousKey []byte
	RotationTime time.Time
}

var keyManager *KeyRotationManager

// InitializeKeyManager initializes the key rotation manager
func InitializeKeyManager() {
	cfg := config.AppConfig
	keyManager = &KeyRotationManager{
		CurrentKey:  []byte(cfg.JWTSecret),
		PreviousKey: []byte(cfg.JWTSecret), // Fallback to current during initialization
		RotationTime: time.Now(),
	}
}

// GenerateToken generates a JWT token with current key
func GenerateToken(userID uint, email, role string, duration time.Duration) (string, error) {
	if keyManager == nil {
		return "", errors.New("key manager not initialized")
	}

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(keyManager.CurrentKey)
}

// ValidateToken validates a JWT token using current or previous key
func ValidateToken(tokenString string) (*JWTClaims, error) {
	if keyManager == nil {
		return nil, errors.New("key manager not initialized")
	}

	// Try current key first
	token, err := parseTokenWithKey(tokenString, keyManager.CurrentKey)
	if err == nil {
		return token, nil
	}

	// Fall back to previous key for rotation period
	if keyManager.PreviousKey != nil {
		token, err := parseTokenWithKey(tokenString, keyManager.PreviousKey)
		if err == nil {
			return token, nil
		}
	}

	return nil, fmt.Errorf("invalid token: %w", err)
}

// parseTokenWithKey parses a token with a specific key
func parseTokenWithKey(tokenString string, key []byte) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return key, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// RotateKeys rotates JWT signing keys
func RotateKeys(newKey string) {
	if keyManager == nil {
		return
	}
	keyManager.PreviousKey = keyManager.CurrentKey
	keyManager.CurrentKey = []byte(newKey)
	keyManager.RotationTime = time.Now()
}

// GetKeyRotationStatus returns key rotation status
func GetKeyRotationStatus() map[string]interface{} {
	return map[string]interface{}{
		"rotation_time": keyManager.RotationTime,
		"has_previous_key": keyManager.PreviousKey != nil,
	}
}
