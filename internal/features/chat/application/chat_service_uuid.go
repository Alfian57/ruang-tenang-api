package application

import (
	"context"
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/google/uuid"
)

func (s *ChatService) GetSessionByUUID(ctx context.Context, uuidStr string, userID uint) (*dto.ChatSessionDTO, error) {
	id, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetSessionByID(ctx, session.ID, userID)
}

func (s *ChatService) SendMessageByUUID(ctx context.Context, sessionUUID string, userID uint, req *dto.SendMessageRequest) (*dto.ChatMessageDTO, *dto.ChatMessageDTO, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, nil, errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return s.SendMessage(ctx, session.ID, userID, req)
}

func (s *ChatService) GetContextStateByUUID(ctx context.Context, sessionUUID string, userID uint) (*dto.ChatContextStateDTO, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.GetContextState(ctx, session.ID, userID)
}

func (s *ChatService) UpdateContextPreferencesByUUID(ctx context.Context, sessionUUID string, userID uint, req *dto.UpdateChatContextPreferencesRequest) (*dto.ChatContextStateDTO, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.UpdateContextPreferences(ctx, session.ID, userID, req)
}

func (s *ChatService) ToggleTrashByUUID(ctx context.Context, sessionUUID string, userID uint) error {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleTrash(ctx, session.ID)
}

func (s *ChatService) ToggleFavoriteByUUID(ctx context.Context, sessionUUID string, userID uint) error {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.ToggleFavorite(ctx, session.ID)
}

func (s *ChatService) DeleteSessionByUUID(ctx context.Context, sessionUUID string, userID uint) error {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}

	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return err
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.sessionRepo.Delete(ctx, session.ID)
}

func (s *ChatService) MoveSessionToFolderByUUID(ctx context.Context, sessionUUID string, userID uint, folderID *uint) error {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return errors.New("invalid uuid")
	}
	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return errors.New("session not found")
	}
	return s.MoveSessionToFolder(ctx, session.ID, userID, folderID)
}

func (s *ChatService) ExportChatByUUID(ctx context.Context, sessionUUID string, userID uint, req *dto.ExportChatRequest) (*dto.ExportChatResponse, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}
	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, errors.New("session not found")
	}
	return s.ExportChat(ctx, session.ID, userID, req)
}

func (s *ChatService) GetPinnedMessagesByUUID(ctx context.Context, sessionUUID string, userID uint) ([]dto.ChatMessageDTO, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}
	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, errors.New("session not found")
	}
	return s.GetPinnedMessages(ctx, session.ID, userID)
}

func (s *ChatService) GenerateSummaryByUUID(ctx context.Context, sessionUUID string, userID uint) (*dto.ChatSessionSummaryDTO, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}
	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, errors.New("session not found")
	}
	return s.GenerateSummary(ctx, session.ID, userID)
}

func (s *ChatService) GetSummaryByUUID(ctx context.Context, sessionUUID string, userID uint) (*dto.ChatSessionSummaryDTO, error) {
	id, err := uuid.Parse(sessionUUID)
	if err != nil {
		return nil, errors.New("invalid uuid")
	}
	session, err := s.sessionRepo.FindByUUID(ctx, id)
	if err != nil {
		return nil, errors.New("session not found")
	}
	return s.GetSummary(ctx, session.ID, userID)
}
