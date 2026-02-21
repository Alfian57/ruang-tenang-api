package service

import (
	"context"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newArticleServiceForTest(t *testing.T, withCache bool) (*ArticleService, *repository.ArticleRepository, *repository.ArticleCategoryRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.ArticleCategory{}, &model.Article{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}

	articleRepo := repository.NewArticleRepository(db)
	categoryRepo := repository.NewArticleCategoryRepository(db)

	var cacheSvc *CacheService
	if withCache {
		cacheSvc = NewCacheService()
	}

	service := NewArticleService(articleRepo, categoryRepo, nil, nil, cacheSvc, nil)
	return service, articleRepo, categoryRepo, db
}

func seedArticleBaseData(t *testing.T, db *gorm.DB) (ownerID uint, otherID uint, categoryID uint) {
	t.Helper()
	owner := model.User{Name: "Owner", Username: "owner", Email: "owner@test.local", Password: "x"}
	other := model.User{Name: "Other", Username: "other", Email: "other@test.local", Password: "x"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatalf("create owner: %v", err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("create other: %v", err)
	}

	category := model.ArticleCategory{Name: "Mental Health", Description: "desc"}
	if err := db.Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	return owner.ID, other.ID, category.ID
}

func TestArticleService_ReadAndCategoryCache(t *testing.T) {
	ctx := context.Background()
	svc, _, _, db := newArticleServiceForTest(t, true)
	ownerID, _, categoryID := seedArticleBaseData(t, db)

	pub := model.Article{Title: "Published", Content: "content", ArticleCategoryID: categoryID, UserID: ownerID, Status: model.ArticleStatusPublished}
	draft := model.Article{Title: "Draft", Content: "content", ArticleCategoryID: categoryID, UserID: ownerID, Status: model.ArticleStatusDraft}
	if err := db.Create(&pub).Error; err != nil {
		t.Fatalf("create published article: %v", err)
	}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatalf("create draft article: %v", err)
	}

	if _, err := svc.GetArticleByID(ctx, pub.ID); err != nil {
		t.Fatalf("GetArticleByID failed: %v", err)
	}
	if _, err := svc.GetPublishedArticleByID(ctx, pub.ID); err != nil {
		t.Fatalf("GetPublishedArticleByID published failed: %v", err)
	}
	if _, err := svc.GetPublishedArticleByID(ctx, draft.ID); err == nil {
		t.Fatal("expected GetPublishedArticleByID to fail for draft article")
	}

	if _, err := svc.GetArticleBySlug(ctx, pub.Slug); err != nil {
		t.Fatalf("GetArticleBySlug failed: %v", err)
	}
	if got, err := svc.GetPublishedArticleBySlug(ctx, pub.Slug); err != nil || got.ID != pub.ID {
		t.Fatalf("GetPublishedArticleBySlug published failed: got=%+v err=%v", got, err)
	}
	if _, err := svc.GetPublishedArticleBySlug(ctx, draft.Slug); err == nil {
		t.Fatal("expected GetPublishedArticleBySlug to fail for draft article")
	}
	if _, err := svc.GetArticleBySlug(ctx, "missing-slug"); err == nil {
		t.Fatal("expected GetArticleBySlug to fail for missing slug")
	}
	if _, err := svc.GetPublishedArticleBySlug(ctx, "missing-slug"); err == nil {
		t.Fatal("expected GetPublishedArticleBySlug to fail for missing slug")
	}

	publishedList, publishedTotal, err := svc.GetPublishedArticles(ctx, &dto.ArticleQueryParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("GetPublishedArticles failed: %v", err)
	}
	if publishedTotal != 1 || len(publishedList) != 1 {
		t.Fatalf("expected 1 published article, total=%d len=%d", publishedTotal, len(publishedList))
	}

	adminList, adminTotal, err := svc.GetArticles(ctx, &dto.ArticleQueryParams{Page: 1, Limit: 10, Status: string(model.ArticleStatusDraft), UserID: ownerID})
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}
	if adminTotal < 1 || len(adminList) < 1 {
		t.Fatalf("expected at least 1 filtered article, total=%d len=%d", adminTotal, len(adminList))
	}

	list, total, err := svc.GetUserArticles(ctx, ownerID, 1, 10)
	if err != nil {
		t.Fatalf("GetUserArticles failed: %v", err)
	}
	if total != 2 || len(list) != 2 {
		t.Fatalf("expected 2 user articles, total=%d len=%d", total, len(list))
	}

	// cache warm-up
	categories1, err := svc.GetCategories(ctx)
	if err != nil {
		t.Fatalf("GetCategories first call failed: %v", err)
	}
	if len(categories1) == 0 {
		t.Fatal("expected categories from db")
	}

	// remove DB rows, second call should still return cached data
	if err := db.Exec("DELETE FROM article_categories").Error; err != nil {
		t.Fatalf("delete categories: %v", err)
	}
	categories2, err := svc.GetCategories(ctx)
	if err != nil {
		t.Fatalf("GetCategories second call failed: %v", err)
	}
	if len(categories2) == 0 {
		t.Fatal("expected categories from cache")
	}
}

func TestArticleService_UserMutationsAndPermissions(t *testing.T) {
	ctx := context.Background()
	svc, articleRepo, _, db := newArticleServiceForTest(t, false)
	ownerID, otherID, categoryID := seedArticleBaseData(t, db)

	article := model.Article{
		Title:             "Owned",
		Content:           "content",
		ArticleCategoryID: categoryID,
		UserID:            ownerID,
		Status:            model.ArticleStatusDraft,
		IsUserGenerated:   true,
		ModerationStatus:  model.ArticleModerationRejected,
	}
	if err := db.Create(&article).Error; err != nil {
		t.Fatalf("create owned article: %v", err)
	}

	_, err := svc.UpdateUserArticle(ctx, otherID, article.ID, &dto.UpdateUserArticleRequest{Title: "N", Content: "C", CategoryID: categoryID})
	if err == nil {
		t.Fatal("expected unauthorized update error")
	}

	article.Status = model.ArticleStatusBlocked
	_ = db.Save(&article).Error
	_, err = svc.UpdateUserArticle(ctx, ownerID, article.ID, &dto.UpdateUserArticleRequest{Title: "N", Content: "C", CategoryID: categoryID})
	if err == nil {
		t.Fatal("expected blocked update error")
	}

	article.Status = model.ArticleStatusDraft
	article.ModerationStatus = model.ArticleModerationRejected
	_ = db.Save(&article).Error
	updated, err := svc.UpdateUserArticle(ctx, ownerID, article.ID, &dto.UpdateUserArticleRequest{Title: "Updated", Content: "Updated content", CategoryID: categoryID})
	if err != nil {
		t.Fatalf("UpdateUserArticle failed: %v", err)
	}
	if updated.ModerationStatus != model.ArticleModerationPending {
		t.Fatalf("expected moderation reset to pending, got %s", updated.ModerationStatus)
	}

	created, err := svc.CreateUserArticle(ctx, ownerID, &dto.CreateUserArticleRequest{Title: "New", Content: "New content", CategoryID: categoryID})
	if err != nil {
		t.Fatalf("CreateUserArticle failed: %v", err)
	}
	if created.Status != model.ArticleStatusDraft || !created.IsUserGenerated {
		t.Fatalf("unexpected created article state: status=%s user_generated=%v", created.Status, created.IsUserGenerated)
	}

	err = svc.DeleteUserArticle(ctx, otherID, created.ID)
	if err == nil {
		t.Fatal("expected unauthorized delete error")
	}
	err = svc.DeleteUserArticle(ctx, ownerID, created.ID)
	if err != nil {
		t.Fatalf("DeleteUserArticle failed: %v", err)
	}
	if _, err := articleRepo.FindByID(ctx, created.ID); err == nil {
		t.Fatal("expected deleted article not found")
	}

	bySlug := model.Article{Title: "BySlug", Content: "content", ArticleCategoryID: categoryID, UserID: ownerID, Status: model.ArticleStatusDraft}
	if err := db.Create(&bySlug).Error; err != nil {
		t.Fatalf("create bySlug article: %v", err)
	}

	_, err = svc.UpdateUserArticleBySlug(ctx, otherID, bySlug.Slug, &dto.UpdateUserArticleRequest{Title: "No", Content: "No", CategoryID: categoryID})
	if err == nil {
		t.Fatal("expected unauthorized update by slug error")
	}
	if _, err := svc.UpdateUserArticleBySlug(ctx, ownerID, bySlug.Slug, &dto.UpdateUserArticleRequest{Title: "Yes", Content: "Yes", CategoryID: categoryID}); err != nil {
		t.Fatalf("UpdateUserArticleBySlug failed: %v", err)
	}

	bySlug.Status = model.ArticleStatusBlocked
	if err := db.Save(&bySlug).Error; err != nil {
		t.Fatalf("set bySlug blocked: %v", err)
	}
	if _, err := svc.UpdateUserArticleBySlug(ctx, ownerID, bySlug.Slug, &dto.UpdateUserArticleRequest{Title: "Blocked", Content: "Blocked", CategoryID: categoryID}); err == nil {
		t.Fatal("expected blocked update by slug error")
	}

	if _, err := svc.UpdateUserArticleBySlug(ctx, ownerID, "missing-slug", &dto.UpdateUserArticleRequest{Title: "No", Content: "No", CategoryID: categoryID}); err == nil {
		t.Fatal("expected update by slug not found error")
	}

	err = svc.DeleteUserArticleBySlug(ctx, otherID, bySlug.Slug)
	if err == nil {
		t.Fatal("expected unauthorized delete by slug error")
	}
	err = svc.DeleteUserArticleBySlug(ctx, ownerID, bySlug.Slug)
	if err != nil {
		t.Fatalf("DeleteUserArticleBySlug failed: %v", err)
	}

	if err := svc.DeleteUserArticleBySlug(ctx, ownerID, "missing-slug"); err == nil {
		t.Fatal("expected delete by slug not found error")
	}

	adminArticle := model.Article{Title: "AdminFlow", Content: "x", ArticleCategoryID: categoryID, UserID: ownerID, Status: model.ArticleStatusDraft}
	if err := db.Create(&adminArticle).Error; err != nil {
		t.Fatalf("create admin flow article: %v", err)
	}
	if err := svc.BlockArticle(ctx, adminArticle.ID); err != nil {
		t.Fatalf("BlockArticle failed: %v", err)
	}
	afterBlock, _ := articleRepo.FindByID(ctx, adminArticle.ID)
	if afterBlock.Status != model.ArticleStatusBlocked {
		t.Fatalf("expected blocked status, got %s", afterBlock.Status)
	}
	if err := svc.UnblockArticle(ctx, adminArticle.ID); err != nil {
		t.Fatalf("UnblockArticle failed: %v", err)
	}

	newCategory := &model.ArticleCategory{Name: "New Category", Description: "desc"}
	if err := svc.CreateCategory(ctx, newCategory); err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	afterBlock.Title = "Updated By Admin"
	if err := svc.UpdateArticle(ctx, afterBlock); err != nil {
		t.Fatalf("UpdateArticle failed: %v", err)
	}
	if err := svc.DeleteArticle(ctx, afterBlock.ID); err != nil {
		t.Fatalf("DeleteArticle failed: %v", err)
	}

	adminCreated := &model.Article{Title: "Admin Created", Content: "z", ArticleCategoryID: categoryID, UserID: ownerID, Status: model.ArticleStatusDraft}
	if err := svc.CreateArticle(ctx, adminCreated); err != nil {
		t.Fatalf("CreateArticle failed: %v", err)
	}
}

func TestArticleService_CreateUserArticle_WithModerationService(t *testing.T) {
	ctx := context.Background()
	svc, articleRepo, _, db := newArticleServiceForTest(t, false)
	ownerID, _, categoryID := seedArticleBaseData(t, db)

	moderationRepo := repository.NewModerationRepository(db)
	aiSvc := &AIModerationService{moderationRepo: moderationRepo}
	svc.moderationService = NewModerationService(
		moderationRepo,
		repository.NewUserRepository(db),
		articleRepo,
		nil,
		aiSvc,
	)

	created, err := svc.CreateUserArticle(ctx, ownerID, &dto.CreateUserArticleRequest{
		Title:      "Moderated New",
		Content:    "Konten artikel untuk diuji moderasi",
		CategoryID: categoryID,
	})
	if err != nil {
		t.Fatalf("CreateUserArticle with moderation service failed: %v", err)
	}
	if created.ModerationStatus != model.ArticleModerationApproved {
		t.Fatalf("expected moderation approved fallback, got %s", created.ModerationStatus)
	}
	if created.ModerationNotes == "" {
		t.Fatal("expected moderation notes to be populated")
	}
}

func TestArticleService_RepositoryErrorBranches(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	svc := NewArticleService(
		repository.NewArticleRepository(db),
		repository.NewArticleCategoryRepository(db),
		nil,
		nil,
		nil,
		nil,
	)

	if _, _, err := svc.GetPublishedArticles(ctx, &dto.ArticleQueryParams{Page: 1, Limit: 10}); err == nil {
		t.Fatal("expected GetPublishedArticles error on missing schema")
	}

	if _, _, err := svc.GetArticles(ctx, &dto.ArticleQueryParams{Page: 1, Limit: 10}); err == nil {
		t.Fatal("expected GetArticles error on missing schema")
	}

	if _, _, err := svc.GetUserArticles(ctx, 1, 1, 10); err == nil {
		t.Fatal("expected GetUserArticles error on missing schema")
	}

	if _, err := svc.GetArticleByID(ctx, 1); err == nil {
		t.Fatal("expected GetArticleByID error on missing schema")
	}

	if _, err := svc.GetArticleBySlug(ctx, "missing"); err == nil {
		t.Fatal("expected GetArticleBySlug error on missing schema")
	}
}
