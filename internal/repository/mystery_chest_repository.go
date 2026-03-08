package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type MysteryChestRepository struct {
	db *gorm.DB
}

func NewMysteryChestRepository(db *gorm.DB) *MysteryChestRepository {
	return &MysteryChestRepository{db: db}
}

func (r *MysteryChestRepository) Create(ctx context.Context, chest *model.UserChest) error {
	return r.db.WithContext(ctx).Create(chest).Error
}

func (r *MysteryChestRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.UserChest, error) {
	var chest model.UserChest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&chest).Error
	return &chest, err
}

func (r *MysteryChestRepository) GetUserChests(ctx context.Context, userID uint, isOpened *bool, rarity string, page, limit int) ([]model.UserChest, int64, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if isOpened != nil {
		query = query.Where("is_opened = ?", *isOpened)
	}
	if rarity != "" {
		query = query.Where("rarity = ?", rarity)
	}

	var total int64
	query.Model(&model.UserChest{}).Count(&total)

	var chests []model.UserChest
	err := query.Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&chests).Error

	return chests, total, err
}

func (r *MysteryChestRepository) CountUnopenedChests(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserChest{}).
		Where("user_id = ? AND is_opened = false", userID).
		Count(&count).Error
	return count, err
}

func (r *MysteryChestRepository) Update(ctx context.Context, chest *model.UserChest) error {
	return r.db.WithContext(ctx).Save(chest).Error
}
