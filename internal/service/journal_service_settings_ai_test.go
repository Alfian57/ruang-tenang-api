package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func journalSummaryResponse(text string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{genai.Text(text)}}}},
	}
}

func setupJournalServiceForSettingsAI(t *testing.T) (*JournalService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schema := []string{
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
		`CREATE TABLE journal_settings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER UNIQUE,
			allow_ai_access BOOLEAN,
			ai_context_days INTEGER,
			ai_context_max_entries INTEGER,
			default_share_with_ai BOOLEAN,
			is_blocked BOOLEAN,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE journal_ai_access_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER,
			journal_id INTEGER,
			chat_session_id INTEGER,
			accessed_at DATETIME,
			context_type TEXT
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("schema error: %v", err)
		}
	}

	now := time.Now()
	if err := db.Exec(`INSERT INTO journals (uuid, user_id, title, content, share_with_ai, created_at, updated_at) VALUES (?, 1, 'A', 'This is journal content for AI context', 1, ?, ?)`, uuid.New().String(), now, now).Error; err != nil {
		t.Fatalf("seed journal: %v", err)
	}

	svc := &JournalService{
		journalRepo:   repository.NewJournalRepository(db),
		settingsRepo:  repository.NewJournalSettingsRepository(db),
		accessLogRepo: repository.NewJournalAIAccessLogRepository(db),
	}
	return svc, db
}

func TestJournalService_SettingsAndAIContext(t *testing.T) {
	svc, _ := setupJournalServiceForSettingsAI(t)
	ctx := context.Background()

	settings, err := svc.GetSettings(ctx, 1)
	if err != nil {
		t.Fatalf("get settings failed: %v", err)
	}
	if settings == nil || settings.AIContextDays == 0 || settings.AIContextMaxEntries == 0 {
		t.Fatalf("unexpected settings: %+v", settings)
	}

	allow := true
	days := 14
	maxEntries := 10
	defaultShare := true
	updated, err := svc.UpdateSettings(ctx, 1, dto.JournalSettingsRequest{
		AllowAIAccess:       &allow,
		AIContextDays:       &days,
		AIContextMaxEntries: &maxEntries,
		DefaultShareWithAI:  &defaultShare,
	})
	if err != nil {
		t.Fatalf("update settings failed: %v", err)
	}
	if !updated.AllowAIAccess || updated.AIContextDays != 14 || updated.AIContextMaxEntries != 10 || !updated.DefaultShareWithAI {
		t.Fatalf("unexpected updated settings: %+v", updated)
	}

	blocked, err := svc.ToggleJournalBlock(ctx, 1)
	if err != nil {
		t.Fatalf("toggle block failed: %v", err)
	}
	if !blocked.IsBlocked {
		t.Fatalf("expected blocked=true, got %+v", blocked)
	}

	unblocked, err := svc.ToggleJournalBlock(ctx, 1)
	if err != nil {
		t.Fatalf("toggle unblock failed: %v", err)
	}
	if unblocked.IsBlocked {
		t.Fatalf("expected blocked=false, got %+v", unblocked)
	}

	ctxRes, err := svc.GetAIContext(ctx, 1, nil, dto.JournalAIContextRequest{IncludeSummary: true, MaxEntries: 2, DaysBack: 30})
	if err != nil {
		t.Fatalf("get ai context failed: %v", err)
	}
	if !ctxRes.HasAccess || ctxRes.EntriesCount < 1 || len(ctxRes.Entries) < 1 || ctxRes.LastEntryDate == nil {
		t.Fatalf("unexpected ai context response: %+v", ctxRes)
	}

	logs, err := svc.GetAIAccessLogs(ctx, 1, 10)
	if err != nil {
		t.Fatalf("get ai access logs failed: %v", err)
	}
	if len(logs) < 1 {
		t.Fatalf("expected ai access logs, got %d", len(logs))
	}

	disable := false
	if _, err := svc.UpdateSettings(ctx, 1, dto.JournalSettingsRequest{AllowAIAccess: &disable}); err != nil {
		t.Fatalf("disable ai access failed: %v", err)
	}
	ctxDisabled, err := svc.GetAIContext(ctx, 1, nil, dto.JournalAIContextRequest{})
	if err != nil {
		t.Fatalf("get ai context disabled failed: %v", err)
	}
	if ctxDisabled.HasAccess {
		t.Fatalf("expected hasAccess=false when disabled, got %+v", ctxDisabled)
	}
}

func TestJournalService_SettingsAIHelpers(t *testing.T) {
	svc := &JournalService{}

	if got := svc.truncateContent(context.Background(), "hello", 10); got != "hello" {
		t.Fatalf("unexpected truncateContent short result: %q", got)
	}
	if got := svc.truncateContent(context.Background(), "0123456789ABCDE", 5); got != "01234..." {
		t.Fatalf("unexpected truncateContent long result: %q", got)
	}

	summary, err := svc.generateJournalSummary(context.Background(), []model.Journal{{Content: "one", CreatedAt: time.Now()}})
	if err != nil {
		t.Fatalf("unexpected summary fallback error: %v", err)
	}
	if summary != "" {
		t.Fatalf("expected empty summary without genai client, got %q", summary)
	}
}

func TestJournalService_GenerateJournalSummary_WithInjectedGenerator(t *testing.T) {
	now := time.Now()
	entries := []model.Journal{{Content: "entry one", CreatedAt: now}}
	svc := &JournalService{
		genaiClient: &genai.Client{},
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return nil, errors.New("gen fail")
	}
	_, err := svc.generateJournalSummary(context.Background(), entries)
	if err == nil {
		t.Fatal("expected generator error")
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{}, nil
	}
	res, err := svc.generateJournalSummary(context.Background(), entries)
	if err != nil || res != "" {
		t.Fatalf("expected empty summary on empty candidate response, got %q err=%v", res, err)
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return journalSummaryResponse("ringkasan jurnal"), nil
	}
	res, err = svc.generateJournalSummary(context.Background(), entries)
	if err != nil || res != "ringkasan jurnal" {
		t.Fatalf("expected parsed summary text, got %q err=%v", res, err)
	}
}

func TestJournalService_GenerateSingleEntrySummary_WithInjectedGenerator(t *testing.T) {
	svc := &JournalService{}

	res, err := svc.generateSingleEntrySummary(context.Background(), "isi jurnal")
	if err != nil || res != "" {
		t.Fatalf("expected empty summary when no client and no generator, got %q err=%v", res, err)
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return nil, errors.New("single summary fail")
	}
	if _, err := svc.generateSingleEntrySummary(context.Background(), "isi jurnal"); err == nil {
		t.Fatal("expected error from injected generator")
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{}, nil
	}
	res, err = svc.generateSingleEntrySummary(context.Background(), "isi jurnal")
	if err != nil || res != "" {
		t.Fatalf("expected empty summary for empty candidates, got %q err=%v", res, err)
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return journalSummaryResponse("  ringkasan entri  "), nil
	}
	res, err = svc.generateSingleEntrySummary(context.Background(), "isi jurnal")
	if err != nil || res != "ringkasan entri" {
		t.Fatalf("expected trimmed text summary, got %q err=%v", res, err)
	}
}
