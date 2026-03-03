package repository

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLevelConfigRepositoryTest(t *testing.T) (*LevelConfigRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.LevelConfig{}); err != nil {
		t.Fatalf("migrate level_configs: %v", err)
	}

	repo := NewLevelConfigRepository(db)
	seed := []model.LevelConfig{
		{Level: 1, MinExp: 0, BadgeName: "Pemula", BadgeIcon: "🌱", TierName: "Bronze", TierColor: "#cd7f32"},
		{Level: 2, MinExp: 100, BadgeName: "Tumbuh", BadgeIcon: "🌿", TierName: "Silver", TierColor: "#c0c0c0"},
		{Level: 3, MinExp: 300, BadgeName: "Maju", BadgeIcon: "🚀", TierName: "Gold", TierColor: "#ffd700"},
	}
	for i := range seed {
		if err := repo.Create(context.Background(), &seed[i]); err != nil {
			t.Fatalf("seed level %d: %v", seed[i].Level, err)
		}
	}

	return repo, db
}

func TestLevelConfigRepository_Branches(t *testing.T) {
	repo, _ := setupLevelConfigRepositoryTest(t)
	ctx := context.Background()

	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(all) != 3 || all[0].Level != 1 {
		t.Fatalf("unexpected GetAll result: %+v", all)
	}

	byID, err := repo.GetByID(ctx, all[1].ID)
	if err != nil || byID.Level != 2 {
		t.Fatalf("GetByID unexpected: cfg=%+v err=%v", byID, err)
	}
	if _, err := repo.GetByID(ctx, 999999); err == nil {
		t.Fatal("expected GetByID missing error")
	}

	byLevel, err := repo.GetByLevel(ctx, 2)
	if err != nil || byLevel.MinExp != 100 {
		t.Fatalf("GetByLevel unexpected: cfg=%+v err=%v", byLevel, err)
	}
	if _, err := repo.GetByLevel(ctx, 99); err == nil {
		t.Fatal("expected GetByLevel missing error")
	}

	for _, tc := range []struct {
		exp   int64
		level int
	}{
		{exp: 0, level: 1},
		{exp: 150, level: 2},
		{exp: 999, level: 3},
	} {
		cfg, getErr := repo.GetLevelByExp(ctx, tc.exp)
		if getErr != nil || cfg.Level != tc.level {
			t.Fatalf("GetLevelByExp(%d) unexpected cfg=%+v err=%v", tc.exp, cfg, getErr)
		}
	}
	if _, err := repo.GetLevelByExp(ctx, -1); err == nil {
		t.Fatal("expected GetLevelByExp no-match error")
	}

	next, err := repo.GetNextLevel(ctx, 1)
	if err != nil || next.Level != 2 {
		t.Fatalf("GetNextLevel unexpected: cfg=%+v err=%v", next, err)
	}
	if _, err := repo.GetNextLevel(ctx, 3); err == nil {
		t.Fatal("expected GetNextLevel no-next error")
	}

	if !repo.ExistsByLevel(ctx, 2) {
		t.Fatal("expected ExistsByLevel true")
	}
	if repo.ExistsByLevel(ctx, 99) {
		t.Fatal("expected ExistsByLevel false")
	}
	if !repo.ExistsByLevelExcept(ctx, 2, 1) {
		t.Fatal("expected ExistsByLevelExcept true")
	}
	if repo.ExistsByLevelExcept(ctx, 2, byLevel.ID) {
		t.Fatal("expected ExistsByLevelExcept false when excluding same row")
	}

	count, err := repo.Count(ctx)
	if err != nil || count != 3 {
		t.Fatalf("Count unexpected count=%d err=%v", count, err)
	}

	byLevel.Description = "updated"
	if err := repo.Update(ctx, byLevel); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := repo.Delete(ctx, byLevel.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, byLevel.ID); err == nil {
		t.Fatal("expected deleted row not found")
	}
}

func TestLevelConfigRepository_Count_ErrorBranch(t *testing.T) {
	repo, db := setupLevelConfigRepositoryTest(t)
	ctx := context.Background()

	if err := db.Exec(`DROP TABLE level_configs`).Error; err != nil {
		t.Fatalf("drop level_configs: %v", err)
	}

	if _, err := repo.Count(ctx); err == nil {
		t.Fatal("expected Count error when level_configs table missing")
	}
}

func TestLevelConfigRepository_GetAll_ErrorBranch(t *testing.T) {
	repo, db := setupLevelConfigRepositoryTest(t)
	ctx := context.Background()

	if err := db.Exec(`DROP TABLE level_configs`).Error; err != nil {
		t.Fatalf("drop level_configs: %v", err)
	}

	if _, err := repo.GetAll(ctx); err == nil {
		t.Fatal("expected GetAll error when level_configs table missing")
	}
}
