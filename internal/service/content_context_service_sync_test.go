package service

import (
	"context"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupContentContextSyncService(t *testing.T, withSchema bool) *ContentContextService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := &ContentContextService{
		articleRepo:      repository.NewArticleRepository(db),
		songRepo:         repository.NewSongRepository(db),
		songCategoryRepo: repository.NewSongCategoryRepository(db),
		forumRepo:        repository.NewForumRepository(db),
		articles:         make(map[uint]*ArticleSummary),
		songCategories:   make(map[uint]*SongCategorySummary),
		forums:           make(map[uint]*ForumSummary),
		syncInterval:     1 * time.Minute,
		stopChan:         make(chan struct{}),
	}

	if !withSchema {
		return svc
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			avatar TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE article_categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			slug TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE articles (
			id INTEGER PRIMARY KEY,
			title TEXT,
			slug TEXT,
			content TEXT,
			thumbnail TEXT,
			article_category_id INTEGER,
			user_id INTEGER,
			status TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE forum_categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			slug TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE forums (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			category_id INTEGER,
			title TEXT,
			slug TEXT,
			content TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE forum_posts (
			id INTEGER PRIMARY KEY,
			forum_id INTEGER,
			user_id INTEGER,
			content TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE song_categories (
			id INTEGER PRIMARY KEY,
			name TEXT,
			slug TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE songs (
			id INTEGER PRIMARY KEY,
			title TEXT,
			slug TEXT,
			file_path TEXT,
			song_category_id INTEGER,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`INSERT INTO users (id, name, avatar, created_at, updated_at) VALUES (1, 'U1', '/u1.png', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO article_categories (id, name, slug) VALUES (1, 'Mental', 'mental')`,
		`INSERT INTO articles (id, title, slug, content, article_category_id, user_id, status, created_at, updated_at)
			VALUES (1, 'Published A', 'published-a', 'c', 1, 1, 'published', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			       (2, 'Draft B', 'draft-b', 'c', 1, 1, 'draft', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO forum_categories (id, name, slug) VALUES (1, 'General', 'general')`,
		`INSERT INTO forums (id, user_id, category_id, title, slug, content, created_at, updated_at)
			VALUES (1, 1, 1, 'Forum One', 'forum-one', 'c', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO forum_posts (id, forum_id, user_id, content, created_at, updated_at)
			VALUES (1, 1, 1, 'Reply', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		`INSERT INTO song_categories (id, name, slug) VALUES (1, 'Focus', 'focus')`,
		`INSERT INTO songs (id, title, slug, file_path, song_category_id, created_at, updated_at)
			VALUES (1, 'Song A', 'song-a', '/a.mp3', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("setup query failed: %v", err)
		}
	}

	return svc
}

func TestContentContextService_IncrementalSyncAndRepoSyncs(t *testing.T) {
	ctx := context.Background()
	svc := setupContentContextSyncService(t, true)

	svc.syncForums(ctx, time.Time{})
	if len(svc.forums) == 0 {
		t.Fatal("expected synced forums")
	}

	svc.syncMusic(ctx)
	if len(svc.songCategories) == 0 {
		t.Fatal("expected synced song categories")
	}

	svc.syncArticles(ctx, time.Time{})
	if len(svc.articles) == 0 {
		t.Fatal("expected synced published articles")
	}

	beforeArticleSync := svc.lastArticleSync
	beforeForumSync := svc.lastForumSync
	beforeMusicSync := svc.lastMusicSync

	svc.incrementalSync(ctx)
	if !svc.lastArticleSync.After(beforeArticleSync) {
		t.Fatal("expected article sync timestamp updated")
	}
	if !svc.lastForumSync.After(beforeForumSync) {
		t.Fatal("expected forum sync timestamp updated")
	}
	if !svc.lastMusicSync.After(beforeMusicSync) {
		t.Fatal("expected music sync timestamp updated")
	}
}

func TestContentContextService_SyncErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupContentContextSyncService(t, false)

	svc.syncArticles(ctx, time.Time{})
	svc.syncForums(ctx, time.Time{})
	svc.syncMusic(ctx)
	svc.incrementalSync(ctx)

	if len(svc.articles) != 0 || len(svc.forums) != 0 || len(svc.songCategories) != 0 {
		t.Fatal("expected no cached data on missing schema")
	}
}

func TestContentContextService_BackgroundSyncTickerAndStop(t *testing.T) {
	ctx := context.Background()
	svc := setupContentContextSyncService(t, true)
	svc.syncInterval = 5 * time.Millisecond

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.backgroundSync(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	svc.Stop(ctx)

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("backgroundSync did not stop in time")
	}

	if !svc.isReady {
		t.Fatal("expected service to become ready after background sync pre-warm")
	}
}
