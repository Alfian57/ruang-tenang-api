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

func setupBreathingRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	statements := []string{
		`CREATE TABLE breathing_techniques (
			id TEXT PRIMARY KEY,
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
			is_system BOOLEAN,
			is_active BOOLEAN,
			user_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE breathing_sessions (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			technique_id TEXT NOT NULL,
			duration_seconds INTEGER,
			target_duration_seconds INTEGER,
			cycles_completed INTEGER,
			voice_guidance_enabled BOOLEAN,
			background_sound TEXT,
			haptic_feedback_enabled BOOLEAN,
			completed BOOLEAN,
			completed_percentage INTEGER,
			started_at DATETIME,
			ended_at DATETIME,
			xp_earned INTEGER,
			mood_before TEXT,
			mood_after TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE breathing_preferences (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE NOT NULL,
			default_duration_seconds INTEGER,
			default_technique_id TEXT,
			voice_guidance TEXT,
			background_sound TEXT,
			default_background_sound TEXT,
			haptic_feedback BOOLEAN,
			animation_speed TEXT,
			theme TEXT,
			reminder_enabled BOOLEAN,
			reminder_time TEXT,
			reminder_days TEXT,
			tutorial_completed BOOLEAN,
			current_streak INTEGER,
			longest_streak INTEGER,
			last_practice_date DATETIME,
			streak_freeze_available BOOLEAN,
			streak_freeze_used_at DATETIME,
			daily_xp_earned INTEGER,
			daily_xp_date DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE breathing_favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			technique_id TEXT NOT NULL,
			sort_order INTEGER,
			created_at DATETIME
		)`,
	}

	for _, stmt := range statements {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}

	return db
}

func seedBreathingData(t *testing.T, db *gorm.DB) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	now := time.Now()
	techSystemID := uuid.New()
	techUserID := uuid.New()
	techOtherID := uuid.New()

	if err := db.Exec(`INSERT INTO breathing_techniques
		(id, name, slug, inhale_duration, inhale_hold_duration, exhale_duration, exhale_hold_duration, icon, color, animation_type, difficulty, category, is_system, is_active, user_id, created_at, updated_at)
		VALUES (?, ?, ?, 4, 0, 4, 0, '🌬️', '#6366F1', 'circle', 'easy', 'general', 1, 1, NULL, ?, ?)`,
		techSystemID.String(), "Box", "box", now, now).Error; err != nil {
		t.Fatalf("seed system technique: %v", err)
	}

	if err := db.Exec(`INSERT INTO breathing_techniques
		(id, name, slug, inhale_duration, inhale_hold_duration, exhale_duration, exhale_hold_duration, icon, color, animation_type, difficulty, category, is_system, is_active, user_id, created_at, updated_at)
		VALUES (?, ?, ?, 4, 2, 6, 2, '🧘', '#123456', 'circle', 'custom', 'custom', 0, 1, 7, ?, ?)`,
		techUserID.String(), "Custom", "custom", now, now).Error; err != nil {
		t.Fatalf("seed user technique: %v", err)
	}

	if err := db.Exec(`INSERT INTO breathing_techniques
		(id, name, slug, inhale_duration, inhale_hold_duration, exhale_duration, exhale_hold_duration, icon, color, animation_type, difficulty, category, is_system, is_active, user_id, created_at, updated_at)
		VALUES (?, ?, ?, 4, 2, 6, 2, '😴', '#999999', 'circle', 'easy', 'general', 0, 0, 8, ?, ?)`,
		techOtherID.String(), "Inactive", "inactive", now, now).Error; err != nil {
		t.Fatalf("seed inactive technique: %v", err)
	}

	s1 := uuid.New()
	s2 := uuid.New()
	yesterday := now.AddDate(0, 0, -1)

	if err := db.Exec(`INSERT INTO breathing_sessions
		(id, user_id, technique_id, duration_seconds, target_duration_seconds, cycles_completed, voice_guidance_enabled, haptic_feedback_enabled, completed, completed_percentage, started_at, xp_earned, created_at)
		VALUES (?, 7, ?, 300, 300, 8, 1, 1, 1, 100, ?, 10, ?)`, s1.String(), techUserID.String(), now, now).Error; err != nil {
		t.Fatalf("seed session today: %v", err)
	}

	if err := db.Exec(`INSERT INTO breathing_sessions
		(id, user_id, technique_id, duration_seconds, target_duration_seconds, cycles_completed, voice_guidance_enabled, haptic_feedback_enabled, completed, completed_percentage, started_at, xp_earned, created_at)
		VALUES (?, 7, ?, 120, 300, 3, 1, 0, 0, 40, ?, 0, ?)`, s2.String(), techSystemID.String(), yesterday, yesterday).Error; err != nil {
		t.Fatalf("seed session yesterday: %v", err)
	}

	return techSystemID, techUserID, s1
}

func TestBreathingRepository_TechniqueAndSessionFlow(t *testing.T) {
	db := setupBreathingRepoDB(t)
	ctx := context.Background()
	repo := NewBreathingRepository(db)
	techSystemID, techUserID, sessionID := seedBreathingData(t, db)

	all, err := repo.GetAllTechniques(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("get all techniques failed: err=%v len=%d", err, len(all))
	}

	systems, err := repo.GetSystemTechniques(ctx)
	if err != nil || len(systems) != 1 {
		t.Fatalf("get system techniques failed: err=%v len=%d", err, len(systems))
	}

	userTechniques, err := repo.GetUserTechniques(ctx, 7)
	if err != nil || len(userTechniques) != 1 {
		t.Fatalf("get user techniques failed: err=%v len=%d", err, len(userTechniques))
	}

	byID, err := repo.GetTechniqueByID(ctx, techUserID)
	if err != nil || byID.ID != techUserID {
		t.Fatalf("get technique by id failed: err=%v", err)
	}

	bySlug, err := repo.GetTechniqueBySlug(ctx, "box")
	if err != nil || bySlug.ID != techSystemID {
		t.Fatalf("get technique by slug failed: err=%v", err)
	}

	newTech := &model.BreathingTechnique{ID: uuid.New(), Name: "Focus", IsActive: true}
	if err := repo.CreateTechnique(ctx, newTech); err != nil {
		t.Fatalf("create technique failed: %v", err)
	}
	newTech.Name = "Focus Updated"
	if err := repo.UpdateTechnique(ctx, newTech); err != nil {
		t.Fatalf("update technique failed: %v", err)
	}
	if err := repo.DeleteTechnique(ctx, newTech.ID); err != nil {
		t.Fatalf("delete technique failed: %v", err)
	}

	foundSession, err := repo.GetSessionByID(ctx, sessionID)
	if err != nil || foundSession.ID != sessionID {
		t.Fatalf("get session by id failed: err=%v", err)
	}

	newSession := &model.BreathingSession{ID: uuid.New(), UserID: 7, TechniqueID: techUserID, DurationSeconds: 60, TargetDurationSeconds: 120, StartedAt: time.Now()}
	if err := repo.CreateSession(ctx, newSession); err != nil {
		t.Fatalf("create session failed: %v", err)
	}
	newSession.Completed = true
	newSession.XPEarned = 5
	if err := repo.UpdateSession(ctx, newSession); err != nil {
		t.Fatalf("update session failed: %v", err)
	}

	start := time.Now().AddDate(0, 0, -2)
	end := time.Now().AddDate(0, 0, 1)
	sessions, total, err := repo.GetUserSessions(ctx, 7, &start, &end, &techUserID, 10, 0)
	if err != nil || total < 1 || len(sessions) < 1 {
		t.Fatalf("get user sessions failed: err=%v total=%d len=%d", err, total, len(sessions))
	}

	todaySessions, err := repo.GetUserSessionsToday(ctx, 7)
	if err != nil || len(todaySessions) < 1 {
		t.Fatalf("get user sessions today failed: err=%v len=%d", err, len(todaySessions))
	}

	rangeSessions, err := repo.GetUserSessionsByDateRange(ctx, 7, start, end)
	if err != nil || len(rangeSessions) < 1 {
		t.Fatalf("get user sessions by range failed: err=%v len=%d", err, len(rangeSessions))
	}
}

func TestBreathingRepository_StatsPreferencesFavoritesAndHelpers(t *testing.T) {
	db := setupBreathingRepoDB(t)
	ctx := context.Background()
	repo := NewBreathingRepository(db)
	techSystemID, techUserID, _ := seedBreathingData(t, db)

	count, err := repo.GetUserTotalSessionsCount(ctx, 7)
	if err != nil || count < 1 {
		t.Fatalf("get total sessions count failed: err=%v count=%d", err, count)
	}

	minutes, err := repo.GetUserTotalMinutes(ctx, 7)
	if err != nil || minutes < 1 {
		t.Fatalf("get total minutes failed: err=%v minutes=%d", err, minutes)
	}

	todayXP, err := repo.GetUserTodayXP(ctx, 7)
	if err != nil || todayXP < 1 {
		t.Fatalf("get today xp failed: err=%v xp=%d", err, todayXP)
	}

	mostUsed, usedCount, err := repo.GetMostUsedTechnique(ctx, 7)
	if err != nil || mostUsed == nil || usedCount < 1 {
		t.Fatalf("get most used technique failed: err=%v count=%d", err, usedCount)
	}

	rate, err := repo.GetCompletionRate(ctx, 7)
	if err != nil || rate <= 0 {
		t.Fatalf("get completion rate failed: err=%v rate=%f", err, rate)
	}

	sinceCount, err := repo.CountSessionsSince(ctx, 7, time.Now().AddDate(0, 0, -2))
	if err != nil || sinceCount < 1 {
		t.Fatalf("count sessions since failed: err=%v count=%d", err, sinceCount)
	}

	if _, err := repo.GetPreferences(ctx, 7); err == nil {
		t.Fatal("expected not found preferences")
	}

	pref := &model.BreathingPreference{UserID: 7, DefaultDurationSeconds: 300, VoiceGuidance: "ask", BackgroundSound: "ask", DefaultBackgroundSound: "none", HapticFeedback: true, AnimationSpeed: "normal", Theme: "default"}
	if err := repo.CreatePreferences(ctx, pref); err != nil {
		t.Fatalf("create preferences failed: %v", err)
	}
	pref.Theme = "dark"
	if err := repo.UpdatePreferences(ctx, pref); err != nil {
		t.Fatalf("update preferences failed: %v", err)
	}

	gotPref, err := repo.GetOrCreatePreferences(ctx, 7)
	if err != nil || gotPref.Theme != "dark" {
		t.Fatalf("get or create existing preferences failed: err=%v theme=%s", err, gotPref.Theme)
	}

	newPref, err := repo.GetOrCreatePreferences(ctx, 8)
	if err != nil || newPref.UserID != 8 {
		t.Fatalf("get or create new preferences failed: err=%v user=%d", err, newPref.UserID)
	}

	if err := repo.AddFavorite(ctx, &model.BreathingFavorite{UserID: 7, TechniqueID: techUserID}); err != nil {
		t.Fatalf("add favorite 1 failed: %v", err)
	}
	if err := repo.AddFavorite(ctx, &model.BreathingFavorite{UserID: 7, TechniqueID: techSystemID}); err != nil {
		t.Fatalf("add favorite 2 failed: %v", err)
	}

	favorites, err := repo.GetFavorites(ctx, 7)
	if err != nil || len(favorites) != 2 {
		t.Fatalf("get favorites failed: err=%v len=%d", err, len(favorites))
	}

	isFav, err := repo.IsFavorite(ctx, 7, techUserID)
	if err != nil || !isFav {
		t.Fatalf("is favorite failed: err=%v isFav=%v", err, isFav)
	}

	if err := repo.UpdateFavoriteOrder(ctx, 7, []uuid.UUID{techSystemID, techUserID}); err != nil {
		t.Fatalf("update favorite order failed: %v", err)
	}

	if err := repo.RemoveFavorite(ctx, 7, techSystemID); err != nil {
		t.Fatalf("remove favorite failed: %v", err)
	}

	dailyCount, dailyMinutes, err := repo.GetDailyStats(ctx, 7, time.Now())
	if err != nil || dailyCount < 1 || dailyMinutes < 1 {
		t.Fatalf("get daily stats failed: err=%v count=%d minutes=%d", err, dailyCount, dailyMinutes)
	}

	usage, err := repo.GetTechniqueUsageStats(ctx, 7)
	if err != nil || len(usage) < 1 {
		t.Fatalf("get technique usage stats failed: err=%v len=%d", err, len(usage))
	}

	calendar, err := repo.GetMonthlyCalendar(ctx, 7, time.Now().Year(), int(time.Now().Month()))
	if err != nil {
		t.Fatalf("expected monthly calendar success on sqlite, got %v", err)
	}
	if len(calendar) < 1 || calendar[0].SessionsCount < 1 {
		t.Fatalf("unexpected monthly calendar result: %#v", calendar)
	}

	if out := splitAndTrim(" a, b,  c "); len(out) != 3 || out[0] != "a" || out[2] != "c" {
		t.Fatalf("splitAndTrim failed: %#v", out)
	}
	if out := splitString("a,b,c", ','); len(out) != 3 {
		t.Fatalf("splitString failed: %#v", out)
	}
	if out := trimSpace("\t hello \t"); out != "hello" {
		t.Fatalf("trimSpace failed: %q", out)
	}
}

func TestBreathingRepository_EdgeBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("most used returns nil when no completed session", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		tech, count, err := repo.GetMostUsedTechnique(ctx, 999)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if tech != nil || count != 0 {
			t.Fatalf("expected nil technique and 0 count, got tech=%v count=%d", tech, count)
		}
	})

	t.Run("most used technique lookup error", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		missingTechID := uuid.New()
		sessionID := uuid.New()
		now := time.Now()
		if err := db.Exec(`INSERT INTO breathing_sessions
			(id, user_id, technique_id, duration_seconds, target_duration_seconds, cycles_completed, completed, completed_percentage, started_at, xp_earned, created_at)
			VALUES (?, 7, ?, 120, 120, 3, 1, 100, ?, 5, ?)`, sessionID.String(), missingTechID.String(), now, now).Error; err != nil {
			t.Fatalf("seed orphan session failed: %v", err)
		}

		tech, count, err := repo.GetMostUsedTechnique(ctx, 7)
		if err == nil {
			t.Fatal("expected lookup error for missing technique")
		}
		if tech != nil || count != 0 {
			t.Fatalf("expected nil technique and zero count on error, got tech=%v count=%d", tech, count)
		}
	})

	t.Run("completion rate zero total", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		rate, err := repo.GetCompletionRate(ctx, 12345)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if rate != 0 {
			t.Fatalf("expected zero completion rate for no sessions, got %f", rate)
		}
	})

	t.Run("completion rate query error", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		if err := db.Exec(`DROP TABLE breathing_sessions`).Error; err != nil {
			t.Fatalf("drop breathing_sessions failed: %v", err)
		}

		if _, err := repo.GetCompletionRate(ctx, 7); err == nil {
			t.Fatal("expected query error when breathing_sessions table is missing")
		}
	})

	t.Run("get or create preferences returns db error", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		if err := db.Exec(`DROP TABLE breathing_preferences`).Error; err != nil {
			t.Fatalf("drop breathing_preferences failed: %v", err)
		}

		if _, err := repo.GetOrCreatePreferences(ctx, 7); err == nil {
			t.Fatal("expected error when preferences table is missing")
		}
	})

	t.Run("get by id slug and session not found", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		if _, err := repo.GetTechniqueByID(ctx, uuid.New()); err == nil {
			t.Fatal("expected not found error for unknown technique id")
		}
		if _, err := repo.GetTechniqueBySlug(ctx, "does-not-exist"); err == nil {
			t.Fatal("expected not found error for unknown slug")
		}
		if _, err := repo.GetSessionByID(ctx, uuid.New()); err == nil {
			t.Fatal("expected not found error for unknown session id")
		}
	})

	t.Run("update favorite order error path", func(t *testing.T) {
		db := setupBreathingRepoDB(t)
		repo := NewBreathingRepository(db)

		if err := db.Exec(`DROP TABLE breathing_favorites`).Error; err != nil {
			t.Fatalf("drop breathing_favorites failed: %v", err)
		}

		if err := repo.UpdateFavoriteOrder(ctx, 7, []uuid.UUID{uuid.New()}); err == nil {
			t.Fatal("expected update favorite order to fail when table is missing")
		}
	})
}
