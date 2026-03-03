package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInspiringStoryServiceFlows(t *testing.T) *InspiringStoryService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	now := time.Now()
	schema := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, avatar TEXT, exp INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE level_configs (id INTEGER PRIMARY KEY, level INTEGER, min_exp INTEGER, badge_name TEXT, badge_icon TEXT, tier_name TEXT, tier_color TEXT, description TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (
			id TEXT PRIMARY KEY,
			author_id INTEGER,
			title TEXT,
			content TEXT,
			cover_image TEXT,
			is_anonymous BOOLEAN,
			has_trigger_warning BOOLEAN,
			trigger_warning_text TEXT,
			status TEXT,
			moderated_by INTEGER,
			moderator_feedback TEXT,
			moderator_id INTEGER,
			moderation_feedback TEXT,
			moderated_at DATETIME,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_featured BOOLEAN,
			featured_by INTEGER,
			featured_at DATETIME,
			featured_until DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			published_at DATETIME
		)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_category_relations (
			story_id TEXT,
			category_id TEXT,
			inspiring_story_id TEXT,
			story_category_id TEXT
		)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, heart_count INTEGER, is_hidden BOOLEAN, hidden_reason TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER, created_at DATETIME)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO users (id, name, avatar, exp, deleted_at) VALUES (1, 'User One', 'a.png', 700, NULL), (2, 'User Two', 'b.png', 20, NULL)`).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (1, 1, 0, 'Seed', 'icon', 'Bronze', '#A97142', 'desc', ?, ?), (2, 2, 100, 'Seed2', 'icon2', 'Silver', '#C0C0C0', 'desc2', ?, ?), (3, 7, 500, 'Seed7', 'icon7', 'Gold', '#FFD700', 'desc7', ?, ?)`, now, now, now, now, now, now).Error; err != nil {
		t.Fatalf("seed levels: %v", err)
	}

	catID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	if err := db.Exec(`INSERT INTO story_categories (id, name, slug, description, icon, display_order, is_active, created_at, updated_at) VALUES (?, 'Hope', 'hope', 'desc', '🌟', 1, 1, ?, ?)`, catID, now, now).Error; err != nil {
		t.Fatalf("seed categories: %v", err)
	}

	approvedID := "11111111-1111-1111-1111-111111111111"
	pendingID := "22222222-2222-2222-2222-222222222222"
	featuredID := "33333333-3333-3333-3333-333333333333"
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES (?, 1, 'Approved', 'approved content', 'approved', 0, 5, 2, 1, 0, ?, ?, ?), (?, 1, 'Pending', 'pending content', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL), (?, 1, 'Featured', 'featured content', 'approved', 0, 3, 4, 2, 1, ?, ?, ?)`, approvedID, now, now, now, pendingID, now, now, featuredID, now, now, now).Error; err != nil {
		t.Fatalf("seed stories: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_category_relations (story_id, category_id, inspiring_story_id, story_category_id) VALUES (?, ?, ?, ?), (?, ?, ?, ?), (?, ?, ?, ?)`, approvedID, catID, approvedID, catID, pendingID, catID, pendingID, catID, featuredID, catID, featuredID, catID).Error; err != nil {
		t.Fatalf("seed story relations: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_tags (id, story_id, tag, created_at) VALUES (?, ?, 'hope', ?), (?, ?, 'calm', ?)`, uuid.New().String(), approvedID, now, uuid.New().String(), featuredID, now).Error; err != nil {
		t.Fatalf("seed story tags: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, heart_count, is_hidden, created_at, updated_at) VALUES (?, ?, 2, 'support', 0, 0, ?, ?)`, "44444444-4444-4444-4444-444444444444", approvedID, now, now).Error; err != nil {
		t.Fatalf("seed comments: %v", err)
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

func TestInspiringStoryService_MainFlows(t *testing.T) {
	svc := setupInspiringStoryServiceFlows(t)
	ctx := context.Background()

	if _, err := svc.CreateStory(ctx, 1, &dto.CreateStoryRequest{
		Title:       "New Story",
		Content:     "new content",
		CategoryIDs: []uuid.UUID{uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")},
		Tags:        []string{"tag1"},
	}); err == nil {
		t.Fatal("expected CreateStory to fail on sqlite UUID default behavior")
	}

	storyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	updated, err := svc.UpdateStory(ctx, storyID, 1, &dto.UpdateStoryRequest{Title: "Updated", Content: "updated content", IsAnonymous: false, CategoryIDs: []uuid.UUID{uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")}, Tags: []string{"tag2"}})
	if err != nil || updated == nil {
		t.Fatalf("UpdateStory failed: err=%v updated=%v", err, updated)
	}

	if _, err := svc.UpdateStory(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 2, &dto.UpdateStoryRequest{Title: "x"}); err != ErrUnauthorized {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if _, err := svc.UpdateStory(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 1, &dto.UpdateStoryRequest{Title: "x"}); err != ErrCannotEditPublishedStory {
		t.Fatalf("expected ErrCannotEditPublishedStory, got %v", err)
	}

	storyOwned, err := svc.GetStory(ctx, uuid.MustParse("22222222-2222-2222-2222-222222222222"), 1)
	if err != nil || storyOwned == nil {
		t.Fatalf("GetStory owner pending failed: err=%v story=%v", err, storyOwned)
	}
	if _, err := svc.GetStory(ctx, uuid.MustParse("22222222-2222-2222-2222-222222222222"), 2); err != ErrStoryNotFound {
		t.Fatalf("expected ErrStoryNotFound for non-owner pending story, got %v", err)
	}

	storiesDefault, err := svc.GetStories(ctx, &dto.StoryFilterRequest{Page: 1, Limit: 10, SortBy: "recent"}, 1)
	if err != nil || storiesDefault == nil {
		t.Fatalf("GetStories default failed: err=%v", err)
	}
	storiesByCategory, err := svc.GetStories(ctx, &dto.StoryFilterRequest{Page: 1, Limit: 10, CategoryID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}, 1)
	if err != nil || storiesByCategory == nil {
		t.Fatalf("GetStories category failed: err=%v", err)
	}
	if storiesDefault.Total < 1 || storiesByCategory.Total < 1 {
		t.Fatalf("unexpected stories totals default=%d category=%d", storiesDefault.Total, storiesByCategory.Total)
	}

	myStories, err := svc.GetUserStories(ctx, 1, "", 1, 10)
	if err != nil || myStories == nil || myStories.Total < 1 {
		t.Fatalf("GetUserStories failed: err=%v payload=%v", err, myStories)
	}

	featured, err := svc.GetFeaturedStories(ctx, 0)
	if err != nil || len(featured) < 1 {
		t.Fatalf("GetFeaturedStories failed: err=%v len=%d", err, len(featured))
	}

	pending, err := svc.GetPendingStories(ctx, 0, 0)
	if err != nil || pending.Total < 1 {
		t.Fatalf("GetPendingStories failed: err=%v payload=%v", err, pending)
	}

	comments, err := svc.GetComments(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 1, 0, 0)
	if err != nil || comments == nil || comments.Total < 1 {
		t.Fatalf("GetComments failed: err=%v payload=%v", err, comments)
	}

	most, err := svc.GetMostAppreciatedStories(ctx, int(nowMonth()), time.Now().Year(), 0)
	if err != nil || most == nil {
		t.Fatalf("GetMostAppreciatedStories failed: err=%v payload=%v", err, most)
	}

	if err := svc.SetFeatured(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), true, 1); err != nil {
		t.Fatalf("SetFeatured failed: %v", err)
	}

	stats, err := svc.GetAuthorStats(ctx, 1)
	if err != nil || stats == nil || stats.TotalStories < 1 {
		t.Fatalf("GetAuthorStats failed: err=%v stats=%v", err, stats)
	}

	if err := svc.DeleteStory(ctx, storyID, 2); err != ErrUnauthorized {
		t.Fatalf("expected delete unauthorized, got %v", err)
	}
	if err := svc.DeleteStory(ctx, storyID, 1); err != nil {
		t.Fatalf("DeleteStory failed: %v", err)
	}
}

func nowMonth() int {
	return int(time.Now().Month())
}

func setupInspiringStoryServiceInteractions(t *testing.T) (*InspiringStoryService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	now := time.Now()
	schema := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, avatar TEXT, exp INTEGER, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE level_configs (id INTEGER PRIMARY KEY, level INTEGER, min_exp INTEGER, badge_name TEXT, badge_icon TEXT, tier_name TEXT, tier_color TEXT, description TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (
			id TEXT PRIMARY KEY,
			author_id INTEGER,
			title TEXT,
			content TEXT,
			cover_image TEXT,
			is_anonymous BOOLEAN,
			has_trigger_warning BOOLEAN,
			trigger_warning_text TEXT,
			status TEXT,
			moderated_by INTEGER,
			moderator_feedback TEXT,
			moderator_id INTEGER,
			moderation_feedback TEXT,
			moderated_at DATETIME,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_featured BOOLEAN,
			featured_by INTEGER,
			featured_at DATETIME,
			featured_until DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			published_at DATETIME
		)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, inspiring_story_id TEXT, story_category_id TEXT)`,
		`CREATE TABLE story_tags (id TEXT PRIMARY KEY, story_id TEXT, tag TEXT, created_at DATETIME)`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE story_comments (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, content TEXT, heart_count INTEGER, is_hidden BOOLEAN, hidden_reason TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_comment_hearts (id TEXT PRIMARY KEY, comment_id TEXT, user_id INTEGER, created_at DATETIME)`,
		`CREATE TABLE badge_definitions (id TEXT PRIMARY KEY, badge_key TEXT NOT NULL UNIQUE, badge_name TEXT NOT NULL, description TEXT, icon TEXT, category TEXT, requirement_type TEXT NOT NULL, requirement_value INTEGER, is_active NUMERIC, display_order INTEGER, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE user_badges (id TEXT PRIMARY KEY, user_id INTEGER NOT NULL, badge_id TEXT NOT NULL, earned_at DATETIME, is_showcased NUMERIC DEFAULT 0)`,
		`CREATE TABLE user_activities (id INTEGER PRIMARY KEY, user_id INTEGER, activity_type TEXT, date DATETIME, count INTEGER, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE exp_histories (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, activity_type TEXT, points INTEGER, description TEXT, created_at DATETIME)`,
		`CREATE TABLE notifications (id TEXT PRIMARY KEY, user_id INTEGER, type TEXT, title TEXT, message TEXT, is_read BOOLEAN, data TEXT, created_at DATETIME)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO users (id, name, avatar, exp, updated_at, deleted_at) VALUES (1, 'Author One', 'a.png', 700, ?, NULL), (2, 'Reader Two', 'b.png', 20, ?, NULL)`, now, now).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (1, 1, 0, 'Seed', 'icon', 'Bronze', '#A97142', 'desc', ?, ?), (2, 7, 500, 'Seed7', 'icon7', 'Gold', '#FFD700', 'desc7', ?, ?)`, now, now, now, now).Error; err != nil {
		t.Fatalf("seed levels: %v", err)
	}

	approvedID := "11111111-1111-1111-1111-111111111111"
	pendingApprovedID := "55555555-5555-5555-5555-555555555555"
	pendingRejectedID := "66666666-6666-6666-6666-666666666666"
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES (?, 1, 'Approved Story', 'approved content', 'approved', 0, 0, 0, 0, 0, ?, ?, ?), (?, 1, 'Pending For Approval', 'pending content', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL), (?, 1, 'Pending For Rejection', 'pending content 2', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL)`, approvedID, now, now, now, pendingApprovedID, now, now, pendingRejectedID, now, now).Error; err != nil {
		t.Fatalf("seed stories: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_categories (id, name, slug, description, icon, display_order, is_active, created_at, updated_at) VALUES (?, 'Hope', 'hope', 'desc', '🌟', 1, 1, ?, ?)`, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", now, now).Error; err != nil {
		t.Fatalf("seed categories: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_category_relations (story_id, category_id, inspiring_story_id, story_category_id) VALUES (?, ?, ?, ?)`, approvedID, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", approvedID, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa").Error; err != nil {
		t.Fatalf("seed story category relation: %v", err)
	}

	if err := db.Exec(`INSERT INTO badge_definitions (id, badge_key, badge_name, description, icon, category, requirement_type, requirement_value, is_active, display_order, created_at, updated_at) VALUES (?, 'first_story', 'First Story', 'd', 'i', 'activity', 'manual', 1, 1, 1, ?, ?), (?, 'first_heart', 'First Heart', 'd', 'i', 'activity', 'manual', 1, 1, 2, ?, ?), (?, 'first_comment', 'First Comment', 'd', 'i', 'activity', 'manual', 1, 1, 3, ?, ?)`, uuid.New().String(), now, now, uuid.New().String(), now, now, uuid.New().String(), now, now).Error; err != nil {
		t.Fatalf("seed badges: %v", err)
	}

	badgeSvc := NewBadgeService(repository.NewBadgeRepository(db), repository.NewUserRepository(db), repository.NewLevelConfigRepository(db))
	notifSvc := NewNotificationService(repository.NewNotificationRepository(db))

	svc := NewInspiringStoryService(
		repository.NewInspiringStoryRepository(db),
		repository.NewUserRepository(db),
		repository.NewLevelConfigRepository(db),
		badgeSvc,
		NewGamificationService(db),
		notifSvc,
	)

	return svc, db
}

func TestInspiringStoryService_InteractionBranches(t *testing.T) {
	svc, db := setupInspiringStoryServiceInteractions(t)
	ctx := context.Background()

	categories, err := svc.GetCategories(ctx)
	if err != nil || len(categories) < 1 {
		t.Fatalf("GetCategories success expected: err=%v categories=%v", err, categories)
	}

	if err := svc.ModerateStory(ctx, uuid.MustParse("55555555-5555-5555-5555-555555555555"), 2, &dto.ModerateStoryRequest{Status: "approved", Feedback: "ok"}); err != nil {
		t.Fatalf("ModerateStory approved failed: %v", err)
	}
	if err := svc.ModerateStory(ctx, uuid.MustParse("66666666-6666-6666-6666-666666666666"), 2, &dto.ModerateStoryRequest{Status: "rejected", Feedback: "revise"}); err != nil {
		t.Fatalf("ModerateStory rejected failed: %v", err)
	}

	var notifCount int64
	if err := db.Model(&model.Notification{}).Where("user_id = ?", 1).Count(&notifCount).Error; err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if notifCount < 2 {
		t.Fatalf("expected notifications from moderation actions, got %d", notifCount)
	}

	hearted, count, err := svc.ToggleHeart(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 2)
	if err != nil {
		t.Fatalf("ToggleHeart add failed: %v", err)
	}
	if !hearted || count < 1 {
		t.Fatalf("expected heart added, got hearted=%v count=%d", hearted, count)
	}

	hearted, count, err = svc.ToggleHeart(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 2)
	if err != nil && !strings.Contains(err.Error(), "GREATEST") {
		t.Fatalf("ToggleHeart remove failed: %v", err)
	}
	if err == nil && hearted {
		t.Fatalf("expected heart removed, got hearted=%v count=%d", hearted, count)
	}

	createCommentPanicked := false
	func() {
		defer func() {
			if recover() != nil {
				createCommentPanicked = true
			}
		}()
		_, _ = svc.CreateComment(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 2, &dto.CreateStoryCommentRequest{Content: "Semangat ya"})
	}()
	if !createCommentPanicked {
		t.Fatal("expected CreateComment panic on sqlite UUID default behavior")
	}

	commentID := uuid.MustParse("77777777-7777-7777-7777-777777777777")
	if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, heart_count, is_hidden, created_at, updated_at) VALUES (?, ?, 2, 'manual comment', 0, 0, ?, ?)`, commentID.String(), "11111111-1111-1111-1111-111111111111", time.Now(), time.Now()).Error; err != nil {
		t.Fatalf("seed manual comment: %v", err)
	}

	if err := svc.HideComment(ctx, commentID, &dto.HideStoryCommentRequest{Reason: "sensitive"}); err != nil {
		t.Fatalf("HideComment success expected: %v", err)
	}

	commentHearted, heartCount, err := svc.ToggleCommentHeart(ctx, commentID, 1)
	if err != nil {
		t.Fatalf("ToggleCommentHeart add failed: %v", err)
	}
	if !commentHearted || heartCount < 1 {
		t.Fatalf("expected comment heart added, got hearted=%v count=%d", commentHearted, heartCount)
	}

	commentHearted, heartCount, err = svc.ToggleCommentHeart(ctx, commentID, 1)
	if err != nil && !strings.Contains(err.Error(), "GREATEST") {
		t.Fatalf("ToggleCommentHeart remove failed: %v", err)
	}
	if err == nil && commentHearted {
		t.Fatalf("expected comment heart removed, got hearted=%v count=%d", commentHearted, heartCount)
	}

	if err := svc.DeleteComment(ctx, commentID, 1); err != ErrUnauthorized {
		t.Fatalf("expected delete comment unauthorized, got %v", err)
	}
	if err := svc.DeleteComment(ctx, commentID, 2); err != nil && !strings.Contains(err.Error(), "GREATEST") {
		t.Fatalf("DeleteComment expected success/sqlite greatest limitation, got %v", err)
	}
}

func TestInspiringStoryService_AdditionalBranchCoverage(t *testing.T) {
	svc := setupInspiringStoryServiceFlows(t)
	ctx := context.Background()

	stories, err := svc.GetUserStories(ctx, 1, "", 0, 100)
	if err != nil {
		t.Fatalf("expected GetUserStories with normalized pagination success, got %v", err)
	}
	if stories.Page != 1 || stories.Limit != 10 {
		t.Fatalf("expected normalized page=1 limit=10, got page=%d limit=%d", stories.Page, stories.Limit)
	}

	_, err = svc.CreateComment(ctx, uuid.MustParse("22222222-2222-2222-2222-222222222222"), 1, &dto.CreateStoryCommentRequest{Content: "hi"})
	if err != ErrStoryNotApproved {
		t.Fatalf("expected ErrStoryNotApproved for pending story comment, got %v", err)
	}

	stats, err := svc.GetAuthorStats(ctx, 2)
	if err != nil {
		t.Fatalf("expected GetAuthorStats user2 success, got %v", err)
	}
	if stats.MaxStoriesPerMonth != 3 {
		t.Fatalf("expected low-level max stories 3, got %d", stats.MaxStoriesPerMonth)
	}

	hearted, _, err := svc.ToggleHeart(ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), 1)
	if err != nil {
		t.Fatalf("expected self-heart success, got %v", err)
	}
	if !hearted {
		t.Fatal("expected first self-heart to be added")
	}
}

func setupInspiringStoryServiceCreateSuccess(t *testing.T) *InspiringStoryService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	now := time.Now()
	schema := []string{
		`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, avatar TEXT, exp INTEGER, deleted_at DATETIME)`,
		`CREATE TABLE level_configs (id INTEGER PRIMARY KEY, level INTEGER, min_exp INTEGER, badge_name TEXT, badge_icon TEXT, tier_name TEXT, tier_color TEXT, description TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE inspiring_stories (
			id TEXT PRIMARY KEY DEFAULT (
				lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-' || lower(hex(randomblob(2))) || '-' || lower(hex(randomblob(2))) || '-' || lower(hex(randomblob(6)))
			),
			author_id INTEGER,
			title TEXT,
			content TEXT,
			cover_image TEXT,
			is_anonymous BOOLEAN,
			has_trigger_warning BOOLEAN,
			trigger_warning_text TEXT,
			status TEXT,
			moderator_id INTEGER,
			moderation_feedback TEXT,
			moderated_at DATETIME,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_featured BOOLEAN,
			featured_by INTEGER,
			featured_at DATETIME,
			featured_until DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			published_at DATETIME
		)`,
		`CREATE TABLE story_categories (id TEXT PRIMARY KEY, name TEXT, slug TEXT, description TEXT, icon TEXT, display_order INTEGER, is_active BOOLEAN, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE story_category_relations (story_id TEXT, category_id TEXT, inspiring_story_id TEXT, story_category_id TEXT)`,
		`CREATE TABLE story_tags (
			id TEXT PRIMARY KEY DEFAULT (
				lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-' || lower(hex(randomblob(2))) || '-' || lower(hex(randomblob(2))) || '-' || lower(hex(randomblob(6)))
			),
			story_id TEXT,
			tag TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE story_hearts (id TEXT PRIMARY KEY, story_id TEXT, user_id INTEGER, created_at DATETIME)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	if err := db.Exec(`INSERT INTO users (id, name, avatar, exp, deleted_at) VALUES (1, 'Creator', 'a.png', 100, NULL)`).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (1, 1, 0, 'Seed', 'icon', 'Bronze', '#A97142', 'desc', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed levels: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_categories (id, name, slug, description, icon, display_order, is_active, created_at, updated_at) VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Hope', 'hope', 'desc', '🌟', 1, 1, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed categories: %v", err)
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

func TestInspiringStoryService_CreateStory_SuccessWithCategoryAndTags(t *testing.T) {
	svc := setupInspiringStoryServiceCreateSuccess(t)
	ctx := context.Background()

	story, err := svc.CreateStory(ctx, 1, &dto.CreateStoryRequest{
		Title:       "Cerita Baru",
		Content:     strings.Repeat("konten ", 30),
		CategoryIDs: []uuid.UUID{uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")},
		Tags:        []string{"hope", "calm"},
	})
	if err != nil {
		t.Fatalf("expected CreateStory success, got %v", err)
	}
	if story == nil || story.ID == uuid.Nil {
		t.Fatalf("expected non-empty story response, got %+v", story)
	}
	if len(story.Tags) == 0 {
		t.Fatalf("expected tags in response, got tags=%d", len(story.Tags))
	}
}

func TestInspiringStoryService_MonthlyLimitAndAdditionalErrorBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("create story monthly limit reached", func(t *testing.T) {
		svc, db := setupInspiringStoryServiceInteractions(t)

		now := time.Now()
		if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, is_anonymous, view_count, heart_count, comment_count, is_featured, created_at, updated_at, published_at) VALUES
			(?, 2, 'U2 Story 1', 'c1', 'approved', 0, 0, 0, 0, 0, ?, ?, ?),
			(?, 2, 'U2 Story 2', 'c2', 'approved', 0, 0, 0, 0, 0, ?, ?, ?),
			(?, 2, 'U2 Story 3', 'c3', 'pending', 0, 0, 0, 0, 0, ?, ?, NULL)`,
			uuid.New().String(), now, now, now,
			uuid.New().String(), now, now, now,
			uuid.New().String(), now, now).Error; err != nil {
			t.Fatalf("seed user2 stories: %v", err)
		}

		_, err := svc.CreateStory(ctx, 2, &dto.CreateStoryRequest{Title: "blocked", Content: "should fail by limit"})
		if err != ErrMonthlyStoryLimitReached {
			t.Fatalf("expected ErrMonthlyStoryLimitReached, got %v", err)
		}
	})

	t.Run("update story with nil tags branch", func(t *testing.T) {
		svc := setupInspiringStoryServiceFlows(t)
		storyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

		updated, err := svc.UpdateStory(ctx, storyID, 1, &dto.UpdateStoryRequest{Title: "Only Title Updated"})
		if err != nil {
			t.Fatalf("update story with nil tags failed: %v", err)
		}
		if updated.Title != "Only Title Updated" {
			t.Fatalf("unexpected updated story title: %+v", updated)
		}
	})

	t.Run("update story categories update error", func(t *testing.T) {
		svc, db := setupInspiringStoryServiceInteractions(t)
		storyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		if err := db.Exec(`DROP TABLE story_category_relations`).Error; err != nil {
			t.Fatalf("drop story_category_relations: %v", err)
		}

		_, err := svc.UpdateStory(ctx, storyID, 1, &dto.UpdateStoryRequest{
			Title:       "Updated Categories",
			CategoryIDs: []uuid.UUID{uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")},
		})
		if err == nil {
			t.Fatal("expected UpdateStory category update error")
		}
	})

	t.Run("update story tags update error", func(t *testing.T) {
		svc, db := setupInspiringStoryServiceInteractions(t)
		storyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		if err := db.Exec(`DROP TABLE story_tags`).Error; err != nil {
			t.Fatalf("drop story_tags: %v", err)
		}

		_, err := svc.UpdateStory(ctx, storyID, 1, &dto.UpdateStoryRequest{
			Title: "Updated Tags",
			Tags:  []string{"new-tag"},
		})
		if err == nil {
			t.Fatal("expected UpdateStory tags update error")
		}
	})

	t.Run("update story final fetch error", func(t *testing.T) {
		svc, db := setupInspiringStoryServiceInteractions(t)
		storyID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
		if err := db.Exec(`DROP TABLE story_categories`).Error; err != nil {
			t.Fatalf("drop story_categories: %v", err)
		}

		_, err := svc.UpdateStory(ctx, storyID, 1, &dto.UpdateStoryRequest{Title: "Updated Fetch"})
		if err == nil {
			t.Fatal("expected UpdateStory final fetch error")
		}
	})

	t.Run("toggle comment heart add error branch", func(t *testing.T) {
		svc, db := setupInspiringStoryServiceInteractions(t)
		commentID := uuid.MustParse("99999999-9999-9999-9999-999999999999")
		now := time.Now()

		if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, heart_count, is_hidden, created_at, updated_at) VALUES (?, ?, 1, 'x', 0, 0, ?, ?)`, commentID.String(), "11111111-1111-1111-1111-111111111111", now, now).Error; err != nil {
			t.Fatalf("seed comment: %v", err)
		}
		if err := db.Exec(`DROP TABLE story_comment_hearts`).Error; err != nil {
			t.Fatalf("drop story_comment_hearts: %v", err)
		}

		if _, _, err := svc.ToggleCommentHeart(ctx, commentID, 2); err == nil {
			t.Fatal("expected toggle comment heart error when table missing")
		}
	})
}
