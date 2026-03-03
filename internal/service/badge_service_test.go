package service

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newBadgeServiceForTest(t *testing.T) (*BadgeService, uint, *model.BadgeDefinition, *model.BadgeDefinition) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("migrate base: %v", err)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS badge_definitions (
		id TEXT PRIMARY KEY,
		badge_key TEXT NOT NULL UNIQUE,
		badge_name TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		category TEXT,
		requirement_type TEXT NOT NULL,
		requirement_value INTEGER,
		is_active NUMERIC,
		display_order INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create badge_definitions: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS user_badges (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		badge_id TEXT NOT NULL,
		earned_at DATETIME,
		is_showcased NUMERIC DEFAULT 0
	)`).Error; err != nil {
		t.Fatalf("create user_badges: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS exp_histories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		activity_type TEXT,
		points INTEGER,
		description TEXT,
		created_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create exp_histories: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS inspiring_stories (
		id TEXT PRIMARY KEY,
		author_id INTEGER,
		status TEXT,
		created_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create inspiring_stories: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)

	ctx := context.Background()
	user := &model.User{Name: "BadgeSvc", Username: "badgesvc", Email: "badgesvc@example.id", Password: "x", Exp: 120, CurrentStreak: 5}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	_ = levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱"})
	_ = levelRepo.Create(ctx, &model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐"})

	b1 := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "streak_3_days", BadgeName: "Streak 3", Category: "streak", RequirementType: model.BadgeRequirementStreak, RequirementValue: 3, IsActive: true}
	b2 := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "level_2", BadgeName: "Level 2", Category: "level", RequirementType: model.BadgeRequirementLevel, RequirementValue: 2, IsActive: true}
	b3 := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "first_chat", BadgeName: "First Chat", Category: "activity", RequirementType: model.BadgeRequirementManual, RequirementValue: 1, IsActive: true}
	if err := badgeRepo.CreateBadgeDefinition(ctx, b1); err != nil {
		t.Fatalf("create badge1: %v", err)
	}
	if err := badgeRepo.CreateBadgeDefinition(ctx, b2); err != nil {
		t.Fatalf("create badge2: %v", err)
	}
	if err := badgeRepo.CreateBadgeDefinition(ctx, b3); err != nil {
		t.Fatalf("create badge3: %v", err)
	}

	return NewBadgeService(badgeRepo, userRepo, levelRepo), user.ID, b1, b2
}

func TestBadgeService_BasicsAndAwarding(t *testing.T) {
	ctx := context.Background()
	svc, userID, b1, b2 := newBadgeServiceForTest(t)

	if cats := svc.GetBadgeCategories(ctx); len(cats) == 0 {
		t.Fatalf("expected non-empty badge categories")
	}

	if all, err := svc.GetAllBadges(ctx); err != nil || len(all) < 2 {
		t.Fatalf("get all badges failed: len=%d err=%v", len(all), err)
	}
	if byCat, err := svc.GetBadgesByCategory(ctx, "streak"); err != nil || len(byCat) != 1 {
		t.Fatalf("get badges by category failed: len=%d err=%v", len(byCat), err)
	}

	if _, err := svc.AwardBadge(ctx, userID, "unknown"); err == nil {
		t.Fatalf("expected award unknown badge error")
	}

	awarded, err := svc.AwardBadge(ctx, userID, b1.BadgeKey)
	if err != nil || awarded == nil {
		t.Fatalf("award badge failed: badge=%+v err=%v", awarded, err)
	}

	if _, err := svc.AwardBadge(ctx, userID, b1.BadgeKey); err == nil {
		t.Fatalf("expected already-earned badge error")
	}

	if _, err := svc.AwardBadgeByID(ctx, userID, uuid.New()); err == nil {
		t.Fatalf("expected award by unknown id error")
	}
	if _, err := svc.AwardBadgeByID(ctx, userID, b2.ID); err != nil {
		t.Fatalf("award by id should succeed, got %v", err)
	}
	if _, err := svc.AwardBadgeByID(ctx, userID, b2.ID); err == nil {
		t.Fatalf("expected already-earned error when awarding same badge id twice")
	}
}

func TestBadgeService_AwardBadgeByID_AwardInsertError(t *testing.T) {
	ctx := context.Background()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("migrate base: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS badge_definitions (
		id TEXT PRIMARY KEY,
		badge_key TEXT NOT NULL UNIQUE,
		badge_name TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		category TEXT,
		requirement_type TEXT NOT NULL,
		requirement_value INTEGER,
		is_active NUMERIC,
		display_order INTEGER,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create badge_definitions: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS user_badges (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		badge_id TEXT NOT NULL,
		earned_at DATETIME,
		is_showcased NUMERIC DEFAULT 0
	)`).Error; err != nil {
		t.Fatalf("create user_badges: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)

	user := &model.User{Name: "Badge Error", Username: "badgeerror", Email: "badgeerror@example.id", Password: "x"}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	badge := &model.BadgeDefinition{ID: uuid.New(), BadgeKey: "error_badge", BadgeName: "Error Badge", Category: "activity", RequirementType: model.BadgeRequirementManual, RequirementValue: 1, IsActive: true}
	if err := badgeRepo.CreateBadgeDefinition(ctx, badge); err != nil {
		t.Fatalf("create badge: %v", err)
	}

	svc := NewBadgeService(badgeRepo, userRepo, levelRepo)

	if err := db.Migrator().DropTable(&model.UserBadge{}); err != nil {
		t.Fatalf("drop user_badges table: %v", err)
	}

	if _, err := svc.AwardBadgeByID(ctx, user.ID, badge.ID); err == nil {
		t.Fatalf("expected insert error when user_badges table is missing")
	}
}

func TestBadgeService_UserViewsAndChecks(t *testing.T) {
	ctx := context.Background()
	svc, userID, _, _ := newBadgeServiceForTest(t)

	newBadges, err := svc.CheckAndAwardBadges(ctx, userID)
	if err != nil {
		t.Fatalf("check and award badges failed: %v", err)
	}
	if len(newBadges) == 0 {
		t.Fatalf("expected at least one new badge")
	}

	userBadges, err := svc.GetUserBadges(ctx, userID)
	if err != nil || userBadges == nil {
		t.Fatalf("get user badges failed: resp=%+v err=%v", userBadges, err)
	}

	progress, err := svc.GetBadgeProgress(ctx, userID)
	if err != nil || len(progress) == 0 {
		t.Fatalf("get badge progress failed: len=%d err=%v", len(progress), err)
	}

	recent, err := svc.GetRecentlyEarnedBadges(ctx, userID, 30)
	if err != nil || len(recent) == 0 {
		t.Fatalf("get recently earned badges failed: len=%d err=%v", len(recent), err)
	}

	display, err := svc.GetDisplayBadges(ctx, userID, 5)
	if err != nil {
		t.Fatalf("get display badges failed: %v", err)
	}
	if len(display) < 0 {
		t.Fatalf("unexpected display badge length")
	}
}

func TestBadgeService_GetBadgeProgress_ErrorBranch(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)
	svc := NewBadgeService(badgeRepo, userRepo, levelRepo)

	if _, err := svc.GetBadgeProgress(ctx, 1); err == nil {
		t.Fatal("expected GetBadgeProgress error when badge_definitions table is missing")
	}
}
