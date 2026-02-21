package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLevelConfigService(t *testing.T) (*LevelConfigService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE level_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level INTEGER UNIQUE,
		min_exp INTEGER,
		badge_name TEXT,
		badge_icon TEXT,
		tier_name TEXT,
		tier_color TEXT,
		description TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create level_configs: %v", err)
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (1, 1, 0, 'Pemula', '🌱', 'Bronze', '#A97142', 'start', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed level config 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (2, 2, 100, 'Naik', '⭐', 'Silver', '#C0C0C0', 'next', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed level config 2: %v", err)
	}

	cache := NewCacheService()
	svc := NewLevelConfigService(repository.NewLevelConfigRepository(db), cache)
	return svc, db
}

func TestLevelConfigService_ReadAndCachePaths(t *testing.T) {
	svc, db := setupLevelConfigService(t)
	ctx := context.Background()

	configs, err := svc.GetAll(ctx)
	if err != nil || len(configs) < 2 {
		t.Fatalf("get all failed: err=%v len=%d", err, len(configs))
	}

	byID, err := svc.GetByID(ctx, 1)
	if err != nil || byID.Level != 1 {
		t.Fatalf("get by id failed: err=%v data=%+v", err, byID)
	}
	byExp, err := svc.GetLevelByExp(ctx, 120)
	if err != nil || byExp.Level != 2 {
		t.Fatalf("get level by exp failed: err=%v data=%+v", err, byExp)
	}
	next, err := svc.GetNextLevel(ctx, 1)
	if err != nil || next.Level != 2 {
		t.Fatalf("get next level failed: err=%v data=%+v", err, next)
	}

	if err := db.Exec(`DROP TABLE level_configs`).Error; err != nil {
		t.Fatalf("drop table: %v", err)
	}
	cachedConfigs, err := svc.GetAll(ctx)
	if err != nil || len(cachedConfigs) < 2 {
		t.Fatalf("expected cache hit for get all: err=%v len=%d", err, len(cachedConfigs))
	}
}

func TestLevelConfigService_WritePaths(t *testing.T) {
	svc, _ := setupLevelConfigService(t)
	ctx := context.Background()

	err := svc.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 10, BadgeName: "dup", BadgeIcon: "x"})
	if !errors.Is(err, ErrLevelExists) {
		t.Fatalf("expected ErrLevelExists on duplicate create, got %v", err)
	}

	if err := svc.Create(ctx, &model.LevelConfig{Level: 3, MinExp: 200, BadgeName: "L3", BadgeIcon: "🔥", TierName: "Gold", TierColor: "#FFD700"}); err != nil {
		t.Fatalf("create level failed: %v", err)
	}

	err = svc.Update(ctx, 1, &model.LevelConfig{Level: 2, MinExp: 50, BadgeName: "U", BadgeIcon: "U"})
	if !errors.Is(err, ErrLevelExists) {
		t.Fatalf("expected ErrLevelExists on duplicate update level, got %v", err)
	}

	if err := svc.Update(ctx, 1, &model.LevelConfig{Level: 1, MinExp: 10, BadgeName: "Updated", BadgeIcon: "🆕"}); err != nil {
		t.Fatalf("update level failed: %v", err)
	}

	if err := svc.Delete(ctx, 2); err != nil {
		t.Fatalf("delete level failed: %v", err)
	}
}

func TestLevelConfigService_GetUserLevelInfoFallback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE level_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level INTEGER UNIQUE,
		min_exp INTEGER,
		badge_name TEXT,
		badge_icon TEXT,
		tier_name TEXT,
		tier_color TEXT,
		description TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create level_configs: %v", err)
	}

	svc := NewLevelConfigService(repository.NewLevelConfigRepository(db), NewCacheService())
	current, next, err := svc.GetUserLevelInfo(context.Background(), 10)
	if err != nil {
		t.Fatalf("get user level info fallback err: %v", err)
	}
	if current == nil || current.Level != 1 || current.BadgeName == "" {
		t.Fatalf("unexpected fallback current level: %+v", current)
	}
	if next != nil {
		t.Fatalf("expected nil next level for fallback, got %+v", next)
	}
}
