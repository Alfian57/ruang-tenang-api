package repository

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFeatureUnlockRepositoryTest(t *testing.T) (*FeatureUnlockRepository, *gorm.DB, uuid.UUID, uuid.UUID) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE feature_definitions (
			id TEXT PRIMARY KEY,
			feature_key TEXT UNIQUE,
			feature_name TEXT,
			description TEXT,
			icon TEXT,
			required_level INTEGER,
			category TEXT,
			is_active BOOLEAN,
			display_order INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE user_feature_unlocks (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			feature_id TEXT,
			unlocked_at DATETIME
		)`,
	}
	for _, stmt := range schema {
		if execErr := db.Exec(stmt).Error; execErr != nil {
			t.Fatalf("create schema failed: %v", execErr)
		}
	}

	repo := NewFeatureUnlockRepository(db)
	f1 := uuid.New()
	f2 := uuid.New()
	fInactive := uuid.New()

	seed := []model.FeatureDefinition{
		{ID: f1, FeatureKey: "chat_pro", FeatureName: "Chat Pro", RequiredLevel: 1, Category: "ai", IsActive: true},
		{ID: f2, FeatureKey: "story_plus", FeatureName: "Story Plus", RequiredLevel: 2, Category: "content", IsActive: true},
		{ID: fInactive, FeatureKey: "old_feature", FeatureName: "Old Feature", RequiredLevel: 1, Category: "legacy", IsActive: false},
	}
	for i := range seed {
		if err := repo.CreateFeatureDefinition(context.Background(), &seed[i]); err != nil {
			t.Fatalf("seed feature %d: %v", i+1, err)
		}
	}
	if err := db.Exec(`UPDATE feature_definitions SET is_active = 0 WHERE feature_key = ?`, "old_feature").Error; err != nil {
		t.Fatalf("force inactive feature state: %v", err)
	}

	if err := db.Exec(`INSERT INTO user_feature_unlocks (id, user_id, feature_id, unlocked_at) VALUES (?, 1, ?, CURRENT_TIMESTAMP)`, uuid.New().String(), f1.String()).Error; err != nil {
		t.Fatalf("seed user unlock: %v", err)
	}

	return repo, db, f1, f2
}

func TestFeatureUnlockRepository_Branches(t *testing.T) {
	repo, db, f1, f2 := setupFeatureUnlockRepositoryTest(t)
	ctx := context.Background()

	all, err := repo.GetAllFeatureDefinitions(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("GetAllFeatureDefinitions unexpected err=%v len=%d", err, len(all))
	}

	byLevel, err := repo.GetFeaturesByLevel(ctx, 1)
	if err != nil || len(byLevel) != 1 || byLevel[0].FeatureKey != "chat_pro" {
		t.Fatalf("GetFeaturesByLevel unexpected err=%v data=%+v", err, byLevel)
	}

	upToLevel, err := repo.GetFeaturesUpToLevel(ctx, 2)
	if err != nil || len(upToLevel) != 2 {
		t.Fatalf("GetFeaturesUpToLevel unexpected err=%v len=%d", err, len(upToLevel))
	}

	featureByKey, err := repo.GetFeatureByKey(ctx, "chat_pro")
	if err != nil || featureByKey.ID != f1 {
		t.Fatalf("GetFeatureByKey unexpected feature=%+v err=%v", featureByKey, err)
	}
	if _, err := repo.GetFeatureByKey(ctx, "missing_key"); err == nil {
		t.Fatal("expected GetFeatureByKey missing error")
	}

	featureByID, err := repo.GetFeatureByID(ctx, f2)
	if err != nil || featureByID.FeatureKey != "story_plus" {
		t.Fatalf("GetFeatureByID unexpected feature=%+v err=%v", featureByID, err)
	}
	if _, err := repo.GetFeatureByID(ctx, uuid.New()); err == nil {
		t.Fatal("expected GetFeatureByID missing error")
	}

	newFeature := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "breath_plus", FeatureName: "Breath Plus", RequiredLevel: 3, Category: "ai", IsActive: true}
	if err := repo.CreateFeatureDefinition(ctx, newFeature); err != nil {
		t.Fatalf("CreateFeatureDefinition: %v", err)
	}
	newFeature.FeatureName = "Breath Plus Updated"
	if err := repo.UpdateFeatureDefinition(ctx, newFeature); err != nil {
		t.Fatalf("UpdateFeatureDefinition: %v", err)
	}

	unlocks, err := repo.GetUserUnlockedFeatures(ctx, 1)
	if err != nil || len(unlocks) == 0 {
		t.Fatalf("GetUserUnlockedFeatures unexpected err=%v len=%d", err, len(unlocks))
	}

	if !repo.IsFeatureUnlocked(ctx, 1, f1) {
		t.Fatal("expected IsFeatureUnlocked true")
	}
	if repo.IsFeatureUnlocked(ctx, 1, uuid.New()) {
		t.Fatal("expected IsFeatureUnlocked false")
	}

	if !repo.IsFeatureUnlockedByKey(ctx, 1, "chat_pro") {
		t.Fatal("expected IsFeatureUnlockedByKey true")
	}
	if repo.IsFeatureUnlockedByKey(ctx, 1, "missing_key") {
		t.Fatal("expected IsFeatureUnlockedByKey false")
	}

	if err := repo.UnlockFeature(ctx, 2, f2); err != nil {
		t.Fatalf("UnlockFeature: %v", err)
	}

	newlyUnlocked, err := repo.UnlockFeaturesForLevel(ctx, 3, 2)
	if err != nil {
		t.Fatalf("UnlockFeaturesForLevel: %v", err)
	}
	if len(newlyUnlocked) == 0 {
		t.Fatal("expected UnlockFeaturesForLevel to return newly unlocked features")
	}

	newlyAvailable, err := repo.GetNewlyAvailableFeatures(ctx, 3, 1, 3)
	if err != nil || len(newlyAvailable) == 0 {
		t.Fatalf("GetNewlyAvailableFeatures unexpected err=%v len=%d", err, len(newlyAvailable))
	}

	unlockedCount, totalCount, err := repo.GetUserFeatureStats(ctx, 1)
	if err != nil || totalCount < 2 || unlockedCount < 1 {
		t.Fatalf("GetUserFeatureStats unexpected unlocked=%d total=%d err=%v", unlockedCount, totalCount, err)
	}

	upcoming, err := repo.GetUpcomingFeatures(ctx, 1, 5)
	if err != nil || len(upcoming) == 0 {
		t.Fatalf("GetUpcomingFeatures unexpected err=%v len=%d", err, len(upcoming))
	}

	byCategory, err := repo.GetFeaturesByCategory(ctx, "ai")
	if err != nil || len(byCategory) == 0 {
		t.Fatalf("GetFeaturesByCategory unexpected err=%v len=%d", err, len(byCategory))
	}

	if err := db.Exec(`DROP TABLE feature_definitions`).Error; err != nil {
		t.Fatalf("drop feature_definitions: %v", err)
	}
	if _, err := repo.UnlockFeaturesForLevel(ctx, 1, 1); err == nil {
		t.Fatal("expected UnlockFeaturesForLevel error when feature_definitions table missing")
	}
	if _, err := repo.GetFeaturesByCategory(ctx, "ai"); err == nil {
		t.Fatal("expected GetFeaturesByCategory error when table missing")
	}
}
