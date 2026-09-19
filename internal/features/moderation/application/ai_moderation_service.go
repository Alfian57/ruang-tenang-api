package application

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/ai"
	"github.com/Alfian57/ruang-tenang-api/prompts"

	"github.com/Alfian57/ruang-tenang-api/internal/features/moderation/infrastructure"
)

type AIModerationService struct {
	moderationRepo *infrastructure.ModerationRepository
	aiClient       ai.Client
	aiModel        string
	generateFn     func(ctx context.Context, prompt string) (*ai.CompletionResponse, error)
}

func NewAIModerationService(moderationRepo *infrastructure.ModerationRepository, aiClient ai.Client, aiModel string) *AIModerationService {
	return &AIModerationService{
		moderationRepo: moderationRepo,
		aiClient:       aiClient,
		aiModel:        aiModel,
	}
}

func (s *AIModerationService) generateContent(ctx context.Context, prompt string) (*ai.CompletionResponse, error) {
	if s.generateFn != nil {
		return s.generateFn(ctx, prompt)
	}
	if s.generateFn == nil && (s.aiClient == nil || !s.aiClient.IsConfigured()) {
		return nil, ai.ErrNotConfigured
	}
	return s.aiClient.Complete(ctx, ai.CompletionRequest{
		Model:          s.aiModel,
		Messages:       []ai.Message{{Role: "user", Content: prompt}},
		ResponseFormat: &ai.ResponseFormat{Type: "json_object"},
		MaxTokens:      2048,
	})
}

func moderationResponseText(resp *ai.CompletionResponse) (string, bool) {
	if resp == nil || len(resp.Choices) == 0 {
		return "", false
	}
	text := strings.TrimSpace(resp.Choices[0].Message.Content)
	return text, text != ""
}

// ModerateArticle uses AI to analyze article content for moderation
func (s *AIModerationService) ModerateArticle(ctx context.Context, title, content string) (*dto.AIModerationResult, error) {
	if s.generateFn == nil && (s.aiClient == nil || !s.aiClient.IsConfigured()) {
		// Fallback: auto-approve if AI is not available
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationApproved,
			Confidence: 0,
			Reasons:    []string{"AI moderation unavailable, auto-approved"},
		}, nil
	}

	prompt := prompts.Format("moderation", "article", title, content)

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		// Fallback to flagged for manual review if AI fails
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationFlagged,
			Confidence: 0,
			Reasons:    []string{"AI analysis failed, manual review required"},
		}, nil
	}

	responseText, ok := moderationResponseText(resp)
	if !ok {
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationFlagged,
			Confidence: 0,
			Reasons:    []string{"AI returned empty response, manual review required"},
		}, nil
	}

	// Parse JSON response
	var result struct {
		Status          string   `json:"status"`
		Confidence      float64  `json:"confidence"`
		Reasons         []string `json:"reasons"`
		FlagCategory    string   `json:"flag_category"`
		Severity        string   `json:"severity"`
		Suggestions     string   `json:"suggestions"`
		TriggerWarnings []string `json:"trigger_warnings"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		// If parsing fails, flag for manual review
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationFlagged,
			Confidence: 0,
			Reasons:    []string{"Failed to parse AI response", responseText},
		}, nil
	}

	// Map status string to constant
	var status model.ArticleModerationStatus
	switch result.Status {
	case "approved":
		status = model.ArticleModerationApproved
	case "rejected":
		status = model.ArticleModerationRejected
	default:
		status = model.ArticleModerationFlagged
	}

	return &dto.AIModerationResult{
		Status:       status,
		Confidence:   result.Confidence,
		Reasons:      result.Reasons,
		FlagCategory: result.FlagCategory,
		Severity:     result.Severity,
		Suggestions:  result.Suggestions,
	}, nil
}

// DetectCrisis checks message content for crisis keywords
func (s *AIModerationService) DetectCrisis(ctx context.Context, message string) (*model.CrisisDetectionResult, error) {
	keywords, err := s.moderationRepo.GetActiveCrisisKeywords(ctx, "id")
	if err != nil {
		return &model.CrisisDetectionResult{IsCrisis: false}, nil
	}

	messageLower := strings.ToLower(message)
	var detectedKeywords []string
	var highestSeverity model.CrisisSeverity = model.CrisisSeverityMedium
	var category model.CrisisCategory

	for _, kw := range keywords {
		if strings.Contains(messageLower, strings.ToLower(kw.Keyword)) {
			detectedKeywords = append(detectedKeywords, kw.Keyword)

			// Track highest severity
			if kw.Severity == model.CrisisSeverityCritical {
				highestSeverity = model.CrisisSeverityCritical
				category = kw.Category
			} else if kw.Severity == model.CrisisSeverityHigh && highestSeverity != model.CrisisSeverityCritical {
				highestSeverity = model.CrisisSeverityHigh
				category = kw.Category
			} else if category == "" {
				category = kw.Category
			}
		}
	}

	if len(detectedKeywords) == 0 {
		return &model.CrisisDetectionResult{IsCrisis: false}, nil
	}

	// Generate crisis response
	crisisResponse := s.generateCrisisResponse(ctx, category, highestSeverity)

	return &model.CrisisDetectionResult{
		IsCrisis:        true,
		Keywords:        detectedKeywords,
		Category:        category,
		Severity:        highestSeverity,
		CrisisResponse:  crisisResponse,
		EmergencyNumber: "119 ext 8", // Indonesia emergency number
	}, nil
}

// generateCrisisResponse creates appropriate crisis intervention message
func (s *AIModerationService) generateCrisisResponse(ctx context.Context, category model.CrisisCategory, severity model.CrisisSeverity) string {
	baseResponse := `Aku mendengarmu dan aku ingin kamu tahu bahwa perasaanmu valid. 💙

Tapi aku perlu bicara serius sebentar - sepertinya kamu sedang mengalami masa yang sangat berat. Aku AI dan kemampuanku terbatas untuk membantu dalam situasi seperti ini.

`

	var specificResponse string
	switch category {
	case model.CrisisCategorySuicide, model.CrisisCategorySelfHarm:
		specificResponse = `**Tolong hubungi bantuan profesional sekarang:**
- 🆘 Hotline Kesehatan Jiwa: **119 ext 8** (24 jam)
- 📞 Into The Light Indonesia: **021-78842580**
- 💬 Yayasan Pulih: **021-788-42580**

Jika kamu dalam bahaya segera, hubungi 112 atau pergi ke IGD rumah sakit terdekat.

`
	case model.CrisisCategorySevereDepression:
		specificResponse = `**Kamu tidak sendirian. Bantuan tersedia:**
- 🆘 Hotline Kesehatan Jiwa: **119 ext 8** (24 jam)
- 📞 Sejiwa (Kemenkes): **119 ext 8**
- 💬 Into The Light: **021-78842580**

Berbicara dengan profesional bisa sangat membantu.

`
	default:
		specificResponse = `**Bantuan tersedia untukmu:**
- 🆘 Hotline Kesehatan Jiwa: **119 ext 8**
- 📞 Sejiwa (Kemenkes): **119 ext 8**

`
	}

	closingResponse := `Kamu berharga dan pantas mendapat dukungan dari orang yang terlatih untuk membantu. Apakah ada seseorang yang kamu percaya - keluarga, teman, atau guru - yang bisa kamu hubungi sekarang?

Aku tetap di sini untuk menemanimu, tapi tolong pertimbangkan untuk menghubungi salah satu layanan di atas. Mereka benar-benar bisa membantu. 🤍`

	return baseResponse + specificResponse + closingResponse
}

// DetectTriggerWarnings uses AI to detect potential trigger content
func (s *AIModerationService) DetectTriggerWarnings(ctx context.Context, content string) ([]string, error) {
	if s.generateFn == nil && (s.aiClient == nil || !s.aiClient.IsConfigured()) {
		return []string{}, nil
	}

	prompt := prompts.Format("moderation", "trigger", content)

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		return []string{}, nil
	}

	responseText, ok := moderationResponseText(resp)
	if !ok {
		return []string{}, nil
	}

	var result struct {
		TriggerWarnings     []string `json:"trigger_warnings"`
		HasSensitiveContent bool     `json:"has_sensitive_content"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return []string{}, nil
	}

	return result.TriggerWarnings, nil
}

// AnalyzeForumContent analyzes forum post content for safety
func (s *AIModerationService) AnalyzeForumContent(ctx context.Context, content string) (bool, string, error) {
	if s.generateFn == nil && (s.aiClient == nil || !s.aiClient.IsConfigured()) {
		return false, "", nil
	}

	prompt := prompts.Format("moderation", "forum", content)

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		return false, "", nil
	}

	responseText, ok := moderationResponseText(resp)
	if !ok {
		return false, "", nil
	}

	var result struct {
		ShouldFlag bool   `json:"should_flag"`
		Reason     string `json:"reason"`
	}

	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return false, "", nil
	}

	return result.ShouldFlag, result.Reason, nil
}
