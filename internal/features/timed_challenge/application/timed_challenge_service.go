package application

import (
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/timed_challenge/infrastructure")

var (
	ErrTimedChallengeNotFound         = errors.New("timed challenge tidak ditemukan")
	ErrTimedChallengeTemplateNotFound = errors.New("template challenge tidak ditemukan")
	ErrAlreadyHasActiveChallenge      = errors.New("kamu sudah memiliki challenge aktif")
	ErrChallengeNotActive             = errors.New("challenge tidak dalam status aktif")
	ErrNotChallengeOwner              = errors.New("kamu bukan pemilik challenge ini")
)

type TimedChallengeService struct {
	challengeRepo *infrastructure.TimedChallengeRepository
	userRepo      *authinfra.UserRepository
}

func NewTimedChallengeService(
	challengeRepo *infrastructure.TimedChallengeRepository,
	userRepo *authinfra.UserRepository,
) *TimedChallengeService {
	return &TimedChallengeService{
		challengeRepo: challengeRepo,
		userRepo:      userRepo,
	}
}

// GetTemplates returns available challenge templates
func (s *TimedChallengeService) GetTemplates(ctx context.Context) ([]dto.TimedChallengeTemplateResponse, error) {
	templates, err := s.challengeRepo.GetActiveTemplates(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.TimedChallengeTemplateResponse, len(templates))
	for i, t := range templates {
		result[i] = s.toTemplateResponse(&t)
	}

	return result, nil
}

// StartChallenge begins a new timed challenge
func (s *TimedChallengeService) StartChallenge(ctx context.Context, userID uint, req dto.StartTimedChallengeRequest) (*dto.UserTimedChallengeResponse, error) {
	// Check for existing active challenge
	_, err := s.challengeRepo.GetActiveChallenge(ctx, userID)
	if err == nil {
		return nil, ErrAlreadyHasActiveChallenge
	}

	// Expire overdue challenges first
	s.challengeRepo.ExpireOverdueChallenges(ctx)

	template, err := s.challengeRepo.GetTemplateByID(ctx, req.TemplateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTimedChallengeTemplateNotFound
		}
		return nil, err
	}

	now := time.Now()
	challenge := &model.UserTimedChallenge{
		UserID:     userID,
		TemplateID: template.ID,
		Status:     model.TimedChallengeActive,
		StartedAt:  now,
		ExpiresAt:  now.Add(time.Duration(template.DurationMinutes) * time.Minute),
	}

	if err := s.challengeRepo.Create(ctx, challenge); err != nil {
		return nil, err
	}

	challenge.Template = template
	return s.toChallengeResponse(challenge), nil
}

// GetActiveChallenge returns user's current active challenge
func (s *TimedChallengeService) GetActiveChallenge(ctx context.Context, userID uint) (*dto.UserTimedChallengeResponse, error) {
	challenge, err := s.challengeRepo.GetActiveChallenge(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTimedChallengeNotFound
		}
		return nil, err
	}

	// Check if expired
	if challenge.IsExpired() {
		challenge.Status = model.TimedChallengeExpired
		s.challengeRepo.Update(ctx, challenge)
		return nil, ErrTimedChallengeNotFound
	}

	return s.toChallengeResponse(challenge), nil
}

// GetMyHistory returns past challenges
func (s *TimedChallengeService) GetMyHistory(ctx context.Context, userID uint, filter dto.TimedChallengeFilterRequest) ([]dto.UserTimedChallengeResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		filter.Limit = 10
	}

	challenges, total, err := s.challengeRepo.GetUserChallenges(ctx, userID, filter.Status, filter.Page, filter.Limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserTimedChallengeResponse, len(challenges))
	for i, c := range challenges {
		result[i] = *s.toChallengeResponse(&c)
	}

	return result, total, nil
}

// CompleteChallenge marks a challenge as completed
func (s *TimedChallengeService) CompleteChallenge(ctx context.Context, userID uint, challengeID uuid.UUID) (*dto.UserTimedChallengeResponse, error) {
	challenge, err := s.challengeRepo.GetByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTimedChallengeNotFound
		}
		return nil, err
	}

	if challenge.UserID != userID {
		return nil, ErrNotChallengeOwner
	}

	if challenge.Status != model.TimedChallengeActive {
		return nil, ErrChallengeNotActive
	}

	if challenge.IsExpired() {
		challenge.Status = model.TimedChallengeExpired
		s.challengeRepo.Update(ctx, challenge)
		return nil, ErrChallengeNotActive
	}

	now := time.Now()
	challenge.Status = model.TimedChallengeCompleted
	challenge.CompletedAt = &now

	if err := s.challengeRepo.Update(ctx, challenge); err != nil {
		return nil, err
	}

	// Award rewards
	if challenge.Template != nil {
		if challenge.Template.XPReward > 0 {
			s.userRepo.AddExp(ctx, userID, int64(challenge.Template.XPReward))
		}
		if challenge.Template.CoinReward > 0 {
			s.userRepo.AddGoldCoins(ctx, userID, int64(challenge.Template.CoinReward))
		}
	}

	return s.toChallengeResponse(challenge), nil
}

func (s *TimedChallengeService) toTemplateResponse(t *model.TimedChallengeTemplate) dto.TimedChallengeTemplateResponse {
	return dto.TimedChallengeTemplateResponse{
		ID:              t.ID,
		Title:           t.Title,
		Description:     t.Description,
		ChallengeType:   t.ChallengeType,
		TargetValue:     t.TargetValue,
		DurationMinutes: t.DurationMinutes,
		XPReward:        t.XPReward,
		CoinReward:      t.CoinReward,
		Icon:            t.Icon,
	}
}

func (s *TimedChallengeService) toChallengeResponse(c *model.UserTimedChallenge) *dto.UserTimedChallengeResponse {
	targetValue := 0
	var templateResp dto.TimedChallengeTemplateResponse
	if c.Template != nil {
		targetValue = c.Template.TargetValue
		templateResp = s.toTemplateResponse(c.Template)
	}

	progressPercent := 0.0
	if targetValue > 0 {
		progressPercent = float64(c.CurrentValue) / float64(targetValue) * 100
		if progressPercent > 100 {
			progressPercent = 100
		}
	}

	return &dto.UserTimedChallengeResponse{
		ID:               c.ID,
		Template:         templateResp,
		CurrentValue:     c.CurrentValue,
		TargetValue:      targetValue,
		ProgressPercent:  progressPercent,
		Status:           string(c.Status),
		StartedAt:        c.StartedAt,
		ExpiresAt:        c.ExpiresAt,
		RemainingSeconds: c.RemainingSeconds(),
		CompletedAt:      c.CompletedAt,
	}
}
