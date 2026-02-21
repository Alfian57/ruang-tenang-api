package service

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInspiringStoryServiceErrorPaths(t *testing.T) *InspiringStoryService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	return NewInspiringStoryService(
		repository.NewInspiringStoryRepository(db),
		repository.NewUserRepository(db),
		repository.NewLevelConfigRepository(db),
		nil,
		nil,
		nil,
	)
}

func TestInspiringStoryService_ErrorHeavyPaths(t *testing.T) {
	svc := setupInspiringStoryServiceErrorPaths(t)
	ctx := context.Background()
	id := uuid.New()

	if _, err := svc.GetCategories(ctx); err == nil {
		t.Fatal("expected GetCategories error")
	}
	if _, err := svc.CreateStory(ctx, 1, &dto.CreateStoryRequest{Title: "x", Content: "y"}); err == nil {
		t.Fatal("expected CreateStory error")
	}
	if _, err := svc.UpdateStory(ctx, id, 1, &dto.UpdateStoryRequest{Title: "x"}); err == nil {
		t.Fatal("expected UpdateStory not found error")
	}
	if err := svc.DeleteStory(ctx, id, 1); err == nil {
		t.Fatal("expected DeleteStory not found error")
	}
	if _, err := svc.GetStory(ctx, id, 1); err == nil {
		t.Fatal("expected GetStory not found error")
	}

	if _, err := svc.GetStories(ctx, &dto.StoryFilterRequest{Page: 0, Limit: 0, CategoryID: id.String()}, 1); err == nil {
		t.Fatal("expected GetStories error")
	}
	if _, err := svc.GetStories(ctx, &dto.StoryFilterRequest{Search: "hope"}, 1); err == nil {
		t.Fatal("expected GetStories search error")
	}
	if _, err := svc.GetStories(ctx, &dto.StoryFilterRequest{AuthorID: 1}, 1); err == nil {
		t.Fatal("expected GetStories author error")
	}
	if _, err := svc.GetUserStories(ctx, 1, "approved", 1, 10); err == nil {
		t.Fatal("expected GetUserStories error")
	}
	if _, err := svc.GetFeaturedStories(ctx, 0); err == nil {
		t.Fatal("expected GetFeaturedStories error")
	}
	if _, err := svc.GetPendingStories(ctx, 0, 0); err == nil {
		t.Fatal("expected GetPendingStories error")
	}

	if err := svc.ModerateStory(ctx, id, 1, &dto.ModerateStoryRequest{Status: "approved"}); err == nil {
		t.Fatal("expected ModerateStory not found error")
	}
	if err := svc.SetFeatured(ctx, id, true, 1); err == nil {
		t.Fatal("expected SetFeatured not found error")
	}
	if _, _, err := svc.ToggleHeart(ctx, id, 1); err == nil {
		t.Fatal("expected ToggleHeart not found error")
	}

	if _, err := svc.CreateComment(ctx, id, 1, &dto.CreateStoryCommentRequest{Content: "nice"}); err == nil {
		t.Fatal("expected CreateComment not found error")
	}
	if _, err := svc.GetComments(ctx, id, 1, 0, 0); err == nil {
		t.Fatal("expected GetComments error")
	}
	if err := svc.DeleteComment(ctx, id, 1); err == nil {
		t.Fatal("expected DeleteComment not found error")
	}
	if err := svc.HideComment(ctx, id, &dto.HideStoryCommentRequest{Reason: "spam"}); err == nil {
		t.Fatal("expected HideComment not found error")
	}
	if _, _, err := svc.ToggleCommentHeart(ctx, id, 1); err == nil {
		t.Fatal("expected ToggleCommentHeart not found error")
	}

	if _, err := svc.GetMostAppreciatedStories(ctx, 1, 2026, 0); err == nil {
		t.Fatal("expected GetMostAppreciatedStories error")
	}
}
