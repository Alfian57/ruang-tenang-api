package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInspiringStoryServiceHelpers(t *testing.T) *InspiringStoryService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			avatar TEXT,
			exp INTEGER,
			deleted_at DATETIME
		)`,
		`CREATE TABLE inspiring_stories (
			id TEXT PRIMARY KEY,
			author_id INTEGER,
			title TEXT,
			content TEXT,
			status TEXT,
			view_count INTEGER,
			heart_count INTEGER,
			comment_count INTEGER,
			is_anonymous BOOLEAN,
			is_featured BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME,
			published_at DATETIME
		)`,
		`CREATE TABLE story_tags (
			id TEXT PRIMARY KEY,
			story_id TEXT,
			tag TEXT,
			created_at DATETIME
		)`,
		`CREATE TABLE story_hearts (
			id TEXT PRIMARY KEY,
			story_id TEXT,
			user_id INTEGER,
			created_at DATETIME
		)`,
		`CREATE TABLE story_comments (
			id TEXT PRIMARY KEY,
			story_id TEXT,
			user_id INTEGER,
			content TEXT,
			heart_count INTEGER,
			is_hidden BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE story_comment_hearts (
			id TEXT PRIMARY KEY,
			comment_id TEXT,
			user_id INTEGER,
			created_at DATETIME
		)`,
		`CREATE TABLE level_configs (
			id INTEGER PRIMARY KEY,
			level INTEGER,
			min_exp INTEGER,
			badge_name TEXT,
			badge_icon TEXT,
			tier_name TEXT,
			tier_color TEXT,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO users (id, name, avatar, exp, deleted_at) VALUES (1, 'Alfi', 'ava.png', 120, NULL)`).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Exec(`INSERT INTO inspiring_stories (id, author_id, title, content, status, view_count, heart_count, comment_count, is_anonymous, is_featured, created_at, updated_at, published_at) VALUES (?, 1, 'Story One', 'Some content', 'approved', 10, 2, 1, 0, 0, ?, ?, ?)`, "11111111-1111-1111-1111-111111111111", now, now, now).Error; err != nil {
		t.Fatalf("seed story: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_tags (id, story_id, tag, created_at) VALUES (?, ?, 'hope', ?)`, uuid.New().String(), "11111111-1111-1111-1111-111111111111", now).Error; err != nil {
		t.Fatalf("seed story tag: %v", err)
	}
	if err := db.Exec(`INSERT INTO story_comments (id, story_id, user_id, content, heart_count, is_hidden, created_at, updated_at) VALUES (?, ?, 1, 'Nice', 1, 0, ?, ?)`, "22222222-2222-2222-2222-222222222222", "11111111-1111-1111-1111-111111111111", now, now).Error; err != nil {
		t.Fatalf("seed story comment: %v", err)
	}
	if err := db.Exec(`INSERT INTO level_configs (id, level, min_exp, badge_name, badge_icon, tier_name, tier_color, description, created_at, updated_at) VALUES (1, 2, 100, 'Badge', 'icon', 'Bronze', '#A97142', 'desc', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed level config: %v", err)
	}

	storyRepo := repository.NewInspiringStoryRepository(db)
	userRepo := repository.NewUserRepository(db)
	levelRepo := repository.NewLevelConfigRepository(db)
	return NewInspiringStoryService(storyRepo, userRepo, levelRepo, nil, nil, nil)
}

func TestInspiringStoryService_HelperMethods(t *testing.T) {
	svc := setupInspiringStoryServiceHelpers(t)
	ctx := context.Background()

	short := svc.createExcerpt(ctx, "short text", 200)
	if short != "short text" {
		t.Fatalf("unexpected short excerpt: %q", short)
	}

	long := svc.createExcerpt(ctx, strings.Repeat("word ", 80), 50)
	if !strings.HasSuffix(long, "...") || len(long) <= 50 {
		t.Fatalf("unexpected long excerpt: %q", long)
	}

	author := svc.buildAuthorResponse(ctx, 1)
	if author == nil || author.Name != "Alfi" || author.TierName == "" {
		t.Fatalf("unexpected author response: %+v", author)
	}
	if missing := svc.buildAuthorResponse(ctx, 999); missing != nil {
		t.Fatalf("expected nil author for missing user, got %+v", missing)
	}

	anonCard := svc.toStoryCard(ctx, model.InspiringStory{
		ID:          uuid.New(),
		Title:       "Anon Story",
		Content:     "content",
		IsAnonymous: true,
		Categories: []model.StoryCategory{{
			ID:   uuid.New(),
			Name: "Hope",
			Slug: "hope",
			Icon: "🌟",
		}},
	})
	if anonCard.Author == nil || anonCard.Author.Name != "Anonim" || len(anonCard.Categories) != 1 {
		t.Fatalf("unexpected anon story card: %+v", anonCard)
	}

	nonAnonCard := svc.toStoryCard(ctx, model.InspiringStory{
		ID:          uuid.New(),
		AuthorID:    1,
		Title:       "Public Story",
		Content:     "public content",
		IsAnonymous: false,
	})
	if nonAnonCard.Author == nil || nonAnonCard.Author.Name != "Alfi" {
		t.Fatalf("unexpected non-anon story card: %+v", nonAnonCard)
	}

	commentResp := svc.toCommentResponse(ctx, &model.StoryComment{
		ID:        uuid.New(),
		UserID:    1,
		Content:   "Great post",
		CreatedAt: time.Now(),
		User:      &model.User{ID: 1},
	}, 1)
	if commentResp == nil || commentResp.Author == nil || commentResp.Author.Name != "Alfi" {
		t.Fatalf("unexpected comment response: %+v", commentResp)
	}
}

func TestInspiringStoryService_BuildResponseAndStats(t *testing.T) {
	svc := setupInspiringStoryServiceHelpers(t)
	ctx := context.Background()

	story := &model.InspiringStory{
		ID:          uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		AuthorID:    1,
		Title:       "Story One",
		Content:     "Some content",
		Status:      model.StoryStatusApproved,
		IsAnonymous: false,
		Categories: []model.StoryCategory{{
			ID:   uuid.New(),
			Name: "Hope",
			Slug: "hope",
			Icon: "🌟",
		}},
		CreatedAt: time.Now(),
	}

	resp, err := svc.buildStoryResponseDTO(ctx, story, 1)
	if err != nil {
		t.Fatalf("buildStoryResponseDTO failed: %v", err)
	}
	if resp == nil || resp.Author == nil || resp.Author.Name != "Alfi" {
		t.Fatalf("unexpected story response: %+v", resp)
	}

	stats, err := svc.GetAuthorStats(ctx, 1)
	if err != nil {
		t.Fatalf("GetAuthorStats failed: %v", err)
	}
	if stats == nil || stats.MaxStoriesPerMonth == 0 {
		t.Fatalf("unexpected stats response: %+v", stats)
	}
}
