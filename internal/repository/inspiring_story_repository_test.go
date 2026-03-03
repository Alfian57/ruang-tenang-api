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

	if _, err := repo.GetAuthorStats(ctx, 1); err == nil {
		t.Fatal("expected GetAuthorStats error on empty db")
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
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, username TEXT, email TEXT, password TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, title TEXT, content TEXT, cover_image TEXT, is_anonymous BOOLEAN, has_trigger_warning BOOLEAN, trigger_warning_text TEXT, status TEXT, moderator_id INTEGER, moderation_feedback TEXT, moderated_at DATETIME, view_count INTEGER DEFAULT 0, heart_count INTEGER DEFAULT 0, comment_count INTEGER DEFAULT 0, is_featured BOOLEAN DEFAULT 0, featured_at DATETIME, featured_by INTEGER, featured_until DATETIME, created_at DATETIME, updated_at DATETIME, published_at DATETIME)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, inspiring_story_id TEXT, story_category_id TEXT, PRIMARY KEY (story_id, category_id))`,
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

func TestInspiringStoryRepository_GetAuthorStats_Branches(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		schema := []string{
			`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, status TEXT, view_count INTEGER, created_at DATETIME)`,
			`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
			`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, created_at DATETIME)`,
		}
		for _, stmt := range schema {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatalf("schema setup failed: %v", err)
			}
		}

		now := time.Now()
		if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, status, view_count, created_at) VALUES
			('s1', 1, 'approved', 10, ?),
			('s2', 1, 'pending', 5, ?),
			('s3', 2, 'approved', 20, ?)`, now, now, now).Error; err != nil {
			t.Fatalf("seed stories: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_hearts (id, story_id, user_id, created_at) VALUES ('h1', 's1', 2, ?), ('h2', 's1', 3, ?)`, now, now).Error; err != nil {
			t.Fatalf("seed hearts: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, created_at) VALUES ('c1', 's1', 2, 'nice', ?), ('c2', 's2', 3, 'keep going', ?)`, now, now).Error; err != nil {
			t.Fatalf("seed comments: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		stats, err := repo.GetAuthorStats(ctx, 1)
		if err != nil {
			t.Fatalf("GetAuthorStats failed: %v", err)
		}
		if stats.TotalStories != 2 || stats.ApprovedStories != 1 || stats.PendingStories != 1 {
			t.Fatalf("unexpected story counts: %+v", stats)
		}
		if stats.TotalHearts != 2 || stats.TotalComments != 2 || stats.TotalViews != 15 {
			t.Fatalf("unexpected engagement counts: %+v", stats)
		}
	})

	t.Run("hearts query error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := db.Exec(`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, status TEXT, view_count INTEGER, created_at DATETIME)`).Error; err != nil {
			t.Fatalf("create inspiring_stories: %v", err)
		}
		repo := NewInspiringStoryRepository(db)
		if _, err := repo.GetAuthorStats(ctx, 1); err == nil {
			t.Fatal("expected hearts query error when story_hearts table missing")
		}
	})

	t.Run("approved query error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		if err := db.Exec(`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, view_count INTEGER, created_at DATETIME)`).Error; err != nil {
			t.Fatalf("create inspiring_stories without status: %v", err)
		}
		repo := NewInspiringStoryRepository(db)
		if _, err := repo.GetAuthorStats(ctx, 1); err == nil {
			t.Fatal("expected approved query error when status column missing")
		}
	})

	t.Run("comments query error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		schema := []string{
			`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, status TEXT, view_count INTEGER, created_at DATETIME)`,
			`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
		}
		for _, stmt := range schema {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatalf("schema setup failed: %v", err)
			}
		}
		repo := NewInspiringStoryRepository(db)
		if _, err := repo.GetAuthorStats(ctx, 1); err == nil {
			t.Fatal("expected comments query error when story_comments table missing")
		}
	})

	t.Run("views scan error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		schema := []string{
			`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, status TEXT, created_at DATETIME)`,
			`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
			`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, created_at DATETIME)`,
		}
		for _, stmt := range schema {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatalf("schema setup failed: %v", err)
			}
		}
		repo := NewInspiringStoryRepository(db)
		if _, err := repo.GetAuthorStats(ctx, 1); err == nil {
			t.Fatal("expected views scan error when view_count column missing")
		}
	})
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

func TestInspiringStoryRepository_AdditionalSuccessAndNotFoundBranches(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	stmts := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, username TEXT, email TEXT, password TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, title TEXT, content TEXT, cover_image TEXT, is_anonymous BOOLEAN, has_trigger_warning BOOLEAN, trigger_warning_text TEXT, status TEXT, moderator_id INTEGER, moderation_feedback TEXT, moderated_at DATETIME, view_count INTEGER DEFAULT 0, heart_count INTEGER DEFAULT 0, comment_count INTEGER DEFAULT 0, is_featured BOOLEAN DEFAULT 0, featured_at DATETIME, featured_by INTEGER, featured_until DATETIME, created_at DATETIME, updated_at DATETIME, published_at DATETIME)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, inspiring_story_id TEXT, story_category_id TEXT, PRIMARY KEY (story_id, category_id))`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, heart_count INTEGER DEFAULT 0, is_hidden BOOLEAN DEFAULT 0, hidden_reason TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER, created_at DATETIME)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema setup failed: %v", err)
		}
	}

	repo := NewInspiringStoryRepository(db)
	ctx := context.Background()
	now := time.Now()

	if err := db.Exec(`INSERT INTO users (id, name, username, email, password, created_at, updated_at) VALUES (1, 'Author', 'author', 'author@test.local', 'x', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	storyID := uuid.New()
	storyID2 := uuid.New()
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES (?, 1, 'A', 'content a', 'approved', 2, 3, 1, 1, ?, ?, ?), (?, 1, 'B', 'content b', 'approved', 10, 1, 0, 0, ?, ?, ?)`, storyID.String(), now, now, now, storyID2.String(), now, now, now).Error; err != nil {
		t.Fatalf("seed stories failed: %v", err)
	}

	if _, _, err := repo.GetApprovedStories(ctx, 1, 10, "hearts"); err != nil {
		t.Fatalf("GetApprovedStories hearts sort failed: %v", err)
	}
	if _, _, err := repo.GetApprovedStories(ctx, 1, 10, "featured"); err != nil {
		t.Fatalf("GetApprovedStories featured sort failed: %v", err)
	}
	if _, _, err := repo.GetApprovedStories(ctx, 1, 10, "views"); err != nil {
		t.Fatalf("GetApprovedStories views sort failed: %v", err)
	}

	catID := uuid.New()
	if err := db.Exec(`INSERT INTO story_categories (id, name, slug, is_active, created_at, updated_at) VALUES (?, 'Hope', 'hope', 1, ?, ?)`, catID.String(), now, now).Error; err != nil {
		t.Fatalf("seed category failed: %v", err)
	}
	if gotCategory, err := repo.GetCategoryByID(ctx, catID); err != nil || gotCategory == nil || gotCategory.ID != catID {
		t.Fatalf("GetCategoryByID success branch failed: err=%v category=%+v", err, gotCategory)
	}
	if gotCategory, err := repo.GetCategoryBySlug(ctx, "hope"); err != nil || gotCategory == nil || gotCategory.ID != catID {
		t.Fatalf("GetCategoryBySlug success branch failed: err=%v category=%+v", err, gotCategory)
	}
	if err := db.Exec(`INSERT INTO story_category_relations (story_id, category_id) VALUES (?, ?)`, storyID.String(), catID.String()).Error; err != nil {
		t.Fatalf("seed category relation failed: %v", err)
	}
	if gotStory, err := repo.GetByID(ctx, storyID); err != nil || gotStory == nil || gotStory.ID != storyID {
		t.Fatalf("GetByID success branch failed: err=%v story=%+v", err, gotStory)
	}
	if gotStory, err := repo.GetByIDWithRelations(ctx, storyID); err != nil || gotStory == nil || gotStory.ID != storyID {
		t.Fatalf("GetByIDWithRelations success branch failed: err=%v story=%+v", err, gotStory)
	}
	if _, _, err := repo.GetStoriesByAuthor(ctx, 1, "approved", 1, 10); err != nil {
		t.Fatalf("GetStoriesByAuthor status branch failed: %v", err)
	}

	if err := repo.SetStoryCategories(ctx, storyID, []uuid.UUID{}); err != nil {
		t.Fatalf("SetStoryCategories empty list failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_hearts (id, story_id, user_id, created_at) VALUES (?, ?, 2, ?)`, uuid.New().String(), storyID.String(), now).Error; err != nil {
		t.Fatalf("seed remove-heart success row failed: %v", err)
	}
	if err := repo.RemoveHeart(ctx, storyID, 2); err != nil {
		t.Fatalf("RemoveHeart success branch failed: %v", err)
	}

	if err := repo.RemoveHeart(ctx, storyID, 9999); err == nil {
		t.Fatal("expected RemoveHeart not found error")
	}

	commentID := uuid.New()
	if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, heart_count, is_hidden, created_at, updated_at) VALUES (?, ?, 1, 'comment', 0, 0, ?, ?)`, commentID.String(), storyID.String(), now, now).Error; err != nil {
		t.Fatalf("seed comment failed: %v", err)
	}
	if gotComment, err := repo.GetCommentByID(ctx, commentID); err != nil || gotComment == nil || gotComment.ID != commentID {
		t.Fatalf("GetCommentByID success branch failed: err=%v comment=%+v", err, gotComment)
	}
	if err := db.Exec(`INSERT INTO story_comment_hearts (id, comment_id, user_id, created_at) VALUES (?, ?, 2, ?)`, uuid.New().String(), commentID.String(), now).Error; err != nil {
		t.Fatalf("seed remove-comment-heart success row failed: %v", err)
	}
	if err := repo.RemoveCommentHeart(ctx, commentID, 2); err != nil {
		t.Fatalf("RemoveCommentHeart success branch failed: %v", err)
	}
	if err := repo.RemoveCommentHeart(ctx, commentID, 9999); err == nil {
		t.Fatal("expected RemoveCommentHeart not found error")
	}
}

func TestInspiringStoryRepository_NotFoundAndUpdateStepFailureBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("getters return record not found on existing schema", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		stmts := []string{
			`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
			`CREATE TABLE users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, username TEXT, email TEXT, password TEXT, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
			`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, author_id INTEGER, title TEXT, content TEXT, status TEXT, created_at DATETIME, updated_at DATETIME, published_at DATETIME)`,
			`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, inspiring_story_id TEXT, story_category_id TEXT)`,
			`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		}
		for _, stmt := range stmts {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatalf("schema setup failed: %v", err)
			}
		}

		repo := NewInspiringStoryRepository(db)
		if _, err := repo.GetCategoryByID(ctx, uuid.New()); err == nil {
			t.Fatal("expected GetCategoryByID not found error")
		}
		if _, err := repo.GetCategoryBySlug(ctx, "missing"); err == nil {
			t.Fatal("expected GetCategoryBySlug not found error")
		}
		if _, err := repo.GetByID(ctx, uuid.New()); err == nil {
			t.Fatal("expected GetByID not found error")
		}
		if _, err := repo.GetByIDWithRelations(ctx, uuid.New()); err == nil {
			t.Fatal("expected GetByIDWithRelations not found error")
		}
		if _, err := repo.GetCommentByID(ctx, uuid.New()); err == nil {
			t.Fatal("expected GetCommentByID not found error")
		}
	})

	t.Run("add heart fails on counter update step", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		if err := db.Exec(`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`).Error; err != nil {
			t.Fatalf("create story_hearts failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.AddHeart(ctx, uuid.New(), 1); err == nil {
			t.Fatal("expected AddHeart error when inspiring_stories table missing")
		}
	})

	t.Run("add comment heart fails on counter update step", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		if err := db.Exec(`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER, created_at DATETIME)`).Error; err != nil {
			t.Fatalf("create story_comment_hearts failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.AddCommentHeart(ctx, uuid.New(), 1); err == nil {
			t.Fatal("expected AddCommentHeart error when story_comments table missing")
		}
	})
}

func TestInspiringStoryRepository_CreateComment_SuccessPath(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	stmts := []string{
		`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, comment_count INTEGER DEFAULT 0)`,
		`CREATE TABLE story_comments (
			id TEXT PRIMARY KEY,
			story_id TEXT,
			user_id INTEGER,
			content TEXT,
			heart_count INTEGER DEFAULT 0,
			is_hidden BOOLEAN DEFAULT 0,
			hidden_reason TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema setup failed: %v", err)
		}
	}

	storyID := uuid.New()
	if err := db.Exec(`INSERT INTO inspiring_stories (id, comment_count) VALUES (?, 0)`, storyID.String()).Error; err != nil {
		t.Fatalf("seed story failed: %v", err)
	}

	repo := NewInspiringStoryRepository(db)
	comment := &model.StoryComment{
		ID:      uuid.New(),
		StoryID: storyID,
		UserID:  1,
		Content: "semangat terus",
	}
	if err := repo.CreateComment(context.Background(), comment); err != nil {
		t.Fatalf("CreateComment success path failed: %v", err)
	}

	var commentCount int64
	if err := db.Raw(`SELECT comment_count FROM inspiring_stories WHERE id = ?`, storyID.String()).Scan(&commentCount).Error; err != nil {
		t.Fatalf("query comment_count failed: %v", err)
	}
	if commentCount != 1 {
		t.Fatalf("expected comment_count=1, got %d", commentCount)
	}
}

func TestInspiringStoryRepository_Delete_SpecificStepFailureBranches(t *testing.T) {
	ctx := context.Background()

	setupDeleteSchema := func(t *testing.T) (*gorm.DB, uuid.UUID) {
		t.Helper()
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
			`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER)`,
		}
		for _, stmt := range stmts {
			if err := db.Exec(stmt).Error; err != nil {
				t.Fatalf("schema setup failed: %v", err)
			}
		}

		storyID := uuid.New()
		commentID := uuid.New()
		if err := db.Exec(`INSERT INTO inspiring_stories (id, comment_count) VALUES (?, 1)`, storyID.String()).Error; err != nil {
			t.Fatalf("seed story failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_category_relations (story_id, category_id) VALUES (?, ?)`, storyID.String(), uuid.New().String()).Error; err != nil {
			t.Fatalf("seed relation failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_tags (id, story_id, tag) VALUES (?, ?, 'tag1')`, uuid.New().String(), storyID.String()).Error; err != nil {
			t.Fatalf("seed tag failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_hearts (id, story_id, user_id) VALUES (?, ?, 1)`, uuid.New().String(), storyID.String()).Error; err != nil {
			t.Fatalf("seed heart failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content) VALUES (?, ?, 1, 'x')`, commentID.String(), storyID.String()).Error; err != nil {
			t.Fatalf("seed comment failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_comment_hearts (id, comment_id, user_id) VALUES (?, ?, 1)`, uuid.New().String(), commentID.String()).Error; err != nil {
			t.Fatalf("seed comment heart failed: %v", err)
		}

		return db, storyID
	}

	t.Run("fails at category relation deletion step", func(t *testing.T) {
		db, storyID := setupDeleteSchema(t)
		if err := db.Exec(`CREATE TRIGGER fail_delete_story_category_relations
			BEFORE DELETE ON story_category_relations
			BEGIN
				SELECT RAISE(FAIL, 'fail delete category relation');
			END`).Error; err != nil {
			t.Fatalf("create trigger failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.Delete(ctx, storyID); err == nil {
			t.Fatal("expected Delete error at category relation deletion step")
		}
	})

	t.Run("fails at story hearts deletion step", func(t *testing.T) {
		db, storyID := setupDeleteSchema(t)
		if err := db.Exec(`CREATE TRIGGER fail_delete_story_hearts
			BEFORE DELETE ON story_hearts
			BEGIN
				SELECT RAISE(FAIL, 'fail delete story hearts');
			END`).Error; err != nil {
			t.Fatalf("create trigger failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.Delete(ctx, storyID); err == nil {
			t.Fatal("expected Delete error at story hearts deletion step")
		}
	})

	t.Run("fails at comment hearts cleanup step", func(t *testing.T) {
		db, storyID := setupDeleteSchema(t)
		if err := db.Exec(`DROP TABLE story_comment_hearts`).Error; err != nil {
			t.Fatalf("drop story_comment_hearts failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.Delete(ctx, storyID); err == nil {
			t.Fatal("expected Delete error at comment hearts cleanup step")
		}
	})

	t.Run("fails at final story deletion step", func(t *testing.T) {
		db, storyID := setupDeleteSchema(t)
		if err := db.Exec(`CREATE TRIGGER fail_delete_inspiring_stories
			BEFORE DELETE ON inspiring_stories
			BEGIN
				SELECT RAISE(FAIL, 'fail delete story');
			END`).Error; err != nil {
			t.Fatalf("create trigger failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.Delete(ctx, storyID); err == nil {
			t.Fatal("expected Delete error at final story deletion step")
		}
	})
}

func TestInspiringStoryRepository_DeleteComment_SpecificStepFailureBranches(t *testing.T) {
	ctx := context.Background()

	setupDeleteCommentSchema := func(t *testing.T) (*gorm.DB, uuid.UUID, uuid.UUID) {
		t.Helper()
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}

		stmts := []string{
			`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY, comment_count INTEGER DEFAULT 1)`,
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
		if err := db.Exec(`INSERT INTO inspiring_stories (id, comment_count) VALUES (?, 1)`, storyID.String()).Error; err != nil {
			t.Fatalf("seed story failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content) VALUES (?, ?, 1, 'x')`, commentID.String(), storyID.String()).Error; err != nil {
			t.Fatalf("seed comment failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO story_comment_hearts (id, comment_id, user_id) VALUES (?, ?, 1)`, uuid.New().String(), commentID.String()).Error; err != nil {
			t.Fatalf("seed comment heart failed: %v", err)
		}

		return db, storyID, commentID
	}

	t.Run("fails at comment hearts deletion step", func(t *testing.T) {
		db, storyID, commentID := setupDeleteCommentSchema(t)
		if err := db.Exec(`CREATE TRIGGER fail_delete_story_comment_hearts
			BEFORE DELETE ON story_comment_hearts
			BEGIN
				SELECT RAISE(FAIL, 'fail delete comment hearts');
			END`).Error; err != nil {
			t.Fatalf("create trigger failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.DeleteComment(ctx, commentID, storyID); err == nil {
			t.Fatal("expected DeleteComment error at comment hearts deletion step")
		}
	})

	t.Run("fails at comment deletion step", func(t *testing.T) {
		db, storyID, commentID := setupDeleteCommentSchema(t)
		if err := db.Exec(`CREATE TRIGGER fail_delete_story_comments
			BEFORE DELETE ON story_comments
			BEGIN
				SELECT RAISE(FAIL, 'fail delete comments');
			END`).Error; err != nil {
			t.Fatalf("create trigger failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.DeleteComment(ctx, commentID, storyID); err == nil {
			t.Fatal("expected DeleteComment error at comment deletion step")
		}
	})

	t.Run("fails at story comment_count update step", func(t *testing.T) {
		db, storyID, commentID := setupDeleteCommentSchema(t)
		if err := db.Exec(`ALTER TABLE inspiring_stories RENAME TO inspiring_stories_old`).Error; err != nil {
			t.Fatalf("rename inspiring_stories failed: %v", err)
		}
		if err := db.Exec(`CREATE TABLE inspiring_stories (id TEXT PRIMARY KEY)`).Error; err != nil {
			t.Fatalf("create inspiring_stories without comment_count failed: %v", err)
		}
		if err := db.Exec(`INSERT INTO inspiring_stories (id) SELECT id FROM inspiring_stories_old`).Error; err != nil {
			t.Fatalf("copy inspiring_stories rows failed: %v", err)
		}

		repo := NewInspiringStoryRepository(db)
		if err := repo.DeleteComment(ctx, commentID, storyID); err == nil {
			t.Fatal("expected DeleteComment error at story update step")
		}
	})
}
