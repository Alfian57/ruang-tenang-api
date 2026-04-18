package infrastructure

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

const (
	defaultRewardBoostMultiplier = 2.0
	defaultRewardBoostDuration   = 24 * time.Hour
)

var (
	ErrInsufficientCoins  = errors.New("insufficient gold coins")
	ErrRewardUnavailable  = errors.New("reward is unavailable")
	ErrRewardOutOfStock   = errors.New("reward is out of stock")
	ErrRewardAlreadyOwned = errors.New("hadiah tema sudah dimiliki")
	xpBoostValuePattern   = regexp.MustCompile(`^\s*\d+(?:\.\d+)?x(?:_\d+[mhd])?\s*$`)
)

type RewardRepository struct {
	db *gorm.DB
}

func NewRewardRepository(db *gorm.DB) *RewardRepository {
	return &RewardRepository{db: db}
}

// GetAllRewards returns all rewards (optionally filter active only)
func (r *RewardRepository) GetAllRewards(ctx context.Context, activeOnly bool) ([]model.Reward, error) {
	var rewards []model.Reward
	query := r.db.WithContext(ctx)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	err := query.Order("created_at DESC").Find(&rewards).Error
	return rewards, err
}

// GetRewardByID returns a single reward by ID
func (r *RewardRepository) GetRewardByID(ctx context.Context, id uint) (*model.Reward, error) {
	var reward model.Reward
	err := r.db.WithContext(ctx).First(&reward, id).Error
	if err != nil {
		return nil, err
	}
	return &reward, nil
}

// CreateReward creates a new reward
func (r *RewardRepository) CreateReward(ctx context.Context, reward *model.Reward) error {
	return r.db.WithContext(ctx).Create(reward).Error
}

// UpdateReward updates an existing reward
func (r *RewardRepository) UpdateReward(ctx context.Context, reward *model.Reward) error {
	return r.db.WithContext(ctx).Save(reward).Error
}

// DeleteReward deletes a reward by ID
func (r *RewardRepository) DeleteReward(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Reward{}, id).Error
}

// ClaimReward handles the full claim transaction: deduct coins, decrement stock, create claim
func (r *RewardRepository) ClaimReward(ctx context.Context, userID uint, rewardID uint) (*model.RewardClaim, error) {
	var claim model.RewardClaim

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get reward
		var reward model.Reward
		if err := tx.First(&reward, rewardID).Error; err != nil {
			return err
		}

		// Check availability
		if !reward.IsActive {
			return ErrRewardUnavailable
		}
		if reward.Stock == 0 {
			return ErrRewardOutOfStock
		}

		// Prevent duplicate claims for theme rewards
		if reward.RewardType == model.RewardTypeTheme {
			var existingCount int64
			tx.Model(&model.RewardClaim{}).
				Where("user_id = ? AND reward_id = ?", userID, rewardID).
				Count(&existingCount)
			if existingCount > 0 {
				return ErrRewardAlreadyOwned
			}
		}

		// Check user coins
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		if user.GoldCoins < int64(reward.CoinCost) {
			return ErrInsufficientCoins
		}

		// Deduct coins
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("gold_coins", gorm.Expr("gold_coins - ?", reward.CoinCost)).Error; err != nil {
			return err
		}

		// Decrement stock (if not unlimited)
		if reward.Stock > 0 {
			if err := tx.Model(&model.Reward{}).
				Where("id = ? AND stock > 0", rewardID).
				Update("stock", gorm.Expr("stock - 1")).Error; err != nil {
				return err
			}
		}

		// Create claim record
		claim = model.RewardClaim{
			UserID:    userID,
			RewardID:  rewardID,
			CoinSpent: reward.CoinCost,
		}
		if err := tx.Create(&claim).Error; err != nil {
			return err
		}

		if err := r.applyRewardSideEffects(tx, userID, &reward); err != nil {
			return err
		}

		// Populate the reward in the claim
		claim.Reward = reward

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &claim, nil
}

func (r *RewardRepository) applyRewardSideEffects(tx *gorm.DB, userID uint, reward *model.Reward) error {
	if reward == nil {
		return nil
	}

	if isXPBoostReward(reward) {
		multiplier, duration := parseRewardXPBoostConfig(reward.RewardValue)
		return r.activateOrExtendXPBoost(tx, userID, multiplier, duration)
	}

	return nil
}

func containsXPBoostKeyword(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}

	return strings.Contains(normalized, "xp boost") ||
		strings.Contains(normalized, "xpboost") ||
		strings.Contains(normalized, "exp boost") ||
		strings.Contains(normalized, "expboost")
}

func isXPBoostReward(reward *model.Reward) bool {
	if reward == nil {
		return false
	}

	if reward.RewardType == model.RewardTypeXPBoost {
		return true
	}

	if containsXPBoostKeyword(reward.Name) || containsXPBoostKeyword(reward.Description) {
		return true
	}

	normalizedValue := strings.ToLower(strings.TrimSpace(reward.RewardValue))
	return xpBoostValuePattern.MatchString(normalizedValue)
}

func (r *RewardRepository) activateOrExtendXPBoost(tx *gorm.DB, userID uint, multiplier float64, duration time.Duration) error {
	now := time.Now()

	var activeBoost model.XPBoost
	err := tx.Where("user_id = ? AND is_active = true AND expires_at > ?", userID, now).
		Order("expires_at DESC").
		First(&activeBoost).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		newBoost := model.XPBoost{
			UserID:      userID,
			Multiplier:  multiplier,
			TriggerType: model.BoostTriggerReward,
			StartedAt:   now,
			ExpiresAt:   now.Add(duration),
			IsActive:    true,
		}

		return tx.Create(&newBoost).Error
	}

	startFrom := activeBoost.ExpiresAt
	if startFrom.Before(now) {
		startFrom = now
	}

	updates := map[string]interface{}{
		"trigger_type": model.BoostTriggerReward,
		"expires_at":   startFrom.Add(duration),
		"is_active":    true,
	}

	if multiplier > activeBoost.Multiplier {
		updates["multiplier"] = multiplier
	}

	return tx.Model(&model.XPBoost{}).
		Where("id = ?", activeBoost.ID).
		Updates(updates).Error
}

func parseRewardXPBoostConfig(raw string) (float64, time.Duration) {
	multiplier := defaultRewardBoostMultiplier
	duration := defaultRewardBoostDuration

	if raw == "" {
		return multiplier, duration
	}

	parts := strings.Split(strings.ToLower(strings.TrimSpace(raw)), "_")
	if len(parts) > 0 {
		rawMultiplier := strings.TrimSuffix(parts[0], "x")
		if parsedMultiplier, err := strconv.ParseFloat(rawMultiplier, 64); err == nil && parsedMultiplier > 1.0 {
			multiplier = parsedMultiplier
		}
	}

	if len(parts) > 1 {
		if parsedDuration, ok := parseRewardDuration(parts[1]); ok {
			duration = parsedDuration
		}
	}

	return multiplier, duration
}

func parseRewardDuration(token string) (time.Duration, bool) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, false
	}

	parseUnit := func(suffix string, unit time.Duration) (time.Duration, bool) {
		if !strings.HasSuffix(token, suffix) {
			return 0, false
		}

		numberPart := strings.TrimSuffix(token, suffix)
		value, err := strconv.Atoi(numberPart)
		if err != nil || value <= 0 {
			return 0, false
		}

		return time.Duration(value) * unit, true
	}

	if d, ok := parseUnit("m", time.Minute); ok {
		return d, true
	}

	if d, ok := parseUnit("h", time.Hour); ok {
		return d, true
	}

	if d, ok := parseUnit("d", 24*time.Hour); ok {
		return d, true
	}

	return 0, false
}

// GetUserClaims returns paginated claims for a specific user
func (r *RewardRepository) GetUserClaims(ctx context.Context, userID uint, limit, offset int) ([]model.RewardClaim, int64, error) {
	var claims []model.RewardClaim
	var total int64

	r.db.WithContext(ctx).Model(&model.RewardClaim{}).
		Where("user_id = ?", userID).
		Count(&total)

	err := r.db.WithContext(ctx).
		Preload("Reward").
		Where("user_id = ?", userID).
		Order("claimed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&claims).Error

	return claims, total, err
}

// GetAllClaims returns paginated claims for all users (admin)
func (r *RewardRepository) GetAllClaims(ctx context.Context, limit, offset int) ([]model.RewardClaim, int64, error) {
	var claims []model.RewardClaim
	var total int64

	r.db.WithContext(ctx).Model(&model.RewardClaim{}).Count(&total)

	err := r.db.WithContext(ctx).
		Preload("Reward").
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email, avatar")
		}).
		Order("claimed_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&claims).Error

	return claims, total, err
}
