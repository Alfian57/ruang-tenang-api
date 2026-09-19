package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/ai"

	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/Alfian57/ruang-tenang-api/prompts"
	"go.uber.org/zap"
)

func (s *ChatService) getJournalContext(ctx context.Context, userID uint, chatSessionID uint, query string) string {
	if s.journalRepo == nil || s.journalSettingsRepo == nil {
		return ""
	}

	settings, err := s.journalSettingsRepo.FindByUserID(ctx, userID)
	if err != nil || !settings.AllowAIAccess {
		return ""
	}

	var journals []model.Journal

	if query != "" {
		journals, err = s.journalRepo.FindRelevantForAIContext(ctx, userID, query, settings.AIContextMaxEntries)
	} else {
		journals, err = s.journalRepo.FindForAIContext(ctx, userID, settings.AIContextDays, settings.AIContextMaxEntries)
	}

	if err != nil || len(journals) == 0 {
		return ""
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("\n\n=== KONTEKS JURNAL PRIBADI USER ===\n")
	contextBuilder.WriteString("(User telah mengizinkan Anda membaca jurnal mereka untuk memberikan dukungan yang lebih personal)\n\n")

	for _, j := range journals {
		if s.journalAccessLogRepo != nil {
			contextType := "chat_context"
			if query != "" {
				contextType = "query_context"
			}

			log := &model.JournalAIAccessLog{
				UserID:        userID,
				JournalID:     j.ID,
				ChatSessionID: &chatSessionID,
				ContextType:   contextType,
				AccessedAt:    time.Now(),
			}
			s.journalAccessLogRepo.Create(ctx, log)
		}

		s.journalRepo.UpdateAIAccessedAt(ctx, j.ID)

		contextBuilder.WriteString(fmt.Sprintf("📅 %s", j.CreatedAt.Format("2 January 2006")))
		if j.Title != "" {
			contextBuilder.WriteString(fmt.Sprintf(" - %s", j.Title))
		}
		contextBuilder.WriteString("\n")

		if j.Mood != nil {
			contextBuilder.WriteString(fmt.Sprintf("Mood: %s %s\n", j.Mood.GetMoodEmoji(), j.Mood.Mood))
		}

		if j.Summary != "" {
			contextBuilder.WriteString(fmt.Sprintf("Ringkasan: %s\n", j.Summary))
		}

		content := j.Content
		if len(content) > 500 {
			content = content[:500] + "..."
		}
		contextBuilder.WriteString(content)
		contextBuilder.WriteString("\n\n")
	}

	contextBuilder.WriteString("=== AKHIR KONTEKS JURNAL ===\n")
	contextBuilder.WriteString("Gunakan informasi ini untuk memberikan respons yang lebih personal dan empati. ")
	contextBuilder.WriteString("Jika relevan, Anda bisa merujuk ke apa yang user tulis di jurnal. ")
	contextBuilder.WriteString("Jangan secara eksplisit menyebut 'saya membaca jurnal Anda' kecuali memang sangat relevan.\n\n")

	return contextBuilder.String()
}

func (s *ChatService) GetSessions(ctx context.Context, userID uint, params dto.ChatSessionQueryParams) ([]dto.ChatSessionListDTO, int64, error) {
	sessions, total, err := s.sessionRepo.FindByUserID(ctx, userID, params.Filter, params.Search, params.FolderID, params.Page, params.Limit)
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
			UUID:        session.UUID.String(),
			Title:       session.Title,
			FolderID:    session.FolderID,
			IsTrash:     session.IsTrash,
			IsFavorite:  session.IsFavorite,
			HasSummary:  session.Summary != nil && *session.Summary != "",
			LastMessage: lastMsg,
			CreatedAt:   session.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, total, nil
}

func (s *ChatService) GetSessionByID(ctx context.Context, id, userID uint) (*dto.ChatSessionDTO, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(ctx, id)
	if err != nil {
		return nil, err
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	var messages []dto.ChatMessageDTO
	var pinnedMessages []dto.ChatMessageDTO
	for _, msg := range session.Messages {
		msgDTO := dto.ChatMessageDTO{
			ID:         msg.ID,
			Role:       string(msg.Role),
			Content:    msg.Content,
			Type:       msg.Type,
			IsLiked:    msg.IsLiked,
			IsDisliked: msg.IsDisliked,
			IsPinned:   msg.IsPinned,
			CreatedAt:  msg.CreatedAt,
		}
		messages = append(messages, msgDTO)
		if msg.IsPinned {
			pinnedMessages = append(pinnedMessages, msgDTO)
		}
	}

	var folderName string
	if session.Folder != nil {
		folderName = session.Folder.Name
	}

	return &dto.ChatSessionDTO{
		ID:                 session.ID,
		UUID:               session.UUID.String(),
		Title:              session.Title,
		FolderID:           session.FolderID,
		FolderName:         folderName,
		Summary:            session.Summary,
		SummaryGeneratedAt: session.SummaryGeneratedAt,
		IsTrash:            session.IsTrash,
		IsFavorite:         session.IsFavorite,
		Messages:           messages,
		PinnedMessages:     pinnedMessages,
		CreatedAt:          session.CreatedAt,
		UpdatedAt:          session.UpdatedAt,
	}, nil
}

// defaultSessionTitle is used when a session is created without a title.
// The real title is generated automatically from the first user message.
const defaultSessionTitle = "Obrolan Baru"

func (s *ChatService) CreateSession(ctx context.Context, userID uint, req *dto.CreateChatSessionRequest) (*model.ChatSession, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = defaultSessionTitle
	}

	session := &model.ChatSession{
		UserID:   userID,
		Title:    title,
		FolderID: req.FolderID,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

// generateSessionTitle creates a concise chat title from the first user message.
// It first tries the AI model for a natural, summarized title and falls back to
// a trimmed version of the message when the AI is unavailable or fails.
func (s *ChatService) generateSessionTitle(ctx context.Context, firstMessage string) string {
	cleaned := strings.TrimSpace(firstMessage)
	if cleaned == "" {
		return defaultSessionTitle
	}

	fallback := buildFallbackTitle(cleaned)

	if !s.modelAvailable() {
		return fallback
	}

	titleCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	prompt := prompts.Format("chat", "title", cleaned)

	resp, err := s.generateContent(titleCtx, prompt)
	if err != nil || resp == nil || len(resp.Choices) == 0 {
		return fallback
	}

	title := strings.TrimSpace(resp.Choices[0].Message.Content)
	if title == "" {
		return fallback
	}

	return sanitizeGeneratedTitle(title, fallback)
}

// buildFallbackTitle trims a message into a short, single-line title.
func buildFallbackTitle(message string) string {
	normalized := strings.Join(strings.Fields(message), " ")
	if normalized == "" {
		return defaultSessionTitle
	}

	const maxLen = 60
	runes := []rune(normalized)
	if len(runes) <= maxLen {
		return normalized
	}

	return strings.TrimSpace(string(runes[:maxLen])) + "..."
}

// sanitizeGeneratedTitle cleans an AI-generated title and enforces a max length.
func sanitizeGeneratedTitle(raw, fallback string) string {
	title := strings.TrimSpace(raw)
	title = strings.Trim(title, "\"'`")
	title = strings.TrimSpace(strings.TrimPrefix(title, "Judul:"))
	title = strings.Join(strings.Fields(title), " ")

	if title == "" {
		return fallback
	}

	const maxLen = 80
	runes := []rune(title)
	if len(runes) > maxLen {
		title = strings.TrimSpace(string(runes[:maxLen]))
	}

	return title
}

func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID uint, req *dto.SendMessageRequest) (*dto.ChatMessageDTO, *dto.ChatMessageDTO, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: session not found: %w", err)
	}

	if session.UserID != userID {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: unauthorized access to session %d", sessionID)
	}

	// Detect whether this is the first message so we can auto-generate the
	// session title from it.
	isFirstMessage := len(session.Messages) == 0

	if s.chatQuotaChecker != nil {
		quotaResult, quotaErr := s.chatQuotaChecker.ConsumeChatQuota(ctx, userID)
		if quotaErr != nil {
			if quotaResult != nil && !quotaResult.Allowed {
				return nil, nil, ErrDailyChatQuotaExceeded
			}
			return nil, nil, quotaErr
		}

		if quotaResult != nil && !quotaResult.Allowed {
			return nil, nil, ErrDailyChatQuotaExceeded
		}
	}

	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}
	aiInputContent := req.Content
	if msgType == "audio" {
		aiInputContent = "Pengguna mengirim pesan suara tetapi transkripsi tidak tersedia. Minta pengguna menuliskan inti pesan dengan singkat."
	}

	userMsg := &model.ChatMessage{
		ChatSessionID: sessionID,
		Role:          model.ChatRoleUser,
		Content:       req.Content,
		Type:          msgType,
	}

	if err := s.messageRepo.Create(ctx, userMsg); err != nil {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: failed to create user message: %w", err)
	}

	var crisisDetected *model.CrisisDetectionResult
	if s.moderationRepo != nil {
		crisisDetected = s.detectCrisis(ctx, aiInputContent)
	}

	aiResponseText := "Maaf, saya sedang mengalami gangguan koneksi. Silakan coba lagi nanti."

	if crisisDetected != nil && crisisDetected.IsCrisis {
		aiResponseText = crisisDetected.CrisisResponse
	} else if s.modelAvailable() {
		// Bound the DeepSeek call with the request context so client cancellation
		// and server-side timeouts propagate.
		ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		preferences := s.resolveContextPreferences(session, req.Context)
		userMessageCount := 1
		for _, msg := range session.Messages {
			if msg.Role == model.ChatRoleUser {
				userMessageCount++
			}
		}

		systemPrompt := s.loadAIPrompt(ctx)
		systemPrompt += s.buildDynamicContextPrompt(ctx, session, userID, req)

		if preferences.EnableJournalContext {
			var journalQuery string
			checkJournalRegex := regexp.MustCompile(`(?i)^(?:cek|check)\s+(?:jurnal|journal)\s+(?:saya\s+)?(?:tentang|about)\s+(.+)`)
			matches := checkJournalRegex.FindStringSubmatch(aiInputContent)
			if len(matches) > 1 {
				journalQuery = strings.TrimSpace(matches[1])
			}

			journalContext := s.getJournalContext(ctx, userID, sessionID, journalQuery)
			if journalContext != "" {
				systemPrompt += journalContext
			}
		}

		startIdx := 0
		if len(session.Messages) > 10 {
			startIdx = len(session.Messages) - 10
		}

		if s.generateChatReplyFn != nil {
			reply, err := s.generateChatReplyFn(ctx, systemPrompt, session.Messages[startIdx:], aiInputContent)
			if err == nil && reply != "" {
				aiResponseText = reply
			} else {
				logger.Warn("DeepSeek reply failed", zap.Error(err))
			}
		} else if s.aiClient != nil && s.aiClient.IsConfigured() {
			messages := []ai.Message{{Role: "system", Content: systemPrompt}}
			for i := startIdx; i < len(session.Messages); i++ {
				msg := session.Messages[i]
				role := "user"
				if msg.Role == model.ChatRoleAI {
					role = "assistant"
				}
				messages = append(messages, ai.Message{
					Role:    role,
					Content: msg.Content,
				})
			}
			messages = append(messages, ai.Message{Role: "user", Content: aiInputContent})

			request := ai.CompletionRequest{
				Model:      s.modelName,
				Messages:   messages,
				Tools:      s.buildRAGTools(),
				ToolChoice: "auto",
			}
			for i := 0; i < 3; i++ {
				resp, err := s.aiClient.Complete(ctx, request)
				if err != nil {
					logger.Warn("DeepSeek reply failed", zap.Error(err))
					break
				}
				if resp == nil || len(resp.Choices) == 0 {
					break
				}

				assistant := resp.Choices[0].Message
				if len(assistant.ToolCalls) == 0 {
					if strings.TrimSpace(assistant.Content) != "" {
						aiResponseText = assistant.Content
					}
					break
				}

				messages = append(messages, assistant)
				for _, toolCall := range assistant.ToolCalls {
					functionCall := toolCall.Function
					args := map[string]any{}
					if strings.TrimSpace(functionCall.Arguments) != "" {
						if err := json.Unmarshal([]byte(functionCall.Arguments), &args); err != nil {
							logger.Warn("DeepSeek tool arguments were invalid", zap.Error(err))
						}
					}
					functionCall.Args = args
					result := s.handleFunctionCall(ctx, functionCall, userID, preferences, userMessageCount)
					resultContent, marshalErr := json.Marshal(result.Response)
					if marshalErr != nil {
						resultContent = []byte(`{"result":"Fungsi tidak dapat mengembalikan hasil."}`)
					}
					messages = append(messages, ai.Message{
						Role:       "tool",
						ToolCallID: toolCall.ID,
						Content:    string(resultContent),
					})
				}
				request.Messages = messages
			}
		}
	}

	aiMsg := &model.ChatMessage{
		ChatSessionID: sessionID,
		Role:          model.ChatRoleAI,
		Content:       aiResponseText,
	}

	if err := s.messageRepo.Create(ctx, aiMsg); err != nil {
		return nil, nil, err
	}

	session.UpdatedAt = time.Now()

	// Auto-generate the session title from the first user message. For audio
	// messages, use the fallback text instead of the raw audio URL because
	// server-side transcription is intentionally not configured.
	if isFirstMessage {
		titleSource := req.Content
		if msgType == "audio" {
			titleSource = aiInputContent
		}

		generatedTitle := s.generateSessionTitle(ctx, titleSource)
		if generatedTitle != "" {
			session.Title = generatedTitle
		}
	}

	_ = s.sessionRepo.Update(ctx, session)

	if err := s.gamificationService.AwardExp(ctx, userID, "chat_ai", 10); err != nil {
		logger.Warn("chat: failed to award exp", zap.Uint("user_id", userID), zap.Error(err))
	}

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

func (s *ChatService) ToggleTrash(ctx context.Context, sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleTrash(ctx, sessionID)
}

func (s *ChatService) ToggleFavorite(ctx context.Context, sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleFavorite(ctx, sessionID)
}

func (s *ChatService) DeleteSession(ctx context.Context, sessionID, userID uint) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.Delete(ctx, sessionID)
}

func (s *ChatService) ToggleMessageLike(ctx context.Context, messageID, userID uint) error {
	return s.messageRepo.ToggleLike(ctx, messageID)
}

func (s *ChatService) ToggleMessageDislike(ctx context.Context, messageID, userID uint) error {
	return s.messageRepo.ToggleDislike(ctx, messageID)
}
