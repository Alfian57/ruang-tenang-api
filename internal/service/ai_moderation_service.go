package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIModerationService struct {
	moderationRepo *repository.ModerationRepository
	genaiClient    *genai.Client
	genaiModel     *genai.GenerativeModel
	generateFn     func(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error)
}

func NewAIModerationService(moderationRepo *repository.ModerationRepository, cfg *config.Config) *AIModerationService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	var model *genai.GenerativeModel
	if err == nil {
		model = client.GenerativeModel("gemini-flash-latest")
		// Configure for JSON output
		model.ResponseMIMEType = "application/json"
	} else {
		fmt.Printf("Failed to create Gemini client for moderation: %v\n", err)
	}

	return &AIModerationService{
		moderationRepo: moderationRepo,
		genaiClient:    client,
		genaiModel:     model,
	}
}

func (s *AIModerationService) generateContent(ctx context.Context, prompt string) (*genai.GenerateContentResponse, error) {
	if s.generateFn != nil {
		return s.generateFn(ctx, prompt)
	}
	return s.genaiModel.GenerateContent(ctx, genai.Text(prompt))
}

// ModerateArticle uses AI to analyze article content for moderation
func (s *AIModerationService) ModerateArticle(ctx context.Context, title, content string) (*dto.AIModerationResult, error) {
	if s.genaiModel == nil {
		// Fallback: auto-approve if AI is not available
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationApproved,
			Confidence: 0,
			Reasons:    []string{"AI moderation unavailable, auto-approved"},
		}, nil
	}

	ctx = context.Background()

	prompt := fmt.Sprintf(`Anda adalah sistem moderasi konten untuk platform kesehatan mental "Ruang Tenang".
Analisis artikel berikut dan tentukan apakah konten ini aman untuk dipublikasikan.

JUDUL: %s

KONTEN:
%s

---

Evaluasi konten berdasarkan kriteria berikut:
1. MISINFORMASI MEDIS: Apakah ada klaim medis yang salah atau berbahaya?
2. SARAN BERBAHAYA: Apakah ada saran yang bisa membahayakan pembaca?
3. KONTEN PEMICU (TRIGGER): Apakah ada konten tentang self-harm, bunuh diri, atau topik sensitif yang perlu peringatan?
4. SPAM/PROMOSI: Apakah konten ini spam atau promosi?
5. HATE SPEECH: Apakah ada ujaran kebencian atau diskriminasi?

Berikan respons dalam format JSON berikut:
{
  "status": "approved" | "flagged" | "rejected",
  "confidence": 0-100,
  "reasons": ["alasan1", "alasan2"],
  "flag_category": "kategori jika flagged/rejected",
  "severity": "low" | "medium" | "high",
  "suggestions": "saran untuk penulis jika perlu revisi",
  "trigger_warnings": ["self_harm", "suicide", "abuse", "trauma", "eating_disorder", "substance"]
}

Kriteria keputusan:
- APPROVED: Konten aman, informatif, dan sesuai guidelines
- FLAGGED: Perlu review manual karena ada konten sensitif tapi tidak jelas-jelas berbahaya
- REJECTED: Jelas melanggar guidelines (misinformasi medis, saran berbahaya, hate speech)

Jika konten mengandung topik sensitif tapi dibahas dengan cara yang supportive dan educational, set status ke "approved" tapi isi trigger_warnings yang sesuai.`, title, content)

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		// Fallback to flagged for manual review if AI fails
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationFlagged,
			Confidence: 0,
			Reasons:    []string{"AI analysis failed, manual review required"},
		}, nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return &dto.AIModerationResult{
			Status:     model.ArticleModerationFlagged,
			Confidence: 0,
			Reasons:    []string{"AI returned empty response, manual review required"},
		}, nil
	}

	// Parse JSON response
	responseText := ""
	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		responseText = string(txt)
	}

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
	if s.genaiModel == nil {
		return []string{}, nil
	}

	ctx = context.Background()

	prompt := fmt.Sprintf(`Analisis konten berikut dan identifikasi apakah mengandung topik sensitif yang memerlukan trigger warning.

KONTEN:
%s

---

Periksa apakah konten mengandung pembahasan tentang:
- self_harm: menyakiti diri sendiri, cutting, melukai diri
- suicide: bunuh diri, mengakhiri hidup, ingin mati
- abuse: kekerasan, pelecehan, penganiayaan
- trauma: pengalaman traumatis, PTSD
- eating_disorder: gangguan makan, anoreksia, bulimia
- substance: penyalahgunaan zat, narkoba, alkohol

Berikan respons dalam format JSON:
{
  "trigger_warnings": ["self_harm", "suicide", "abuse", "trauma", "eating_disorder", "substance"],
  "has_sensitive_content": true/false
}

Hanya masukkan kategori yang benar-benar ada dalam konten. Array kosong jika tidak ada.`, content)

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		return []string{}, nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return []string{}, nil
	}

	responseText := ""
	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		responseText = string(txt)
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
	if s.genaiModel == nil {
		return false, "", nil
	}

	ctx = context.Background()

	prompt := fmt.Sprintf(`Analisis postingan forum berikut untuk platform kesehatan mental.
Tentukan apakah konten ini perlu ditandai (flagged) untuk review moderator.

KONTEN:
%s

---

Flag konten jika mengandung:
1. Informasi medis yang menyesatkan
2. Saran berbahaya
3. Ujaran kebencian atau pelecehan
4. Spam atau promosi
5. Konten yang bisa memicu self-harm

Berikan respons dalam format JSON:
{
  "should_flag": true/false,
  "reason": "alasan jika perlu di-flag"
}`, content)

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		return false, "", nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return false, "", nil
	}

	responseText := ""
	if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
		responseText = string(txt)
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

// Close cleans up resources
func (s *AIModerationService) Close(ctx context.Context) error {
	if s.genaiClient != nil {
		return s.genaiClient.Close()
	}
	return nil
}
