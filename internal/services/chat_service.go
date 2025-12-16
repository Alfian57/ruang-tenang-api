package services

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ChatService struct {
	sessionRepo         *repositories.ChatSessionRepository
	messageRepo         *repositories.ChatMessageRepository
	genaiClient         *genai.Client
	genaiModel          *genai.GenerativeModel
	gamificationService *GamificationService
}

func NewChatService(sessionRepo *repositories.ChatSessionRepository, messageRepo *repositories.ChatMessageRepository, cfg *config.Config, gamificationService *GamificationService) *ChatService {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GeminiAPIKey))
	var model *genai.GenerativeModel
	if err == nil {
		// Use gemini-1.5-flash which is more stable and widely supported
		model = client.GenerativeModel("gemini-1.5-flash")
	} else {
		fmt.Printf("Failed to create Gemini client: %v\n", err)
	}

	return &ChatService{
		sessionRepo:         sessionRepo,
		messageRepo:         messageRepo,
		genaiClient:         client,
		genaiModel:          model,
		gamificationService: gamificationService,
	}
}

func (s *ChatService) GetSessions(userID uint, params dto.ChatSessionQueryParams) ([]dto.ChatSessionListDTO, int64, error) {
	sessions, total, err := s.sessionRepo.FindByUserID(userID, params.Filter, params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.ChatSessionListDTO
	for _, session := range sessions {
		lastMsg := ""
		if len(session.Messages) > 0 {
			lastMsg = session.Messages[0].Content
		}
		result = append(result, dto.ChatSessionListDTO{
			ID:          session.ID,
			Title:       session.Title,
			IsTrash:     session.IsTrash,
			IsFavorite:  session.IsFavorite,
			LastMessage: lastMsg,
			CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, total, nil
}

func (s *ChatService) GetSessionByID(id, userID uint) (*dto.ChatSessionDTO, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(id)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	var messages []dto.ChatMessageDTO
	for _, msg := range session.Messages {
		messages = append(messages, dto.ChatMessageDTO{
			ID:         msg.ID,
			Role:       string(msg.Role),
			Content:    msg.Content,
			Type:       msg.Type,
			IsLiked:    msg.IsLiked,
			IsDisliked: msg.IsDisliked,
			CreatedAt:  msg.CreatedAt,
		})
	}

	return &dto.ChatSessionDTO{
		ID:         session.ID,
		Title:      session.Title,
		IsTrash:    session.IsTrash,
		IsFavorite: session.IsFavorite,
		Messages:   messages,
		CreatedAt:  session.CreatedAt,
		UpdatedAt:  session.UpdatedAt,
	}, nil
}

func (s *ChatService) CreateSession(userID uint, req *dto.CreateChatSessionRequest) (*models.ChatSession, error) {
	session := &models.ChatSession{
		UserID: userID,
		Title:  req.Title,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *ChatService) SendMessage(sessionID, userID uint, req *dto.SendMessageRequest) (*dto.ChatMessageDTO, *dto.ChatMessageDTO, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: session not found: %w", err)
	}

	if session.UserID != userID {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: unauthorized access to session %d", sessionID)
	}

	// Determine message type, default to "text"
	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	// Create user message
	userMsg := &models.ChatMessage{
		ChatSessionID: sessionID,
		Role:          models.ChatRoleUser,
		Content:       req.Content,
		Type:          msgType,
	}

	if err := s.messageRepo.Create(userMsg); err != nil {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: failed to create user message: %w", err)
	}

	// Generate AI response
	aiResponseText := "Maaf, saya sedang mengalami gangguan koneksi. Silakan coba lagi nanti."
	if s.genaiModel != nil {
		ctx := context.Background()

		// Build history
		cs := s.genaiModel.StartChat()

		// Load system prompt from file
		systemPrompt := "Anda adalah asisten kesehatan mental yang empatik, suportif, dan menenangkan bernama Ruang Tenang AI. Tugas Anda adalah mendengarkan keluh kesah pengguna, memberikan validasi emosional, dan saran-saran praktis untuk manajemen stres atau kecemasan. Jangan memberikan diagnosis medis. Gunakan bahasa Indonesia yang sopan, hangat, dan tidak menghakimi."
		if promptData, err := os.ReadFile("prompts/ai_prompt.txt"); err == nil {
			systemPrompt = string(promptData)
		}

		// Note: gemini-pro text-only input often takes history by just appending.
		// However, creating a chat session properly is better.
		// We need to map our history to genai history.
		// For simplicity/safety with current SDK version, we'll just send the current message with system prompt prepended context if history is empty,
		// or iterate history.

		// Let's rely on StartChat and manually populate history if needed,
		// but simple call for now:
		cs.History = []*genai.Content{}
		// Prepend system prompt as the first part of context if possible, or just instruction.
		// Simplest valid approach for mental health context:

		// Map existing messages to history
		// Limit to last 10 messages for context window efficiency
		startIdx := 0
		if len(session.Messages) > 10 {
			startIdx = len(session.Messages) - 10
		}

		// Add system instruction as first user part conceptually (or relies on model instruction)
		// For this implementation, we will append recent history.

		// Add System Prompt as the very first history item from "user" role to set behavior
		cs.History = append(cs.History, &genai.Content{
			Role: "user",
			Parts: []genai.Part{
				genai.Text(systemPrompt),
			},
		})
		cs.History = append(cs.History, &genai.Content{
			Role: "model",
			Parts: []genai.Part{
				genai.Text("Baik, saya mengerti. Saya siap mendengarkan dan membantu Anda dengan penuh empati."),
			},
		})

		for i := startIdx; i < len(session.Messages); i++ {
			msg := session.Messages[i]
			role := "user"
			if msg.Role == models.ChatRoleAI {
				role = "model"
			}
			cs.History = append(cs.History, &genai.Content{
				Role: role,
				Parts: []genai.Part{
					genai.Text(msg.Content),
				},
			})
		}

		resp, err := cs.SendMessage(ctx, genai.Text(req.Content))
		if err == nil && len(resp.Candidates) > 0 {
			if len(resp.Candidates[0].Content.Parts) > 0 {
				if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
					aiResponseText = string(txt)
				}
			}
		} else {
			fmt.Printf("Gemini Error: %v\n", err)
		}
	}

	aiMsg := &models.ChatMessage{
		ChatSessionID: sessionID,
		Role:          models.ChatRoleAI,
		Content:       aiResponseText,
	}

	if err := s.messageRepo.Create(aiMsg); err != nil {
		return nil, nil, err
	}

	// Update session timestamp
	session.UpdatedAt = time.Now()
	_ = s.sessionRepo.Update(session)

	// Award EXP
	_ = s.gamificationService.AwardExp(userID, "chat_ai", 10) // Should use constant, importing pkg/gamification

	return &dto.ChatMessageDTO{
			ID:        userMsg.ID,
			Role:      string(userMsg.Role),
			Content:   userMsg.Content,
			Type:      userMsg.Type,
			CreatedAt: userMsg.CreatedAt,
		}, &dto.ChatMessageDTO{
			ID:        aiMsg.ID,
			Role:      string(aiMsg.Role),
			Content:   aiMsg.Content,
			Type:      "text",
			CreatedAt: aiMsg.CreatedAt,
		}, nil
}

func (s *ChatService) ToggleTrash(sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleTrash(sessionID)
}

func (s *ChatService) ToggleFavorite(sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleFavorite(sessionID)
}

func (s *ChatService) DeleteSession(sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.Delete(sessionID)
}

func (s *ChatService) ToggleMessageLike(messageID, userID uint) error {
	// Verification logic could be added here (e.g., check if message belongs to user's session)
	// For now, assuming ID access check is sufficient or will be handled by repo finding
	return s.messageRepo.ToggleLike(messageID)
}

func (s *ChatService) ToggleMessageDislike(messageID, userID uint) error {
	return s.messageRepo.ToggleDislike(messageID)
}

// generateAIResponse generates a placeholder AI response
// TODO: Integrate with OpenAI/Gemini API
func (s *ChatService) generateAIResponse(userMessage string) string {
	responses := []string{
		"Terima kasih sudah mempercayai saya. Mari kita bicarakan apa yang sedang kamu rasakan.",
		"Saya di sini untukmu. Tidak apa-apa untuk merasa seperti ini. Apa yang ingin kamu ceritakan?",
		"Perasaanmu sangat berarti. Saya harap kamu tahu bahwa selalu ada harapan dan bantuan tersedia.",
		"Kamu sangat berani untuk berbagi ini. Bagaimana jika kita coba teknik pernapasan sederhana bersama?",
	}

	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s 💚", responses[rand.Intn(len(responses))])
}
