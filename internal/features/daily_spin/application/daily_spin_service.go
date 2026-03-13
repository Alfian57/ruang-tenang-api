package application

import (
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/internal/features/daily_spin/infrastructure")

var (
	ErrAlreadySpunToday = errors.New("kamu sudah spin hari ini, coba lagi besok")
	ErrNoSpinRewards    = errors.New("tidak ada reward spin yang tersedia")
)

type DailySpinService struct {
	spinRepo *infrastructure.DailySpinRepository
	userRepo *authinfra.UserRepository
}

func NewDailySpinService(
	spinRepo *infrastructure.DailySpinRepository,
	userRepo *authinfra.UserRepository,
) *DailySpinService {
	return &DailySpinService{
		spinRepo: spinRepo,
		userRepo: userRepo,
	}
}

// GetWheel returns the spin wheel with all slots and user status
func (s *DailySpinService) GetWheel(ctx context.Context, userID uint) (*dto.SpinWheelResponse, error) {
	rewards, err := s.spinRepo.GetActiveRewards(ctx)
	if err != nil {
		return nil, err
	}

	slots := make([]dto.SpinRewardSlotResponse, len(rewards))
	for i, r := range rewards {
		slots[i] = dto.SpinRewardSlotResponse{
			ID:          r.ID,
			Name:        r.Name,
			Icon:        r.Icon,
			RewardType:  string(r.RewardType),
			RewardValue: r.RewardValue,
			Rarity:      r.Rarity,
		}
	}

	hasSpun, _ := s.spinRepo.HasSpunToday(ctx, userID)

	var lastSpinAt *time.Time
	lastSpin, err := s.spinRepo.GetLastSpin(ctx, userID)
	if err == nil {
		lastSpinAt = &lastSpin.CreatedAt
	}

	return &dto.SpinWheelResponse{
		Slots:        slots,
		HasSpunToday: hasSpun,
		LastSpinAt:   lastSpinAt,
	}, nil
}

// Spin performs the daily spin
func (s *DailySpinService) Spin(ctx context.Context, userID uint) (*dto.SpinResultResponse, error) {
	hasSpun, err := s.spinRepo.HasSpunToday(ctx, userID)
	if err != nil {
		return nil, err
	}
	if hasSpun {
		return nil, ErrAlreadySpunToday
	}

	rewards, err := s.spinRepo.GetActiveRewards(ctx)
	if err != nil {
		return nil, err
	}
	if len(rewards) == 0 {
		return nil, ErrNoSpinRewards
	}

	// Weighted random selection
	selectedIdx, selected := s.weightedRandom(rewards)

	// Record spin
	today := time.Now().Truncate(24 * time.Hour)
	spin := &model.UserSpin{
		UserID:   userID,
		RewardID: selected.ID,
		SpinDate: today,
	}
	if err := s.spinRepo.CreateSpin(ctx, spin); err != nil {
		return nil, err
	}

	// Apply reward
	s.applySpinReward(ctx, userID, &selected)

	return &dto.SpinResultResponse{
		SlotIndex:   selectedIdx,
		RewardName:  selected.Name,
		RewardIcon:  selected.Icon,
		RewardType:  string(selected.RewardType),
		RewardValue: selected.RewardValue,
		Rarity:      selected.Rarity,
	}, nil
}

func (s *DailySpinService) weightedRandom(rewards []model.SpinReward) (int, model.SpinReward) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	totalWeight := 0
	for _, rw := range rewards {
		totalWeight += rw.Weight
	}

	pick := r.Intn(totalWeight)
	cumulative := 0
	for i, rw := range rewards {
		cumulative += rw.Weight
		if pick < cumulative {
			return i, rw
		}
	}

	return 0, rewards[0]
}

func (s *DailySpinService) applySpinReward(ctx context.Context, userID uint, reward *model.SpinReward) {
	switch reward.RewardType {
	case model.SpinRewardXP:
		s.userRepo.AddExp(ctx, userID, int64(reward.RewardValue))
	case model.SpinRewardCoins:
		s.userRepo.AddGoldCoins(ctx, userID, int64(reward.RewardValue))
	}
	// streak_freeze, xp_boost, nothing — handled by caller or silently ignored
}
