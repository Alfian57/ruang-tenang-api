package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestChatService_GenerateAIResponse(t *testing.T) {
	svc := &ChatService{}
	resp := svc.generateAIResponse(context.Background(), "halo")
	if strings.TrimSpace(resp) == "" {
		t.Fatal("expected non-empty AI response")
	}
	if !strings.Contains(resp, "💚") {
		t.Fatalf("expected heart emoji suffix, got: %s", resp)
	}
}

func TestChatService_LoadAIPrompt_DefaultAndCustom(t *testing.T) {
	svc := &ChatService{}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	defaultPrompt := svc.loadAIPrompt(context.Background())
	if !strings.Contains(defaultPrompt, "Ruang Tenang AI") {
		t.Fatalf("expected default prompt fallback, got: %s", defaultPrompt)
	}

	promptDir := filepath.Join(tmp, "prompts")
	if err := os.MkdirAll(promptDir, 0755); err != nil {
		t.Fatalf("mkdir prompts failed: %v", err)
	}
	content := `system:
  name: Test AI
  context: Context Test
  persona: Empatic
goals:
  - Goal 1
instructions: Selalu ramah
restrictions:
  allowed_topics: [a, b]
  forbidden_topics: [x]
  rejection_response: Tolak
security:
  ignore_prompt_injection: true
  rules: [Rule 1, Rule 2]
  injection_response: Aman
crisis_handling:
  description: desc
  self_harm_response: self-harm
  professional_disclaimer: disclaimer
`
	if err := os.WriteFile(filepath.Join(promptDir, "ai_prompt.yml"), []byte(content), 0644); err != nil {
		t.Fatalf("write prompt file failed: %v", err)
	}

	loaded := svc.loadAIPrompt(context.Background())
	if !strings.Contains(loaded, "Test AI") || !strings.Contains(loaded, "Goal 1") || !strings.Contains(loaded, "Rule 1") {
		t.Fatalf("expected loaded YAML prompt content, got: %s", loaded)
	}
}

func TestChatService_DetectCrisis(t *testing.T) {
	ctx := context.Background()

	// no moderation repo branch
	if got := (&ChatService{}).detectCrisis(ctx, "aku sedih"); got != nil {
		t.Fatal("expected nil when moderation repo is nil")
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.CrisisKeyword{}); err != nil {
		t.Fatalf("migrate crisis keyword failed: %v", err)
	}

	kw1 := model.CrisisKeyword{Keyword: "bunuh diri", Category: model.CrisisCategorySuicide, Severity: model.CrisisSeverityCritical, Language: "id", IsActive: true}
	kw2 := model.CrisisKeyword{Keyword: "putus asa", Category: model.CrisisCategorySevereDepression, Severity: model.CrisisSeverityHigh, Language: "id", IsActive: true}
	if err := db.Create(&kw1).Error; err != nil {
		t.Fatalf("create kw1 failed: %v", err)
	}
	if err := db.Create(&kw2).Error; err != nil {
		t.Fatalf("create kw2 failed: %v", err)
	}

	modRepo := repository.NewModerationRepository(db)
	svc := &ChatService{moderationRepo: modRepo}

	res := svc.detectCrisis(ctx, "aku merasa putus asa dan ingin bunuh diri")
	if res == nil || !res.IsCrisis {
		t.Fatal("expected crisis detection result")
	}
	if res.Severity != model.CrisisSeverityCritical || res.Category != model.CrisisCategorySuicide {
		t.Fatalf("unexpected severity/category: %s/%s", res.Severity, res.Category)
	}
	if len(res.Keywords) == 0 || !strings.Contains(res.CrisisResponse, "119 ext 8") {
		t.Fatalf("unexpected crisis response payload: %+v", res)
	}

	noMatch := svc.detectCrisis(ctx, "hari ini cuaca cerah")
	if noMatch != nil {
		t.Fatal("expected nil for non-crisis content")
	}
}
