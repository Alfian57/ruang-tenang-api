package infrastructure

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
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
	today := timeutil.Today()
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

// CreateSpin inserts a daily spin record. It relies on the unique index
// (user_id, spin_date) to make the insert atomic: a concurrent second spin for
// the same user/day fails at the DB level and is reported as ErrDuplicateSpin.
func (r *DailySpinRepository) CreateSpin(ctx context.Context, spin *model.UserSpin) error {
	err := r.db.WithContext(ctx).Create(spin).Error
	if err != nil && isDuplicateKeyError(err) {
		return ErrDuplicateSpin
	}
	return err
}

// ErrDuplicateSpin signals a concurrent double-spin attempt rejected by the
// unique constraint on (user_id, spin_date).
var ErrDuplicateSpin = gorm.ErrDuplicatedKey

// isDuplicateKeyError reports whether err is a unique-constraint violation
// surfaced by GORM/Postgres. We check the typed sentinel first (preferred) and
// fall back to substring matching for older driver error shapes.
func isDuplicateKeyError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}
