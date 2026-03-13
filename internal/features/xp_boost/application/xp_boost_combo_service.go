package application

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/infrastructure")

const comboTimeoutMinutes = 30

var (
	ErrBoostAlreadyActive = errors.New("kamu sudah memiliki XP boost aktif")
	ErrNoActiveBoost      = errors.New("tidak ada XP boost aktif")
)

type XPBoostComboService struct {
	boostRepo *infrastructure.XPBoostComboRepository
}

func NewXPBoostComboService(boostRepo *infrastructure.XPBoostComboRepository) *XPBoostComboService {
	return &XPBoostComboService{boostRepo: boostRepo}
}

// === XP Boost ===

// GetActiveBoost returns current active boost
func (s *XPBoostComboService) GetActiveBoost(ctx context.Context, userID uint) (*dto.XPBoostResponse, error) {
	boost, err := s.boostRepo.GetActiveBoost(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoActiveBoost
		}
		return nil, err
	}

	remaining := int(time.Until(boost.ExpiresAt).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	return &dto.XPBoostResponse{
		ID:               boost.ID,
		Multiplier:       boost.Multiplier,
		TriggerType:      string(boost.TriggerType),
		StartedAt:        boost.StartedAt,
		ExpiresAt:        boost.ExpiresAt,
		RemainingSeconds: remaining,
	}, nil
}

// ActivateBoost creates a new XP boost
func (s *XPBoostComboService) ActivateBoost(ctx context.Context, userID uint, multiplier float64, durationMins int, trigger model.XPBoostTrigger) (*dto.XPBoostResponse, error) {
	_, err := s.boostRepo.GetActiveBoost(ctx, userID)
	if err == nil {
		return nil, ErrBoostAlreadyActive
	}

	now := time.Now()
	boost := &model.XPBoost{
		UserID:      userID,
		Multiplier:  multiplier,
		TriggerType: trigger,
		StartedAt:   now,
		ExpiresAt:   now.Add(time.Duration(durationMins) * time.Minute),
		IsActive:    true,
	}

	if err := s.boostRepo.CreateBoost(ctx, boost); err != nil {
		return nil, err
	}

	return &dto.XPBoostResponse{
		ID:               boost.ID,
		Multiplier:       boost.Multiplier,
		TriggerType:      string(boost.TriggerType),
		StartedAt:        boost.StartedAt,
		ExpiresAt:        boost.ExpiresAt,
		RemainingSeconds: durationMins * 60,
	}, nil
}

// GetEffectiveMultiplier returns the combined boost+combo multiplier
func (s *XPBoostComboService) GetEffectiveMultiplier(ctx context.Context, userID uint) float64 {
	mult := 1.0

	boost, err := s.boostRepo.GetActiveBoost(ctx, userID)
	if err == nil && !boost.IsExpired() {
		mult *= boost.Multiplier
	}

	combo, err := s.boostRepo.GetCombo(ctx, userID)
	if err == nil && combo.Multiplier > 1.0 {
		if combo.LastActivityAt != nil && time.Since(*combo.LastActivityAt).Minutes() < comboTimeoutMinutes {
			mult *= combo.Multiplier
		}
	}

	return mult
}

// === Combo ===

// GetComboStatus returns current combo state
func (s *XPBoostComboService) GetComboStatus(ctx context.Context, userID uint) (*dto.ComboStatusResponse, error) {
	combo, err := s.boostRepo.GetCombo(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.ComboStatusResponse{
				ComboCount:     0,
				Multiplier:     1.0,
				NextMultiplier: 1.5,
				ExpiresInSecs:  0,
			}, nil
		}
		return nil, err
	}

	expiresIn := 0
	if combo.LastActivityAt != nil {
		remaining := comboTimeoutMinutes*60 - int(time.Since(*combo.LastActivityAt).Seconds())
		if remaining > 0 {
			expiresIn = remaining
		}
	}

	return &dto.ComboStatusResponse{
		ComboCount:     combo.ComboCount,
		Multiplier:     combo.Multiplier,
		NextMultiplier: model.ComboMultiplier(combo.ComboCount + 1),
		LastActivity:   combo.LastActivityType,
		LastActivityAt: combo.LastActivityAt,
		ExpiresInSecs:  expiresIn,
	}, nil
}

// IncrementCombo advances the combo chain
func (s *XPBoostComboService) IncrementCombo(ctx context.Context, userID uint, activityType string) (*dto.ComboStatusResponse, error) {
	now := time.Now()

	combo, err := s.boostRepo.GetCombo(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			combo = &model.UserCombo{
				UserID:           userID,
				ComboCount:       1,
				Multiplier:       model.ComboMultiplier(1),
				LastActivityType: activityType,
				LastActivityAt:   &now,
				SessionStartedAt: &now,
			}
			if err := s.boostRepo.UpsertCombo(ctx, combo); err != nil {
				return nil, err
			}
			return s.GetComboStatus(ctx, userID)
		}
		return nil, err
	}

	// Check if combo expired
	if combo.LastActivityAt != nil && time.Since(*combo.LastActivityAt).Minutes() >= comboTimeoutMinutes {
		combo.ComboCount = 0
		combo.SessionStartedAt = &now
	}

	// Different activity type advances combo
	if combo.LastActivityType != activityType {
		combo.ComboCount++
	}

	combo.Multiplier = model.ComboMultiplier(combo.ComboCount)
	combo.LastActivityType = activityType
	combo.LastActivityAt = &now

	if err := s.boostRepo.UpsertCombo(ctx, combo); err != nil {
		return nil, err
	}

	return s.GetComboStatus(ctx, userID)
}
