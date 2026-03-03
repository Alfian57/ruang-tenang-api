package utils

import (
	"crypto/rand"
	"encoding/hex"
)

var randReadFn = rand.Read

func GenerateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := randReadFn(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
