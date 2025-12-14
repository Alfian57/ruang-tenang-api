package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type ChatService struct {
	sessionRepo *repositories.ChatSessionRepository
	messageRepo *repositories.ChatMessageRepository
}

func NewChatService(sessionRepo *repositories.ChatSessionRepository, messageRepo *repositories.ChatMessageRepository) *ChatService {
	return &ChatService{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
	}
}

func (s *ChatService) GetSessions(userID uint, params *dto.ChatSessionQueryParams) ([]dto.ChatSessionListDTO, int64, error) {
	sessions, total, err := s.sessionRepo.FindByUserID(userID, params.Filter, params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, 0, err
	}

	var result []dto.ChatSessionListDTO
	for _, session := range sessions {
		lastMsg := ""
		if len(session.Messages) > 0 {
			lastMsg = session.Messages[0].Content
			if len(lastMsg) > 50 {
				lastMsg = lastMsg[:50] + "..."
			}
		}

		result = append(result, dto.ChatSessionListDTO{
			ID:           session.ID,
			Title:        session.Title,
			IsBookmarked: session.IsBookmarked,
			IsFavorite:   session.IsFavorite,
			LastMessage:  lastMsg,
			CreatedAt:    session.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
			IsLiked:    msg.IsLiked,
			IsDisliked: msg.IsDisliked,
			CreatedAt:  msg.CreatedAt,
		})
	}

	return &dto.ChatSessionDTO{
		ID:           session.ID,
		Title:        session.Title,
		IsBookmarked: session.IsBookmarked,
		IsFavorite:   session.IsFavorite,
		Messages:     messages,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
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
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return nil, nil, errors.New("session not found")
	}

	if session.UserID != userID {
		return nil, nil, errors.New("unauthorized")
	}

	// Create user message
	userMsg := &models.ChatMessage{
		ChatSessionID: sessionID,
		Role:          models.ChatRoleUser,
		Content:       req.Content,
	}

	if err := s.messageRepo.Create(userMsg); err != nil {
		return nil, nil, err
	}

	// Generate AI response (placeholder - integrate with real AI later)
	aiResponse := s.generateAIResponse(req.Content)

	aiMsg := &models.ChatMessage{
		ChatSessionID: sessionID,
		Role:          models.ChatRoleAI,
		Content:       aiResponse,
	}

	if err := s.messageRepo.Create(aiMsg); err != nil {
		return nil, nil, err
	}

	// Update session timestamp
	session.UpdatedAt = time.Now()
	_ = s.sessionRepo.Update(session)

	return &dto.ChatMessageDTO{
			ID:        userMsg.ID,
			Role:      string(userMsg.Role),
			Content:   userMsg.Content,
			CreatedAt: userMsg.CreatedAt,
		}, &dto.ChatMessageDTO{
			ID:        aiMsg.ID,
			Role:      string(aiMsg.Role),
			Content:   aiMsg.Content,
			CreatedAt: aiMsg.CreatedAt,
		}, nil
}

func (s *ChatService) ToggleBookmark(sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleBookmark(sessionID)
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

func (s *ChatService) ToggleMessageLike(messageID uint) error {
	return s.messageRepo.ToggleLike(messageID)
}

func (s *ChatService) ToggleMessageDislike(messageID uint) error {
	return s.messageRepo.ToggleDislike(messageID)
}

// generateAIResponse generates a placeholder AI response
// TODO: Integrate with OpenAI/Gemini API
func (s *ChatService) generateAIResponse(userMessage string) string {
	responses := []string{
		"Terima kasih sudah berbagi. Saya di sini untuk mendengarkan kamu. Bagaimana perasaanmu saat ini?",
		"Saya mengerti. Ini pasti tidak mudah untukmu. Ceritakan lebih lanjut jika kamu mau.",
		"Perasaanmu valid dan penting. Apa yang bisa saya bantu untuk membuatmu merasa lebih baik?",
		"Saya senang kamu mau berbagi denganku. Ingat, kamu tidak sendiri dalam menghadapi ini.",
		"Terima kasih sudah mempercayai saya. Mari kita bicarakan apa yang sedang kamu rasakan.",
		"Saya di sini untukmu. Tidak apa-apa untuk merasa seperti ini. Apa yang ingin kamu ceritakan?",
		"Perasaanmu sangat berarti. Saya harap kamu tahu bahwa selalu ada harapan dan bantuan tersedia.",
		"Kamu sangat berani untuk berbagi ini. Bagaimana jika kita coba teknik pernapasan sederhana bersama?",
	}

	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%s 💚", responses[rand.Intn(len(responses))])
}
