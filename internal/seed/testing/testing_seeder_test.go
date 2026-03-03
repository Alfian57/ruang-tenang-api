package testing

import (
	"errors"
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

func TestTestingSeeder_SeedErrorBranches(t *testing.T) {
	if err := logger.Init("development"); err != nil {
		t.Fatalf("failed init logger: %v", err)
	}
	defer logger.Sync()

	t.Run("roles error", func(t *testing.T) {
		s := NewTestingSeeder(nil)
		s.seedRolesFn = func() error { return errors.New("roles failed") }
		if err := s.Seed(); err == nil || err.Error() != "roles failed" {
			t.Fatalf("expected roles failed error, got %v", err)
		}
	})

	t.Run("users error", func(t *testing.T) {
		s := NewTestingSeeder(nil)
		s.seedUsersFn = func() error { return errors.New("users failed") }
		if err := s.Seed(); err == nil || err.Error() != "users failed" {
			t.Fatalf("expected users failed error, got %v", err)
		}
	})
}
