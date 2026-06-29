package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/generative-ai-go/genai"

	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"go.uber.org/zap"
)

func audioMimeTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".ogg":
		return "audio/ogg"
	default:
		return "audio/mpeg"
	}
}

func resolveAudioUploadPath(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "", errors.New("audio content is empty")
	}

	pathPart := trimmed
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("invalid audio url: %w", err)
		}
		pathPart = u.Path
	}

	pathPart = strings.Split(pathPart, "?")[0]
	if !strings.HasPrefix(pathPart, "/uploads/audio/") {
		return "", fmt.Errorf("unsupported audio path: %s", pathPart)
	}

	relativePath := strings.TrimPrefix(pathPart, "/")
	cleanPath := filepath.Clean(relativePath)
	if !strings.HasPrefix(cleanPath, "uploads/audio/") {
		return "", fmt.Errorf("invalid audio path: %s", cleanPath)
	}

	return cleanPath, nil
}

func (s *ChatService) transcribeAudioContent(ctx context.Context, content string) (string, error) {
	aiModel := s.modelForRequest()
	if aiModel == nil {
		return "", errors.New("gemini model is not available")
	}

	audioPath, err := resolveAudioUploadPath(content)
	if err != nil {
		return "", err
	}

	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	transcribePrompt := "Transkripsikan audio berikut ke dalam bahasa Indonesia. Kembalikan hanya teks transkrip tanpa penjelasan tambahan."
	resp, err := aiModel.GenerateContent(ctx,
		genai.Text(transcribePrompt),
		genai.Blob{MIMEType: audioMimeTypeFromPath(audioPath), Data: audioBytes},
	)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty transcription response")
	}

	var builder strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			builder.WriteString(string(txt))
		}
	}

	result := strings.TrimSpace(builder.String())
	if result == "" {
		return "", errors.New("empty transcription text")
	}

	return result, nil
}

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

func (s *ChatService) CreateSession(ctx context.Context, userID uint, req *dto.CreateChatSessionRequest) (*model.ChatSession, error) {
	session := &model.ChatSession{
		UserID:   userID,
		Title:    req.Title,
		FolderID: req.FolderID,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID uint, req *dto.SendMessageRequest) (*dto.ChatMessageDTO, *dto.ChatMessageDTO, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(ctx, sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: session not found: %w", err)
	}

	if session.UserID != userID {
		return nil, nil, fmt.Errorf("ChatService.SendMessage: unauthorized access to session %d", sessionID)
	}

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
		transcript, err := s.transcribeAudioContent(ctx, req.Content)
		if err == nil && strings.TrimSpace(transcript) != "" {
			aiInputContent = transcript
		} else {
			aiInputContent = "Pengguna mengirim pesan suara tetapi transkripsi tidak tersedia. Minta pengguna menuliskan inti pesan dengan singkat."
		}
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
	} else if s.genaiClient != nil {
		// Bound the Gemini call with the request context so client cancellation
		// and server-side timeouts propagate (previously context.Background()
		// made calls un-cancellable and could hang indefinitely).
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
				logger.Warn("gemini reply failed", zap.Error(err))
			}
		} else {
			// Derive a fresh model per request so RAG tools are scoped to this
			// chat only. Mutating the shared s.genaiModel here would race with
			// concurrent requests (tools leaking between users or being nilled
			// mid-call).
			aiModel := s.modelForRequest()
			aiModel.Tools = s.buildRAGTools()

			cs := aiModel.StartChat()

			cs.History = []*genai.Content{}
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
				if msg.Role == model.ChatRoleAI {
					role = "model"
				}
				cs.History = append(cs.History, &genai.Content{
					Role: role,
					Parts: []genai.Part{
						genai.Text(msg.Content),
					},
				})
			}

			// Send message and handle function calling loop
			resp, err := cs.SendMessage(ctx, genai.Text(aiInputContent))
			if err != nil {
				logger.Warn("gemini reply failed", zap.Error(err))
			} else {
				// Function calling loop: max 3 iterations to prevent infinite loops
				for i := 0; i < 3; i++ {
					if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
						break
					}

					funcCalls := resp.Candidates[0].FunctionCalls()
					if len(funcCalls) == 0 {
						// No function calls — extract text response
						if len(resp.Candidates[0].Content.Parts) > 0 {
							if txt, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
								aiResponseText = string(txt)
							}
						}
						break
					}

					// Process all function calls and send results back
					var funcResponseParts []genai.Part
					for _, fc := range funcCalls {
						result := s.handleFunctionCall(ctx, fc, userID, preferences, userMessageCount)
						funcResponseParts = append(funcResponseParts, result)
					}

					// Send function responses back to Gemini
					resp, err = cs.SendMessage(ctx, funcResponseParts...)
					if err != nil {
						logger.Warn("gemini function response failed", zap.Error(err))
						break
					}
				}
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
