package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"
)

type aiSummaryPayload struct {
	Summary     string   `json:"summary"`
	MainTopics  []string `json:"main_topics"`
	KeyInsights []string `json:"key_insights"`
	ActionItems []string `json:"action_items"`
	Sentiment   string   `json:"sentiment"`
}

func normalizeSummarySentiment(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case "positive", "neutral", "negative", "mixed":
		return s
	default:
		return "neutral"
	}
}

func cleanAISummaryResponse(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}

func parseAISummaryPayload(raw string) (*aiSummaryPayload, string, error) {
	cleaned := cleanAISummaryResponse(raw)
	if cleaned == "" {
		return nil, "", errors.New("empty summary response")
	}

	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start >= 0 && end > start {
		cleaned = cleaned[start : end+1]
	}

	var payload aiSummaryPayload
	if err := json.Unmarshal([]byte(cleaned), &payload); err != nil {
		return nil, cleaned, err
	}

	payload.Summary = strings.TrimSpace(payload.Summary)
	payload.Sentiment = normalizeSummarySentiment(payload.Sentiment)
	return &payload, cleaned, nil
}

func (s *ChatService) GenerateSummary(ctx context.Context, sessionID, userID uint) (*dto.ChatSessionSummaryDTO, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if len(session.Messages) < 4 {
		return nil, errors.New("tidak cukup pesan untuk membuat ringkasan (minimal 4 pesan)")
	}

	if s.genaiModel == nil {
		return nil, errors.New("AI service tidak tersedia")
	}

	var convBuilder strings.Builder
	for _, msg := range session.Messages {
		role := "User"
		if msg.Role == model.ChatRoleAI {
			role = "AI"
		}
		convBuilder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	prompt := fmt.Sprintf(`Buatkan ringkasan untuk percakapan kesehatan mental berikut dalam Bahasa Indonesia. 
Format output HARUS dalam JSON dengan struktur:
{
  "summary": "ringkasan singkat 2-3 kalimat",
  "main_topics": ["topik1", "topik2"],
  "key_insights": ["insight1", "insight2"],
  "action_items": ["action1", "action2"],
  "sentiment": "positive/neutral/negative/mixed"
}

Percakapan:
%s

PENTING: Hanya kembalikan JSON, tanpa teks tambahan.`, convBuilder.String())

	resp, err := s.generateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("gagal generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("gagal mendapat respons AI")
	}

	aiResponse, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, errors.New("format respons AI tidak valid")
	}

	summaryResult := &dto.ChatSessionSummaryDTO{
		SessionID:   sessionID,
		GeneratedAt: time.Now(),
	}

	responseStr := string(aiResponse)
	payload, cleanResponse, err := parseAISummaryPayload(responseStr)
	toStore := cleanResponse

	if err == nil && payload != nil {
		summaryResult.Summary = payload.Summary
		summaryResult.MainTopics = payload.MainTopics
		summaryResult.KeyInsights = payload.KeyInsights
		summaryResult.ActionItems = payload.ActionItems
		summaryResult.Sentiment = payload.Sentiment

		if normalized, mErr := json.Marshal(payload); mErr == nil {
			toStore = string(normalized)
		}
	} else {
		summaryResult.Summary = cleanAISummaryResponse(responseStr)
		summaryResult.Sentiment = "neutral"
	}

	if err := s.sessionRepo.UpdateSummary(ctx, sessionID, toStore); err != nil {
		return nil, fmt.Errorf("gagal menyimpan summary: %w", err)
	}

	return summaryResult, nil
}

func (s *ChatService) GetSummary(ctx context.Context, sessionID, userID uint) (*dto.ChatSessionSummaryDTO, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if session.Summary == nil || *session.Summary == "" {
		return nil, errors.New("ringkasan belum tersedia")
	}

	generatedAt := session.SummaryGeneratedAt
	if generatedAt == nil {
		now := time.Now()
		generatedAt = &now
	}

	payload, _, err := parseAISummaryPayload(*session.Summary)
	if err == nil && payload != nil {
		return &dto.ChatSessionSummaryDTO{
			SessionID:   sessionID,
			Summary:     payload.Summary,
			MainTopics:  payload.MainTopics,
			KeyInsights: payload.KeyInsights,
			ActionItems: payload.ActionItems,
			Sentiment:   payload.Sentiment,
			GeneratedAt: *generatedAt,
		}, nil
	}

	return &dto.ChatSessionSummaryDTO{
		SessionID:   sessionID,
		Summary:     *session.Summary,
		Sentiment:   "neutral",
		GeneratedAt: *generatedAt,
	}, nil
}

func (s *ChatService) GetSuggestedPrompts(ctx context.Context, userID uint, params *dto.GetSuggestedPromptsRequest) (*dto.SuggestedPromptsResponse, error) {
	if userID == 0 {
		return nil, errors.New("unauthorized")
	}

	var prompts []dto.SuggestedPromptDTO

	hour := time.Now().Hour()
	timeOfDay := params.TimeOfDay
	if timeOfDay == "" {
		switch {
		case hour >= 5 && hour < 12:
			timeOfDay = "morning"
		case hour >= 12 && hour < 17:
			timeOfDay = "afternoon"
		case hour >= 17 && hour < 21:
			timeOfDay = "evening"
		default:
			timeOfDay = "night"
		}
	}

	if !params.HasMessages {
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "empty_1",
			Text:     "Aku merasa cemas akhir-akhir ini dan butuh berbicara",
			Category: "general",
			Icon:     "💭",
		})
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "empty_2",
			Text:     "Ceritakan teknik pernapasan untuk menenangkan diri",
			Category: "general",
			Icon:     "🧘",
		})
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "empty_3",
			Text:     "Bagaimana cara mengatasi stres di tempat kerja?",
			Category: "general",
			Icon:     "💼",
		})
	}

	switch timeOfDay {
	case "morning":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "time_morning",
			Text:     "Bagaimana cara memulai hari dengan pikiran positif?",
			Category: "time_based",
			Icon:     "🌅",
		})
	case "afternoon":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "time_afternoon",
			Text:     "Aku merasa lelah di tengah hari, apa yang bisa aku lakukan?",
			Category: "time_based",
			Icon:     "☀️",
		})
	case "evening":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "time_evening",
			Text:     "Bagaimana cara melepas stres setelah bekerja?",
			Category: "time_based",
			Icon:     "🌆",
		})
	case "night":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "time_night",
			Text:     "Aku sulit tidur, ada tips untuk relaksasi malam?",
			Category: "time_based",
			Icon:     "🌙",
		})
	}

	switch params.Mood {
	case "sad", "crying":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "mood_sad",
			Text:     "Aku sedang sedih dan butuh seseorang untuk mendengarkan",
			Category: "mood",
			Icon:     "💙",
		})
	case "angry":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "mood_angry",
			Text:     "Bagaimana cara mengelola kemarahan dengan sehat?",
			Category: "mood",
			Icon:     "🔥",
		})
	case "disappointed":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "mood_disappointed",
			Text:     "Aku merasa kecewa dan tidak tahu harus berbuat apa",
			Category: "mood",
			Icon:     "😔",
		})
	case "happy":
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "mood_happy",
			Text:     "Aku ingin berbagi kebahagiaan hari ini!",
			Category: "mood",
			Icon:     "😊",
		})
	}

	if params.HasMessages {
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "followup_1",
			Text:     "Bisa jelaskan lebih lanjut tentang perasaanku?",
			Category: "follow_up",
			Icon:     "🔍",
		})
		prompts = append(prompts, dto.SuggestedPromptDTO{
			ID:       "followup_2",
			Text:     "Apa langkah konkret yang bisa aku ambil?",
			Category: "follow_up",
			Icon:     "📋",
		})
	}

	return &dto.SuggestedPromptsResponse{
		Prompts: prompts,
	}, nil
}
