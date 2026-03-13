package infrastructure

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type XPBoostComboRepository struct {
	db *gorm.DB
}

func NewXPBoostComboRepository(db *gorm.DB) *XPBoostComboRepository {
	return &XPBoostComboRepository{db: db}
}

// === XP Boosts ===

func (r *XPBoostComboRepository) GetActiveBoost(ctx context.Context, userID uint) (*model.XPBoost, error) {
	var boost model.XPBoost
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = true AND expires_at > ?", userID, time.Now()).
		First(&boost).Error
	return &boost, err
}

func (r *XPBoostComboRepository) CreateBoost(ctx context.Context, boost *model.XPBoost) error {
	return r.db.WithContext(ctx).Create(boost).Error
}

func (r *XPBoostComboRepository) DeactivateExpiredBoosts(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&model.XPBoost{}).
		Where("is_active = true AND expires_at <= ?", time.Now()).
		UpdateColumn("is_active", false).Error
}

func (r *XPBoostComboRepository) DeactivateBoost(ctx context.Context, boostID interface{}) error {
	return r.db.WithContext(ctx).
		Model(&model.XPBoost{}).
		Where("id = ?", boostID).
		UpdateColumn("is_active", false).Error
}

// === Combos ===

func (r *XPBoostComboRepository) GetCombo(ctx context.Context, userID uint) (*model.UserCombo, error) {
	var combo model.UserCombo
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&combo).Error
	return &combo, err
}

func (r *XPBoostComboRepository) UpsertCombo(ctx context.Context, combo *model.UserCombo) error {
	return r.db.WithContext(ctx).Save(combo).Error
}

func (r *XPBoostComboRepository) ResetCombo(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.UserCombo{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"combo_count":        0,
			"multiplier":         1.0,
			"last_activity_type": "",
			"last_activity_at":   nil,
			"session_started_at": nil,
		}).Error
}
