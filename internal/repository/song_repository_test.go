package repository

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSongRepository_ErrorPathsOnEmptyDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	categoryRepo := NewSongCategoryRepository(db)
	songRepo := NewSongRepository(db)
	ctx := context.Background()

	if _, err := categoryRepo.FindAll(ctx); err == nil {
		t.Fatal("expected FindAll categories error")
	}
	if _, err := categoryRepo.FindAllWithSongCount(ctx); err == nil {
		t.Fatal("expected FindAllWithSongCount categories error")
	}
	if _, err := categoryRepo.FindByID(ctx, 1); err == nil {
		t.Fatal("expected FindByID category error")
	}
	if err := categoryRepo.Create(ctx, &model.SongCategory{}); err == nil {
		t.Fatal("expected Create category error")
	}
	if _, err := categoryRepo.FindBySlug(ctx, "calm"); err == nil {
		t.Fatal("expected FindBySlug category error")
	}

	if _, err := songRepo.FindByCategoryID(ctx, 1); err == nil {
		t.Fatal("expected FindByCategoryID song error")
	}
	if _, err := songRepo.FindByID(ctx, 1); err == nil {
		t.Fatal("expected FindByID song error")
	}
	if _, err := songRepo.FindAll(ctx); err == nil {
		t.Fatal("expected FindAll songs error")
	}
	if err := songRepo.Create(ctx, &model.Song{}); err == nil {
		t.Fatal("expected Create song error")
	}
	if _, err := songRepo.FindBySlug(ctx, "song-a"); err == nil {
		t.Fatal("expected FindBySlug song error")
	}

	if count := songRepo.CountByCategoryID(ctx, 1); count != 0 {
		t.Fatalf("expected count 0 on empty db error path, got %d", count)
	}
	if _, err := songRepo.Search(ctx, "calm"); err == nil {
		t.Fatal("expected Search error on empty db")
	}
}

func TestSongRepository_SuccessAndMissingBranches(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	queries := []string{
		`CREATE TABLE song_categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			slug TEXT,
			thumbnail TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			slug TEXT,
			file_path TEXT,
			thumbnail TEXT,
			song_category_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
	}
	for _, q := range queries {
		if execErr := db.Exec(q).Error; execErr != nil {
			t.Fatalf("create schema failed: %v", execErr)
		}
	}

	ctx := context.Background()
	categoryRepo := NewSongCategoryRepository(db)
	songRepo := NewSongRepository(db)

	category := &model.SongCategory{Name: "Relax", Slug: "relax"}
	if err := categoryRepo.Create(ctx, category); err != nil {
		t.Fatalf("create category: %v", err)
	}

	song := &model.Song{Title: "Ocean", Slug: "ocean", FilePath: "ocean.mp3", SongCategoryID: category.ID}
	if err := songRepo.Create(ctx, song); err != nil {
		t.Fatalf("create song: %v", err)
	}

	allCategories, err := categoryRepo.FindAll(ctx)
	if err != nil || len(allCategories) != 1 {
		t.Fatalf("FindAll categories unexpected err=%v len=%d", err, len(allCategories))
	}
	allCategoriesWithCount, err := categoryRepo.FindAllWithSongCount(ctx)
	if err != nil || len(allCategoriesWithCount) != 1 {
		t.Fatalf("FindAllWithSongCount unexpected err=%v len=%d", err, len(allCategoriesWithCount))
	}

	catByID, err := categoryRepo.FindByID(ctx, category.ID)
	if err != nil || catByID.Slug != "relax" {
		t.Fatalf("FindByID category unexpected category=%+v err=%v", catByID, err)
	}
	if _, err := categoryRepo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected category FindByID missing error")
	}

	catBySlug, err := categoryRepo.FindBySlug(ctx, "relax")
	if err != nil || catBySlug.ID != category.ID {
		t.Fatalf("FindBySlug category unexpected category=%+v err=%v", catBySlug, err)
	}
	if _, err := categoryRepo.FindBySlug(ctx, "missing-category"); err == nil {
		t.Fatal("expected category FindBySlug missing error")
	}

	byCategory, err := songRepo.FindByCategoryID(ctx, category.ID)
	if err != nil || len(byCategory) != 1 {
		t.Fatalf("FindByCategoryID unexpected err=%v len=%d", err, len(byCategory))
	}

	byID, err := songRepo.FindByID(ctx, song.ID)
	if err != nil || byID.Slug != "ocean" {
		t.Fatalf("FindByID song unexpected song=%+v err=%v", byID, err)
	}
	if _, err := songRepo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected song FindByID missing error")
	}

	allSongs, err := songRepo.FindAll(ctx)
	if err != nil || len(allSongs) != 1 {
		t.Fatalf("FindAll songs unexpected err=%v len=%d", err, len(allSongs))
	}

	bySlug, err := songRepo.FindBySlug(ctx, "ocean")
	if err != nil || bySlug.ID != song.ID {
		t.Fatalf("FindBySlug song unexpected song=%+v err=%v", bySlug, err)
	}
	if _, err := songRepo.FindBySlug(ctx, "missing-song"); err == nil {
		t.Fatal("expected song FindBySlug missing error")
	}

	if count := songRepo.CountByCategoryID(ctx, category.ID); count != 1 {
		t.Fatalf("expected CountByCategoryID 1, got %d", count)
	}
	if count := songRepo.CountByCategoryID(ctx, 999999); count != 0 {
		t.Fatalf("expected CountByCategoryID 0 for unknown category, got %d", count)
	}

	results, err := songRepo.Search(ctx, "Ocean")
	if err != nil || len(results) != 1 {
		t.Fatalf("expected Search success on sqlite fallback, err=%v len=%d", err, len(results))
	}
}
