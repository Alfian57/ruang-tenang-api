package service

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewCommunityProgressService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := NewCommunityProgressService(
		repository.NewCommunityProgressRepository(db),
		repository.NewLevelConfigRepository(db),
		repository.NewFeatureUnlockRepository(db),
		repository.NewBadgeRepository(db),
		repository.NewUserRepository(db),
	)

	if svc == nil {
		t.Fatal("expected community progress service instance")
	}
}

func TestNewForumService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := NewForumService(repository.NewForumRepository(db), nil, NewGamificationService(db), nil)
	if svc == nil {
		t.Fatal("expected forum service instance")
	}
}

func TestNewContentContextService(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := NewContentContextService(
		repository.NewArticleRepository(db),
		repository.NewSongRepository(db),
		repository.NewSongCategoryRepository(db),
		repository.NewForumRepository(db),
	)

	if svc == nil {
		t.Fatal("expected content context service instance")
	}

	_ = svc.GetContentContext(context.Background())
	svc.Stop(context.Background())
}
