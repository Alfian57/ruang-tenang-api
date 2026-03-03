package production

import (
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProductionSeedTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.LevelConfig{},
		&model.ArticleCategory{},
		&model.SongCategory{},
		&model.ForumCategory{},
		&model.CrisisKeyword{},
		&model.User{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	manualTables := []string{
		`CREATE TABLE IF NOT EXISTS story_categories (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			name TEXT NOT NULL,
			slug TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			display_order INTEGER DEFAULT 0,
			is_active NUMERIC DEFAULT 1,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS breathing_techniques (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			name TEXT NOT NULL,
			slug TEXT,
			description TEXT,
			benefits TEXT,
			best_for TEXT,
			inhale_duration INTEGER,
			inhale_hold_duration INTEGER,
			exhale_duration INTEGER,
			exhale_hold_duration INTEGER,
			icon TEXT,
			color TEXT,
			animation_type TEXT,
			difficulty TEXT,
			category TEXT,
			origin TEXT,
			is_system NUMERIC,
			is_active NUMERIC,
			user_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS feature_definitions (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			feature_key TEXT NOT NULL,
			feature_name TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			required_level INTEGER,
			category TEXT,
			is_active NUMERIC,
			display_order INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS badge_definitions (
			id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			badge_key TEXT NOT NULL,
			badge_name TEXT NOT NULL,
			description TEXT,
			icon TEXT,
			category TEXT,
			requirement_type TEXT,
			requirement_value INTEGER,
			is_active NUMERIC,
			display_order INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}

	for _, stmt := range manualTables {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create manual table: %v", err)
		}
	}
	return db
}

func TestProductionSeeders_RunAndIdempotent(t *testing.T) {
	db := newProductionSeedTestDB(t)
	t.Setenv("SEED_ADMIN_EMAIL", "admin-seed@test.local")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("ADMIN_NAME", "Seeder Admin")

	seeders := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"level-configs", SeedLevelConfigs},
		{"article-categories", SeedArticleCategories},
		{"song-categories", SeedSongCategories},
		{"forum-categories", SeedForumCategories},
		{"story-categories", SeedStoryCategories},
		{"breathing-techniques", SeedBreathingTechniques},
		{"feature-definitions", SeedFeatureDefinitions},
		{"badge-definitions", SeedBadgeDefinitions},
		{"crisis-keywords", SeedCrisisKeywords},
		{"admin-user", SeedAdminUser},
	}

	for _, seeder := range seeders {
		t.Run(seeder.name, func(t *testing.T) {
			if err := seeder.fn(db); err != nil {
				t.Fatalf("first run failed: %v", err)
			}
			if err := seeder.fn(db); err != nil {
				t.Fatalf("second run failed: %v", err)
			}
		})
	}
}

func TestSeedAdminUser_EnvFallbackAndLegacyBranches(t *testing.T) {
	db := newProductionSeedTestDB(t)

	t.Setenv("SEED_ADMIN_EMAIL", "")
	t.Setenv("SEED_ADMIN_PASSWORD", "")
	t.Setenv("ADMIN_EMAIL", "legacy-admin@test.local")
	t.Setenv("ADMIN_PASSWORD", "legacy-password")
	t.Setenv("ADMIN_NAME", "Legacy Admin")

	if err := SeedAdminUser(db); err != nil {
		t.Fatalf("seed admin with legacy env failed: %v", err)
	}

	var legacyUser model.User
	if err := db.Where("email = ?", "legacy-admin@test.local").First(&legacyUser).Error; err != nil {
		t.Fatalf("expected legacy admin user to exist: %v", err)
	}
	if legacyUser.Name != "Legacy Admin" {
		t.Fatalf("expected admin name from ADMIN_NAME, got %q", legacyUser.Name)
	}

	if err := SeedAdminUser(db); err != nil {
		t.Fatalf("second run should be idempotent, got %v", err)
	}

	t.Setenv("SEED_ADMIN_EMAIL", "seed-admin@test.local")
	t.Setenv("SEED_ADMIN_PASSWORD", "seed-password")
	t.Setenv("ADMIN_EMAIL", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("ADMIN_NAME", "")

	if err := SeedAdminUser(db); err != nil {
		t.Fatalf("seed admin with SEED_* env failed: %v", err)
	}

	var seedUser model.User
	if err := db.Where("email = ?", "seed-admin@test.local").First(&seedUser).Error; err != nil {
		t.Fatalf("expected seed admin user to exist: %v", err)
	}
	if seedUser.Name != "Admin" {
		t.Fatalf("expected default admin name when ADMIN_NAME is empty, got %q", seedUser.Name)
	}
}

func TestProductionSeeders_ErrorOnMissingSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	t.Setenv("SEED_ADMIN_EMAIL", "admin-seed@test.local")
	t.Setenv("SEED_ADMIN_PASSWORD", "password123")
	t.Setenv("ADMIN_NAME", "Seeder Admin")

	seeders := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"level-configs", SeedLevelConfigs},
		{"article-categories", SeedArticleCategories},
		{"song-categories", SeedSongCategories},
		{"forum-categories", SeedForumCategories},
		{"story-categories", SeedStoryCategories},
		{"breathing-techniques", SeedBreathingTechniques},
		{"feature-definitions", SeedFeatureDefinitions},
		{"badge-definitions", SeedBadgeDefinitions},
		{"crisis-keywords", SeedCrisisKeywords},
		{"admin-user", SeedAdminUser},
	}

	for _, seeder := range seeders {
		t.Run(seeder.name, func(t *testing.T) {
			if err := seeder.fn(db); err == nil {
				t.Fatalf("expected %s to fail without schema", seeder.name)
			}
		})
	}
}
