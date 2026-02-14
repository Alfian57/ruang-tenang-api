package repository

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
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

func (r *ChatFolderRepository) FindByUserID(userID uint) ([]model.ChatFolder, error) {
	var folders []model.ChatFolder
	err := r.db.Where("user_id = ?", userID).Order("position ASC, created_at ASC").Find(&folders).Error
	return folders, err
}

func (r *ChatFolderRepository) FindByID(id uint) (*model.ChatFolder, error) {
	var folder model.ChatFolder
	err := r.db.First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *ChatFolderRepository) FindByIDWithSessions(id uint) (*model.ChatFolder, error) {
	var folder model.ChatFolder
	err := r.db.Preload("Sessions", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_trash = ?", false).Order("updated_at DESC")
	}).First(&folder, id).Error
	if err != nil {
		return nil, err
	}
	return &folder, nil
}

func (r *ChatFolderRepository) Create(folder *model.ChatFolder) error {
	return r.db.Create(folder).Error
}

func (r *ChatFolderRepository) Update(folder *model.ChatFolder) error {
	return r.db.Save(folder).Error
}

func (r *ChatFolderRepository) Delete(id uint) error {
	// First, remove folder_id from all sessions in this folder
	if err := r.db.Model(&model.ChatSession{}).Where("folder_id = ?", id).Update("folder_id", nil).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.ChatFolder{}, id).Error
}

func (r *ChatFolderRepository) CountSessionsInFolder(folderID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ChatSession{}).Where("folder_id = ? AND is_trash = ?", folderID, false).Count(&count).Error
	return count, err
}

func (r *ChatFolderRepository) GetMaxPosition(userID uint) (int, error) {
	var maxPos int
	err := r.db.Model(&model.ChatFolder{}).Where("user_id = ?", userID).Select("COALESCE(MAX(position), 0)").Scan(&maxPos).Error
	return maxPos, err
}

func (r *ChatFolderRepository) ReorderFolders(userID uint, folderIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *ChatSessionRepository) FindByUserID(userID uint, filter, search string, folderID *uint, page, limit int) ([]model.ChatSession, int64, error) {
	var sessions []model.ChatSession
	var total int64

	query := r.db.Model(&model.ChatSession{}).Where("user_id = ?", userID)

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
func (r *ChatSessionRepository) FindByUserIDGroupedByFolder(userID uint, filter string) ([]model.ChatSession, error) {
	var sessions []model.ChatSession

	query := r.db.Model(&model.ChatSession{}).Where("user_id = ?", userID)

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

func (r *ChatSessionRepository) FindByID(id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) FindByIDWithMessages(id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Preload("Folder").First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByIDWithPinnedMessages returns session with only pinned messages
func (r *ChatSessionRepository) FindByIDWithPinnedMessages(id uint) (*model.ChatSession, error) {
	var session model.ChatSession
	err := r.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_pinned = ?", true).Order("created_at ASC")
	}).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) Create(session *model.ChatSession) error {
	return r.db.Create(session).Error
}

func (r *ChatSessionRepository) Update(session *model.ChatSession) error {
	return r.db.Save(session).Error
}

func (r *ChatSessionRepository) Delete(id uint) error {
	return r.db.Delete(&model.ChatSession{}, id).Error
}

func (r *ChatSessionRepository) ToggleTrash(id uint) error {
	return r.db.Model(&model.ChatSession{}).Where("id = ?", id).
		Update("is_trash", gorm.Expr("NOT is_trash")).Error
}

func (r *ChatSessionRepository) ToggleFavorite(id uint) error {
	return r.db.Model(&model.ChatSession{}).Where("id = ?", id).
		Update("is_favorite", gorm.Expr("NOT is_favorite")).Error
}

// MoveToFolder moves a session to a folder (or removes from folder if folderID is nil)
func (r *ChatSessionRepository) MoveToFolder(sessionID uint, folderID *uint) error {
	return r.db.Model(&model.ChatSession{}).Where("id = ?", sessionID).Update("folder_id", folderID).Error
}

// UpdateSummary updates the session summary
func (r *ChatSessionRepository) UpdateSummary(sessionID uint, summary string) error {
	return r.db.Model(&model.ChatSession{}).Where("id = ?", sessionID).
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
func (r *ChatMessageRepository) Create(message *model.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *ChatMessageRepository) FindByID(id uint) (*model.ChatMessage, error) {
	var message model.ChatMessage
	err := r.db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *ChatMessageRepository) Update(message *model.ChatMessage) error {
	return r.db.Save(message).Error
}

func (r *ChatMessageRepository) ToggleLike(id uint) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_liked":    gorm.Expr("NOT is_liked"),
			"is_disliked": false,
		}).Error
}

func (r *ChatMessageRepository) ToggleDislike(id uint) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_disliked": gorm.Expr("NOT is_disliked"),
			"is_liked":    false,
		}).Error
}

// TogglePin toggles the pinned status of a message
func (r *ChatMessageRepository) TogglePin(id uint) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).
		Update("is_pinned", gorm.Expr("NOT is_pinned")).Error
}

// SetPinned sets the pinned status of a message
func (r *ChatMessageRepository) SetPinned(id uint, pinned bool) error {
	return r.db.Model(&model.ChatMessage{}).Where("id = ?", id).Update("is_pinned", pinned).Error
}

// FindPinnedBySessionID returns all pinned messages in a session
func (r *ChatMessageRepository) FindPinnedBySessionID(sessionID uint) ([]model.ChatMessage, error) {
	var messages []model.ChatMessage
	err := r.db.Where("chat_session_id = ? AND is_pinned = ?", sessionID, true).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// CountPinnedBySessionID counts pinned messages in a session
func (r *ChatMessageRepository) CountPinnedBySessionID(sessionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ChatMessage{}).Where("chat_session_id = ? AND is_pinned = ?", sessionID, true).Count(&count).Error
	return count, err
}
