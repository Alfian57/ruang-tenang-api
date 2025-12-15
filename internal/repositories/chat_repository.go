package repositories

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

type ChatSessionRepository struct {
	db *gorm.DB
}

func NewChatSessionRepository(db *gorm.DB) *ChatSessionRepository {
	return &ChatSessionRepository{db: db}
}

func (r *ChatSessionRepository) FindByUserID(userID uint, filter, search string, page, limit int) ([]models.ChatSession, int64, error) {
	var sessions []models.ChatSession
	var total int64

	query := r.db.Model(&models.ChatSession{}).Where("user_id = ?", userID)

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

func (r *ChatSessionRepository) FindByID(id uint) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.db.First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) FindByIDWithMessages(id uint) (*models.ChatSession, error) {
	var session models.ChatSession
	err := r.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *ChatSessionRepository) Create(session *models.ChatSession) error {
	return r.db.Create(session).Error
}

func (r *ChatSessionRepository) Update(session *models.ChatSession) error {
	return r.db.Save(session).Error
}

func (r *ChatSessionRepository) Delete(id uint) error {
	return r.db.Delete(&models.ChatSession{}, id).Error
}

func (r *ChatSessionRepository) ToggleTrash(id uint) error {
	return r.db.Model(&models.ChatSession{}).Where("id = ?", id).
		Update("is_trash", gorm.Expr("NOT is_trash")).Error
}

func (r *ChatSessionRepository) ToggleFavorite(id uint) error {
	return r.db.Model(&models.ChatSession{}).Where("id = ?", id).
		Update("is_favorite", gorm.Expr("NOT is_favorite")).Error
}

// ChatMessageRepository
type ChatMessageRepository struct {
	db *gorm.DB
}

func NewChatMessageRepository(db *gorm.DB) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

// Message methods
func (r *ChatMessageRepository) Create(message *models.ChatMessage) error {
	return r.db.Create(message).Error
}

func (r *ChatMessageRepository) FindByID(id uint) (*models.ChatMessage, error) {
	var message models.ChatMessage
	err := r.db.First(&message, id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *ChatMessageRepository) Update(message *models.ChatMessage) error {
	return r.db.Save(message).Error
}

func (r *ChatMessageRepository) ToggleLike(id uint) error {
	return r.db.Model(&models.ChatMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_liked":    gorm.Expr("NOT is_liked"),
			"is_disliked": false,
		}).Error
}

func (r *ChatMessageRepository) ToggleDislike(id uint) error {
	return r.db.Model(&models.ChatMessage{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_disliked": gorm.Expr("NOT is_disliked"),
			"is_liked":    false,
		}).Error
}
