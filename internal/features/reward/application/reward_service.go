package application

import (
	"context"
	"errors"

	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/features/reward/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

var (
	ErrRewardNotFound     = errors.New("reward not found")
	ErrThemeNotOwned      = errors.New("tema belum dimiliki")
	ErrThemeLockedForRole = errors.New("tema dashboard terkunci untuk role ini")
)

type RewardService interface {
	// Member endpoints
	GetAvailableRewards(ctx context.Context) ([]model.Reward, error)
	GetRewardByID(ctx context.Context, id uint) (*model.Reward, error)
	ClaimReward(ctx context.Context, userID uint, rewardID uint) (*RewardClaimResult, error)
	GetUserClaims(ctx context.Context, userID uint, page, pageSize int) (*RewardClaimListResult, error)
	GetCoinBalance(ctx context.Context, userID uint) (int64, error)

	// Theme endpoints
	GetOwnedThemes(ctx context.Context, userID uint) ([]string, string, error)
	ActivateTheme(ctx context.Context, userID uint, theme string) error

	// Admin endpoints
	GetAllRewards(ctx context.Context) ([]model.Reward, error)
	CreateReward(ctx context.Context, reward *model.Reward) error
	UpdateReward(ctx context.Context, id uint, input UpdateRewardInput) (*model.Reward, error)
	DeleteReward(ctx context.Context, id uint) error
	GetAllClaims(ctx context.Context, page, pageSize int) (*RewardClaimListResult, error)
}

type RewardClaimResult struct {
	Claim          model.RewardClaim `json:"claim"`
	RemainingCoins int64             `json:"remaining_coins"`
}

type RewardClaimListResult struct {
	Claims     []model.RewardClaim `json:"claims"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

type UpdateRewardInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Image       *string `json:"image"`
	CoinCost    *int    `json:"coin_cost"`
	Stock       *int    `json:"stock"`
	IsActive    *bool   `json:"is_active"`
}

type rewardService struct {
	rewardRepo *infrastructure.RewardRepository
	userRepo   *authinfra.UserRepository
}

func NewRewardService(rewardRepo *infrastructure.RewardRepository, userRepo *authinfra.UserRepository) RewardService {
	return &rewardService{
		rewardRepo: rewardRepo,
		userRepo:   userRepo,
	}
}

func (s *rewardService) GetAvailableRewards(ctx context.Context) ([]model.Reward, error) {
	return s.rewardRepo.GetAllRewards(ctx, true)
}

func (s *rewardService) GetAllRewards(ctx context.Context) ([]model.Reward, error) {
	return s.rewardRepo.GetAllRewards(ctx, false)
}

func (s *rewardService) GetRewardByID(ctx context.Context, id uint) (*model.Reward, error) {
	reward, err := s.rewardRepo.GetRewardByID(ctx, id)
	if err != nil {
		return nil, ErrRewardNotFound
	}
	return reward, nil
}

func (s *rewardService) ClaimReward(ctx context.Context, userID uint, rewardID uint) (*RewardClaimResult, error) {
	claim, err := s.rewardRepo.ClaimReward(ctx, userID, rewardID)
	if err != nil {
		switch {
		case errors.Is(err, infrastructure.ErrInsufficientCoins):
			return nil, infrastructure.ErrInsufficientCoins
		case errors.Is(err, infrastructure.ErrRewardUnavailable):
			return nil, infrastructure.ErrRewardUnavailable
		case errors.Is(err, infrastructure.ErrRewardOutOfStock):
			return nil, infrastructure.ErrRewardOutOfStock
		case errors.Is(err, infrastructure.ErrRewardAlreadyOwned):
			return nil, infrastructure.ErrRewardAlreadyOwned
		}
		return nil, err
	}

	// Get updated coin balance
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Auto-activate dashboard themes for regular users only.
	if user.Role == model.RoleUser && claim.Reward.RewardType == model.RewardTypeTheme && claim.Reward.RewardValue != "" {
		_ = s.userRepo.UpdateField(ctx, userID, "profile_theme", claim.Reward.RewardValue)
	}

	return &RewardClaimResult{
		Claim:          *claim,
		RemainingCoins: user.GoldCoins,
	}, nil
}

func (s *rewardService) GetUserClaims(ctx context.Context, userID uint, page, pageSize int) (*RewardClaimListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize
	claims, total, err := s.rewardRepo.GetUserClaims(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &RewardClaimListResult{
		Claims:     claims,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *rewardService) GetCoinBalance(ctx context.Context, userID uint) (int64, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return 0, err
	}
	return user.GoldCoins, nil
}

func (s *rewardService) CreateReward(ctx context.Context, reward *model.Reward) error {
	return s.rewardRepo.CreateReward(ctx, reward)
}

func (s *rewardService) UpdateReward(ctx context.Context, id uint, input UpdateRewardInput) (*model.Reward, error) {
	reward, err := s.rewardRepo.GetRewardByID(ctx, id)
	if err != nil {
		return nil, ErrRewardNotFound
	}

	if input.Name != nil {
		reward.Name = *input.Name
	}
	if input.Description != nil {
		reward.Description = *input.Description
	}
	if input.Image != nil {
		reward.Image = *input.Image
	}
	if input.CoinCost != nil {
		reward.CoinCost = *input.CoinCost
	}
	if input.Stock != nil {
		reward.Stock = *input.Stock
	}
	if input.IsActive != nil {
		reward.IsActive = *input.IsActive
	}

	if err := s.rewardRepo.UpdateReward(ctx, reward); err != nil {
		return nil, err
	}
	return reward, nil
}

func (s *rewardService) DeleteReward(ctx context.Context, id uint) error {
	_, err := s.rewardRepo.GetRewardByID(ctx, id)
	if err != nil {
		return ErrRewardNotFound
	}
	return s.rewardRepo.DeleteReward(ctx, id)
}

func (s *rewardService) GetOwnedThemes(ctx context.Context, userID uint) ([]string, string, error) {
	// "default" is always owned
	owned := []string{"default"}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return owned, "default", err
	}
	if user.Role != model.RoleUser {
		return owned, "default", nil
	}

	claims, _, err := s.rewardRepo.GetUserClaims(ctx, userID, 1000, 0)
	if err != nil {
		return owned, "default", err
	}

	for _, claim := range claims {
		if claim.Reward.RewardType == model.RewardTypeTheme && claim.Reward.RewardValue != "" {
			owned = append(owned, claim.Reward.RewardValue)
		}
	}

	activeTheme := user.ProfileTheme
	if activeTheme == "" {
		activeTheme = "default"
	}

	return owned, activeTheme, nil
}

func (s *rewardService) ActivateTheme(ctx context.Context, userID uint, theme string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role != model.RoleUser && theme != "default" {
		return ErrThemeLockedForRole
	}

	// "default" is always available
	if theme == "default" {
		return s.userRepo.UpdateField(ctx, userID, "profile_theme", "default")
	}

	// Check if user owns this theme
	claims, _, err := s.rewardRepo.GetUserClaims(ctx, userID, 1000, 0)
	if err != nil {
		return err
	}

	for _, claim := range claims {
		if claim.Reward.RewardType == model.RewardTypeTheme && claim.Reward.RewardValue == theme {
			return s.userRepo.UpdateField(ctx, userID, "profile_theme", theme)
		}
	}

	return ErrThemeNotOwned
}

func (s *rewardService) GetAllClaims(ctx context.Context, page, pageSize int) (*RewardClaimListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize
	claims, total, err := s.rewardRepo.GetAllClaims(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &RewardClaimListResult{
		Claims:     claims,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}
