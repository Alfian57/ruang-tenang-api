package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type StreakSocietyRepository struct {
	db *gorm.DB
}

func NewStreakSocietyRepository(db *gorm.DB) *StreakSocietyRepository {
	return &StreakSocietyRepository{db: db}
}

// === Societies ===

func (r *StreakSocietyRepository) GetAllSocieties(ctx context.Context) ([]model.StreakSociety, error) {
	var societies []model.StreakSociety
	err := r.db.WithContext(ctx).Order("min_streak ASC").Find(&societies).Error
	return societies, err
}

func (r *StreakSocietyRepository) GetSocietyByMinStreak(ctx context.Context, streak int) (*model.StreakSociety, error) {
	var society model.StreakSociety
	err := r.db.WithContext(ctx).
		Where("min_streak <= ?", streak).
		Order("min_streak DESC").
		First(&society).Error
	return &society, err
}

// === Memberships ===

func (r *StreakSocietyRepository) GetUserMembership(ctx context.Context, userID uint, societyID int) (*model.UserSocietyMembership, error) {
	var m model.UserSocietyMembership
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND society_id = ? AND is_active = true", userID, societyID).
		First(&m).Error
	return &m, err
}

func (r *StreakSocietyRepository) GetUserActiveMemberships(ctx context.Context, userID uint) ([]model.UserSocietyMembership, error) {
	var memberships []model.UserSocietyMembership
	err := r.db.WithContext(ctx).
		Preload("Society").
		Where("user_id = ? AND is_active = true", userID).
		Find(&memberships).Error
	return memberships, err
}

func (r *StreakSocietyRepository) CreateMembership(ctx context.Context, m *model.UserSocietyMembership) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *StreakSocietyRepository) DeactivateMembership(ctx context.Context, userID uint, societyID int) error {
	return r.db.WithContext(ctx).
		Model(&model.UserSocietyMembership{}).
		Where("user_id = ? AND society_id = ?", userID, societyID).
		UpdateColumn("is_active", false).Error
}

func (r *StreakSocietyRepository) CountMembers(ctx context.Context, societyID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserSocietyMembership{}).
		Where("society_id = ? AND is_active = true", societyID).
		Count(&count).Error
	return count, err
}

func (r *StreakSocietyRepository) GetSocietyMembers(ctx context.Context, societyID int, page, limit int) ([]model.UserSocietyMembership, int64, error) {
	query := r.db.WithContext(ctx).
		Preload("User").
		Where("society_id = ? AND is_active = true", societyID)

	var total int64
	query.Model(&model.UserSocietyMembership{}).Count(&total)

	var members []model.UserSocietyMembership
	err := query.Order("joined_at ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&members).Error

	return members, total, err
}
