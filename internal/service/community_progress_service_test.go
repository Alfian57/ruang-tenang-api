package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCommunityProgressService(t *testing.T, withSchema bool) *CommunityProgressService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := NewCommunityProgressService(
		repository.NewCommunityProgressRepository(db),
		repository.NewLevelConfigRepository(db),
		repository.NewFeatureUnlockRepository(db),
		repository.NewBadgeRepository(db),
		repository.NewUserRepository(db),
	)

	if !withSchema {
		return svc
	}

	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			avatar TEXT,
			exp INTEGER,
			current_streak INTEGER,
			longest_streak INTEGER,
			total_activities INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE level_configs (
			id INTEGER PRIMARY KEY,
			level INTEGER,
			min_exp INTEGER,
			badge_name TEXT,
			badge_icon TEXT,
			tier_name TEXT,
			tier_color TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE community_stats (
			id TEXT PRIMARY KEY,
			month INTEGER,
			year INTEGER,
			total_xp_earned INTEGER,
			active_members INTEGER,
			total_achievements INTEGER,
			new_members INTEGER,
			total_stories_published INTEGER,
			total_articles_published INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE monthly_hall_of_fame (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			level INTEGER,
			month INTEGER,
			year INTEGER,
			rank INTEGER,
			monthly_xp INTEGER,
			message TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE exp_histories (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			exp_earned INTEGER,
			created_at DATETIME
		)`,
		`CREATE TABLE user_badges (
			id TEXT PRIMARY KEY,
			user_id INTEGER,
			badge_id TEXT,
			earned_at DATETIME,
			is_showcased BOOLEAN
		)`,
		`CREATE TABLE badge_definitions (
			id TEXT PRIMARY KEY,
			name TEXT,
			description TEXT,
			icon TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE feature_definitions (
			id TEXT PRIMARY KEY,
			feature_key TEXT,
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
		`INSERT INTO users (id, name, avatar, exp, current_streak, longest_streak, total_activities, created_at, updated_at)
			VALUES (1, 'User One', '/u1.png', 150, 3, 5, 12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			       (2, 'User Two', '/u2.png', 120, 1, 2, 8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color)
			VALUES (1, 1, 0, 'Pemula', '🌱', 'Bronze', '#cd7f32'),
			       (2, 2, 100, 'Tumbuh', '🌿', 'Silver', '#c0c0c0'),
			       (3, 3, 200, 'Maju', '🌳', 'Gold', '#ffd700')`,
		fmt.Sprintf(`INSERT INTO community_stats (id, month, year, total_xp_earned, active_members, total_achievements, new_members, total_stories_published, total_articles_published, created_at, updated_at)
			VALUES ('00000000-0000-0000-0000-000000000001', %d, %d, 999, 12, 7, 4, 3, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, month, year),
		fmt.Sprintf(`INSERT INTO monthly_hall_of_fame (id, user_id, level, month, year, rank, monthly_xp, message, created_at)
			VALUES ('00000000-0000-0000-0000-000000000010', 1, 2, %d, %d, 1, 300, 'Great support', CURRENT_TIMESTAMP)`, month, year),
		`INSERT INTO exp_histories (id, user_id, activity_type, exp_earned, created_at)
			VALUES (1, 1, 'chat_ai', 10, CURRENT_TIMESTAMP),
			       (2, 1, 'forum_comment', 5, CURRENT_TIMESTAMP)`,
		`INSERT INTO user_badges (id, user_id, badge_id, earned_at, is_showcased)
			VALUES ('00000000-0000-0000-0000-000000000100', 1, '00000000-0000-0000-0000-000000000200', CURRENT_TIMESTAMP, 0)`,
		`INSERT INTO badge_definitions (id, name, description, icon, created_at, updated_at)
			VALUES ('00000000-0000-0000-0000-000000000200', 'Badge One', 'desc', 'icon', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO feature_definitions (id, feature_key, feature_name, description, icon, required_level, category, is_active, display_order, created_at, updated_at)
			VALUES ('00000000-0000-0000-0000-000000000300', 'f_level_2', 'Feature L2', 'desc', 'icon', 2, 'general', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("setup query failed: %v", err)
		}
	}

	return svc
}

func TestCommunityProgressService_MainMethods(t *testing.T) {
	ctx := context.Background()
	svc := setupCommunityProgressService(t, true)

	stats, err := svc.GetCommunityStats(ctx)
	if err != nil || stats == nil || stats.TotalXPEarned != 999 {
		t.Fatalf("unexpected community stats: %+v err=%v", stats, err)
	}

	hof, err := svc.GetLevelHallOfFame(ctx, 2, 0)
	if err != nil || hof == nil || hof.TotalMembers == 0 {
		t.Fatalf("unexpected level hall of fame: %+v err=%v", hof, err)
	}

	monthly, err := svc.GetMonthlyHallOfFame(ctx, int(time.Now().Month()), time.Now().Year(), "")
	if err != nil || len(monthly) == 0 {
		t.Fatalf("unexpected monthly hall of fame: len=%d err=%v", len(monthly), err)
	}

	cats := svc.GetAvailableHallOfFameCategories(ctx)
	if len(cats) == 0 {
		t.Fatal("expected hall of fame categories")
	}

	journey, err := svc.GetPersonalJourney(ctx, 1)
	if err != nil || journey == nil || journey.CurrentLevel == 0 {
		t.Fatalf("unexpected personal journey: %+v err=%v", journey, err)
	}

	weekly, err := svc.GetWeeklyProgress(ctx, 1)
	if err != nil || weekly == nil {
		t.Fatalf("unexpected weekly progress: %+v err=%v", weekly, err)
	}

	monthlyProgress, err := svc.GetMonthlyProgress(ctx, 1)
	if err != nil || monthlyProgress == nil {
		t.Fatalf("unexpected monthly progress: %+v err=%v", monthlyProgress, err)
	}

	allTime, err := svc.GetAllTimeStats(ctx, 1)
	if err != nil || allTime == nil || allTime.CurrentLevel == 0 {
		t.Fatalf("unexpected all time stats: %+v err=%v", allTime, err)
	}

	celebration, err := svc.GetLevelUpCelebration(ctx, 1, 2)
	if err != nil || celebration == nil || celebration.NewLevel != 2 {
		t.Fatalf("unexpected level up celebration: %+v err=%v", celebration, err)
	}
}

func TestCommunityProgressService_ErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupCommunityProgressService(t, false)

	if _, err := svc.GetCommunityStats(ctx); err == nil {
		t.Fatal("expected get community stats error on missing schema")
	}
	if _, err := svc.GetPersonalJourney(ctx, 1); err == nil {
		t.Fatal("expected get personal journey error on missing schema")
	}
	if _, err := svc.GetWeeklyProgress(ctx, 1); err == nil {
		t.Fatal("expected get weekly progress error on missing schema")
	}
	if _, err := svc.GetMonthlyProgress(ctx, 1); err == nil {
		t.Fatal("expected get monthly progress error on missing schema")
	}
	if _, err := svc.GetLevelHallOfFame(ctx, 2, 10); err == nil {
		t.Fatal("expected get level hall of fame error on missing schema")
	}
	if _, err := svc.GetLevelUpCelebration(ctx, 1, 2); err == nil {
		t.Fatal("expected get level up celebration error on missing schema")
	}
}

func TestCommunityProgressService_GetLevelHallOfFame_UserQueryError(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE level_configs (
		id INTEGER PRIMARY KEY,
		level INTEGER,
		min_exp INTEGER,
		badge_name TEXT,
		badge_icon TEXT,
		tier_name TEXT,
		tier_color TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create level_configs table: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color) VALUES (1, 2, 100, 'Tumbuh', '🌿', 'Silver', '#c0c0c0')`).Error; err != nil {
		t.Fatalf("insert level config: %v", err)
	}

	svc := NewCommunityProgressService(
		repository.NewCommunityProgressRepository(db),
		repository.NewLevelConfigRepository(db),
		repository.NewFeatureUnlockRepository(db),
		repository.NewBadgeRepository(db),
		repository.NewUserRepository(db),
	)

	if _, err := svc.GetLevelHallOfFame(ctx, 2, 10); err == nil {
		t.Fatal("expected get level hall of fame user query error")
	}
}

func TestCommunityProgressService_GetCommunityStats_RecalculateWhenStale(t *testing.T) {
	ctx := context.Background()
	svc := setupCommunityProgressService(t, true)

	staleMonth := int(time.Now().Month()) - 1
	staleYear := time.Now().Year()
	if staleMonth <= 0 {
		staleMonth = 12
		staleYear--
	}

	stored, err := svc.communityRepo.GetCommunityStats(ctx)
	if err != nil {
		t.Fatalf("load stored stats: %v", err)
	}
	stored.Month = staleMonth
	stored.Year = staleYear
	if err := svc.communityRepo.UpdateCommunityStats(ctx, stored); err != nil {
		t.Fatalf("set stale month/year: %v", err)
	}

	stats, err := svc.GetCommunityStats(ctx)
	if err != nil {
		t.Fatalf("GetCommunityStats with stale row failed: %v", err)
	}
	if stats.Month != int(time.Now().Month()) || stats.Year != time.Now().Year() {
		t.Fatalf("expected recalculated current period, got month=%d year=%d", stats.Month, stats.Year)
	}
}
