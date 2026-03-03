package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func summaryResponse(text string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: &genai.Content{Parts: []genai.Part{genai.Text(text)}}},
		},
	}
}

func hasPromptID(prompts []dto.SuggestedPromptDTO, id string) bool {
	for _, p := range prompts {
		if p.ID == id {
			return true
		}
	}
	return false
}

func TestGetSuggestedPrompts_NoMessages(t *testing.T) {
	svc := &ChatService{}
	resp, err := svc.GetSuggestedPrompts(context.Background(), 1, &dto.GetSuggestedPromptsRequest{
		Mood:        "sad",
		TimeOfDay:   "morning",
		HasMessages: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Prompts) == 0 || len(resp.Prompts) > 6 {
		t.Fatalf("unexpected prompts length: %d", len(resp.Prompts))
	}
	if !hasPromptID(resp.Prompts, "empty_1") {
		t.Fatal("expected empty-state prompt")
	}
	if !hasPromptID(resp.Prompts, "time_morning") {
		t.Fatal("expected morning prompt")
	}
	if !hasPromptID(resp.Prompts, "mood_sad") {
		t.Fatal("expected mood sad prompt")
	}
}

func TestGetSuggestedPrompts_WithMessages(t *testing.T) {
	svc := &ChatService{}
	resp, err := svc.GetSuggestedPrompts(context.Background(), 1, &dto.GetSuggestedPromptsRequest{
		Mood:        "happy",
		TimeOfDay:   "night",
		HasMessages: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hasPromptID(resp.Prompts, "empty_1") {
		t.Fatal("did not expect empty-state prompts when has messages")
	}
	if !hasPromptID(resp.Prompts, "followup_1") || !hasPromptID(resp.Prompts, "followup_2") {
		t.Fatal("expected follow-up prompts")
	}
	if !hasPromptID(resp.Prompts, "time_night") {
		t.Fatal("expected night prompt")
	}
	if !hasPromptID(resp.Prompts, "mood_happy") {
		t.Fatal("expected happy prompt")
	}
}

func setupSummaryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	createSessions := `CREATE TABLE chat_sessions (
		id INTEGER PRIMARY KEY,
		uuid TEXT,
		user_id INTEGER,
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
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, title, created_at, updated_at) VALUES (1, '11111111-1111-1111-1111-111111111111', 7, 'session enough', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed session 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, title, created_at, updated_at) VALUES (2, '22222222-2222-2222-2222-222222222222', 7, 'session short', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed session 2: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_sessions (id, uuid, user_id, title, summary, created_at, updated_at) VALUES (3, '33333333-3333-3333-3333-333333333333', 7, 'session with summary', 'ringkasan tersimpan', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed session 3: %v", err)
	}

	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, created_at, updated_at) VALUES (1, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', 1, 'user', 'u1', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed msg 1: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, created_at, updated_at) VALUES (2, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', 1, 'ai', 'a1', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed msg 2: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, created_at, updated_at) VALUES (3, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', 1, 'user', 'u2', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed msg 3: %v", err)
	}
	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, created_at, updated_at) VALUES (4, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa4', 1, 'ai', 'a2', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed msg 4: %v", err)
	}

	if err := db.Exec(`INSERT INTO chat_messages (id, uuid, chat_session_id, role, content, created_at, updated_at) VALUES (5, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1', 2, 'user', 'only one', ?, ?)`, now, now).Error; err != nil {
		t.Fatalf("seed msg 5: %v", err)
	}

	return db
}

func TestGenerateSummary_Branches(t *testing.T) {
	db := setupSummaryDB(t)
	svc := &ChatService{sessionRepo: repository.NewChatSessionRepository(db)}

	_, err := svc.GenerateSummary(context.Background(), 999, 7)
	if err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found, got %v", err)
	}

	_, err = svc.GenerateSummary(context.Background(), 1, 8)
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized, got %v", err)
	}

	_, err = svc.GenerateSummary(context.Background(), 2, 7)
	if err == nil || err.Error() != "tidak cukup pesan untuk membuat ringkasan (minimal 4 pesan)" {
		t.Fatalf("expected insufficient messages error, got %v", err)
	}

	_, err = svc.GenerateSummary(context.Background(), 1, 7)
	if err == nil || err.Error() != "AI service tidak tersedia" {
		t.Fatalf("expected AI unavailable, got %v", err)
	}
}

func TestGenerateSummary_GenAIBranchesWithStub(t *testing.T) {
	db := setupSummaryDB(t)
	svc := &ChatService{sessionRepo: repository.NewChatSessionRepository(db), genaiModel: &genai.GenerativeModel{}}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return nil, errors.New("gen error")
	}
	_, err := svc.GenerateSummary(context.Background(), 1, 7)
	if err == nil || !strings.Contains(err.Error(), "gagal generate summary") {
		t.Fatalf("expected generate error branch, got %v", err)
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{}, nil
	}
	_, err = svc.GenerateSummary(context.Background(), 1, 7)
	if err == nil || err.Error() != "gagal mendapat respons AI" {
		t.Fatalf("expected empty candidate error, got %v", err)
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{genai.Blob{MIMEType: "text/plain", Data: []byte("x")}}}}},
		}, nil
	}
	_, err = svc.GenerateSummary(context.Background(), 1, 7)
	if err == nil || err.Error() != "format respons AI tidak valid" {
		t.Fatalf("expected invalid format error, got %v", err)
	}

	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return summaryResponse("```json\n{\"summary\":\"ok\"}\n```"), nil
	}
	res, err := svc.GenerateSummary(context.Background(), 1, 7)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if res == nil || res.Summary != "{\"summary\":\"ok\"}" {
		t.Fatalf("unexpected cleaned summary: %+v", res)
	}
}

func TestGenerateSummary_UpdateSummaryError(t *testing.T) {
	db := setupSummaryDB(t)
	if err := db.Exec(`CREATE TRIGGER fail_summary_update
	BEFORE UPDATE ON chat_sessions
	BEGIN
		SELECT RAISE(FAIL, 'summary update blocked');
	END;`).Error; err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	svc := &ChatService{sessionRepo: repository.NewChatSessionRepository(db), genaiModel: &genai.GenerativeModel{}}
	svc.generateContentFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return summaryResponse("{\"summary\":\"ok\"}"), nil
	}

	_, err := svc.GenerateSummary(context.Background(), 1, 7)
	if err == nil || !strings.Contains(err.Error(), "gagal menyimpan summary") {
		t.Fatalf("expected update summary error, got %v", err)
	}
}

func TestGetSummary_Branches(t *testing.T) {
	db := setupSummaryDB(t)
	now := time.Now().Add(-time.Hour)
	if err := db.Exec(`UPDATE chat_sessions SET summary_generated_at = ? WHERE id = 3`, now).Error; err != nil {
		t.Fatalf("set generated_at: %v", err)
	}

	svc := &ChatService{sessionRepo: repository.NewChatSessionRepository(db)}

	_, err := svc.GetSummary(context.Background(), 999, 7)
	if err == nil || err.Error() != "session not found" {
		t.Fatalf("expected session not found, got %v", err)
	}

	_, err = svc.GetSummary(context.Background(), 3, 8)
	if err == nil || err.Error() != "unauthorized" {
		t.Fatalf("expected unauthorized, got %v", err)
	}

	_, err = svc.GetSummary(context.Background(), 1, 7)
	if err == nil || err.Error() != "ringkasan belum tersedia" {
		t.Fatalf("expected no summary error, got %v", err)
	}

	res, err := svc.GetSummary(context.Background(), 3, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Summary != "ringkasan tersimpan" || res.SessionID != 3 {
		t.Fatalf("unexpected summary response: %+v", res)
	}
	if res.GeneratedAt.IsZero() {
		t.Fatal("expected generated_at to be set")
	}
}

func TestGetSummary_DefaultGeneratedAtWhenNil(t *testing.T) {
	db := setupSummaryDB(t)
	if err := db.Exec(`UPDATE chat_sessions SET summary_generated_at = NULL WHERE id = 3`).Error; err != nil {
		t.Fatalf("clear generated_at: %v", err)
	}

	svc := &ChatService{sessionRepo: repository.NewChatSessionRepository(db)}
	res, err := svc.GetSummary(context.Background(), 3, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GeneratedAt.IsZero() {
		t.Fatal("expected fallback generated_at")
	}
}

func TestGetSuggestedPrompts_MoreMoodAndTimeBranches(t *testing.T) {
	svc := &ChatService{}

	tests := []struct {
		name         string
		req          dto.GetSuggestedPromptsRequest
		expectIDs    []string
		notExpectIDs []string
	}{
		{
			name: "afternoon-angry-no-messages",
			req: dto.GetSuggestedPromptsRequest{
				Mood:        "angry",
				TimeOfDay:   "afternoon",
				HasMessages: false,
			},
			expectIDs: []string{"time_afternoon", "mood_angry", "empty_1"},
		},
		{
			name: "evening-disappointed-with-messages",
			req: dto.GetSuggestedPromptsRequest{
				Mood:        "disappointed",
				TimeOfDay:   "evening",
				HasMessages: true,
			},
			expectIDs:    []string{"time_evening", "mood_disappointed", "followup_1"},
			notExpectIDs: []string{"empty_1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetSuggestedPrompts(context.Background(), 1, &tc.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.Prompts) == 0 || len(resp.Prompts) > 6 {
				t.Fatalf("unexpected prompts length: %d", len(resp.Prompts))
			}
			for _, id := range tc.expectIDs {
				if !hasPromptID(resp.Prompts, id) {
					t.Fatalf("expected prompt id %s", id)
				}
			}
			for _, id := range tc.notExpectIDs {
				if hasPromptID(resp.Prompts, id) {
					t.Fatalf("did not expect prompt id %s", id)
				}
			}
		})
	}
}

func TestGetSuggestedPrompts_AutoTimeOfDayAndUnknownMoodBranch(t *testing.T) {
	svc := &ChatService{}

	resp, err := svc.GetSuggestedPrompts(context.Background(), 1, &dto.GetSuggestedPromptsRequest{
		Mood:        "neutral_unknown",
		TimeOfDay:   "",
		HasMessages: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !hasPromptID(resp.Prompts, "empty_1") {
		t.Fatal("expected empty-state prompt")
	}

	hour := time.Now().Hour()
	expectedTimePrompt := "time_night"
	switch {
	case hour >= 5 && hour < 12:
		expectedTimePrompt = "time_morning"
	case hour >= 12 && hour < 17:
		expectedTimePrompt = "time_afternoon"
	case hour >= 17 && hour < 21:
		expectedTimePrompt = "time_evening"
	}

	if !hasPromptID(resp.Prompts, expectedTimePrompt) {
		t.Fatalf("expected auto time prompt %s", expectedTimePrompt)
	}

	if hasPromptID(resp.Prompts, "mood_sad") || hasPromptID(resp.Prompts, "mood_angry") || hasPromptID(resp.Prompts, "mood_disappointed") || hasPromptID(resp.Prompts, "mood_happy") {
		t.Fatalf("did not expect mood-specific prompts for unknown mood, got %+v", resp.Prompts)
	}
}

func TestGetSuggestedPrompts_UnauthorizedUserIDBranch(t *testing.T) {
	svc := &ChatService{}

	resp, err := svc.GetSuggestedPrompts(context.Background(), 0, &dto.GetSuggestedPromptsRequest{HasMessages: true})
	if err == nil {
		t.Fatalf("expected unauthorized error, got resp=%+v", resp)
	}
}
