package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func TestContentContextService_BuildContextFromMaps(t *testing.T) {
	s := &ContentContextService{
		articles: map[uint]*ArticleSummary{
			1: {ID: 1, Title: strings.Repeat("A", 70), Category: "Edu"},
		},
		songCategories: map[uint]*SongCategorySummary{
			2: {ID: 2, Name: "Calm", SongCount: 4},
		},
		forums: map[uint]*ForumSummary{
			3: {ID: 3, Title: strings.Repeat("F", 60), RepliesCount: 9},
		},
	}

	ctxText := s.buildContextFromMaps(context.Background())
	if !strings.Contains(ctxText, "Data sedang dimuat") {
		t.Fatalf("expected loading header for not-ready cache")
	}
	if !strings.Contains(ctxText, "ARTIKEL KESEHATAN MENTAL") {
		t.Fatalf("expected article section")
	}
	if !strings.Contains(ctxText, "MUSIK RELAKSASI") {
		t.Fatalf("expected music section")
	}
	if !strings.Contains(ctxText, "FORUM KOMUNITAS") {
		t.Fatalf("expected forum section")
	}
	if !strings.Contains(ctxText, "...") {
		t.Fatalf("expected long title truncation marker")
	}

	s.isReady = true
	ctxText = s.buildContextFromMaps(context.Background())
	if !strings.Contains(ctxText, "KONTEN APLIKASI (Untuk Rekomendasi ke Pengguna)") {
		t.Fatalf("expected ready header")
	}
	if !strings.Contains(ctxText, "FITUR APLIKASI RUANG TENANG") {
		t.Fatalf("expected app features section")
	}
}

func TestContentContextService_NotifyAndDelete(t *testing.T) {
	s := &ContentContextService{
		articles:       make(map[uint]*ArticleSummary),
		songCategories: make(map[uint]*SongCategorySummary),
		forums:         make(map[uint]*ForumSummary),
		stopChan:       make(chan struct{}),
	}

	s.NotifyArticleChange(context.Background(), nil)
	if len(s.articles) != 0 {
		t.Fatalf("nil article must not be cached")
	}

	unpublished := &model.Article{ID: 1, Title: "Draft", Status: model.ArticleStatusDraft}
	s.NotifyArticleChange(context.Background(), unpublished)
	if len(s.articles) != 0 {
		t.Fatalf("non-published article must not be cached")
	}

	published := &model.Article{ID: 2, Title: "Live", Status: model.ArticleStatusPublished}
	s.NotifyArticleChange(context.Background(), published)
	if got := s.articles[2]; got == nil || got.Category != "Umum" {
		t.Fatalf("expected published article in cache with default category")
	}

	forum := &model.Forum{ID: 10, Title: "Topik", RepliesCount: 2}
	s.NotifyForumChange(context.Background(), forum)
	if s.forums[10] == nil {
		t.Fatalf("expected forum in cache")
	}

	s.NotifyArticleDelete(context.Background(), 2)
	if _, ok := s.articles[2]; ok {
		t.Fatalf("article must be removed")
	}

	s.NotifyForumDelete(context.Background(), 10)
	if _, ok := s.forums[10]; ok {
		t.Fatalf("forum must be removed")
	}

	s.Stop(context.Background())
}
