package testing

import (
	"testing"

	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
)

func TestTestingSeeder_BasicFlows(t *testing.T) {
	if err := logger.Init("development"); err != nil {
		t.Fatalf("failed init logger: %v", err)
	}
	defer logger.Sync()

	s := NewTestingSeeder(nil)
	if s == nil {
		t.Fatalf("expected seeder instance")
	}

	if err := s.Seed(); err != nil {
		t.Fatalf("expected seed success, got %v", err)
	}

	if err := s.Reset(); err != nil {
		t.Fatalf("expected reset success, got %v", err)
	}

	if err := s.seedRoles(); err != nil {
		t.Fatalf("expected seedRoles success, got %v", err)
	}

	if err := s.seedUsers(); err != nil {
		t.Fatalf("expected seedUsers success, got %v", err)
	}
}
