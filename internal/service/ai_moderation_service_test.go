package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func mockGenAIResponse(text string) *genai.GenerateContentResponse {
	return &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: &genai.Content{Parts: []genai.Part{genai.Text(text)}}},
		},
	}
}

func setupModerationRepoForAIService(t *testing.T) *repository.ModerationRepository {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.CrisisKeyword{}); err != nil {
		t.Fatalf("failed to migrate crisis keyword table: %v", err)
	}
	return repository.NewModerationRepository(db)
}

func TestAIModerationService_FallbacksWithoutAIModel(t *testing.T) {
	svc := &AIModerationService{}

	result, err := svc.ModerateArticle(context.Background(), "Judul", "Konten")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != model.ArticleModerationApproved {
		t.Fatalf("expected approved fallback, got %s", result.Status)
	}

	warnings, err := svc.DetectTriggerWarnings(context.Background(), "konten biasa")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected empty warnings fallback, got %v", warnings)
	}

	flagged, reason, err := svc.AnalyzeForumContent(context.Background(), "halo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flagged || reason != "" {
		t.Fatalf("expected non-flag fallback, got flagged=%v reason=%q", flagged, reason)
	}

	if err := svc.Close(context.Background()); err != nil {
		t.Fatalf("expected close nil, got %v", err)
	}
}

func TestAIModerationService_DetectCrisis(t *testing.T) {
	repo := setupModerationRepoForAIService(t)
	ctx := context.Background()

	if err := repo.CreateCrisisKeyword(ctx, &model.CrisisKeyword{Keyword: "bunuh diri", Category: model.CrisisCategorySuicide, Severity: model.CrisisSeverityCritical, Language: "id", IsActive: true}); err != nil {
		t.Fatalf("failed to seed keyword: %v", err)
	}
	if err := repo.CreateCrisisKeyword(ctx, &model.CrisisKeyword{Keyword: "putus asa", Category: model.CrisisCategorySevereDepression, Severity: model.CrisisSeverityHigh, Language: "id", IsActive: true}); err != nil {
		t.Fatalf("failed to seed keyword: %v", err)
	}

	svc := &AIModerationService{moderationRepo: repo}

	none, err := svc.DetectCrisis(ctx, "aku baik-baik saja")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if none.IsCrisis {
		t.Fatal("expected non crisis result")
	}

	res, err := svc.DetectCrisis(ctx, "aku merasa putus asa dan ingin bunuh diri")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.IsCrisis {
		t.Fatal("expected crisis result")
	}
	if res.Severity != model.CrisisSeverityCritical {
		t.Fatalf("expected critical severity, got %s", res.Severity)
	}
	if res.Category != model.CrisisCategorySuicide {
		t.Fatalf("expected suicide category, got %s", res.Category)
	}
	if !strings.Contains(res.CrisisResponse, "119 ext 8") {
		t.Fatalf("expected hotline in response, got %q", res.CrisisResponse)
	}
}

func TestAIModerationService_generateCrisisResponseCategories(t *testing.T) {
	svc := &AIModerationService{}

	resp1 := svc.generateCrisisResponse(context.Background(), model.CrisisCategorySelfHarm, model.CrisisSeverityHigh)
	if !strings.Contains(resp1, "Hotline Kesehatan Jiwa") {
		t.Fatalf("expected hotline in self-harm response, got %q", resp1)
	}

	resp2 := svc.generateCrisisResponse(context.Background(), model.CrisisCategorySevereDepression, model.CrisisSeverityMedium)
	if !strings.Contains(resp2, "Kamu tidak sendirian") {
		t.Fatalf("expected depression messaging, got %q", resp2)
	}

	resp3 := svc.generateCrisisResponse(context.Background(), model.CrisisCategoryEmergency, model.CrisisSeverityHigh)
	if !strings.Contains(resp3, "Bantuan tersedia untukmu") {
		t.Fatalf("expected default emergency messaging, got %q", resp3)
	}
}

func TestNewAIModerationService_Constructs(t *testing.T) {
	svc := NewAIModerationService(nil, &config.Config{GeminiAPIKey: ""})
	if svc == nil {
		t.Fatal("expected service instance")
	}

	if err := svc.Close(context.Background()); err != nil {
		t.Fatalf("close should not error: %v", err)
	}
}

func TestNewAIModerationService_WithAPIKey_InitializesModel(t *testing.T) {
	repo := setupModerationRepoForAIService(t)
	svc := NewAIModerationService(repo, &config.Config{GeminiAPIKey: "dummy-test-key"})
	if svc == nil {
		t.Fatal("expected service instance")
	}
	if svc.genaiClient == nil {
		t.Fatal("expected genai client to be initialized")
	}
	if svc.genaiModel == nil {
		t.Fatal("expected genai model to be initialized")
	}
	if svc.genaiModel.ResponseMIMEType != "application/json" {
		t.Fatalf("expected json mime type, got %q", svc.genaiModel.ResponseMIMEType)
	}
	if err := svc.Close(context.Background()); err != nil {
		t.Fatalf("close should not error: %v", err)
	}
}

func TestAIModerationService_DetectCrisis_RepoErrorFallback(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// no crisis_keywords table created -> repo method returns error
	repo := repository.NewModerationRepository(db)
	svc := &AIModerationService{moderationRepo: repo}

	res, err := svc.DetectCrisis(context.Background(), "aku ingin menyerah")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.IsCrisis {
		t.Fatal("expected non-crisis fallback when repo fails")
	}
}

func TestAIModerationService_NonNilModelErrorFallbackBranches(t *testing.T) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("dummy-test-key"))
	if err != nil {
		t.Skipf("unable to construct genai client in this environment: %v", err)
	}
	defer client.Close()

	repo := setupModerationRepoForAIService(t)
	svc := &AIModerationService{
		moderationRepo: repo,
		genaiClient:    client,
		genaiModel:     client.GenerativeModel("gemini-flash-latest"),
	}

	res, err := svc.ModerateArticle(ctx, "Judul Tes", "Konten tes untuk fallback error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != model.ArticleModerationFlagged {
		t.Fatalf("expected flagged fallback when AI call fails, got %s", res.Status)
	}
	if len(res.Reasons) == 0 {
		t.Fatalf("expected fallback reason when AI call fails")
	}

	warnings, err := svc.DetectTriggerWarnings(ctx, "konten dengan kata sensitif")
	if err != nil {
		t.Fatalf("unexpected error on detect warnings fallback: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("expected empty warnings on AI error fallback, got %v", warnings)
	}

	flagged, reason, err := svc.AnalyzeForumContent(ctx, "konten forum")
	if err != nil {
		t.Fatalf("unexpected error on forum analyze fallback: %v", err)
	}
	if flagged || reason != "" {
		t.Fatalf("expected false/empty fallback on AI error, got flagged=%v reason=%q", flagged, reason)
	}
}

func TestAIModerationService_Close_WithClient(t *testing.T) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey("dummy-test-key"))
	if err != nil {
		t.Skipf("unable to construct genai client in this environment: %v", err)
	}

	svc := &AIModerationService{genaiClient: client}
	if err := svc.Close(ctx); err != nil {
		t.Fatalf("expected close nil, got %v", err)
	}
}

func TestAIModerationService_ModerateArticle_ParseBranches(t *testing.T) {
	svc := &AIModerationService{genaiModel: &genai.GenerativeModel{}}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return nil, errors.New("gen fail")
	}
	res, err := svc.ModerateArticle(context.Background(), "t", "c")
	if err != nil || res.Status != model.ArticleModerationFlagged {
		t.Fatalf("expected flagged fallback on generate error, got res=%+v err=%v", res, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return &genai.GenerateContentResponse{}, nil
	}
	res, err = svc.ModerateArticle(context.Background(), "t", "c")
	if err != nil || res.Status != model.ArticleModerationFlagged {
		t.Fatalf("expected flagged fallback on empty response, got res=%+v err=%v", res, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse("not-json"), nil
	}
	res, err = svc.ModerateArticle(context.Background(), "t", "c")
	if err != nil || res.Status != model.ArticleModerationFlagged {
		t.Fatalf("expected flagged fallback on invalid json, got res=%+v err=%v", res, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse(`{"status":"approved","confidence":91,"reasons":["ok"],"flag_category":"","severity":"low","suggestions":""}`), nil
	}
	res, err = svc.ModerateArticle(context.Background(), "t", "c")
	if err != nil || res.Status != model.ArticleModerationApproved {
		t.Fatalf("expected approved mapping, got res=%+v err=%v", res, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse(`{"status":"rejected","confidence":81,"reasons":["bad"],"flag_category":"spam","severity":"high","suggestions":"fix"}`), nil
	}
	res, err = svc.ModerateArticle(context.Background(), "t", "c")
	if err != nil || res.Status != model.ArticleModerationRejected {
		t.Fatalf("expected rejected mapping, got res=%+v err=%v", res, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse(`{"status":"unknown","confidence":50,"reasons":["x"]}`), nil
	}
	res, err = svc.ModerateArticle(context.Background(), "t", "c")
	if err != nil || res.Status != model.ArticleModerationFlagged {
		t.Fatalf("expected default flagged mapping, got res=%+v err=%v", res, err)
	}
}

func TestAIModerationService_TriggerAndForumParseBranches(t *testing.T) {
	svc := &AIModerationService{genaiModel: &genai.GenerativeModel{}}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return nil, errors.New("gen fail")
	}
	warns, err := svc.DetectTriggerWarnings(context.Background(), "x")
	if err != nil || len(warns) != 0 {
		t.Fatalf("expected empty warnings on error fallback, got %v err=%v", warns, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse("bad-json"), nil
	}
	warns, err = svc.DetectTriggerWarnings(context.Background(), "x")
	if err != nil || len(warns) != 0 {
		t.Fatalf("expected empty warnings on parse fallback, got %v err=%v", warns, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse(`{"trigger_warnings":["self_harm","trauma"],"has_sensitive_content":true}`), nil
	}
	warns, err = svc.DetectTriggerWarnings(context.Background(), "x")
	if err != nil || len(warns) != 2 {
		t.Fatalf("expected parsed trigger warnings, got %v err=%v", warns, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse(`{"should_flag":true,"reason":"spam"}`), nil
	}
	flagged, reason, err := svc.AnalyzeForumContent(context.Background(), "x")
	if err != nil || !flagged || reason != "spam" {
		t.Fatalf("expected parsed forum analysis, got flagged=%v reason=%q err=%v", flagged, reason, err)
	}

	svc.generateFn = func(context.Context, string) (*genai.GenerateContentResponse, error) {
		return mockGenAIResponse("bad-json"), nil
	}
	flagged, reason, err = svc.AnalyzeForumContent(context.Background(), "x")
	if err != nil || flagged || reason != "" {
		t.Fatalf("expected parse fallback false/empty, got flagged=%v reason=%q err=%v", flagged, reason, err)
	}
}
