package service

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSongService(t *testing.T) (*SongService, *gorm.DB, *CacheService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE song_categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			slug TEXT,
			thumbnail TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE songs (
			id INTEGER PRIMARY KEY,
			title TEXT,
			slug TEXT,
			file_path TEXT,
			thumbnail TEXT,
			song_category_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`INSERT INTO song_categories (id, name, slug, thumbnail, created_at, updated_at) VALUES (1, 'Calm', 'calm', 'thumb.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO songs (id, title, slug, file_path, thumbnail, song_category_id, created_at, updated_at) VALUES (1, 'Song A', 'song-a', '/a.mp3', 'a.png', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
	}
	for _, q := range schema {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("setup song service schema failed: %v", err)
		}
	}

	cache := NewCacheService()
	svc := NewSongService(repository.NewSongRepository(db), repository.NewSongCategoryRepository(db), cache)
	return svc, db, cache
}

func TestSongService_ReadPathsAndCache(t *testing.T) {
	svc, db, _ := setupSongService(t)
	ctx := context.Background()

	cats, err := svc.GetCategories(ctx)
	if err != nil || len(cats) == 0 {
		t.Fatalf("get categories failed: err=%v len=%d", err, len(cats))
	}

	songs, err := svc.GetSongsByCategoryBySlug(ctx, "calm")
	if err != nil || len(songs) == 0 {
		t.Fatalf("get songs by slug failed: err=%v len=%d", err, len(songs))
	}
	if _, err := svc.GetSongsByCategoryBySlug(ctx, "missing"); err == nil {
		t.Fatal("expected missing category error")
	}

	song, err := svc.GetSongByID(ctx, 1)
	if err != nil || song.ID != 1 {
		t.Fatalf("get song by id failed: err=%v song=%+v", err, song)
	}
	if _, err := svc.GetSongByID(ctx, 999); err == nil {
		t.Fatal("expected song not found error")
	}

	songBySlug, err := svc.GetSongBySlug(ctx, "song-a")
	if err != nil || songBySlug.ID != 1 {
		t.Fatalf("get song by slug failed: err=%v song=%+v", err, songBySlug)
	}
	if _, err := svc.GetSongBySlug(ctx, "missing"); err == nil {
		t.Fatal("expected song by slug not found error")
	}

	if err := db.Exec(`DROP TABLE song_categories`).Error; err != nil {
		t.Fatalf("drop categories table: %v", err)
	}
	cachedCats, err := svc.GetCategories(ctx)
	if err != nil || len(cachedCats) == 0 {
		t.Fatalf("expected cache hit categories, err=%v len=%d", err, len(cachedCats))
	}
}

func TestSongService_CreatePathsInvalidateCache(t *testing.T) {
	svc, _, cache := setupSongService(t)
	ctx := context.Background()

	cache.Set(CacheKeySongCategories, []string{"cached"})
	if err := svc.CreateCategory(ctx, &model.SongCategory{Name: "Focus", Slug: "focus"}); err != nil {
		t.Fatalf("create category failed: %v", err)
	}
	if cache.Get(CacheKeySongCategories) != nil {
		t.Fatal("expected categories cache to be invalidated after create category")
	}

	cache.Set(CacheKeySongCategories, []string{"cached"})
	if err := svc.CreateSong(ctx, &model.Song{Title: "Song B", Slug: "song-b", FilePath: "/b.mp3", SongCategoryID: 1}); err != nil {
		t.Fatalf("create song failed: %v", err)
	}
	if cache.Get(CacheKeySongCategories) != nil {
		t.Fatal("expected categories cache to be invalidated after create song")
	}
}
