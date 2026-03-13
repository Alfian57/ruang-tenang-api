package application

import (
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/mystery_chest/infrastructure")

var (
	ErrChestNotFound    = errors.New("chest tidak ditemukan")
	ErrChestAlreadyOpen = errors.New("chest sudah dibuka")
	ErrNotChestOwner    = errors.New("kamu bukan pemilik chest ini")
)

type MysteryChestService struct {
	chestRepo *infrastructure.MysteryChestRepository
	userRepo  *authinfra.UserRepository
}

func NewMysteryChestService(
	chestRepo *infrastructure.MysteryChestRepository,
	userRepo *authinfra.UserRepository,
) *MysteryChestService {
	return &MysteryChestService{
		chestRepo: chestRepo,
		userRepo:  userRepo,
	}
}

// AwardChest gives a chest to a user
func (s *MysteryChestService) AwardChest(ctx context.Context, userID uint, rarity model.ChestRarity, triggerType, triggerDesc string) (*dto.UserChestResponse, error) {
	chest := &model.UserChest{
		UserID:             userID,
		Rarity:             rarity,
		TriggerType:        triggerType,
		TriggerDescription: triggerDesc,
	}

	if err := s.chestRepo.Create(ctx, chest); err != nil {
		return nil, err
	}

	return s.toChestResponse(chest), nil
}

// GetMyChests returns paginated chests
func (s *MysteryChestService) GetMyChests(ctx context.Context, userID uint, filter dto.ChestFilterRequest) ([]dto.UserChestResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		filter.Limit = 10
	}

	chests, total, err := s.chestRepo.GetUserChests(ctx, userID, filter.IsOpened, filter.Rarity, filter.Page, filter.Limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserChestResponse, len(chests))
	for i, c := range chests {
		result[i] = *s.toChestResponse(&c)
	}

	return result, total, nil
}

// OpenChest opens a chest and awards random rewards
func (s *MysteryChestService) OpenChest(ctx context.Context, userID uint, chestID uuid.UUID) (*dto.OpenChestResponse, error) {
	chest, err := s.chestRepo.GetByID(ctx, chestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChestNotFound
		}
		return nil, err
	}

	if chest.UserID != userID {
		return nil, ErrNotChestOwner
	}

	if chest.IsOpened {
		return nil, ErrChestAlreadyOpen
	}

	// Generate reward based on rarity
	rewardType, rewardValue, rewardLabel := s.generateReward(chest.Rarity)

	now := time.Now()
	chest.IsOpened = true
	chest.RewardType = rewardType
	chest.RewardValue = rewardValue
	chest.RewardLabel = rewardLabel
	chest.OpenedAt = &now

	if err := s.chestRepo.Update(ctx, chest); err != nil {
		return nil, err
	}

	// Apply reward to user
	s.applyReward(ctx, userID, rewardType, rewardValue)

	return &dto.OpenChestResponse{
		ChestID:     chest.ID,
		Rarity:      string(chest.Rarity),
		RewardType:  string(rewardType),
		RewardValue: rewardValue,
		RewardLabel: rewardLabel,
	}, nil
}

func (s *MysteryChestService) generateReward(rarity model.ChestRarity) (model.ChestRewardType, int, string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	switch rarity {
	case model.ChestLegendary:
		rewards := []struct {
			t model.ChestRewardType
			v int
			l string
		}{
			{model.ChestRewardXP, 500, "+500 XP"},
			{model.ChestRewardCoins, 200, "+200 Koin"},
			{model.ChestRewardXPBoost, 60, "XP Boost 60 menit"},
			{model.ChestRewardStreakFreeze, 3, "3x Streak Freeze"},
		}
		pick := rewards[r.Intn(len(rewards))]
		return pick.t, pick.v, pick.l
	case model.ChestEpic:
		rewards := []struct {
			t model.ChestRewardType
			v int
			l string
		}{
			{model.ChestRewardXP, 200, "+200 XP"},
			{model.ChestRewardCoins, 100, "+100 Koin"},
			{model.ChestRewardXPBoost, 30, "XP Boost 30 menit"},
		}
		pick := rewards[r.Intn(len(rewards))]
		return pick.t, pick.v, pick.l
	case model.ChestRare:
		rewards := []struct {
			t model.ChestRewardType
			v int
			l string
		}{
			{model.ChestRewardXP, 100, "+100 XP"},
			{model.ChestRewardCoins, 50, "+50 Koin"},
			{model.ChestRewardStreakFreeze, 1, "1x Streak Freeze"},
		}
		pick := rewards[r.Intn(len(rewards))]
		return pick.t, pick.v, pick.l
	default: // common
		rewards := []struct {
			t model.ChestRewardType
			v int
			l string
		}{
			{model.ChestRewardXP, 25, "+25 XP"},
			{model.ChestRewardCoins, 10, "+10 Koin"},
		}
		pick := rewards[r.Intn(len(rewards))]
		return pick.t, pick.v, pick.l
	}
}

func (s *MysteryChestService) applyReward(ctx context.Context, userID uint, rewardType model.ChestRewardType, value int) {
	switch rewardType {
	case model.ChestRewardXP:
		s.userRepo.AddExp(ctx, userID, int64(value))
	case model.ChestRewardCoins:
		s.userRepo.AddGoldCoins(ctx, userID, int64(value))
	}
	// streak_freeze and xp_boost handled by caller or separate service
}

func (s *MysteryChestService) toChestResponse(chest *model.UserChest) *dto.UserChestResponse {
	return &dto.UserChestResponse{
		ID:                 chest.ID,
		Rarity:             string(chest.Rarity),
		RarityIcon:         chest.RarityIcon(),
		IsOpened:           chest.IsOpened,
		RewardType:         string(chest.RewardType),
		RewardValue:        chest.RewardValue,
		RewardLabel:        chest.RewardLabel,
		TriggerType:        chest.TriggerType,
		TriggerDescription: chest.TriggerDescription,
		OpenedAt:           chest.OpenedAt,
		CreatedAt:          chest.CreatedAt,
	}
}
