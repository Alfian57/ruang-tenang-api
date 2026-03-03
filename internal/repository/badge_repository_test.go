package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBadgeRepositoryTest(t *testing.T) (*BadgeRepository, *gorm.DB, uint, uuid.UUID, uuid.UUID) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			username TEXT,
			email TEXT,
			password TEXT,
			exp INTEGER,
			current_streak INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE badge_definitions (
			id TEXT PRIMARY KEY,
			badge_key TEXT UNIQUE,
			badge_name TEXT,
			description TEXT,
			icon TEXT,
			category TEXT,
			requirement_type TEXT,
			requirement_value INTEGER,
			is_active BOOLEAN,
			display_order INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE user_badges (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			badge_id TEXT,
			earned_at DATETIME,
			is_showcased BOOLEAN
		)`,
	}
	for _, stmt := range schema {
		if execErr := db.Exec(stmt).Error; execErr != nil {
			t.Fatalf("create schema failed: %v", execErr)
		}
	}

	repo := NewBadgeRepository(db)
	if err := db.Exec(`INSERT INTO users (id, name, username, email, password, exp, current_streak, created_at, updated_at, deleted_at)
		VALUES (1, 'Badge User', 'badgeuser', 'badge@test.local', 'x', 150, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)`).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	userID := uint(1)

	badgeEarnedID := uuid.New()
	badgeManualID := uuid.New()
	seedBadges := []model.BadgeDefinition{
		{
			ID:               badgeEarnedID,
			BadgeKey:         "first_steps",
			BadgeName:        "First Steps",
			Category:         "progress",
			RequirementType:  model.BadgeRequirementManual,
			RequirementValue: 1,
			IsActive:         true,
		},
		{
			ID:               badgeManualID,
			BadgeKey:         "writer_manual",
			BadgeName:        "Writer Manual",
			Category:         "writing",
			RequirementType:  model.BadgeRequirementManual,
			RequirementValue: 3,
			IsActive:         true,
		},
	}
	for i := range seedBadges {
		if err := db.Create(&seedBadges[i]).Error; err != nil {
			t.Fatalf("seed badge %d: %v", i+1, err)
		}
	}

	if err := db.Create(&model.UserBadge{ID: uuid.New(), UserID: userID, BadgeID: badgeEarnedID, EarnedAt: time.Now().Add(-time.Hour)}).Error; err != nil {
		t.Fatalf("seed user badge: %v", err)
	}

	return repo, db, userID, badgeEarnedID, badgeManualID
}

func TestBadgeRepository_Branches(t *testing.T) {
	repo, db, userID, badgeEarnedID, badgeManualID := setupBadgeRepositoryTest(t)
	ctx := context.Background()

	all, err := repo.GetAllBadgeDefinitions(ctx)
	if err != nil || len(all) < 2 {
		t.Fatalf("GetAllBadgeDefinitions unexpected err=%v len=%d", err, len(all))
	}

	byCategory, err := repo.GetBadgesByCategory(ctx, "writing")
	if err != nil || len(byCategory) != 1 {
		t.Fatalf("GetBadgesByCategory unexpected err=%v len=%d", err, len(byCategory))
	}

	byKey, err := repo.GetBadgeByKey(ctx, "first_steps")
	if err != nil || byKey.ID != badgeEarnedID {
		t.Fatalf("GetBadgeByKey unexpected badge=%+v err=%v", byKey, err)
	}
	if _, err := repo.GetBadgeByKey(ctx, "missing_badge"); err == nil {
		t.Fatal("expected GetBadgeByKey missing error")
	}

	byID, err := repo.GetBadgeByID(ctx, badgeManualID)
	if err != nil || byID.BadgeKey != "writer_manual" {
		t.Fatalf("GetBadgeByID unexpected badge=%+v err=%v", byID, err)
	}
	if _, err := repo.GetBadgeByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected GetBadgeByID missing error")
	}

	byReq, err := repo.GetBadgesByRequirementType(ctx, string(model.BadgeRequirementManual))
	if err != nil || len(byReq) < 2 {
		t.Fatalf("GetBadgesByRequirementType unexpected err=%v len=%d", err, len(byReq))
	}

	newBadge := &model.BadgeDefinition{
		ID:               uuid.New(),
		BadgeKey:         "xp_runner",
		BadgeName:        "XP Runner",
		Category:         "progress",
		RequirementType:  model.BadgeRequirementXP,
		RequirementValue: 100,
		IsActive:         true,
	}
	if err := repo.CreateBadgeDefinition(ctx, newBadge); err != nil {
		t.Fatalf("CreateBadgeDefinition: %v", err)
	}
	newBadge.BadgeName = "XP Runner Updated"
	if err := repo.UpdateBadgeDefinition(ctx, newBadge); err != nil {
		t.Fatalf("UpdateBadgeDefinition: %v", err)
	}

	userBadges, err := repo.GetUserBadges(ctx, userID)
	if err != nil || len(userBadges) == 0 {
		t.Fatalf("GetUserBadges unexpected err=%v len=%d", err, len(userBadges))
	}
	userWritingBadges, err := repo.GetUserBadgesByCategory(ctx, userID, "progress")
	if err != nil || len(userWritingBadges) == 0 {
		t.Fatalf("GetUserBadgesByCategory unexpected err=%v len=%d", err, len(userWritingBadges))
	}

	if !repo.HasBadge(ctx, userID, badgeEarnedID) {
		t.Fatal("expected HasBadge true")
	}
	if repo.HasBadge(ctx, userID, uuid.New()) {
		t.Fatal("expected HasBadge false")
	}

	if !repo.HasBadgeByKey(ctx, userID, "first_steps") {
		t.Fatal("expected HasBadgeByKey true")
	}
	if repo.HasBadgeByKey(ctx, userID, "missing_badge") {
		t.Fatal("expected HasBadgeByKey false")
	}

	if err := repo.AwardBadge(ctx, userID, badgeManualID); err != nil {
		t.Fatalf("AwardBadge: %v", err)
	}
	if err := repo.AwardBadgeByKey(ctx, userID, "writer_manual"); err != nil {
		t.Fatalf("AwardBadgeByKey already-earned should be nil: %v", err)
	}
	if err := repo.AwardBadgeByKey(ctx, userID, "missing_badge"); err == nil {
		t.Fatal("expected AwardBadgeByKey missing badge error")
	}

	badgeCount, err := repo.GetUserBadgeCount(ctx, userID)
	if err != nil || badgeCount < 2 {
		t.Fatalf("GetUserBadgeCount unexpected count=%d err=%v", badgeCount, err)
	}

	progress, err := repo.GetBadgeProgress(ctx, userID)
	if err != nil || len(progress) < 2 {
		t.Fatalf("GetBadgeProgress unexpected err=%v len=%d", err, len(progress))
	}

	recent, err := repo.GetRecentlyEarnedBadges(ctx, userID, time.Now().Add(-24*time.Hour))
	if err != nil || len(recent) == 0 {
		t.Fatalf("GetRecentlyEarnedBadges unexpected err=%v len=%d", err, len(recent))
	}

	if err := db.Exec(`DROP TABLE badge_definitions`).Error; err != nil {
		t.Fatalf("drop badge_definitions: %v", err)
	}
	if _, err := repo.GetAllBadgeDefinitions(ctx); err == nil {
		t.Fatal("expected GetAllBadgeDefinitions error when table missing")
	}
}
