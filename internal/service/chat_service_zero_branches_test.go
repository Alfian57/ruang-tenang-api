package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestChatService_CoreConstructorAndSetters(t *testing.T) {
	svc := NewChatService(nil, nil, &config.Config{GeminiAPIKey: "", AppEnv: "test"}, nil, nil)
	if svc == nil {
		t.Fatal("expected chat service instance")
	}

	folderRepo := &repository.ChatFolderRepository{}
	moderationRepo := &repository.ModerationRepository{}
	journalRepo := &repository.JournalRepository{}
	settingsRepo := &repository.JournalSettingsRepository{}
	accessLogRepo := &repository.JournalAIAccessLogRepository{}

	svc.SetFolderRepo(folderRepo)
	svc.SetModerationRepo(moderationRepo)
	svc.SetJournalRepos(journalRepo, settingsRepo, accessLogRepo)

	if svc.folderRepo != folderRepo {
		t.Fatal("expected folder repo to be set")
	}
	if svc.moderationRepo != moderationRepo {
		t.Fatal("expected moderation repo to be set")
	}
	if svc.journalRepo != journalRepo || svc.journalSettingsRepo != settingsRepo || svc.journalAccessLogRepo != accessLogRepo {
		t.Fatal("expected journal repos to be set")
	}

	_ = svc.GetGenAIClient()
}

func TestChatService_GenerateContent_Branches(t *testing.T) {
	t.Run("uses injected generateContentFn", func(t *testing.T) {
		svc := &ChatService{}
		svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
			return &genai.GenerateContentResponse{}, nil
		}

		resp, err := svc.generateContent(context.Background(), "hello")
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp == nil {
			t.Fatal("expected response from injected generateContentFn")
		}
	})

	t.Run("fallback branch executes when no injected fn", func(t *testing.T) {
		svc := &ChatService{}
		resp, err := svc.generateContent(context.Background(), "hello")
		if err == nil {
			t.Fatalf("expected error from fallback branch with nil genaiModel, got resp=%v", resp)
		}
	})
}

func TestChatService_FolderMethods_NilRepo(t *testing.T) {
	svc := &ChatService{}
	ctx := context.Background()

	if _, err := svc.GetFolders(ctx, 1); err == nil || err.Error() != "folder repository not initialized" {
		t.Fatalf("expected folder repo not initialized, got %v", err)
	}
	if _, err := svc.CreateFolder(ctx, 1, &dto.CreateChatFolderRequest{Name: "F"}); err == nil || err.Error() != "folder repository not initialized" {
		t.Fatalf("expected folder repo not initialized, got %v", err)
	}
	if _, err := svc.UpdateFolder(ctx, 1, 1, &dto.UpdateChatFolderRequest{Name: "U"}); err == nil || err.Error() != "folder repository not initialized" {
		t.Fatalf("expected folder repo not initialized, got %v", err)
	}
	if err := svc.DeleteFolder(ctx, 1, 1); err == nil || err.Error() != "folder repository not initialized" {
		t.Fatalf("expected folder repo not initialized, got %v", err)
	}
	if err := svc.ReorderFolders(ctx, 1, &dto.ReorderFoldersRequest{FolderIDs: []uint{1}}); err == nil || err.Error() != "folder repository not initialized" {
		t.Fatalf("expected folder repo not initialized, got %v", err)
	}
}

func TestChatService_UUIDWrappers_InvalidUUID(t *testing.T) {
	svc := &ChatService{}
	ctx := context.Background()

	if _, err := svc.GetSessionByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if _, _, err := svc.SendMessageByUUID(ctx, "invalid", 1, &dto.SendMessageRequest{Content: "hi"}); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if err := svc.ToggleTrashByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if err := svc.ToggleFavoriteByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if err := svc.DeleteSessionByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if err := svc.MoveSessionToFolderByUUID(ctx, "invalid", 1, nil); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if _, err := svc.ExportChatByUUID(ctx, "invalid", 1, &dto.ExportChatRequest{Format: dto.ExportFormatTXT}); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if _, err := svc.GetPinnedMessagesByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if _, err := svc.GenerateSummaryByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
	if _, err := svc.GetSummaryByUUID(ctx, "invalid", 1); err == nil || err.Error() != "invalid uuid" {
		t.Fatalf("expected invalid uuid, got %v", err)
	}
}

func setupChatServiceSessionDB(t *testing.T, withSchema bool) *ChatService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := &ChatService{
		sessionRepo:         repository.NewChatSessionRepository(db),
		messageRepo:         repository.NewChatMessageRepository(db),
		folderRepo:          repository.NewChatFolderRepository(db),
		gamificationService: NewGamificationService(db),
	}

	if !withSchema {
		return svc
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			exp INTEGER,
			deleted_at DATETIME
		)`,
		`CREATE TABLE user_activities (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			date DATETIME,
			count INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE exp_history (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			points INTEGER,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE chat_folders (
			id INTEGER PRIMARY KEY,
			uuid TEXT,
			user_id INTEGER,
			name TEXT,
			color TEXT,
			icon TEXT,
			position INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE chat_sessions (
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
		)`,
		`CREATE TABLE chat_messages (
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
		)`,
		`INSERT INTO chat_folders (id, user_id, name, color, icon, position) VALUES (1, 1, 'F1', '#111111', 'folder', 0), (2, 2, 'F2', '#222222', 'folder', 0)`,
		`INSERT INTO users (id, exp, deleted_at) VALUES (1, 0, NULL), (2, 0, NULL)`,
		`INSERT INTO chat_sessions (id, uuid, user_id, folder_id, title, is_favorite, is_trash) VALUES
			(1, '11111111-1111-1111-1111-111111111111', 1, 1, 'S1', 0, 0),
			(2, '22222222-2222-2222-2222-222222222222', 2, 2, 'S2', 0, 0)`,
		`INSERT INTO chat_messages (id, chat_session_id, role, content, type, is_liked, is_disliked, is_pinned) VALUES
			(1, 1, 'ai', 'Pinned one', 'text', 0, 0, 1),
			(2, 1, 'user', 'Hello', 'text', 0, 0, 0),
			(3, 2, 'ai', 'Other user msg', 'text', 0, 0, 0)`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("setup query failed: %v", err)
		}
	}

	return svc
}

func setupChatServiceSessionDBWithDB(t *testing.T, withSchema bool) (*ChatService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	svc := &ChatService{
		sessionRepo:         repository.NewChatSessionRepository(db),
		messageRepo:         repository.NewChatMessageRepository(db),
		folderRepo:          repository.NewChatFolderRepository(db),
		gamificationService: NewGamificationService(db),
	}

	if !withSchema {
		return svc, db
	}

	queries := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			exp INTEGER,
			deleted_at DATETIME
		)`,
		`CREATE TABLE user_activities (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			date DATETIME,
			count INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE exp_history (
			id INTEGER PRIMARY KEY,
			user_id INTEGER,
			activity_type TEXT,
			points INTEGER,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE chat_folders (
			id INTEGER PRIMARY KEY,
			uuid TEXT,
			user_id INTEGER,
			name TEXT,
			color TEXT,
			icon TEXT,
			position INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE chat_sessions (
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
		)`,
		`CREATE TABLE chat_messages (
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
		)`,
		`INSERT INTO chat_folders (id, user_id, name, color, icon, position) VALUES (1, 1, 'F1', '#111111', 'folder', 0), (2, 2, 'F2', '#222222', 'folder', 0)`,
		`INSERT INTO users (id, exp, deleted_at) VALUES (1, 0, NULL), (2, 0, NULL)`,
		`INSERT INTO chat_sessions (id, uuid, user_id, folder_id, title, is_favorite, is_trash) VALUES
			(1, '11111111-1111-1111-1111-111111111111', 1, 1, 'S1', 0, 0),
			(2, '22222222-2222-2222-2222-222222222222', 2, 2, 'S2', 0, 0)`,
		`INSERT INTO chat_messages (id, chat_session_id, role, content, type, is_liked, is_disliked, is_pinned) VALUES
			(1, 1, 'ai', 'Pinned one', 'text', 0, 0, 1),
			(2, 1, 'user', 'Hello', 'text', 0, 0, 0),
			(3, 2, 'ai', 'Other user msg', 'text', 0, 0, 0)`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("setup query failed: %v", err)
		}
	}

	return svc, db
}

func TestChatService_SendMessage_SuccessFallbackPath(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)

	userMsg, aiMsg, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "hello there"})
	if err != nil {
		t.Fatalf("send message success path failed: %v", err)
	}
	if userMsg == nil || aiMsg == nil {
		t.Fatalf("expected both user and ai messages, got user=%v ai=%v", userMsg, aiMsg)
	}
	if userMsg.Role != string(model.ChatRoleUser) {
		t.Fatalf("unexpected user role: %s", userMsg.Role)
	}
	if userMsg.Type != "text" {
		t.Fatalf("expected default text type, got %s", userMsg.Type)
	}
	if aiMsg.Role != string(model.ChatRoleAI) {
		t.Fatalf("unexpected ai role: %s", aiMsg.Role)
	}
	if !strings.Contains(aiMsg.Content, "gangguan koneksi") {
		t.Fatalf("expected fallback ai response, got %q", aiMsg.Content)
	}
}

func TestChatService_SendMessage_CrisisAndErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc, db := setupChatServiceSessionDBWithDB(t, true)

	moderationRepo := repository.NewModerationRepository(db)
	svc.SetModerationRepo(moderationRepo)

	if err := db.Exec(`CREATE TABLE crisis_keywords (
		id INTEGER PRIMARY KEY,
		keyword TEXT,
		category TEXT,
		severity TEXT,
		language TEXT,
		is_active BOOLEAN,
		notes TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create crisis_keywords table: %v", err)
	}
	if err := db.Exec(`INSERT INTO crisis_keywords (id, keyword, category, severity, language, is_active, notes, created_at, updated_at) VALUES (1, 'bunuh diri', 'suicide', 'critical', 'id', 1, '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("insert crisis keyword: %v", err)
	}

	userMsg, aiMsg, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "aku ingin bunuh diri", Type: "voice"})
	if err != nil {
		t.Fatalf("send message crisis branch failed: %v", err)
	}
	if userMsg.Type != "voice" {
		t.Fatalf("expected custom message type to be preserved, got %s", userMsg.Type)
	}
	if !strings.Contains(aiMsg.Content, "119 ext 8") {
		t.Fatalf("expected crisis hotline response, got %q", aiMsg.Content)
	}

	if _, _, err := svc.SendMessage(ctx, 999, 1, &dto.SendMessageRequest{Content: "x"}); err == nil || !strings.Contains(err.Error(), "session not found") {
		t.Fatalf("expected session not found error, got %v", err)
	}

	brokenDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open broken sqlite: %v", err)
	}
	svc.messageRepo = repository.NewChatMessageRepository(brokenDB)

	if _, _, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "x"}); err == nil || !strings.Contains(err.Error(), "failed to create user message") {
		t.Fatalf("expected create user message error, got %v", err)
	}
}

func TestChatService_SendMessage_AiMessageCreateFailure(t *testing.T) {
	ctx := context.Background()
	svc, db := setupChatServiceSessionDBWithDB(t, true)

	if err := db.Exec(`CREATE TRIGGER fail_ai_message_insert
	BEFORE INSERT ON chat_messages
	WHEN NEW.role = 'ai'
	BEGIN
		SELECT RAISE(FAIL, 'ai insert blocked');
	END;`).Error; err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	if _, _, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "hello"}); err == nil || !strings.Contains(err.Error(), "ai insert blocked") {
		t.Fatalf("expected ai message insert failure, got %v", err)
	}
}

func TestChatService_SendMessage_WithInjectedAIReply(t *testing.T) {
	ctx := context.Background()
	svc, db := setupChatServiceSessionDBWithDB(t, true)

	for i := 10; i <= 24; i++ {
		role := "user"
		if i%2 == 0 {
			role = "ai"
		}
		if err := db.Exec(`INSERT INTO chat_messages (id, chat_session_id, role, content, type, is_liked, is_disliked, is_pinned) VALUES (?, 1, ?, ?, 'text', 0, 0, 0)`, i, role, "msg history").Error; err != nil {
			t.Fatalf("seed extra message: %v", err)
		}
	}

	svc.genaiModel = &genai.GenerativeModel{}
	capturedHistoryLen := 0
	svc.generateChatReplyFn = func(_ context.Context, _ string, history []model.ChatMessage, userInput string) (string, error) {
		capturedHistoryLen = len(history)
		if userInput != "hello injected" {
			t.Fatalf("unexpected user input passed to generator: %q", userInput)
		}
		return "AI via stub", nil
	}

	_, aiMsg, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "hello injected"})
	if err != nil {
		t.Fatalf("send message with injected ai failed: %v", err)
	}
	if aiMsg == nil || aiMsg.Content != "AI via stub" {
		t.Fatalf("expected injected ai content, got %+v", aiMsg)
	}
	if capturedHistoryLen != 10 {
		t.Fatalf("expected trimmed history length 10, got %d", capturedHistoryLen)
	}
}

func TestChatService_SendMessage_InjectedAIErrorFallback(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)
	svc.genaiModel = &genai.GenerativeModel{}
	svc.generateChatReplyFn = func(context.Context, string, []model.ChatMessage, string) (string, error) {
		return "", context.DeadlineExceeded
	}

	_, aiMsg, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "hello"})
	if err != nil {
		t.Fatalf("send message should still succeed with fallback response, got %v", err)
	}
	if aiMsg == nil || !strings.Contains(aiMsg.Content, "gangguan koneksi") {
		t.Fatalf("expected fallback ai content, got %+v", aiMsg)
	}
}

func TestChatService_SendMessage_DefaultAIBranchFallback(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)

	client, err := genai.NewClient(ctx, option.WithAPIKey("dummy-test-key"))
	if err != nil {
		t.Skipf("unable to construct genai client in this environment: %v", err)
	}
	defer client.Close()

	svc.genaiModel = client.GenerativeModel("gemini-flash-latest")
	svc.generateChatReplyFn = nil

	_, aiMsg, err := svc.SendMessage(ctx, 1, 1, &dto.SendMessageRequest{Content: "halo, bantu saya"})
	if err != nil {
		t.Fatalf("send message should succeed with fallback when genai call fails, got %v", err)
	}
	if aiMsg == nil || !strings.Contains(aiMsg.Content, "gangguan koneksi") {
		t.Fatalf("expected fallback ai response from default AI path, got %+v", aiMsg)
	}
}

func TestChatService_SessionAndPinMethods_Branches(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)

	sessions, total, err := svc.GetSessions(ctx, 1, dto.ChatSessionQueryParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("get sessions failed: %v", err)
	}
	if total == 0 || len(sessions) == 0 {
		t.Fatalf("expected sessions, total=%d len=%d", total, len(sessions))
	}

	if _, err := svc.GetSessionByID(ctx, 2, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized for get session, got %v", err)
	}
	if _, err := svc.GetSessionByID(ctx, 1, 1); err != nil {
		t.Fatalf("expected get session success: %v", err)
	}

	if _, err := svc.CreateSession(ctx, 1, &dto.CreateChatSessionRequest{Title: "New Session"}); err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	if _, _, err := svc.SendMessage(ctx, 2, 1, &dto.SendMessageRequest{Content: "hello"}); err == nil || !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected unauthorized send message error, got %v", err)
	}

	if err := svc.ToggleTrash(ctx, 2, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized toggle trash, got %v", err)
	}
	if err := svc.ToggleTrash(ctx, 999, 1); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found toggle trash, got %v", err)
	}
	if err := svc.ToggleTrash(ctx, 1, 1); err != nil {
		t.Fatalf("toggle trash success expected: %v", err)
	}

	if err := svc.ToggleFavorite(ctx, 2, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized toggle favorite, got %v", err)
	}
	if err := svc.ToggleFavorite(ctx, 999, 1); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found toggle favorite, got %v", err)
	}
	if err := svc.ToggleFavorite(ctx, 1, 1); err != nil {
		t.Fatalf("toggle favorite success expected: %v", err)
	}

	if err := svc.MoveSessionToFolder(ctx, 2, 1, nil); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized move session, got %v", err)
	}
	if err := svc.MoveSessionToFolder(ctx, 1, 1, func() *uint { v := uint(999); return &v }()); err == nil || err.Error() != "folder not found" {
		t.Fatalf("expected folder not found, got %v", err)
	}
	if err := svc.MoveSessionToFolder(ctx, 1, 1, func() *uint { v := uint(1); return &v }()); err != nil {
		t.Fatalf("move session success expected: %v", err)
	}
	if err := svc.MoveSessionToFolder(ctx, 1, 1, func() *uint { v := uint(2); return &v }()); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized on foreign folder, got %v", err)
	}

	if err := svc.ToggleMessagePin(ctx, 3, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized toggle pin, got %v", err)
	}
	if err := svc.ToggleMessagePin(ctx, 2, 1); err != nil {
		t.Fatalf("toggle message pin success expected: %v", err)
	}

	if _, err := svc.GetPinnedMessages(ctx, 2, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized get pinned messages, got %v", err)
	}
	if pinned, err := svc.GetPinnedMessages(ctx, 1, 1); err != nil || len(pinned) == 0 {
		t.Fatalf("expected pinned messages, err=%v len=%d", err, len(pinned))
	}

	if err := svc.ToggleMessageLike(ctx, 2, 1); err != nil {
		t.Fatalf("toggle like success expected: %v", err)
	}
	if err := svc.ToggleMessageDislike(ctx, 2, 1); err != nil {
		t.Fatalf("toggle dislike success expected: %v", err)
	}

	if err := svc.DeleteSession(ctx, 2, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized delete session, got %v", err)
	}
	if err := svc.DeleteSession(ctx, 999, 1); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found delete session, got %v", err)
	}
	if err := svc.DeleteSession(ctx, 1, 1); err != nil {
		t.Fatalf("delete session success expected: %v", err)
	}
}

func TestChatService_GetJournalContext_Guards(t *testing.T) {
	ctx := context.Background()

	svcNil := &ChatService{}
	if got := svcNil.getJournalContext(ctx, 1, 1, ""); got != "" {
		t.Fatalf("expected empty context when repos nil, got %q", got)
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`CREATE TABLE journal_settings (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		allow_ai_access BOOLEAN,
		ai_context_days INTEGER,
		ai_context_max_entries INTEGER,
		default_share_with_ai BOOLEAN,
		is_blocked BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create journal_settings table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE user_moods (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		mood TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create user_moods table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE journals (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
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
	)`).Error; err != nil {
		t.Fatalf("create journals table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE journal_ai_access_logs (
		id INTEGER PRIMARY KEY,
		user_id INTEGER,
		journal_id INTEGER,
		chat_session_id INTEGER,
		accessed_at DATETIME,
		context_type TEXT
	)`).Error; err != nil {
		t.Fatalf("create journal_ai_access_logs table: %v", err)
	}

	svc := &ChatService{
		journalRepo:          repository.NewJournalRepository(db),
		journalSettingsRepo:  repository.NewJournalSettingsRepository(db),
		journalAccessLogRepo: repository.NewJournalAIAccessLogRepository(db),
	}

	if err := db.Create(&model.JournalSettings{UserID: 1, AllowAIAccess: false, AIContextDays: 7, AIContextMaxEntries: 5}).Error; err != nil {
		t.Fatalf("insert settings disallow: %v", err)
	}
	if got := svc.getJournalContext(ctx, 1, 1, ""); got != "" {
		t.Fatalf("expected empty context when AI access disabled, got %q", got)
	}

	if err := db.Model(&model.JournalSettings{}).Where("user_id = ?", 1).Update("allow_ai_access", true).Error; err != nil {
		t.Fatalf("update settings allow: %v", err)
	}
	now := time.Now()
	if err := db.Exec(`INSERT INTO journals (id, uuid, user_id, title, content, summary, share_with_ai, created_at, updated_at) VALUES (1, '99999999-9999-9999-9999-999999999999', 1, 'J1', 'konten jurnal personal', 'ringkas', 1, ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed journals: %v", err)
	}
	if got := svc.getJournalContext(ctx, 1, 1, ""); !strings.Contains(got, "KONTEKS JURNAL PRIBADI USER") {
		t.Fatalf("expected non-empty journal context when data exists, got %q", got)
	}
	if got := svc.getJournalContext(ctx, 1, 1, "something"); got != "" {
		t.Fatalf("expected empty context when journals unavailable, got %q", got)
	}
}

func TestChatService_CreateSessionAndToggleFavorite_ErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, false)

	if _, err := svc.CreateSession(ctx, 1, &dto.CreateChatSessionRequest{Title: "X"}); err == nil {
		t.Fatal("expected create session error when schema is missing")
	}
	if err := svc.ToggleFavorite(ctx, 1, 1); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found for toggle favorite, got %v", err)
	}
}

func TestChatService_FolderMethods_WithRepo(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)

	folders, err := svc.GetFolders(ctx, 1)
	if err != nil || len(folders) == 0 {
		t.Fatalf("expected folders, err=%v len=%d", err, len(folders))
	}

	created, err := svc.CreateFolder(ctx, 1, &dto.CreateChatFolderRequest{Name: "New Folder"})
	if err != nil {
		t.Fatalf("create folder failed: %v", err)
	}
	if created.Color == "" || created.Icon == "" {
		t.Fatalf("expected default color/icon, got color=%q icon=%q", created.Color, created.Icon)
	}

	if _, err := svc.UpdateFolder(ctx, 2, 1, &dto.UpdateChatFolderRequest{Name: "x"}); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized update folder, got %v", err)
	}
	if _, err := svc.UpdateFolder(ctx, 999, 1, &dto.UpdateChatFolderRequest{Name: "missing"}); err == nil || err.Error() != "folder not found" {
		t.Fatalf("expected folder not found update folder, got %v", err)
	}
	if _, err := svc.UpdateFolder(ctx, created.ID, 1, &dto.UpdateChatFolderRequest{Name: "Renamed"}); err != nil {
		t.Fatalf("update folder failed: %v", err)
	}
	newPos := 9
	updated, err := svc.UpdateFolder(ctx, created.ID, 1, &dto.UpdateChatFolderRequest{Name: "Renamed Again", Color: "#ffffff", Icon: "star", Position: &newPos})
	if err != nil {
		t.Fatalf("update folder with all fields failed: %v", err)
	}
	if updated.Color != "#ffffff" || updated.Icon != "star" || updated.Position != newPos {
		t.Fatalf("expected color/icon/position updated, got %+v", updated)
	}

	if err := svc.DeleteFolder(ctx, 2, 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized delete folder, got %v", err)
	}
	if err := svc.DeleteFolder(ctx, 999, 1); err == nil || err.Error() != "folder not found" {
		t.Fatalf("expected folder not found delete folder, got %v", err)
	}
	if err := svc.DeleteFolder(ctx, created.ID, 1); err != nil {
		t.Fatalf("delete folder failed: %v", err)
	}

	if err := svc.ReorderFolders(ctx, 1, &dto.ReorderFoldersRequest{FolderIDs: []uint{1}}); err != nil {
		t.Fatalf("reorder folders failed: %v", err)
	}
}

func TestChatService_PinsAdditionalErrorBranches(t *testing.T) {
	ctx := context.Background()
	svc, db := setupChatServiceSessionDBWithDB(t, true)

	if err := svc.ToggleMessagePin(ctx, 999, 1); err == nil || err.Error() != "message not found" {
		t.Fatalf("expected message not found on toggle pin, got %v", err)
	}

	if err := db.Exec(`INSERT INTO chat_messages (id, chat_session_id, role, content, type, is_pinned) VALUES (99, 9999, 'ai', 'orphan', 'text', 0)`).Error; err != nil {
		t.Fatalf("insert orphan message failed: %v", err)
	}
	if err := svc.ToggleMessagePin(ctx, 99, 1); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found on toggle pin, got %v", err)
	}

	if _, err := svc.GetPinnedMessages(ctx, 999, 1); err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found pinned query, got %v", err)
	}

	if err := db.Exec(`DROP TABLE chat_messages`).Error; err != nil {
		t.Fatalf("drop chat_messages failed: %v", err)
	}
	if _, err := svc.GetPinnedMessages(ctx, 1, 1); err == nil {
		t.Fatal("expected pinned messages repository error")
	}
}

func TestChatService_UUIDWrappers_ValidPathsAndAuth(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)

	if _, err := svc.GetSessionByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("get session by uuid failed: %v", err)
	}
	if _, _, err := svc.SendMessageByUUID(ctx, "22222222-2222-2222-2222-222222222222", 1, &dto.SendMessageRequest{Content: "hi"}); err == nil || !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected unauthorized send by uuid, got %v", err)
	}
	if err := svc.ToggleTrashByUUID(ctx, "22222222-2222-2222-2222-222222222222", 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized toggle trash by uuid, got %v", err)
	}
	if err := svc.ToggleFavoriteByUUID(ctx, "22222222-2222-2222-2222-222222222222", 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized toggle favorite by uuid, got %v", err)
	}
	if err := svc.DeleteSessionByUUID(ctx, "22222222-2222-2222-2222-222222222222", 1); err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized delete by uuid, got %v", err)
	}
	if err := svc.MoveSessionToFolderByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, nil); err != nil {
		t.Fatalf("move session to folder by uuid failed: %v", err)
	}

	if _, err := svc.ExportChatByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1, &dto.ExportChatRequest{Format: dto.ExportFormatTXT}); err != nil {
		t.Fatalf("export chat by uuid failed: %v", err)
	}
	if pinned, err := svc.GetPinnedMessagesByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil || len(pinned) == 0 {
		t.Fatalf("expected pinned by uuid, err=%v len=%d", err, len(pinned))
	}
	if _, err := svc.GenerateSummaryByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err == nil {
		t.Fatal("expected generate summary error for insufficient messages")
	}

	if err := svc.sessionRepo.UpdateSummary(ctx, 1, "ringkasan"); err != nil {
		t.Fatalf("update summary seed failed: %v", err)
	}
	if _, err := svc.GetSummaryByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("get summary by uuid failed: %v", err)
	}
}

func TestChatService_UUIDWrappers_SuccessAndNotFoundBranches(t *testing.T) {
	ctx := context.Background()
	svc := setupChatServiceSessionDB(t, true)

	if err := svc.ToggleTrashByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("toggle trash by uuid success expected: %v", err)
	}
	if err := svc.ToggleFavoriteByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("toggle favorite by uuid success expected: %v", err)
	}
	if err := svc.DeleteSessionByUUID(ctx, "11111111-1111-1111-1111-111111111111", 1); err != nil {
		t.Fatalf("delete session by uuid success expected: %v", err)
	}

	if err := svc.ToggleTrashByUUID(ctx, "33333333-3333-3333-3333-333333333333", 1); err == nil {
		t.Fatal("expected toggle trash by uuid not-found error")
	}
	if err := svc.ToggleFavoriteByUUID(ctx, "33333333-3333-3333-3333-333333333333", 1); err == nil {
		t.Fatal("expected toggle favorite by uuid not-found error")
	}
	if err := svc.DeleteSessionByUUID(ctx, "33333333-3333-3333-3333-333333333333", 1); err == nil {
		t.Fatal("expected delete session by uuid not-found error")
	}
}
