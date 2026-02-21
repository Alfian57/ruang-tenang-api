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
		t.Fatal("expected Search error on sqlite ILIKE")
	}
}
