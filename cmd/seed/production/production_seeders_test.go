package production

import (
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
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
		{"default-accounts", SeedDefaultAccounts},
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

func TestSeedDefaultAccounts_IdempotentAndPasswordEnv(t *testing.T) {
	db := newProductionSeedTestDB(t)
	t.Setenv("SEED_DEFAULT_USER_PASSWORD", "seeded-password")

	if err := SeedDefaultAccounts(db); err != nil {
		t.Fatalf("seed default accounts failed: %v", err)
	}
	if err := SeedDefaultAccounts(db); err != nil {
		t.Fatalf("second run should be idempotent, got %v", err)
	}

	var users []model.User
	if err := db.Where("email IN ?", []string{
		"admin@ruang-tenang.com",
		"moderator@ruang-tenang.com",
		"gading@gmail.com",
		"dery@gmail.com",
		"andhika@gmail.com",
	}).Find(&users).Error; err != nil {
		t.Fatalf("query seeded users failed: %v", err)
	}

	if len(users) != 5 {
		t.Fatalf("expected 5 default accounts, got %d", len(users))
	}

	for _, user := range users {
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("seeded-password")); err != nil {
			t.Fatalf("expected bcrypt password from env for %s", user.Email)
		}
	}
}

func TestSeedDefaultAccounts_AdminEnvOverride(t *testing.T) {
	db := newProductionSeedTestDB(t)
	t.Setenv("SEED_DEFAULT_USER_PASSWORD", "member-pass")
	t.Setenv("SEED_ADMIN_EMAIL", "admin-seed@test.local")
	t.Setenv("SEED_ADMIN_PASSWORD", "admin-pass")
	t.Setenv("SEED_ADMIN_NAME", "Seed Admin")

	if err := SeedDefaultAccounts(db); err != nil {
		t.Fatalf("seed default accounts failed: %v", err)
	}

	var admin model.User
	if err := db.Where("email = ?", "admin-seed@test.local").First(&admin).Error; err != nil {
		t.Fatalf("expected overridden admin to exist: %v", err)
	}
	if admin.Name != "Seed Admin" {
		t.Fatalf("expected overridden admin name, got %q", admin.Name)
	}
	if admin.Role != model.RoleAdmin {
		t.Fatalf("expected admin role, got %s", admin.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte("admin-pass")); err != nil {
		t.Fatalf("expected overridden admin password hash")
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
		{"default-accounts", SeedDefaultAccounts},
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
