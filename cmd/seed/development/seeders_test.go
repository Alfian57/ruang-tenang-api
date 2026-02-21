package development

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSeedDevDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ArticleCategory{}, &model.Article{}, &model.SongCategory{}, &model.Song{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func createDummySeedAssets(t *testing.T) {
	t.Helper()
	_ = os.MkdirAll(filepath.Join("assets", "images"), 0o755)
	_ = os.MkdirAll(filepath.Join("assets", "audio"), 0o755)

	images := []string{
		"avatar-1.jpg", "avatar-2.jpg", "avatar-3.jpg", "avatar-4.jpg",
		"article-mental.jpg", "article-tips.jpg", "article-meditation.jpg", "article-stress.jpg", "article-sleep.jpg",
		"category-alam.jpg", "category-piano.jpg", "category-hujan.jpg", "category-laut.jpg", "category-meditasi.jpg",
		"song-forest.jpg", "song-river.jpg", "song-piano.jpg", "song-soft-piano.jpg", "song-rain.jpg", "song-thunder.jpg",
	}
	for _, f := range images {
		p := filepath.Join("assets", "images", f)
		if err := os.WriteFile(p, []byte("img"), 0o644); err != nil {
			t.Fatalf("write image asset %s: %v", f, err)
		}
	}

	audios := []string{"song-1.mp3", "song-2.mp3", "song-3.mp3", "song-4.mp3", "song-5.mp3", "song-6.mp3"}
	for _, f := range audios {
		p := filepath.Join("assets", "audio", f)
		if err := os.WriteFile(p, []byte("audio"), 0o644); err != nil {
			t.Fatalf("write audio asset %s: %v", f, err)
		}
	}
}

func TestDevelopmentSeeders_UsersArticlesSongs(t *testing.T) {
	withTempWorkingDir(t)
	createDummySeedAssets(t)
	db := setupSeedDevDB(t)

	if err := SeedTestUsers(db); err != nil {
		t.Fatalf("SeedTestUsers failed: %v", err)
	}
	var usersCount int64
	if err := db.Model(&model.User{}).Count(&usersCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if usersCount < 4 {
		t.Fatalf("expected seeded users >= 4, got %d", usersCount)
	}

	admin := model.User{Name: "Admin", Username: "admin", Email: "admin@ruangtenang.id", Password: "x", Role: model.RoleAdmin}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	cats := []model.ArticleCategory{
		{Name: "Kesehatan Mental"},
		{Name: "Tips & Trik"},
		{Name: "Meditasi"},
	}
	for i := range cats {
		if err := db.Create(&cats[i]).Error; err != nil {
			t.Fatalf("seed article category: %v", err)
		}
	}

	if err := SeedArticles(db); err != nil {
		t.Fatalf("SeedArticles failed: %v", err)
	}
	var articlesCount int64
	if err := db.Model(&model.Article{}).Count(&articlesCount).Error; err != nil {
		t.Fatalf("count articles: %v", err)
	}
	if articlesCount < 5 {
		t.Fatalf("expected seeded articles >= 5, got %d", articlesCount)
	}

	songCats := []model.SongCategory{{Name: "Alam"}, {Name: "Piano"}, {Name: "Hujan"}, {Name: "Laut"}, {Name: "Meditasi"}}
	for i := range songCats {
		if err := db.Create(&songCats[i]).Error; err != nil {
			t.Fatalf("seed song category: %v", err)
		}
	}

	if err := SeedSongs(db); err != nil {
		t.Fatalf("SeedSongs failed: %v", err)
	}
	var songsCount int64
	if err := db.Model(&model.Song{}).Count(&songsCount).Error; err != nil {
		t.Fatalf("count songs: %v", err)
	}
	if songsCount < 10 {
		t.Fatalf("expected seeded songs >= 10, got %d", songsCount)
	}
}
