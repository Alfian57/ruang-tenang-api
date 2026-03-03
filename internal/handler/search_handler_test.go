package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSearchDB(t *testing.T, migrateArticles bool, migrateSongs bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	if migrateArticles {
		if err := db.AutoMigrate(&model.User{}, &model.ArticleCategory{}, &model.Article{}); err != nil {
			t.Fatalf("migrate article tables: %v", err)
		}
	}
	if migrateSongs {
		if err := db.AutoMigrate(&model.SongCategory{}, &model.Song{}); err != nil {
			t.Fatalf("migrate song tables: %v", err)
		}
	}
	return db
}

func TestSearchHandler_Search(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("empty query", func(t *testing.T) {
		h := NewSearchHandler(repository.NewArticleRepository(newSearchDB(t, true, true)), repository.NewSongRepository(newSearchDB(t, true, true)))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/search?q=", nil)

		h.Search(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("article error", func(t *testing.T) {
		articleDB := newSearchDB(t, false, true)
		songDB := newSearchDB(t, true, true)
		h := NewSearchHandler(repository.NewArticleRepository(articleDB), repository.NewSongRepository(songDB))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/search?q=test", nil)

		h.Search(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("song error", func(t *testing.T) {
		articleDB := newSearchDB(t, true, true)
		songDB := newSearchDB(t, true, false)
		h := NewSearchHandler(repository.NewArticleRepository(articleDB), repository.NewSongRepository(songDB))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/search?q=test", nil)

		h.Search(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("search success", func(t *testing.T) {
		db := newSearchDB(t, true, true)
		now := "2026-01-01T00:00:00Z"

		if err := db.Exec(`INSERT INTO users (id, name, username, email, password, role, exp, created_at, updated_at)
			VALUES (1, 'Search User', 'search_user', 'search@test.local', 'x', 'member', 0, ?, ?)`, now, now).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
		if err := db.Exec(`INSERT INTO article_categories (id, name, slug, description, created_at, updated_at)
			VALUES (1, 'Mindfulness', 'mindfulness', 'desc', ?, ?)`, now, now).Error; err != nil {
			t.Fatalf("seed article category: %v", err)
		}
		if err := db.Exec(`INSERT INTO articles (id, title, slug, content, article_category_id, user_id, status, moderation_status, is_user_generated, created_at, updated_at)
			VALUES (1, 'Calm Breathing', ?, 'content', 1, 1, 'published', 'approved', 0, ?, ?)`, uuid.New().String(), now, now).Error; err != nil {
			t.Fatalf("seed article: %v", err)
		}
		if err := db.Exec(`INSERT INTO song_categories (id, name, slug, thumbnail, created_at, updated_at)
			VALUES (1, 'Calm', 'calm', '', ?, ?)`, now, now).Error; err != nil {
			t.Fatalf("seed song category: %v", err)
		}
		if err := db.Exec(`INSERT INTO songs (id, title, slug, file_path, song_category_id, created_at, updated_at)
			VALUES (1, 'Calm Waves', 'calm-waves', '/calm.mp3', 1, ?, ?)`, now, now).Error; err != nil {
			t.Fatalf("seed song: %v", err)
		}

		h := NewSearchHandler(repository.NewArticleRepository(db), repository.NewSongRepository(db))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/search?q=Calm", nil)

		h.Search(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
