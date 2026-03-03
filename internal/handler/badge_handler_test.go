package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newBadgeHandlerForGuardTest() *BadgeHandler {
	return NewBadgeHandler(&service.BadgeService{})
}

func TestBadgeHandler_GetBadgeCategories_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newBadgeHandlerForGuardTest()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/badges/categories", nil)

	h.GetBadgeCategories(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBadgeHandler_AuthGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newBadgeHandlerForGuardTest()

	tests := []struct {
		name string
		call func(*gin.Context)
	}{
		{"GetUserBadges", h.GetUserBadges},
		{"GetBadgeProgress", h.GetBadgeProgress},
		{"GetRecentlyEarnedBadges", h.GetRecentlyEarnedBadges},
		{"CheckNewBadges", h.CheckNewBadges},
		{"GetDisplayBadges", h.GetDisplayBadges},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			tt.call(c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401 for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func newBadgeHandlerWithData(t *testing.T) *BadgeHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	createUsers := `CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT,
		username TEXT,
		email TEXT,
		password TEXT,
		role TEXT,
		exp INTEGER,
		current_streak INTEGER,
		deleted_at DATETIME
	)`
	createBadges := `CREATE TABLE badge_definitions (
		id TEXT PRIMARY KEY,
		badge_key TEXT,
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
	)`
	createUserBadges := `CREATE TABLE user_badges (
		id TEXT PRIMARY KEY,
		user_id INTEGER,
		badge_id TEXT,
		earned_at DATETIME,
		is_showcased BOOLEAN
	)`
	createLevelConfigs := `CREATE TABLE level_configs (
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
	)`
	createExpHistories := `CREATE TABLE exp_histories (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		activity_type TEXT,
		points INTEGER,
		description TEXT,
		created_at DATETIME
	)`
	createStories := `CREATE TABLE inspiring_stories (
		id TEXT PRIMARY KEY,
		author_id INTEGER,
		status TEXT
	)`

	for _, query := range []string{createUsers, createBadges, createUserBadges, createLevelConfigs, createExpHistories, createStories} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatalf("create table failed: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO badge_definitions (id, badge_key, badge_name, description, icon, category, requirement_type, requirement_value, is_active) VALUES ('11111111-1111-1111-1111-111111111111', 'streak_3_days', 'Streak 3 Hari', 'desc', 'flame', 'streak', 'streak', 3, 1)`).Error; err != nil {
		t.Fatalf("seed badge: %v", err)
	}
	if err := db.Exec(`INSERT INTO users (id, name, username, email, password, role, exp, current_streak) VALUES (1, 'Member', 'member', 'member@test.local', 'hashed', 'member', 10, 1)`).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	badgeRepo := repository.NewBadgeRepository(db)
	userRepo := repository.NewUserRepository(db)
	levelConfigRepo := repository.NewLevelConfigRepository(db)
	badgeService := service.NewBadgeService(badgeRepo, userRepo, levelConfigRepo)

	return NewBadgeHandler(badgeService)
}

func TestBadgeHandler_SuccessPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newBadgeHandlerWithData(t)

	tests := []struct {
		name   string
		path   string
		setup  func(*gin.Context)
		call   func(*gin.Context)
		status int
	}{
		{
			name:   "GetAllBadges",
			path:   "/api/v1/badges",
			call:   h.GetAllBadges,
			status: http.StatusOK,
		},
		{
			name: "GetBadgesByCategory",
			path: "/api/v1/badges/category/streak",
			setup: func(c *gin.Context) {
				c.Params = gin.Params{{Key: "category", Value: "streak"}}
			},
			call:   h.GetBadgesByCategory,
			status: http.StatusOK,
		},
		{
			name: "GetUserBadges",
			path: "/api/v1/badges/my-badges",
			setup: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			call:   h.GetUserBadges,
			status: http.StatusOK,
		},
		{
			name: "GetBadgeProgress",
			path: "/api/v1/badges/progress",
			setup: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			call:   h.GetBadgeProgress,
			status: http.StatusOK,
		},
		{
			name: "GetRecentlyEarnedBadges",
			path: "/api/v1/badges/recent?days=999",
			setup: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			call:   h.GetRecentlyEarnedBadges,
			status: http.StatusOK,
		},
		{
			name: "GetDisplayBadges",
			path: "/api/v1/badges/display?limit=99",
			setup: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			call:   h.GetDisplayBadges,
			status: http.StatusOK,
		},
		{
			name: "CheckNewBadges",
			path: "/api/v1/badges/check",
			setup: func(c *gin.Context) {
				c.Set("user_id", uint(1))
			},
			call:   h.CheckNewBadges,
			status: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.setup != nil {
				tt.setup(c)
			}

			tt.call(c)

			if w.Code != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestBadgeHandler_DBErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	badgeService := service.NewBadgeService(
		repository.NewBadgeRepository(db),
		repository.NewUserRepository(db),
		repository.NewLevelConfigRepository(db),
	)
	h := NewBadgeHandler(badgeService)

	w1 := httptest.NewRecorder()
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = httptest.NewRequest(http.MethodGet, "/api/v1/badges", nil)
	h.GetAllBadges(c1)
	if w1.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for GetAllBadges db error, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/api/v1/badges/category/streak", nil)
	c2.Params = gin.Params{{Key: "category", Value: "streak"}}
	h.GetBadgesByCategory(c2)
	if w2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for GetBadgesByCategory db error, got %d", w2.Code)
	}

	tests := []struct {
		name string
		path string
		call func(*gin.Context)
		set  func(*gin.Context)
	}{
		{name: "GetUserBadges", path: "/api/v1/badges/my-badges", call: h.GetUserBadges, set: func(c *gin.Context) { c.Set("user_id", uint(1)) }},
		{name: "GetBadgeProgress", path: "/api/v1/badges/progress", call: h.GetBadgeProgress, set: func(c *gin.Context) { c.Set("user_id", uint(1)) }},
		{name: "GetRecentlyEarnedBadges", path: "/api/v1/badges/recent?days=3", call: h.GetRecentlyEarnedBadges, set: func(c *gin.Context) { c.Set("user_id", uint(1)) }},
		{name: "CheckNewBadges", path: "/api/v1/badges/check", call: h.CheckNewBadges, set: func(c *gin.Context) { c.Set("user_id", uint(1)) }},
		{name: "GetDisplayBadges", path: "/api/v1/badges/display?limit=3", call: h.GetDisplayBadges, set: func(c *gin.Context) { c.Set("user_id", uint(1)) }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.set != nil {
				tc.set(c)
			}
			tc.call(c)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected 500 for %s db error, got %d", tc.name, w.Code)
			}
		})
	}
}
