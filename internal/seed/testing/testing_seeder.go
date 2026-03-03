package testing

import (
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"gorm.io/gorm"
)

// TestingSeeder provides minimal fixture data for automated tests.
// Uses transaction-per-test + rollback or --reset per suite.
type TestingSeeder struct {
	db          *gorm.DB
	seedRolesFn func() error
	seedUsersFn func() error
}

// NewTestingSeeder creates a new testing seeder instance.
func NewTestingSeeder(db *gorm.DB) *TestingSeeder {
	s := &TestingSeeder{db: db}
	s.seedRolesFn = s.seedRoles
	s.seedUsersFn = s.seedUsers
	return s
}

// Seed runs the minimal fixture seeding for tests.
func (s *TestingSeeder) Seed() error {
	logger.Info("Running testing seeder...")

	// Seed roles/permissions (reference data)
	if err := s.seedRolesFn(); err != nil {
		return err
	}

	// Seed minimal test users
	if err := s.seedUsersFn(); err != nil {
		return err
	}

	logger.Info("Testing seeder completed.")
	return nil
}

// Reset truncates all tables for a clean test slate (only for testing).
func (s *TestingSeeder) Reset() error {
	logger.Info("Resetting test database...")
	// TODO: implement TRUNCATE ... RESTART IDENTITY CASCADE
	return nil
}

func (s *TestingSeeder) seedRoles() error {
	// TODO: seed minimal role fixtures
	return nil
}

func (s *TestingSeeder) seedUsers() error {
	// TODO: seed minimal user fixtures (admin + test user)
	return nil
}
