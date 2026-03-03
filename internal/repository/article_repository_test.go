package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type postgresNamedDialectorForArticle struct {
	gorm.Dialector
}

func (d postgresNamedDialectorForArticle) Name() string {
	return "postgres"
}

func setupArticleRepoDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ArticleCategory{}, &model.Article{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
	return db
}

func seedArticleBase(t *testing.T, db *gorm.DB) (model.User, model.ArticleCategory, model.Article, model.Article) {
	t.Helper()
	user := model.User{Name: "Author", Username: "author", Email: "author@example.com", Password: "secret", Role: model.RoleMember}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	category := model.ArticleCategory{Name: "Mindfulness", Slug: "mindfulness"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	draft := model.Article{
		Title:             "Draft Article",
		Slug:              "draft-article",
		Content:           "draft content",
		ArticleCategoryID: category.ID,
		UserID:            user.ID,
		Status:            model.ArticleStatusDraft,
	}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatalf("failed to seed draft article: %v", err)
	}

	published := model.Article{
		Title:             "Published Article",
		Slug:              "published-article",
		Content:           "published content",
		ArticleCategoryID: category.ID,
		UserID:            user.ID,
		Status:            model.ArticleStatusPublished,
	}
	if err := db.Create(&published).Error; err != nil {
		t.Fatalf("failed to seed published article: %v", err)
	}

	return user, category, draft, published
}

func TestArticleRepository_FullFlow(t *testing.T) {
	db := setupArticleRepoDB(t)
	ctx := context.Background()
	repo := NewArticleRepository(db)
	user, category, draft, published := seedArticleBase(t, db)

	all, total, err := repo.FindAll(ctx, 0, "", 1, 10, "", 0)
	if err != nil {
		t.Fatalf("find all failed: %v", err)
	}
	if total != 2 || len(all) != 2 {
		t.Fatalf("expected 2 articles, got total=%d len=%d", total, len(all))
	}

	filtered, filteredTotal, err := repo.FindAll(ctx, category.ID, "", 1, 10, string(model.ArticleStatusDraft), user.ID)
	if err != nil {
		t.Fatalf("find all with filters failed: %v", err)
	}
	if filteredTotal != 1 || len(filtered) != 1 || filtered[0].ID != draft.ID {
		t.Fatalf("expected only draft article, got total=%d len=%d", filteredTotal, len(filtered))
	}

	pub, pubTotal, err := repo.FindPublished(ctx, 0, "", 1, 10)
	if err != nil {
		t.Fatalf("find published failed: %v", err)
	}
	if pubTotal != 1 || len(pub) != 1 || pub[0].ID != published.ID {
		t.Fatalf("expected only published article, got total=%d len=%d", pubTotal, len(pub))
	}

	byUser, byUserTotal, err := repo.FindByUserID(ctx, user.ID, 1, 10)
	if err != nil {
		t.Fatalf("find by user failed: %v", err)
	}
	if byUserTotal != 2 || len(byUser) != 2 {
		t.Fatalf("expected 2 user articles, got total=%d len=%d", byUserTotal, len(byUser))
	}

	byID, err := repo.FindByID(ctx, draft.ID)
	if err != nil {
		t.Fatalf("find by id failed: %v", err)
	}
	if byID.ID != draft.ID {
		t.Fatalf("expected id %d, got %d", draft.ID, byID.ID)
	}

	bySlug, err := repo.FindBySlug(ctx, published.Slug)
	if err != nil {
		t.Fatalf("find by slug failed: %v", err)
	}
	if bySlug.ID != published.ID {
		t.Fatalf("expected id %d, got %d", published.ID, bySlug.ID)
	}
	if _, err := repo.FindBySlug(ctx, "missing-slug"); err == nil {
		t.Fatal("expected FindBySlug missing error")
	}

	created := model.Article{
		Title:             "New Article",
		Slug:              "new-article",
		Content:           "new content",
		ArticleCategoryID: category.ID,
		UserID:            user.ID,
		Status:            model.ArticleStatusDraft,
	}
	if err := repo.Create(ctx, &created); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	created.Title = "Updated Article"
	if err := repo.Update(ctx, &created); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find updated failed: %v", err)
	}
	if updated.Title != "Updated Article" {
		t.Fatalf("expected updated title, got %s", updated.Title)
	}

	if err := repo.UpdateStatus(ctx, created.ID, model.ArticleStatusPublished); err != nil {
		t.Fatalf("update status failed: %v", err)
	}

	updatedStatus, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find updated status failed: %v", err)
	}
	if updatedStatus.Status != model.ArticleStatusPublished {
		t.Fatalf("expected published status, got %s", updatedStatus.Status)
	}

	since := time.Now().Add(-1 * time.Minute)
	updatedArticles, err := repo.FindUpdatedSince(ctx, since)
	if err != nil {
		t.Fatalf("find updated since failed: %v", err)
	}
	if len(updatedArticles) < 1 {
		t.Fatalf("expected updated articles, got %d", len(updatedArticles))
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if _, err := repo.FindByID(ctx, created.ID); err == nil {
		t.Fatal("expected deleted article not found")
	}
}

func TestArticleCategoryRepository_FullFlow(t *testing.T) {
	db := setupArticleRepoDB(t)
	ctx := context.Background()
	repo := NewArticleCategoryRepository(db)

	created := model.ArticleCategory{Name: "Anxiety", Slug: "anxiety", Description: "anxiety resources"}
	if err := repo.Create(ctx, &created); err != nil {
		t.Fatalf("create category failed: %v", err)
	}

	all, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("find all categories failed: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 category, got %d", len(all))
	}

	byID, err := repo.FindByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("find category by id failed: %v", err)
	}
	if byID.Name != "Anxiety" {
		t.Fatalf("expected Anxiety, got %s", byID.Name)
	}

	if _, err := repo.FindByID(ctx, 99999); err == nil {
		t.Fatal("expected not found error for unknown category id")
	}
}

func TestArticleRepository_FindAll_PostgresSearchBranchOnSqlite(t *testing.T) {
	dialector := postgresNamedDialectorForArticle{Dialector: sqlite.Open(":memory:")}
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite with postgres name dialector: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.ArticleCategory{}, &model.Article{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}

	repo := NewArticleRepository(db)
	if _, _, err := repo.FindAll(context.Background(), 0, "mind", 1, 5, "", 0); err == nil {
		t.Fatal("expected postgres ILIKE branch to fail on sqlite backend")
	}
}
