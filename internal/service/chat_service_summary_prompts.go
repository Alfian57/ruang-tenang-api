package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"
)

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

	resp, err := s.genaiModel.GenerateContent(ctx, genai.Text(prompt))
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

	cleanResponse := strings.TrimPrefix(responseStr, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	if err := s.sessionRepo.UpdateSummary(ctx, sessionID, cleanResponse); err != nil {
		return nil, fmt.Errorf("gagal menyimpan summary: %w", err)
	}

	summaryResult.Summary = cleanResponse
	summaryResult.Sentiment = "neutral"

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

	return &dto.ChatSessionSummaryDTO{
		SessionID:   sessionID,
		Summary:     *session.Summary,
		GeneratedAt: *generatedAt,
	}, nil
}

func (s *ChatService) GetSuggestedPrompts(ctx context.Context, userID uint, params *dto.GetSuggestedPromptsRequest) (*dto.SuggestedPromptsResponse, error) {
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

	if len(prompts) > 6 {
		prompts = prompts[:6]
	}

	return &dto.SuggestedPromptsResponse{
		Prompts: prompts,
	}, nil
}
