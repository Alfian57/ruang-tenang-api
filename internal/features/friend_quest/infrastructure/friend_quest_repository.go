package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type FriendQuestRepository struct {
	db *gorm.DB
}

func NewFriendQuestRepository(db *gorm.DB) *FriendQuestRepository {
	return &FriendQuestRepository{db: db}
}

func (r *FriendQuestRepository) Create(ctx context.Context, quest *model.FriendQuest) error {
	return r.db.WithContext(ctx).Create(quest).Error
}

func (r *FriendQuestRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.FriendQuest, error) {
	var quest model.FriendQuest
	err := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Partner").
		Where("id = ?", id).First(&quest).Error
	return &quest, err
}

func (r *FriendQuestRepository) GetUserQuests(ctx context.Context, userID uint, status string, page, limit int) ([]model.FriendQuest, int64, error) {
	query := r.db.WithContext(ctx).
		Preload("Requester").
		Preload("Partner").
		Where("requester_id = ? OR partner_id = ?", userID, userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Model(&model.FriendQuest{}).Count(&total)

	var quests []model.FriendQuest
	err := query.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&quests).Error

	return quests, total, err
}

func (r *FriendQuestRepository) Update(ctx context.Context, quest *model.FriendQuest) error {
	return r.db.WithContext(ctx).Save(quest).Error
}

func (r *FriendQuestRepository) CountActiveQuests(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.FriendQuest{}).
		Where("(requester_id = ? OR partner_id = ?) AND status = 'active'", userID, userID).
		Count(&count).Error
	return count, err
}

func (r *FriendQuestRepository) GetActiveQuestsBetweenUsers(ctx context.Context, userA, userB uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.FriendQuest{}).
		Where("status IN ('pending','active') AND ((requester_id = ? AND partner_id = ?) OR (requester_id = ? AND partner_id = ?))",
			userA, userB, userB, userA).
		Count(&count).Error
	return count, err
}

func (r *FriendQuestRepository) UpdateProgress(ctx context.Context, questID uuid.UUID, field string, value int) error {
	return r.db.WithContext(ctx).
		Model(&model.FriendQuest{}).
		Where("id = ?", questID).
		UpdateColumn(field, gorm.Expr(field+" + ?", value)).Error
}
