package xpboost

import (
	"context"
	"math"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

const comboActiveWindow = 30 * time.Minute

type Snapshot struct {
	Effective float64
	Boost     float64
	Combo     float64
}

// GetEffectiveMultiplier resolves the current XP multiplier from active boost + combo.
func GetEffectiveMultiplier(ctx context.Context, db *gorm.DB, userID uint, now time.Time) Snapshot {
	snapshot := Snapshot{
		Effective: 1.0,
		Boost:     1.0,
		Combo:     1.0,
	}

	var boost model.XPBoost
	if err := db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, now).
		Order("expires_at DESC").
		First(&boost).Error; err == nil && boost.Multiplier > 1.0 {
		snapshot.Boost = boost.Multiplier
		snapshot.Effective *= boost.Multiplier
	}

	var combo model.UserCombo
	if err := db.WithContext(ctx).
		Where("user_id = ?", userID).
		First(&combo).Error; err == nil && combo.Multiplier > 1.0 {
		if combo.LastActivityAt != nil && now.Sub(*combo.LastActivityAt) < comboActiveWindow {
			snapshot.Combo = combo.Multiplier
			snapshot.Effective *= combo.Multiplier
		}
	}

	return snapshot
}

// Apply converts base XP into boosted XP using the provided multiplier.
func Apply(baseXP int64, multiplier float64) int64 {
	if baseXP <= 0 {
		return 0
	}

	if multiplier <= 1.0 {
		return baseXP
	}

	boosted := int64(math.Round(float64(baseXP) * multiplier))
	if boosted < baseXP {
		return baseXP
	}

	return boosted
}
