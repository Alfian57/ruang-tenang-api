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

func setupChatExportDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	createSessions := `CREATE TABLE chat_sessions (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
		folder_id INTEGER,
		title TEXT,
		summary TEXT,
		summary_generated_at DATETIME,
		is_favorite BOOLEAN,
		is_trash BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`
	createMessages := `CREATE TABLE chat_messages (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		chat_session_id INTEGER,
		role TEXT,
		content TEXT,
		type TEXT,
		is_liked BOOLEAN,
		is_disliked BOOLEAN,
		is_pinned BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME
	)`

	if err := db.Exec(createSessions).Error; err != nil {
		t.Fatalf("create chat_sessions: %v", err)
	}
	if err := db.Exec(createMessages).Error; err != nil {
		t.Fatalf("create chat_messages: %v", err)
	}

	now := time.Now()
	sessionUUID := uuid.NewString()
	msg1UUID := uuid.NewString()
	msg2UUID := uuid.NewString()
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, title, created_at, updated_at) VALUES (1, ?, 7, 'My Session', ?, ?)`, sessionUUID, now, now).Error; err != nil {
		t.Fatalf("seed chat session: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, is_pinned, created_at, updated_at) VALUES (1, ?, 1, 'user', 'hello', 0, ?, ?)`, msg1UUID, now, now).Error; err != nil {
		t.Fatalf("seed chat message 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, is_pinned, created_at, updated_at) VALUES (2, ?, 1, 'ai', 'hi there', 1, ?, ?)`, msg2UUID, now, now).Error; err != nil {
		t.Fatalf("seed chat message 2: %v", err)
	}

	return db
}

func TestChatService_ExportAsTXT(t *testing.T) {
	svc := &ChatService{}
	now := time.Now()
	summary := "Ringkasan singkat"
	session := &model.ChatSession{
		ID:        1,
		Title:     "Sesi Test / Emosi",
		Summary:   &summary,
		CreatedAt: now,
	}
	messages := []model.ChatMessage{
		{Role: model.ChatRoleUser, Content: "Halo", CreatedAt: now},
		{Role: model.ChatRoleAI, Content: "Hai juga", CreatedAt: now, IsPinned: true},
	}

	resp, err := svc.exportAsTXT(context.Background(), session, messages, true)
	if err != nil {
		t.Fatalf("exportAsTXT failed: %v", err)
	}
	if resp == nil || resp.Filename == "" || !strings.HasSuffix(resp.Filename, ".txt") {
		t.Fatalf("unexpected TXT response filename: %+v", resp)
	}
	if !strings.Contains(resp.Content, "Ruang Tenang - Chat Export") {
		t.Fatal("expected export header in TXT content")
	}
	if !strings.Contains(resp.Content, "(Pinned)") {
		t.Fatal("expected pinned marker in TXT export")
	}
}

func TestChatService_ExportAsPDF(t *testing.T) {
	svc := &ChatService{}
	now := time.Now()
	session := &model.ChatSession{
		ID:        1,
		Title:     "PDF Session",
		CreatedAt: now,
	}
	messages := []model.ChatMessage{
		{Role: model.ChatRoleUser, Content: "Pesan user", CreatedAt: now},
		{Role: model.ChatRoleAI, Content: "Pesan AI", CreatedAt: now},
	}

	resp, err := svc.exportAsPDF(context.Background(), session, messages, true)
	if err != nil {
		t.Fatalf("exportAsPDF failed: %v", err)
	}
	if resp == nil || resp.Filename == "" || !strings.HasSuffix(resp.Filename, ".pdf") {
		t.Fatalf("unexpected PDF response filename: %+v", resp)
	}
	decoded, err := base64.StdEncoding.DecodeString(resp.Content)
	if err != nil {
		t.Fatalf("expected valid base64 PDF content: %v", err)
	}
	if len(decoded) == 0 {
		t.Fatal("expected non-empty decoded PDF content")
	}
}

func TestChatService_ExportChat_Branches(t *testing.T) {
	db := setupChatExportDB(t)
	svc := &ChatService{sessionRepo: repository.NewChatSessionRepository(db)}

	_, err := svc.ExportChat(context.Background(), 999, 7, &dto.ExportChatRequest{Format: dto.ExportFormatTXT})
	if err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found, got %v", err)
	}

	_, err = svc.ExportChat(context.Background(), 1, 8, &dto.ExportChatRequest{Format: dto.ExportFormatTXT})
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized, got %v", err)
	}

	_, err = svc.ExportChat(context.Background(), 1, 7, &dto.ExportChatRequest{Format: dto.ExportFormat("csv")})
	if err == nil || err.Error() != "unsupported export format" {
		t.Fatalf("expected unsupported format, got %v", err)
	}

	respTXT, err := svc.ExportChat(context.Background(), 1, 7, &dto.ExportChatRequest{Format: dto.ExportFormatTXT, IncludePinned: true, IncludeMetadata: true})
	if err != nil {
		t.Fatalf("export chat txt failed: %v", err)
	}
	if !strings.Contains(respTXT.Content, "(Pinned)") || strings.Contains(respTXT.Content, "hello") {
		t.Fatalf("expected pinned-only TXT export, got: %s", respTXT.Content)
	}

	respPDF, err := svc.ExportChat(context.Background(), 1, 7, &dto.ExportChatRequest{Format: dto.ExportFormatPDF, IncludePinned: false, IncludeMetadata: false})
	if err != nil {
		t.Fatalf("export chat pdf failed: %v", err)
	}
	if _, err := base64.StdEncoding.DecodeString(respPDF.Content); err != nil {
		t.Fatalf("expected valid base64 pdf content: %v", err)
	}
}
