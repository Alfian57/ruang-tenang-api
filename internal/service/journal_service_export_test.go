package service

import (
	"context"
	"encoding/base64"
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

func setupJournalServiceForExport(t *testing.T) *JournalService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
		`CREATE TABLE user_moods (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, mood TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE journals (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE,
			user_id INTEGER NOT NULL,
			title TEXT,
			content TEXT,
			summary TEXT,
			mood_id INTEGER,
			tags TEXT,
			is_private BOOLEAN,
			share_with_ai BOOLEAN,
			ai_accessed_at DATETIME,
			word_count INTEGER,
			sentiment_score REAL,
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
	if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, uuid.New().String(), 1, "Entry", "<p>Today<br/>good</p>", now, now).Error; err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	return &JournalService{journalRepo: repository.NewJournalRepository(db)}
}

func TestJournalService_ExportHelpersAndConstructor(t *testing.T) {
	svc := NewJournalService(nil, nil, nil, nil, nil)
	if svc == nil {
		t.Fatal("expected constructor to return service")
	}

	plain := svc.stripHTML("<p>Hello<br/>World</p>&amp;")
	if plain != "Hello\nWorld\n\n&" {
		t.Fatalf("unexpected stripHTML output: %q", plain)
	}

	journals := []model.Journal{{
		Title:     "A",
		Content:   "<div>Body</div>",
		CreatedAt: time.Now(),
		Mood:      &model.UserMood{Mood: model.MoodHappy},
		Tags:      []string{"calm", "focus"},
	}}
	txt := svc.exportToTXT(context.Background(), journals)
	if !strings.Contains(txt, "RUANG TENANG - JOURNAL EXPORT") || !strings.Contains(txt, "Body") || !strings.Contains(txt, "Mood:") || !strings.Contains(txt, "Tags:") {
		t.Fatalf("unexpected txt export content: %q", txt)
	}

	html := svc.exportToHTML(context.Background(), journals)
	if !strings.Contains(html, "<!DOCTYPE html>") || !strings.Contains(html, "Journal Export") {
		t.Fatalf("unexpected html export content: %q", html)
	}

	pdfBytes, err := svc.exportToPDF(context.Background(), journals)
	if err != nil || len(pdfBytes) == 0 {
		t.Fatalf("expected pdf bytes, err=%v len=%d", err, len(pdfBytes))
	}
}

func TestJournalService_ExportJournalsFormats(t *testing.T) {
	svc := setupJournalServiceForExport(t)
	ctx := context.Background()

	txtRes, err := svc.ExportJournals(ctx, 1, dto.JournalExportRequest{Format: "txt"})
	if err != nil {
		t.Fatalf("txt export failed: %v", err)
	}
	if !strings.HasSuffix(txtRes.Filename, ".txt") || !strings.Contains(txtRes.Content, "RUANG TENANG - JOURNAL EXPORT") {
		t.Fatalf("unexpected txt export response: %+v", txtRes)
	}

	pdfRes, err := svc.ExportJournals(ctx, 1, dto.JournalExportRequest{Format: "pdf"})
	if err != nil {
		t.Fatalf("pdf export failed: %v", err)
	}
	if !strings.HasSuffix(pdfRes.Filename, ".pdf") {
		t.Fatalf("unexpected pdf filename: %s", pdfRes.Filename)
	}
	decoded, decErr := base64.StdEncoding.DecodeString(pdfRes.Content)
	if decErr != nil || len(decoded) == 0 {
		t.Fatalf("invalid base64 pdf content: err=%v len=%d", decErr, len(decoded))
	}

	if _, err := svc.ExportJournals(ctx, 1, dto.JournalExportRequest{Format: "csv"}); err == nil {
		t.Fatal("expected unsupported format error")
	}

	dateStr := time.Now().Format("2006-01-02")
	dateFilteredRes, err := svc.ExportJournals(ctx, 1, dto.JournalExportRequest{Format: "txt", StartDate: dateStr, EndDate: dateStr})
	if err != nil {
		t.Fatalf("date filtered txt export failed: %v", err)
	}
	if !strings.HasSuffix(dateFilteredRes.Filename, ".txt") {
		t.Fatalf("unexpected date filtered txt filename: %s", dateFilteredRes.Filename)
	}

	if _, err := svc.ExportJournals(ctx, 1, dto.JournalExportRequest{Format: "txt", Tags: []string{"calm"}, StartDate: "invalid", EndDate: "invalid"}); err == nil {
		t.Fatal("expected export error when tags filter uses postgres operator on sqlite")
	}
}
