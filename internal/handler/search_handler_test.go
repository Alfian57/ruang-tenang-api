package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSearchDB(t *testing.T, migrateArticles bool, migrateSongs bool) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
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
		db := newSearchDB(t, true, true)
		h := NewSearchHandler(repository.NewArticleRepository(db), repository.NewSongRepository(db))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/search?q=test", nil)

		h.Search(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})
}
