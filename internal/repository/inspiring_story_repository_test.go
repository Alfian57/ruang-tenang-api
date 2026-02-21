package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInspiringStoryRepository_ErrorPathsOnEmptyDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	repo := NewInspiringStoryRepository(db)
	if repo == nil {
		t.Fatal("expected repository instance")
	}

	ctx := context.Background()
	id := uuid.New()
	categoryID := uuid.New()
	commentID := uuid.New()

	if _, err := repo.GetAllCategories(ctx); err == nil {
		t.Fatal("expected GetAllCategories error on empty db")
	}
	if _, err := repo.GetCategoryByID(ctx, categoryID); err == nil {
		t.Fatal("expected GetCategoryByID error on empty db")
	}
	if _, err := repo.GetCategoryBySlug(ctx, "hope"); err == nil {
		t.Fatal("expected GetCategoryBySlug error on empty db")
	}
	if _, err := repo.GetCategoriesWithCount(ctx); err == nil {
		t.Fatal("expected GetCategoriesWithCount error on empty db")
	}

	if err := repo.Create(ctx, &model.InspiringStory{}); err == nil {
		t.Fatal("expected Create error on empty db")
	}
	if _, err := repo.GetByID(ctx, id); err == nil {
		t.Fatal("expected GetByID error on empty db")
	}
	if _, err := repo.GetByIDWithRelations(ctx, id); err == nil {
		t.Fatal("expected GetByIDWithRelations error on empty db")
	}
	if err := repo.Update(ctx, &model.InspiringStory{ID: id}); err == nil {
		t.Fatal("expected Update error on empty db")
	}
	if err := repo.Delete(ctx, id); err == nil {
		t.Fatal("expected Delete error on empty db")
	}

	if _, _, err := repo.GetApprovedStories(ctx, 1, 10, "recent"); err == nil {
		t.Fatal("expected GetApprovedStories error on empty db")
	}
	if _, _, err := repo.GetStoriesByCategory(ctx, categoryID, 1, 10); err == nil {
		t.Fatal("expected GetStoriesByCategory error on empty db")
	}
	if _, _, err := repo.GetStoriesByAuthor(ctx, 1, "", 1, 10); err == nil {
		t.Fatal("expected GetStoriesByAuthor error on empty db")
	}
	if _, _, err := repo.SearchStories(ctx, "test", 1, 10); err == nil {
		t.Fatal("expected SearchStories error on empty db")
	}
	if _, err := repo.GetFeaturedStories(ctx, 5); err == nil {
		t.Fatal("expected GetFeaturedStories error on empty db")
	}
	if _, _, err := repo.GetPendingStories(ctx, 1, 10); err == nil {
		t.Fatal("expected GetPendingStories error on empty db")
	}
	if err := repo.UpdateStatus(ctx, id, string(model.StoryStatusApproved), 1, "ok"); err == nil {
		t.Fatal("expected UpdateStatus error on empty db")
	}
	if err := repo.IncrementViewCount(ctx, id); err == nil {
		t.Fatal("expected IncrementViewCount error on empty db")
	}
	if err := repo.SetFeatured(ctx, id, true, 1); err == nil {
		t.Fatal("expected SetFeatured error on empty db")
	}
	if _, err := repo.GetAuthorStoriesCount(ctx, 1, 1, 2025); err == nil {
		t.Fatal("expected GetAuthorStoriesCount error on empty db")
	}

	if err := repo.SetStoryTags(ctx, id, []string{"hope"}); err == nil {
		t.Fatal("expected SetStoryTags error on empty db")
	}
	if _, err := repo.GetStoryTags(ctx, id); err == nil {
		t.Fatal("expected GetStoryTags error on empty db")
	}
	if err := repo.SetStoryCategories(ctx, id, []uuid.UUID{categoryID}); err == nil {
		t.Fatal("expected SetStoryCategories error on empty db")
	}

	if err := repo.AddHeart(ctx, id, 1); err == nil {
		t.Fatal("expected AddHeart error on empty db")
	}
	if err := repo.RemoveHeart(ctx, id, 1); err == nil {
		t.Fatal("expected RemoveHeart error on empty db")
	}
	if repo.HasHearted(ctx, id, 1) {
		t.Fatal("expected HasHearted false on empty db")
	}
	if _, err := repo.GetStoryHeartCount(ctx, id); err == nil {
		t.Fatal("expected GetStoryHeartCount error on empty db")
	}

	if err := repo.CreateComment(ctx, &model.StoryComment{StoryID: id}); err == nil {
		t.Fatal("expected CreateComment error on empty db")
	}
	if _, err := repo.GetCommentByID(ctx, commentID); err == nil {
		t.Fatal("expected GetCommentByID error on empty db")
	}
	if _, _, err := repo.GetStoryComments(ctx, id, 1, 10, true); err == nil {
		t.Fatal("expected GetStoryComments error on empty db")
	}
	if err := repo.HideComment(ctx, commentID, true, "spam"); err == nil {
		t.Fatal("expected HideComment error on empty db")
	}
	if err := repo.DeleteComment(ctx, commentID, id); err == nil {
		t.Fatal("expected DeleteComment error on empty db")
	}
	if err := repo.AddCommentHeart(ctx, commentID, 1); err == nil {
		t.Fatal("expected AddCommentHeart error on empty db")
	}
	if err := repo.RemoveCommentHeart(ctx, commentID, 1); err == nil {
		t.Fatal("expected RemoveCommentHeart error on empty db")
	}
	if repo.HasHeartedComment(ctx, commentID, 1) {
		t.Fatal("expected HasHeartedComment false on empty db")
	}

	if stats, err := repo.GetAuthorStats(ctx, 1); err != nil || stats == nil {
		t.Fatalf("expected GetAuthorStats non-error stats, err=%v stats=%v", err, stats)
	}
	if _, err := repo.GetMostAppreciatedStories(ctx, 1, 2025, 10); err == nil {
		t.Fatal("expected GetMostAppreciatedStories error on empty db")
	}
}

func TestInspiringStoryRepository_SuccessPathsForLowCoverageFunctions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, username TEXT, email TEXT, password TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, title TEXT, content TEXT, cover_image TEXT, is_anonymous BOOLEAN, has_trigger_warning BOOLEAN, trigger_warning_text TEXT, status TEXT, moderator_id INTEGER, moderation_feedback TEXT, moderated_at DATETIME, view_count INTEGER DEFAULT 0, heart_count INTEGER DEFAULT 0, comment_count INTEGER DEFAULT 0, is_featured BOOLEAN DEFAULT 0, featured_at DATETIME, featured_by INTEGER, featured_until DATETIME, created_at DATETIME, updated_at DATETIME, published_at DATETIME)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, PRIMARY KEY (story_id, category_id))`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, parent_comment_id TEXT, content TEXT, heart_count INTEGER DEFAULT 0, is_hidden BOOLEAN DEFAULT 0, hidden_reason TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER, created_at DATETIME)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema setup failed: %v", err)
		}
	}

	repo := NewInspiringStoryRepository(db)
	ctx := context.Background()

	if err := db.Exec(`INSERT INTO users (id, name, username, email, password, created_at, updated_at) VALUES (1, 'Author', 'story_author', 'story_author@test.local', 'x', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	cat1 := model.StoryCategory{ID: uuid.New(), Slug: "hope", Name: "Hope", IsActive: true}
	cat2 := model.StoryCategory{ID: uuid.New(), Slug: "growth", Name: "Growth", IsActive: true}
	if err := db.Create(&cat1).Error; err != nil {
		t.Fatalf("create category1: %v", err)
	}
	if err := db.Create(&cat2).Error; err != nil {
		t.Fatalf("create category2: %v", err)
	}

	storyID := uuid.New()
	now := time.Now()
	story := &model.InspiringStory{ID: storyID, AuthorID: 1, Title: "My Story", Content: "content", Status: model.StoryStatusApproved, CreatedAt: now, UpdatedAt: now, PublishedAt: &now}
	if err := db.Create(story).Error; err != nil {
		t.Fatalf("create story: %v", err)
	}

	if err := repo.SetStoryTags(ctx, storyID, []string{"hope", "calm"}); err != nil {
		t.Fatalf("set story tags failed: %v", err)
	}
	tags, err := repo.GetStoryTags(ctx, storyID)
	if err != nil {
		t.Fatalf("get story tags failed: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	if err := repo.SetStoryCategories(ctx, storyID, []uuid.UUID{cat1.ID, cat2.ID}); err != nil {
		t.Fatalf("set story categories failed: %v", err)
	}
	var relCount int64
	if err := db.Model(&model.StoryCategoryRelation{}).Where("story_id = ?", storyID).Count(&relCount).Error; err != nil {
		t.Fatalf("count story category relations failed: %v", err)
	}
	if relCount != 2 {
		t.Fatalf("expected 2 category relations, got %d", relCount)
	}

	if err := repo.SetFeatured(ctx, storyID, true, 1); err != nil {
		t.Fatalf("set featured true failed: %v", err)
	}
	if err := repo.SetFeatured(ctx, storyID, false, 1); err != nil {
		t.Fatalf("set featured false failed: %v", err)
	}

	if err := db.Create(&model.StoryHeart{StoryID: storyID, UserID: 1}).Error; err != nil {
		t.Fatalf("create story heart: %v", err)
	}
	commentID := uuid.New()
	if err := db.Create(&model.StoryComment{ID: commentID, StoryID: storyID, UserID: 1, Content: "great"}).Error; err != nil {
		t.Fatalf("create story comment: %v", err)
	}
	if err := db.Create(&model.StoryCommentHeart{CommentID: commentID, UserID: 1}).Error; err != nil {
		t.Fatalf("create story comment heart: %v", err)
	}

	if err := repo.DeleteComment(ctx, commentID, storyID); err != nil {
		t.Fatalf("delete comment failed: %v", err)
	}

	if err := repo.Delete(ctx, storyID); err != nil {
		t.Fatalf("delete story failed: %v", err)
	}
	if _, err := repo.GetByID(ctx, storyID); err == nil {
		t.Fatal("expected deleted story not found")
	}
}

func TestInspiringStoryRepository_Delete_MidTransactionFailureBranch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	stmts := []string{
		`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, comment_count INTEGER DEFAULT 0)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT)`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER)`,
		`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT)`,
		// intentionally do not create story_comment_hearts to force mid-step failure
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema setup failed: %v", err)
		}
	}

	storyID := uuid.New()
	if err := db.Exec(`INSERT INTO inspiring_stories (id, comment_count) VALUES (?, 1)`, storyID.String()).Error; err != nil {
		t.Fatalf("seed story failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content) VALUES (?, ?, 1, 'x')`, uuid.New().String(), storyID.String()).Error; err != nil {
		t.Fatalf("seed story comment failed: %v", err)
	}

	repo := NewInspiringStoryRepository(db)
	if err := repo.Delete(context.Background(), storyID); err == nil {
		t.Fatal("expected delete to fail when story_comment_hearts table is missing")
	}
}

func TestInspiringStoryRepository_CreateCommentAndDelete_AdditionalErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("create comment fails on story counter update", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		if err := db.Exec(`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, created_at DATETIME, updated_at DATETIME)`).Error; err != nil {
			t.Fatalf("create story_comments failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		comment := &model.StoryComment{ID: uuid.New(), StoryID: uuid.New(), UserID: 1, Content: "x"}
		if err := repo.CreateComment(ctx, comment); err == nil {
			t.Fatal("expected CreateComment to fail when inspiring_stories table is missing")
		}
	})

	t.Run("delete fails at final story deletion step", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		stmts := []string{
			`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT)`,
			`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT)`,
			`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER)`,
			`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT)`,
			`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER)`,
		}
		for _, stmt := range stmts {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatalf("schema setup failed: %v", err)
			}
		}

		storyID := uuid.New()
		commentID := uuid.New()
		heartID := uuid.New()
		if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content) VALUES (?, ?, 1, 'x')`, commentID.String(), storyID.String()).Error; err != nil {
			t.Fatalf("seed comment failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_comment_hearts (id, comment_id, user_id) VALUES (?, ?, 1)`, heartID.String(), commentID.String()).Error; err != nil {
			t.Fatalf("seed comment heart failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.Delete(ctx, storyID); err == nil {
			t.Fatal("expected Delete to fail at final step when inspiring_stories table is missing")
		}
	})
}
