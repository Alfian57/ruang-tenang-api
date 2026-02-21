package service

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/pkg/gamification"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGamificationDB(t *testing.T, withSchema bool) *gorm.DB {
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
			exp INTEGER,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE user_activities (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			date DATETIME,
			count INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE exp_histories (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			points INTEGER,
			description TEXT,
			created_at DATETIME
		)`,
		`INSERT INTO users (id, exp) VALUES (1, 0)`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("setup query failed: %v", err)
		}
	}

	return db
}

func TestGamificationService_AwardExp_NoLimitActivity(t *testing.T) {
	ctx := context.Background()
	db := setupGamificationDB(t, true)
	svc := NewGamificationService(db)

	if err := svc.AwardExp(ctx, 1, gamification.ActivityUploadArticle, 20); err != nil {
		t.Fatalf("award exp no-limit failed: %v", err)
	}

	var exp int64
	if err := db.Raw(`SELECT exp FROM users WHERE id = 1`).Scan(&exp).Error; err != nil {
		t.Fatalf("read user exp: %v", err)
	}
	if exp != 20 {
		t.Fatalf("expected exp 20, got %d", exp)
	}

	var historyCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM exp_histories WHERE user_id = 1`).Scan(&historyCount).Error; err != nil {
		t.Fatalf("read history count: %v", err)
	}
	if historyCount != 1 {
		t.Fatalf("expected 1 exp history row, got %d", historyCount)
	}
}

func TestGamificationService_AwardExp_LimitReached(t *testing.T) {
	ctx := context.Background()
	db := setupGamificationDB(t, true)
	svc := NewGamificationService(db)

	today := time.Now().Truncate(24 * time.Hour)
	if err := db.Exec(
		`INSERT INTO user_activities (user_id, activity_type, date, count) VALUES (?, ?, ?, ?)`,
		1, string(gamification.ActivityChatAI), today, gamification.LimitChatAI,
	).Error; err != nil {
		t.Fatalf("seed user activity: %v", err)
	}

	if err := svc.AwardExp(ctx, 1, gamification.ActivityChatAI, gamification.ExpChatAI); err != nil {
		t.Fatalf("award exp limit reached should not error: %v", err)
	}

	var exp int64
	if err := db.Raw(`SELECT exp FROM users WHERE id = 1`).Scan(&exp).Error; err != nil {
		t.Fatalf("read user exp: %v", err)
	}
	if exp != 0 {
		t.Fatalf("expected exp unchanged 0 when limit reached, got %d", exp)
	}

	var historyCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM exp_histories WHERE user_id = 1`).Scan(&historyCount).Error; err != nil {
		t.Fatalf("read history count: %v", err)
	}
	if historyCount != 0 {
		t.Fatalf("expected no exp history when limit reached, got %d", historyCount)
	}
}

func TestGamificationService_AwardExp_WithLimitAndDBError(t *testing.T) {
	ctx := context.Background()
	db := setupGamificationDB(t, true)
	svc := NewGamificationService(db)

	if err := svc.AwardExp(ctx, 1, gamification.ActivityChatAI, gamification.ExpChatAI); err != nil {
		t.Fatalf("award exp with limit activity failed: %v", err)
	}

	var activityCount int64
	if err := db.Raw(`SELECT count FROM user_activities WHERE user_id = 1 AND activity_type = ?`, string(gamification.ActivityChatAI)).Scan(&activityCount).Error; err != nil {
		t.Fatalf("read activity count: %v", err)
	}
	if activityCount != 1 {
		t.Fatalf("expected user activity count 1, got %d", activityCount)
	}

	dbErr := setupGamificationDB(t, false)
	svcErr := NewGamificationService(dbErr)
	if err := svcErr.AwardExp(ctx, 1, gamification.ActivityUploadArticle, 20); err == nil {
		t.Fatal("expected db error on missing schema")
	}
}

func TestGamificationService_AwardExp_LimitIncrementAndErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("limit activity increments existing daily counter", func(t *testing.T) {
		db := setupGamificationDB(t, true)
		svc := NewGamificationService(db)

		today := time.Now().Truncate(24 * time.Hour)
		if err := db.Exec(
			`INSERT INTO user_activities (user_id, activity_type, date, count) VALUES (?, ?, ?, ?)`,
			1, string(gamification.ActivityForumComment), today, 1,
		).Error; err != nil {
			t.Fatalf("seed existing user activity: %v", err)
		}

		if err := svc.AwardExp(ctx, 1, gamification.ActivityForumComment, gamification.ExpForumComment); err != nil {
			t.Fatalf("award exp with existing activity failed: %v", err)
		}

		var activityRows int64
		if err := db.Raw(`SELECT COUNT(*) FROM user_activities WHERE user_id = 1 AND activity_type = ?`, string(gamification.ActivityForumComment)).Scan(&activityRows).Error; err != nil {
			t.Fatalf("read activity rows: %v", err)
		}
		if activityRows < 1 {
			t.Fatalf("expected activity rows >= 1, got %d", activityRows)
		}

		var exp int64
		if err := db.Raw(`SELECT exp FROM users WHERE id = 1`).Scan(&exp).Error; err != nil {
			t.Fatalf("read user exp: %v", err)
		}
		if exp != int64(gamification.ExpForumComment) {
			t.Fatalf("expected exp increased to %d, got %d", gamification.ExpForumComment, exp)
		}
	})

	t.Run("limit activity count query error", func(t *testing.T) {
		db := setupGamificationDB(t, true)
		svc := NewGamificationService(db)

		if err := db.Exec(`DROP TABLE user_activities`).Error; err != nil {
			t.Fatalf("drop user_activities: %v", err)
		}

		if err := svc.AwardExp(ctx, 1, gamification.ActivityChatAI, gamification.ExpChatAI); err == nil {
			t.Fatal("expected error when user_activities table is missing")
		}
	})

	t.Run("exp history create error rolls back transaction", func(t *testing.T) {
		db := setupGamificationDB(t, true)
		svc := NewGamificationService(db)

		if err := db.Exec(`DROP TABLE exp_histories`).Error; err != nil {
			t.Fatalf("drop exp_histories: %v", err)
		}

		err := svc.AwardExp(ctx, 1, gamification.ActivityUploadArticle, 20)
		if err == nil {
			t.Fatal("expected error when exp_histories table is missing")
		}

		var exp int64
		if err := db.Raw(`SELECT exp FROM users WHERE id = 1`).Scan(&exp).Error; err != nil {
			t.Fatalf("read user exp: %v", err)
		}
		if exp != 0 {
			t.Fatalf("expected transaction rollback to keep exp=0, got %d", exp)
		}
	})

	t.Run("firstorcreate fallback count-greater-than-zero branch returns error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		queries := []string{
			`CREATE TABLE users (id INTEGER PRIMARY KEY, exp INTEGER, updated_at DATETIME, deleted_at DATETIME)`,
			`CREATE TABLE user_activities (id INTEGER PRIMARY KEY, user_id INTEGER, activity_type TEXT, date DATETIME, count INTEGER)`,
			`CREATE TABLE exp_histories (id INTEGER PRIMARY KEY, user_id INTEGER, activity_type TEXT, points INTEGER, description TEXT, created_at DATETIME)`,
			`INSERT INTO users (id, exp) VALUES (1, 0)`,
		}
		for _, q := range queries {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup query failed: %v", err)
			}
		}

		today := time.Now().Truncate(24 * time.Hour)
		if err := db.Exec(`INSERT INTO user_activities (user_id, activity_type, date, count) VALUES (?, ?, ?, ?)`, 1, string(gamification.ActivityForumComment), today, 1).Error; err != nil {
			t.Fatalf("seed user_activities: %v", err)
		}

		svc := NewGamificationService(db)
		if err := svc.AwardExp(ctx, 1, gamification.ActivityForumComment, gamification.ExpForumComment); err == nil {
			t.Fatal("expected fallback update path to return error with malformed user_activities schema")
		}
	})

	t.Run("firstorcreate fallback count-zero branch returns error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		queries := []string{
			`CREATE TABLE users (id INTEGER PRIMARY KEY, exp INTEGER, updated_at DATETIME, deleted_at DATETIME)`,
			`CREATE TABLE user_activities (id INTEGER PRIMARY KEY, user_id INTEGER, activity_type TEXT, date DATETIME, count INTEGER)`,
			`CREATE TABLE exp_histories (id INTEGER PRIMARY KEY, user_id INTEGER, activity_type TEXT, points INTEGER, description TEXT, created_at DATETIME)`,
			`INSERT INTO users (id, exp) VALUES (1, 0)`,
		}
		for _, q := range queries {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup query failed: %v", err)
			}
		}

		svc := NewGamificationService(db)
		if err := svc.AwardExp(ctx, 1, gamification.ActivityForumComment, gamification.ExpForumComment); err == nil {
			t.Fatal("expected fallback create path to return error with malformed user_activities schema")
		}
	})
}
