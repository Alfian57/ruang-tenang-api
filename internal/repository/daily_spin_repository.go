package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type DailySpinRepository struct {
	db *gorm.DB
}

func NewDailySpinRepository(db *gorm.DB) *DailySpinRepository {
	return &DailySpinRepository{db: db}
}

// === Rewards ===

func (r *DailySpinRepository) GetActiveRewards(ctx context.Context) ([]model.SpinReward, error) {
	var rewards []model.SpinReward
	err := r.db.WithContext(ctx).Where("is_active = true").Order("id ASC").Find(&rewards).Error
	return rewards, err
}

func (r *DailySpinRepository) GetRewardByID(ctx context.Context, id int) (*model.SpinReward, error) {
	var reward model.SpinReward
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&reward).Error
	return &reward, err
}

// === User Spins ===

func (r *DailySpinRepository) HasSpunToday(ctx context.Context, userID uint) (bool, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserSpin{}).
		Where("user_id = ? AND spin_date = ?", userID, today).
		Count(&count).Error
	return count > 0, err
}

func (r *DailySpinRepository) GetLastSpin(ctx context.Context, userID uint) (*model.UserSpin, error) {
	var spin model.UserSpin
	err := r.db.WithContext(ctx).
		Preload("Reward").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&spin).Error
	return &spin, err
}

func (r *DailySpinRepository) CreateSpin(ctx context.Context, spin *model.UserSpin) error {
	return r.db.WithContext(ctx).Create(spin).Error
}
