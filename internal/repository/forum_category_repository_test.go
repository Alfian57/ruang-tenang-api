package repository

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupForumCategoryRepositoryTest(t *testing.T) (ForumCategoryRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.ForumCategory{}); err != nil {
		t.Fatalf("migrate forum_categories: %v", err)
	}

	repo := NewForumCategoryRepository(db)
	seed := &model.ForumCategory{Name: "General", Slug: "general"}
	if err := repo.Create(context.Background(), seed); err != nil {
		t.Fatalf("seed forum category: %v", err)
	}

	return repo, db
}

func TestForumCategoryRepository_Branches(t *testing.T) {
	repo, db := setupForumCategoryRepositoryTest(t)
	ctx := context.Background()

	all, err := repo.FindAll(ctx)
	if err != nil || len(all) != 1 {
		t.Fatalf("FindAll unexpected err=%v len=%d", err, len(all))
	}

	byID, err := repo.FindByID(ctx, all[0].ID)
	if err != nil || byID.Slug != "general" {
		t.Fatalf("FindByID unexpected category=%+v err=%v", byID, err)
	}
	if _, err := repo.FindByID(ctx, 999999); err == nil {
		t.Fatal("expected FindByID missing error")
	}

	bySlug, err := repo.FindBySlug(ctx, "general")
	if err != nil || bySlug.ID != all[0].ID {
		t.Fatalf("FindBySlug unexpected category=%+v err=%v", bySlug, err)
	}
	if _, err := repo.FindBySlug(ctx, "missing-slug"); err == nil {
		t.Fatal("expected FindBySlug missing error")
	}

	created := &model.ForumCategory{Name: "Kesehatan Mental"}
	if err := repo.Create(ctx, created); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Slug == "" {
		t.Fatal("expected slug generated on create")
	}

	created.Name = "Kesehatan Mental Updated"
	if err := repo.Update(ctx, created); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, created.ID); err == nil {
		t.Fatal("expected deleted category not found")
	}

	if err := db.Exec(`DROP TABLE forum_categories`).Error; err != nil {
		t.Fatalf("drop forum_categories: %v", err)
	}
	if _, err := repo.FindAll(ctx); err == nil {
		t.Fatal("expected FindAll error when table missing")
	}
}
