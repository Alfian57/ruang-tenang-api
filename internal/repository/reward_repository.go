package repository

import (
	"context"
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

var (
	ErrInsufficientCoins  = errors.New("insufficient gold coins")
	ErrRewardUnavailable  = errors.New("reward is unavailable")
	ErrRewardOutOfStock   = errors.New("reward is out of stock")
	ErrRewardAlreadyOwned = errors.New("reward already owned")
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

		// Populate the reward in the claim
		claim.Reward = reward

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &claim, nil
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
