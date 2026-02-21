package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBasicServiceDB(t *testing.T, withSchema bool) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if !withSchema {
		return db
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			username TEXT,
			email TEXT,
			password TEXT,
			role TEXT,
			exp INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE user_moods (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			mood TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE level_configs (
			id INTEGER PRIMARY KEY,
			level INTEGER,
			min_exp INTEGER,
			badge_name TEXT,
			badge_icon TEXT,
			tier_name TEXT,
			tier_color TEXT,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}

	seed := []string{
		`INSERT INTO users (id, name, username, email, password, role, exp) VALUES
			(1, 'A', 'a', 'a@test.com', 'p', 'member', 100),
			(2, 'B', 'b', 'b@test.com', 'p', 'member', 300),
			(3, 'C', 'c', 'c@test.com', 'p', 'member', 200)`,
		`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon) VALUES
			(1, 1, 0, 'Pemula', '🌱'),
			(2, 2, 100, 'Tumbuh', '🌿')`,
	}

	for _, q := range seed {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("seed failed: %v", err)
		}
	}

	return db
}

func TestUserService_GetLeaderboard(t *testing.T) {
	ctx := context.Background()

	db := setupBasicServiceDB(t, true)
	svc := NewUserService(repository.NewUserRepository(db))

	users, err := svc.GetLeaderboard(ctx, 2)
	if err != nil {
		t.Fatalf("get leaderboard failed: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Exp < users[1].Exp {
		t.Fatalf("expected descending exp order, got %d then %d", users[0].Exp, users[1].Exp)
	}

	dbErr := setupBasicServiceDB(t, false)
	svcErr := NewUserService(repository.NewUserRepository(dbErr))
	if _, err := svcErr.GetLeaderboard(ctx, 2); err == nil {
		t.Fatal("expected leaderboard error on missing schema")
	}
}

func TestMoodService_RecordAndFetchBranches(t *testing.T) {
	ctx := context.Background()
	db := setupBasicServiceDB(t, true)
	svc := NewMoodService(repository.NewUserMoodRepository(db))

	created, err := svc.RecordMood(ctx, 1, &dto.CreateMoodRequest{Mood: "happy"})
	if err != nil {
		t.Fatalf("record mood create failed: %v", err)
	}
	if created.Mood != "happy" {
		t.Fatalf("expected happy mood, got %s", created.Mood)
	}

	updated, err := svc.RecordMood(ctx, 1, &dto.CreateMoodRequest{Mood: "sad"})
	if err != nil {
		t.Fatalf("record mood update failed: %v", err)
	}
	if updated.Mood != "sad" {
		t.Fatalf("expected updated sad mood, got %s", updated.Mood)
	}

	history, err := svc.GetMoodHistory(ctx, 1, &dto.MoodQueryParams{StartDate: "bad", EndDate: "bad", Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("get mood history failed: %v", err)
	}
	if history.TotalCount == 0 || len(history.Moods) == 0 {
		t.Fatal("expected non-empty mood history")
	}

	latest, err := svc.GetLatestMood(ctx, 1)
	if err != nil {
		t.Fatalf("get latest mood failed: %v", err)
	}
	if latest.Mood != "sad" {
		t.Fatalf("expected latest sad mood, got %s", latest.Mood)
	}

	stats, err := svc.GetMoodStats(ctx, 1, 30)
	if err != nil {
		t.Fatalf("get mood stats failed: %v", err)
	}
	if stats["sad"] == 0 {
		t.Fatal("expected sad stat count > 0")
	}

	today, err := svc.GetTodayMood(ctx, 1)
	if err != nil {
		t.Fatalf("get today mood failed: %v", err)
	}
	if !today.HasChecked || today.Mood == nil {
		t.Fatal("expected checked today mood")
	}

	emptyToday, err := svc.GetTodayMood(ctx, 999)
	if err != nil {
		t.Fatalf("get today mood empty failed: %v", err)
	}
	if emptyToday.HasChecked || emptyToday.Mood != nil {
		t.Fatal("expected no today mood")
	}
}

func TestMoodService_DBErrorFallbacks(t *testing.T) {
	ctx := context.Background()
	db := setupBasicServiceDB(t, false)
	svc := NewMoodService(repository.NewUserMoodRepository(db))

	if _, err := svc.RecordMood(ctx, 1, &dto.CreateMoodRequest{Mood: "happy"}); err == nil {
		t.Fatal("expected record mood error on missing schema")
	}
	if _, err := svc.GetLatestMood(ctx, 1); err == nil {
		t.Fatal("expected latest mood error on missing schema")
	}
	if _, err := svc.GetMoodHistory(ctx, 1, &dto.MoodQueryParams{Page: 1, Limit: 10}); err == nil {
		t.Fatal("expected mood history error on missing schema")
	}
}

func TestLevelConfigService_Branches(t *testing.T) {
	ctx := context.Background()
	db := setupBasicServiceDB(t, true)
	repo := repository.NewLevelConfigRepository(db)
	cache := NewCacheService()
	svc := NewLevelConfigService(repo, cache)

	all, err := svc.GetAll(ctx)
	if err != nil || len(all) == 0 {
		t.Fatalf("get all failed: %v", err)
	}

	if err := db.Exec(`DROP TABLE level_configs`).Error; err != nil {
		t.Fatalf("drop table failed: %v", err)
	}

	cached, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("get all from cache failed: %v", err)
	}
	if len(cached) != len(all) {
		t.Fatalf("expected cached len %d, got %d", len(all), len(cached))
	}

	svcErr := NewLevelConfigService(repository.NewLevelConfigRepository(setupBasicServiceDB(t, false)), NewCacheService())
	if _, err := svcErr.GetAll(ctx); err == nil {
		t.Fatal("expected get all error on missing schema")
	}
}

func TestLevelConfigService_CreateUpdateDeleteAndFallback(t *testing.T) {
	ctx := context.Background()
	db := setupBasicServiceDB(t, true)
	repo := repository.NewLevelConfigRepository(db)
	cache := NewCacheService()
	svc := NewLevelConfigService(repo, cache)

	cache.Set(CacheKeyLevelConfigs, []model.LevelConfig{{ID: 99, Level: 99}})
	if err := svc.Create(ctx, &model.LevelConfig{Level: 1, MinExp: 0, BadgeName: "dup", BadgeIcon: "x"}); !errors.Is(err, ErrLevelExists) {
		t.Fatalf("expected ErrLevelExists, got %v", err)
	}

	if err := svc.Create(ctx, &model.LevelConfig{Level: 3, MinExp: 200, BadgeName: "L3", BadgeIcon: "🔥"}); err != nil {
		t.Fatalf("create level failed: %v", err)
	}
	if cache.Get(CacheKeyLevelConfigs) != nil {
		t.Fatal("expected level cache invalidated after create")
	}

	if err := svc.Update(ctx, 999, &model.LevelConfig{Level: 9}); err == nil {
		t.Fatal("expected update error for missing level")
	}

	if err := svc.Update(ctx, 2, &model.LevelConfig{Level: 3, MinExp: 150, BadgeName: "L2", BadgeIcon: "⭐"}); !errors.Is(err, ErrLevelExists) {
		t.Fatalf("expected ErrLevelExists on duplicate update, got %v", err)
	}

	if err := svc.Update(ctx, 2, &model.LevelConfig{Level: 2, MinExp: 120, BadgeName: "L2+", BadgeIcon: "⭐"}); err != nil {
		t.Fatalf("update level failed: %v", err)
	}

	if err := svc.Delete(ctx, 2); err != nil {
		t.Fatalf("delete level failed: %v", err)
	}

	current, next, err := svc.GetUserLevelInfo(ctx, 150)
	if err != nil {
		t.Fatalf("get user level info failed: %v", err)
	}
	if current == nil || current.Level == 0 {
		t.Fatal("expected current level info")
	}
	if next != nil && next.Level <= current.Level {
		t.Fatal("expected next level to be greater than current")
	}

	svcFallback := NewLevelConfigService(repository.NewLevelConfigRepository(setupBasicServiceDB(t, false)), NewCacheService())
	def, nxt, err := svcFallback.GetUserLevelInfo(ctx, 123)
	if err != nil {
		t.Fatalf("expected fallback no error, got %v", err)
	}
	if def == nil || def.Level != 1 || nxt != nil {
		t.Fatalf("expected default level fallback, got %+v next=%+v", def, nxt)
	}
}
