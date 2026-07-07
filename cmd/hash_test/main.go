package main

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func main() {
	orderID := "RT-3-1783415073416086755"
	statusCode := "200"
	serverKey := "SB-Mid-server-X66QpIKc8dZ5cpqHUH5IQVJ9"

	amounts := []string{"279000.00", "279000", "279000.0"}
	for _, amt := range amounts {
		raw := orderID + statusCode + amt + serverKey
		hash := sha512.Sum512([]byte(raw))
		fmt.Printf("Amount: %s -> Hash: %s\n", amt, hex.EncodeToString(hash[:]))
	}
}
