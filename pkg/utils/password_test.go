package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "Sup3rStrong!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}
	if hash == password {
		t.Fatal("hash should not equal original password")
	}

	if !CheckPassword(password, hash) {
		t.Fatal("expected password check to pass")
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("expected wrong password check to fail")
	}
}
