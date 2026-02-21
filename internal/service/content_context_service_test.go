package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func newContentSvcForTest() *ContentContextService {
	return &ContentContextService{
		articles:       make(map[uint]*ArticleSummary),
		songCategories: make(map[uint]*SongCategorySummary),
		forums:         make(map[uint]*ForumSummary),
		stopChan:       make(chan struct{}),
	}
}

func TestContentContextService_NotifyAndDelete_Extended(t *testing.T) {
	ctx := context.Background()
	s := newContentSvcForTest()

	s.NotifyArticleChange(ctx, nil)
	s.NotifyArticleChange(ctx, &model.Article{ID: 1, Title: "Draft", Status: model.ArticleStatusDraft})
	if len(s.articles) != 0 {
		t.Fatalf("expected no cached article for nil/draft")
	}

	s.NotifyArticleChange(ctx, &model.Article{ID: 1, Title: "Published", Status: model.ArticleStatusPublished})
	if len(s.articles) != 1 || s.articles[1].Category != "Umum" {
		t.Fatalf("expected article cached with default category")
	}

	s.NotifyArticleChange(ctx, &model.Article{ID: 2, Title: "With Category", Status: model.ArticleStatusPublished, Category: model.ArticleCategory{Name: "Relaksasi"}})
	if s.articles[2].Category != "Relaksasi" {
		t.Fatalf("expected category from relation")
	}

	s.NotifyArticleDelete(ctx, 1)
	if _, ok := s.articles[1]; ok {
		t.Fatalf("expected article removed")
	}

	s.NotifyForumChange(ctx, nil)
	s.NotifyForumChange(ctx, &model.Forum{ID: 10, Title: "Topik", RepliesCount: 3})
	if len(s.forums) != 1 || s.forums[10].RepliesCount != 3 {
		t.Fatalf("expected forum cached")
	}
	s.NotifyForumDelete(ctx, 10)
	if _, ok := s.forums[10]; ok {
		t.Fatalf("expected forum removed")
	}

	if !s.contextDirty {
		t.Fatalf("expected contextDirty true after notifications")
	}
}

func TestContentContextService_GetContentContext_CacheExtended(t *testing.T) {
	ctx := context.Background()
	s := newContentSvcForTest()
	s.isReady = true
	s.contextDirty = true
	s.articles[1] = &ArticleSummary{ID: 1, Title: "Artikel Pertama", Category: "Umum"}
	s.songCategories[1] = &SongCategorySummary{ID: 1, Name: "Sleep", SongCount: 4}
	s.forums[1] = &ForumSummary{ID: 1, Title: "Forum A", RepliesCount: 2}

	first := s.GetContentContext(ctx)
	if first == "" {
		t.Fatalf("expected built content context")
	}
	if !strings.Contains(first, "ARTIKEL KESEHATAN MENTAL") || !strings.Contains(first, "MUSIK RELAKSASI") || !strings.Contains(first, "FORUM KOMUNITAS") {
		t.Fatalf("unexpected context sections: %s", first)
	}
	if s.contextDirty {
		t.Fatalf("expected contextDirty false after build")
	}

	second := s.GetContentContext(ctx)
	if second != first {
		t.Fatalf("expected cached context to be reused")
	}
}

func TestContentContextService_BuildersAndStopExtended(t *testing.T) {
	ctx := context.Background()
	s := newContentSvcForTest()

	longArticleTitle := strings.Repeat("A", 70)
	longForumTitle := strings.Repeat("F", 60)

	s.articles[1] = &ArticleSummary{ID: 1, Title: longArticleTitle, Category: "Self Care"}
	s.songCategories[1] = &SongCategorySummary{ID: 1, Name: "Tenang", SongCount: 7}
	s.forums[1] = &ForumSummary{ID: 1, Title: longForumTitle, RepliesCount: 5}

	built := s.buildContextFromMaps(ctx)
	if !strings.Contains(built, "FITUR APLIKASI RUANG TENANG") {
		t.Fatalf("expected app features section")
	}
	if !strings.Contains(built, "...") {
		t.Fatalf("expected truncated long title indicators")
	}

	s.isReady = false
	loading := s.buildContextFromMaps(ctx)
	if !strings.Contains(loading, "Data sedang dimuat") {
		t.Fatalf("expected loading text when cache not ready")
	}

	s.Stop(ctx)
	select {
	case <-s.stopChan:
	default:
		t.Fatalf("expected stop channel closed")
	}
}
