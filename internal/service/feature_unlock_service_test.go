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

func newFeatureServiceDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.LevelConfig{}); err != nil {
		t.Fatalf("auto migrate base tables: %v", err)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS feature_definitions (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		feature_key TEXT NOT NULL UNIQUE,
		feature_name TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		required_level INTEGER NOT NULL DEFAULT 1,
		category TEXT DEFAULT 'general',
		is_active NUMERIC DEFAULT 1,
		display_order INTEGER DEFAULT 0,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create feature_definitions table: %v", err)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS user_feature_unlocks (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		user_id INTEGER NOT NULL,
		feature_id TEXT NOT NULL,
		unlocked_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create user_feature_unlocks table: %v", err)
	}

	return db
}

func TestFeatureUnlockService_Flow(t *testing.T) {
	ctx := context.Background()
	db := newFeatureServiceDB(t)

	featureRepo := repository.NewFeatureUnlockRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	userRepo := repository.NewUserRepository(db)
	svc := NewFeatureUnlockService(featureRepo, levelRepo, userRepo)

	user := &model.User{Name: "FU", Username: "fu_user", Email: "fu@example.id", Password: "x", Exp: 150}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱", TierName: "Bronze", TierColor: "#aaa"}); err != nil {
		t.Fatalf("create level 1: %v", err)
	}
	if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 2, MinExp: 100, BadgeName: "Naik", BadgeIcon: "⭐", TierName: "Silver", TierColor: "#bbb"}); err != nil {
		t.Fatalf("create level 2: %v", err)
	}
	if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 3, MinExp: 200, BadgeName: "Lanjut", BadgeIcon: "🚀", TierName: "Gold", TierColor: "#ccc"}); err != nil {
		t.Fatalf("create level 3: %v", err)
	}

	f1 := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "chat_pro", FeatureName: "Chat Pro", RequiredLevel: 1, Category: "ai", IsActive: true}
	f2 := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "story_mode", FeatureName: "Story Mode", RequiredLevel: 2, Category: "content", IsActive: true}
	f3 := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "badge_room", FeatureName: "Badge Room", RequiredLevel: 3, Category: "special", IsActive: true}
	if err := featureRepo.CreateFeatureDefinition(ctx, f1); err != nil {
		t.Fatalf("create feature1: %v", err)
	}
	if err := featureRepo.CreateFeatureDefinition(ctx, f2); err != nil {
		t.Fatalf("create feature2: %v", err)
	}
	if err := featureRepo.CreateFeatureDefinition(ctx, f3); err != nil {
		t.Fatalf("create feature3: %v", err)
	}

	if grouped, err := svc.GetAllFeatures(ctx); err != nil || len(grouped) < 2 {
		t.Fatalf("get all features failed: len=%d err=%v", len(grouped), err)
	}

	if byCategory, err := svc.GetFeaturesByCategory(ctx, "ai"); err != nil || len(byCategory) != 1 {
		t.Fatalf("get features by category failed: len=%d err=%v", len(byCategory), err)
	}

	if categories := svc.GetFeatureCategories(ctx); len(categories) == 0 {
		t.Fatalf("get feature categories failed: len=%d", len(categories))
	}

	if userFeatures, err := svc.GetUserFeatures(ctx, user.ID); err != nil || userFeatures.CurrentLevel != 2 {
		t.Fatalf("get user features failed: %+v err=%v", userFeatures, err)
	}

	if access, err := svc.CheckFeatureAccess(ctx, user.ID, "unknown_feature"); err != nil || access.HasAccess {
		t.Fatalf("unknown feature access should be false: access=%+v err=%v", access, err)
	}

	if access, err := svc.CheckFeatureAccess(ctx, user.ID, "chat_pro"); err != nil || !access.HasAccess {
		t.Fatalf("chat_pro should be accessible by level: access=%+v err=%v", access, err)
	}

	if access, err := svc.CheckFeatureAccess(ctx, user.ID, "badge_room"); err != nil || access.HasAccess {
		t.Fatalf("badge_room should still be locked: access=%+v err=%v", access, err)
	}

	if newly, err := svc.UnlockFeaturesOnLevelUp(ctx, user.ID, 3); err != nil || len(newly) < 1 {
		t.Fatalf("unlock features on level up failed: len=%d err=%v", len(newly), err)
	}

	if upcoming, err := svc.GetUpcomingFeatures(ctx, user.ID, 10); err != nil {
		t.Fatalf("get upcoming features failed: %v", err)
	} else if len(upcoming) < 0 {
		t.Fatalf("unexpected upcoming features length")
	}
}

func TestFeatureUnlockService_ErrorAndAdditionalBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("get features by category error", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		svc := NewFeatureUnlockService(repository.NewFeatureUnlockRepository(db), repository.NewLevelConfigRepository(db), repository.NewUserRepository(db))

		if err := db.Exec(`DROP TABLE feature_definitions`).Error; err != nil {
			t.Fatalf("drop feature_definitions: %v", err)
		}
		if _, err := svc.GetFeaturesByCategory(ctx, "ai"); err == nil {
			t.Fatal("expected GetFeaturesByCategory error")
		}
	})

	t.Run("get user features with explicit unlock", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		featureRepo := repository.NewFeatureUnlockRepository(db)
		levelRepo := repository.NewLevelConfigRepository(db)
		userRepo := repository.NewUserRepository(db)
		svc := NewFeatureUnlockService(featureRepo, levelRepo, userRepo)

		user := &model.User{Name: "Unlock User", Username: "unlock_u", Email: "unlock_u@test.local", Password: "x", Exp: 10}
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Seed", BadgeIcon: "🌱"}); err != nil {
			t.Fatalf("create level: %v", err)
		}

		feat := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "unlock_only", FeatureName: "Unlock Only", RequiredLevel: 5, Category: "special", IsActive: true}
		if err := featureRepo.CreateFeatureDefinition(ctx, feat); err != nil {
			t.Fatalf("create feature: %v", err)
		}
		if err := featureRepo.UnlockFeature(ctx, user.ID, feat.ID); err != nil {
			t.Fatalf("unlock feature: %v", err)
		}

		resp, err := svc.GetUserFeatures(ctx, user.ID)
		if err != nil {
			t.Fatalf("GetUserFeatures failed: %v", err)
		}
		if resp.TotalUnlocked < 1 {
			t.Fatalf("expected unlocked features > 0, got %+v", resp)
		}
	})

	t.Run("get user features user not found", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		svc := NewFeatureUnlockService(repository.NewFeatureUnlockRepository(db), repository.NewLevelConfigRepository(db), repository.NewUserRepository(db))

		if _, err := svc.GetUserFeatures(ctx, 99999); err == nil {
			t.Fatal("expected GetUserFeatures user not found error")
		}
	})

	t.Run("get user features level lookup error", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		featureRepo := repository.NewFeatureUnlockRepository(db)
		levelRepo := repository.NewLevelConfigRepository(db)
		userRepo := repository.NewUserRepository(db)
		svc := NewFeatureUnlockService(featureRepo, levelRepo, userRepo)

		user := &model.User{Name: "Level Err", Username: "lvl_err", Email: "lvl_err@test.local", Password: "x", Exp: 100}
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if err := db.Exec(`DROP TABLE level_configs`).Error; err != nil {
			t.Fatalf("drop level_configs: %v", err)
		}

		if _, err := svc.GetUserFeatures(ctx, user.ID); err == nil {
			t.Fatal("expected GetUserFeatures level lookup error")
		}
	})

	t.Run("get user features feature list error", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		featureRepo := repository.NewFeatureUnlockRepository(db)
		levelRepo := repository.NewLevelConfigRepository(db)
		userRepo := repository.NewUserRepository(db)
		svc := NewFeatureUnlockService(featureRepo, levelRepo, userRepo)

		user := &model.User{Name: "Feat Err", Username: "feat_err", Email: "feat_err@test.local", Password: "x", Exp: 0}
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Seed", BadgeIcon: "🌱"}); err != nil {
			t.Fatalf("create level: %v", err)
		}
		if err := db.Exec(`DROP TABLE feature_definitions`).Error; err != nil {
			t.Fatalf("drop feature_definitions: %v", err)
		}

		if _, err := svc.GetUserFeatures(ctx, user.ID); err == nil {
			t.Fatal("expected GetUserFeatures feature list error")
		}
	})

	t.Run("unlock features on level up error", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		featureRepo := repository.NewFeatureUnlockRepository(db)
		svc := NewFeatureUnlockService(featureRepo, repository.NewLevelConfigRepository(db), repository.NewUserRepository(db))

		if err := db.Exec(`DROP TABLE feature_definitions`).Error; err != nil {
			t.Fatalf("drop feature_definitions: %v", err)
		}
		if _, err := svc.UnlockFeaturesOnLevelUp(ctx, 1, 2); err == nil {
			t.Fatal("expected UnlockFeaturesOnLevelUp error")
		}
	})

	t.Run("get upcoming features branches", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		featureRepo := repository.NewFeatureUnlockRepository(db)
		levelRepo := repository.NewLevelConfigRepository(db)
		userRepo := repository.NewUserRepository(db)
		svc := NewFeatureUnlockService(featureRepo, levelRepo, userRepo)

		if _, err := svc.GetUpcomingFeatures(ctx, 99999, 5); err == nil {
			t.Fatal("expected GetUpcomingFeatures user not found error")
		}

		user := &model.User{Name: "Upcoming", Username: "upc_u", Email: "upc_u@test.local", Password: "x", Exp: 0}
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Seed", BadgeIcon: "🌱"}); err != nil {
			t.Fatalf("create level: %v", err)
		}

		if err := db.Exec(`DROP TABLE feature_definitions`).Error; err != nil {
			t.Fatalf("drop feature_definitions: %v", err)
		}
		if _, err := svc.GetUpcomingFeatures(ctx, user.ID, 5); err == nil {
			t.Fatal("expected GetUpcomingFeatures repository error")
		}
	})

	t.Run("check feature access additional branches", func(t *testing.T) {
		db := newFeatureServiceDB(t)
		featureRepo := repository.NewFeatureUnlockRepository(db)
		levelRepo := repository.NewLevelConfigRepository(db)
		userRepo := repository.NewUserRepository(db)
		svc := NewFeatureUnlockService(featureRepo, levelRepo, userRepo)

		user := &model.User{Name: "Check Access", Username: "check_access_u", Email: "check_access_u@test.local", Password: "x", Exp: 0}
		if err := userRepo.Create(ctx, user); err != nil {
			t.Fatalf("create user: %v", err)
		}
		if err := levelRepo.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "Seed", BadgeIcon: "🌱"}); err != nil {
			t.Fatalf("create level: %v", err)
		}

		locked := &model.FeatureDefinition{ID: uuid.New(), FeatureKey: "need_level_5", FeatureName: "Need Level 5", RequiredLevel: 5, Category: "special", IsActive: true}
		if err := featureRepo.CreateFeatureDefinition(ctx, locked); err != nil {
			t.Fatalf("create locked feature: %v", err)
		}

		resp, err := svc.CheckFeatureAccess(ctx, user.ID, "need_level_5")
		if err != nil {
			t.Fatalf("CheckFeatureAccess locked path failed: %v", err)
		}
		if resp.HasAccess || resp.RequiredLevel != 5 || resp.CurrentLevel != 1 {
			t.Fatalf("expected locked response with level details, got %+v", resp)
		}

		if _, err := svc.CheckFeatureAccess(ctx, 99999, "need_level_5"); err == nil {
			t.Fatal("expected CheckFeatureAccess user not found error")
		}

		if err := db.Exec(`DROP TABLE level_configs`).Error; err != nil {
			t.Fatalf("drop level_configs: %v", err)
		}
		if _, err := svc.CheckFeatureAccess(ctx, user.ID, "need_level_5"); err == nil {
			t.Fatal("expected CheckFeatureAccess level lookup error")
		}
	})
}
