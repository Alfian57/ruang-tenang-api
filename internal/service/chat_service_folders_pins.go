package service

import (
	"context"
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

func (s *ChatService) GetFolders(ctx context.Context, userID uint) ([]dto.ChatFolderDTO, error) {
	if s.folderRepo == nil {
		return nil, errors.New("folder repository not initialized")
	}

	folders, err := s.folderRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []dto.ChatFolderDTO
	for _, folder := range folders {
		count, _ := s.folderRepo.CountSessionsInFolder(ctx, folder.ID)
		result = append(result, dto.ChatFolderDTO{
			ID:           folder.ID,
			UUID:         folder.UUID.String(),
			Name:         folder.Name,
			Color:        folder.Color,
			Icon:         folder.Icon,
			Position:     folder.Position,
			SessionCount: int(count),
			CreatedAt:    folder.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return result, nil
}

func (s *ChatService) CreateFolder(ctx context.Context, userID uint, req *dto.CreateChatFolderRequest) (*dto.ChatFolderDTO, error) {
	if s.folderRepo == nil {
		return nil, errors.New("folder repository not initialized")
	}

	maxPos, _ := s.folderRepo.GetMaxPosition(ctx, userID)

	folder := &model.ChatFolder{
		UserID:   userID,
		Name:     req.Name,
		Color:    req.Color,
		Icon:     req.Icon,
		Position: maxPos + 1,
	}

	if folder.Color == "" {
		folder.Color = "#6366f1"
	}
	if folder.Icon == "" {
		folder.Icon = "folder"
	}

	if err := s.folderRepo.Create(ctx, folder); err != nil {
		return nil, err
	}

	return &dto.ChatFolderDTO{
		ID:           folder.ID,
		UUID:         folder.UUID.String(),
		Name:         folder.Name,
		Color:        folder.Color,
		Icon:         folder.Icon,
		Position:     folder.Position,
		SessionCount: 0,
		CreatedAt:    folder.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *ChatService) UpdateFolder(ctx context.Context, folderID, userID uint, req *dto.UpdateChatFolderRequest) (*dto.ChatFolderDTO, error) {
	if s.folderRepo == nil {
		return nil, errors.New("folder repository not initialized")
	}

	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil {
		return nil, errors.New("folder not found")
	}

	if folder.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if req.Name != "" {
		folder.Name = req.Name
	}
	if req.Color != "" {
		folder.Color = req.Color
	}
	if req.Icon != "" {
		folder.Icon = req.Icon
	}
	if req.Position != nil {
		folder.Position = *req.Position
	}

	if err := s.folderRepo.Update(ctx, folder); err != nil {
		return nil, err
	}

	count, _ := s.folderRepo.CountSessionsInFolder(ctx, folder.ID)

	return &dto.ChatFolderDTO{
		ID:           folder.ID,
		UUID:         folder.UUID.String(),
		Name:         folder.Name,
		Color:        folder.Color,
		Icon:         folder.Icon,
		Position:     folder.Position,
		SessionCount: int(count),
		CreatedAt:    folder.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (s *ChatService) DeleteFolder(ctx context.Context, folderID, userID uint) error {
	if s.folderRepo == nil {
		return errors.New("folder repository not initialized")
	}

	folder, err := s.folderRepo.FindByID(ctx, folderID)
	if err != nil {
		return errors.New("folder not found")
	}

	if folder.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.folderRepo.Delete(ctx, folderID)
}

func (s *ChatService) ReorderFolders(ctx context.Context, userID uint, req *dto.ReorderFoldersRequest) error {
	if s.folderRepo == nil {
		return errors.New("folder repository not initialized")
	}

	return s.folderRepo.ReorderFolders(ctx, userID, req.FolderIDs)
}

func (s *ChatService) MoveSessionToFolder(ctx context.Context, sessionID, userID uint, folderID *uint) error {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	if folderID != nil && s.folderRepo != nil {
		folder, err := s.folderRepo.FindByID(ctx, *folderID)
		if err != nil {
			return errors.New("folder not found")
		}
		if folder.UserID != userID {
			return errors.New("unauthorized")
		}
	}

	return s.sessionRepo.MoveToFolder(ctx, sessionID, folderID)
}

func (s *ChatService) ToggleMessagePin(ctx context.Context, messageID, userID uint) error {
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return errors.New("message not found")
	}

	session, err := s.sessionRepo.FindByID(ctx, message.ChatSessionID)
	if err != nil {
		return errors.New("session not found")
	}

	if session.UserID != userID {
		return errors.New("unauthorized")
	}

	return s.messageRepo.TogglePin(ctx, messageID)
}

func (s *ChatService) GetPinnedMessages(ctx context.Context, sessionID, userID uint) ([]dto.ChatMessageDTO, error) {
	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	messages, err := s.messageRepo.FindPinnedBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var result []dto.ChatMessageDTO
	for _, msg := range messages {
		result = append(result, dto.ChatMessageDTO{
			ID:         msg.ID,
			Role:       string(msg.Role),
			Content:    msg.Content,
			Type:       msg.Type,
			IsLiked:    msg.IsLiked,
			IsDisliked: msg.IsDisliked,
			IsPinned:   msg.IsPinned,
			CreatedAt:  msg.CreatedAt,
		})
	}

	return result, nil
}
