package repository

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ================================
// ChatFolderRepository
// ================================

type ChatFolderRepository struct {
	db *gorm.DB
}

func NewChatFolderRepository(db *gorm.DB) *ChatFolderRepository {
	return &ChatFolderRepository{db: db}
}

func (r *ChatFolderRepository) FindByUserID(ctx context.Context, userID uint) ([]model.ChatFolder, error) {
	var folders []model.ChatFolder
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("position ASC, created_at ASC").Find(&folders).Error
	return folders, err
}

func (r *ChatFolderRepository) FindByID(ctx context.Context, id uint) (*model.ChatFolder, error) {
	var folder model.ChatFolder
	err := r.db.WithContext(ctx).First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *ChatFolderRepository) FindByIDWithSessions(ctx context.Context, id uint) (*model.ChatFolder, error) {
	var folder model.ChatFolder
	err := r.db.WithContext(ctx).Preload("Sessions", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_trash = ?", false).Order("updated_at DESC")
	}).First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *ChatFolderRepository) Create(ctx context.Context, folder *model.ChatFolder) error {
	return r.db.WithContext(ctx).Create(folder).Error
}

func (r *ChatFolderRepository) Update(ctx context.Context, folder *model.ChatFolder) error {
	return r.db.WithContext(ctx).Save(folder).Error
}

func (r *ChatFolderRepository) Delete(ctx context.Context, id uint) error {
	// First, remove folder_id from all sessions in this folder
	if err := r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("folder_id = ?", id).Update("folder_id", nil).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&model.ChatFolder{}, id).Error
}

func (r *ChatFolderRepository) CountSessionsInFolder(ctx context.Context, folderID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("folder_id = ? AND is_trash = ?", folderID, false).Count(&count).Error
	return count, err
}

func (r *ChatFolderRepository) GetMaxPosition(ctx context.Context, userID uint) (int, error) {
	var maxPos int
	err := r.db.WithContext(ctx).Model(&model.ChatFolder{}).Where("user_id = ?", userID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos).Error
	return maxPos, err
}

func (r *ChatFolderRepository) ReorderFolders(ctx context.Context, userID uint, folderIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range folderIDs {
			if err := tx.Model(&model.ChatFolder{}).Where("id = ? AND user_id = ?", id, userID).Update("position", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ================================
// ChatSessionRepository
// ================================

type ChatSessionRepository struct {
	db *gorm.DB
}

func NewChatSessionRepository(db *gorm.DB) *ChatSessionRepository {
	return &ChatSessionRepository{db: db}
}

func (r *ChatSessionRepository) FindByUserID(ctx context.Context, userID uint, filter, search string, folderID *uint, page, limit int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("user_id = ?", userID)

	// Apply filters
	switch filter {
	case "trash":
		query = query.Where("is_trash = ?", true)
	case "favorites":
		query = query.Where("is_favorite = ?", true).Where("is_trash = ?", false)
	default:
		// Default: Not trash
		query = query.Where("is_trash = ?", false)
	}

	// Filter by folder
	if folderID != nil {
		query = query.Where("folder_id = ?", *folderID)
	}

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).Order("updated_at DESC").Offset(offset).Limit(limit).Find(&sessions).Error

	return sessions, total, err
}

// FindByUserIDGroupedByFolder returns sessions grouped by folder
func (r *ChatSessionRepository) FindByUserIDGroupedByFolder(ctx context.Context, userID uint, filter string) ([]model.ChatSession, error) {
	var sessions []model.ChatSession

	query := r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("user_id = ?", userID)

	// Apply filters
	switch filter {
	case "trash":
		query = query.Where("is_trash = ?", true)
	case "favorites":
		query = query.Where("is_favorite = ?", true).Where("is_trash = ?", false)
	default:
		query = query.Where("is_trash = ?", false)
	}

	err := query.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).Preload("Folder").Order("folder_id NULLS LAST, updated_at DESC").Find(&sessions).Error

	return sessions, err
}

func (r *ChatSessionRepository) FindByID(ctx context.Context, id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) FindByUUID(ctx context.Context, sessionUUID uuid.UUID) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.WithContext(ctx).Where("uuid = ?", sessionUUID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) FindByIDWithMessages(ctx context.Context, id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.WithContext(ctx).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Preload("Folder").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByIDWithPinnedMessages returns session with only pinned messages
func (r *ChatSessionRepository) FindByIDWithPinnedMessages(ctx context.Context, id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.WithContext(ctx).Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_pinned = ?", true).Order("created_at ASC")
	}).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) Create(ctx context.Context, session *model.ChatSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *ChatSessionRepository) Update(ctx context.Context, session *model.ChatSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *ChatSessionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ChatSession{}, id).Error
}

func (r *ChatSessionRepository) ToggleTrash(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("id = ?", id).
		Update("is_trash", gorm.Expr("NOT is_trash")).Error
}

func (r *ChatSessionRepository) ToggleFavorite(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("id = ?", id).
		Update("is_favorite", gorm.Expr("NOT is_favorite")).Error
}

// MoveToFolder moves a session to a folder (or removes from folder if folderID is nil)
func (r *ChatSessionRepository) MoveToFolder(ctx context.Context, sessionID uint, folderID *uint) error {
	return r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("id = ?", sessionID).Update("folder_id", folderID).Error
}

// UpdateSummary updates the session summary
func (r *ChatSessionRepository) UpdateSummary(ctx context.Context, sessionID uint, summary string) error {
	return r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"summary":              summary,
			"summary_generated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).Error
}

// ================================
// ChatMessageRepository
// ================================

type ChatMessageRepository struct {
	db *gorm.DB
}

func NewChatMessageRepository(db *gorm.DB) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

// Message methods
func (r *ChatMessageRepository) Create(ctx context.Context, message *model.ChatMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *ChatMessageRepository) FindByID(ctx context.Context, id uint) (*model.ChatMessage, error) {
	var message model.ChatMessage
	err := r.db.WithContext(ctx).First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *ChatMessageRepository) Update(ctx context.Context, message *model.ChatMessage) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *ChatMessageRepository) ToggleLike(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_liked":    gorm.Expr("NOT is_liked"),
			"is_disliked": false,
		}).Error
}

func (r *ChatMessageRepository) ToggleDislike(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_disliked": gorm.Expr("NOT is_disliked"),
			"is_liked":    false,
		}).Error
}

// TogglePin toggles the pinned status of a message
func (r *ChatMessageRepository) TogglePin(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("id = ?", id).
		Update("is_pinned", gorm.Expr("NOT is_pinned")).Error
}

// SetPinned sets the pinned status of a message
func (r *ChatMessageRepository) SetPinned(ctx context.Context, id uint, pinned bool) error {
	return r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("id = ?", id).Update("is_pinned", pinned).Error
}

// FindPinnedBySessionID returns all pinned messages in a session
func (r *ChatMessageRepository) FindPinnedBySessionID(ctx context.Context, sessionID uint) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.WithContext(ctx).Where("chat_session_id = ? AND is_pinned = ?", sessionID, true).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// CountPinnedBySessionID counts pinned messages in a session
func (r *ChatMessageRepository) CountPinnedBySessionID(ctx context.Context, sessionID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("chat_session_id = ? AND is_pinned = ?", sessionID, true).Count(&count).Error
	return count, err
}
