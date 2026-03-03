package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCommunityHandler(t *testing.T, withSchema bool) *CommunityProgressHandler {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if withSchema {
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
			`CREATE TABLE exp_histories (id INTEGER PRIMARY KEY, user_id INTEGER, activity_type TEXT, exp_earned INTEGER, created_at DATETIME)`,
			`CREATE TABLE user_badges (id TEXT PRIMARY KEY, user_id INTEGER, badge_id TEXT, earned_at DATETIME, is_showcased BOOLEAN)`,
			`CREATE TABLE feature_definitions (id TEXT PRIMARY KEY, feature_key TEXT, feature_name TEXT, description TEXT, icon TEXT, required_level INTEGER, category TEXT, is_active BOOLEAN, display_order INTEGER, created_at DATETIME, updated_at DATETIME)`,
			`INSERT INTO users (id, name, avatar, exp, current_streak, longest_streak, total_activities, created_at, updated_at) VALUES (1, 'U1', '/u1.png', 150, 3, 5, 10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color) VALUES (1,1,0,'Pemula','🌱','Bronze','#cd7f32'), (2,2,100,'Tumbuh','🌿','Silver','#c0c0c0'), (3,3,200,'Maju','🌳','Gold','#ffd700')`,
			`INSERT INTO community_stats (id, month, year, total_xp_earned, active_members, total_achievements, new_members, total_stories_published, total_articles_published, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000001', 1, 2026, 100, 10, 5, 2, 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			`INSERT INTO monthly_hall_of_fame (id, user_id, level, month, year, rank, monthly_xp, message, created_at) VALUES ('00000000-0000-0000-0000-000000000010', 1, 2, 1, 2026, 1, 300, 'Nice', CURRENT_TIMESTAMP)`,
			`INSERT INTO exp_histories (id, user_id, activity_type, exp_earned, created_at) VALUES (1,1,'chat_ai',10,CURRENT_TIMESTAMP)`,
			`INSERT INTO feature_definitions (id, feature_key, feature_name, description, icon, required_level, category, is_active, display_order, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000300', 'f2', 'Feature 2', 'desc', 'icon', 2, 'general', 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		}
		for _, q := range queries {
			if err := db.Exec(q).Error; err != nil {
				t.Fatalf("setup query failed: %v", err)
			}
		}
	}

	svc := service.NewCommunityProgressService(
		repository.NewCommunityProgressRepository(db),
		repository.NewLevelConfigRepository(db),
		repository.NewFeatureUnlockRepository(db),
		repository.NewBadgeRepository(db),
		repository.NewUserRepository(db),
	)

	return NewCommunityProgressHandler(svc)
}

func TestCommunityProgressHandler_SuccessAndErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hOK := setupCommunityHandler(t, true)
	hErr := setupCommunityHandler(t, false)

	r := gin.New()
	r.GET("/stats", hOK.GetCommunityStats)
	r.GET("/stats-err", hErr.GetCommunityStats)
	r.GET("/level/:level", hOK.GetLevelHallOfFame)
	r.GET("/monthly", hOK.GetMonthlyHallOfFame)
	r.GET("/cats", hOK.GetHallOfFameCategories)

	r.GET("/journey", func(c *gin.Context) { c.Set("user_id", uint(1)); hOK.GetPersonalJourney(c) })
	r.GET("/weekly", func(c *gin.Context) { c.Set("user_id", uint(1)); hOK.GetWeeklyProgress(c) })
	r.GET("/monthly-progress", func(c *gin.Context) { c.Set("user_id", uint(1)); hOK.GetMonthlyProgress(c) })
	r.GET("/all-time", func(c *gin.Context) { c.Set("user_id", uint(1)); hOK.GetAllTimeStats(c) })
	r.GET("/all-time-err", func(c *gin.Context) { c.Set("user_id", uint(999)); hOK.GetAllTimeStats(c) })
	r.GET("/celebrate/:level", func(c *gin.Context) { c.Set("user_id", uint(1)); hOK.GetLevelUpCelebration(c) })
	r.GET("/journey-err", func(c *gin.Context) { c.Set("user_id", uint(1)); hErr.GetPersonalJourney(c) })
	r.GET("/weekly-err", func(c *gin.Context) { c.Set("user_id", uint(1)); hErr.GetWeeklyProgress(c) })
	r.GET("/monthly-progress-err", func(c *gin.Context) { c.Set("user_id", uint(1)); hErr.GetMonthlyProgress(c) })
	r.GET("/level-err/:level", hErr.GetLevelHallOfFame)
	r.GET("/monthly-err", hErr.GetMonthlyHallOfFame)

	checks := []struct {
		path string
		code int
	}{
		{"/stats", http.StatusOK},
		{"/stats-err", http.StatusInternalServerError},
		{"/level/2?limit=100", http.StatusOK},
		{"/monthly?month=1&year=2026", http.StatusOK},
		{"/cats", http.StatusOK},
		{"/journey", http.StatusOK},
		{"/weekly", http.StatusOK},
		{"/monthly-progress", http.StatusOK},
		{"/all-time", http.StatusOK},
		{"/all-time-err", http.StatusInternalServerError},
		{"/journey-err", http.StatusInternalServerError},
		{"/weekly-err", http.StatusInternalServerError},
		{"/monthly-progress-err", http.StatusInternalServerError},
		{"/level-err/2?limit=5", http.StatusInternalServerError},
		{"/monthly-err?month=1&year=2026", http.StatusInternalServerError},
		{"/celebrate/2", http.StatusOK},
		{"/celebrate/99", http.StatusBadRequest},
	}

	for _, tt := range checks {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tt.path, nil))
		if w.Code != tt.code {
			t.Fatalf("path %s expected %d got %d", tt.path, tt.code, w.Code)
		}
	}

	_ = time.Now()
}
