package service

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrFriendQuestNotFound = errors.New("friend quest tidak ditemukan")
	ErrCannotQuestYourself = errors.New("tidak bisa membuat quest dengan dirimu sendiri")
	ErrQuestAlreadyExists  = errors.New("sudah ada quest aktif dengan teman ini")
	ErrNotQuestParticipant = errors.New("kamu bukan peserta quest ini")
	ErrQuestNotPending     = errors.New("quest tidak dalam status pending")
	ErrQuestNotActive      = errors.New("quest tidak dalam status aktif")
	ErrMaxActiveQuests     = errors.New("kamu sudah memiliki maksimal 5 quest aktif")
)

type FriendQuestService struct {
	questRepo *repository.FriendQuestRepository
	userRepo  *repository.UserRepository
}

func NewFriendQuestService(
	questRepo *repository.FriendQuestRepository,
	userRepo *repository.UserRepository,
) *FriendQuestService {
	return &FriendQuestService{
		questRepo: questRepo,
		userRepo:  userRepo,
	}
}

// CreateQuest sends a quest invitation to a friend
func (s *FriendQuestService) CreateQuest(ctx context.Context, userID uint, req dto.CreateFriendQuestRequest) (*dto.FriendQuestResponse, error) {
	if userID == req.PartnerID {
		return nil, ErrCannotQuestYourself
	}

	// Check max active quests
	count, _ := s.questRepo.CountActiveQuests(ctx, userID)
	if count >= 5 {
		return nil, ErrMaxActiveQuests
	}

	// Check existing quest between these users
	existing, _ := s.questRepo.GetActiveQuestsBetweenUsers(ctx, userID, req.PartnerID)
	if existing > 0 {
		return nil, ErrQuestAlreadyExists
	}

	now := time.Now()
	endsAt := now.Add(time.Duration(req.DurationHrs) * time.Hour)

	quest := &model.FriendQuest{
		RequesterID: userID,
		PartnerID:   req.PartnerID,
		Title:       req.Title,
		Description: req.Description,
		QuestType:   model.FriendQuestType(req.QuestType),
		TargetValue: req.TargetValue,
		XPReward:    req.XPReward,
		CoinReward:  req.CoinReward,
		Status:      model.FriendQuestPending,
		StartsAt:    &now,
		EndsAt:      &endsAt,
	}

	if err := s.questRepo.Create(ctx, quest); err != nil {
		return nil, err
	}

	// Reload with preloads
	quest, err := s.questRepo.GetByID(ctx, quest.ID)
	if err != nil {
		return nil, err
	}

	return s.toQuestResponse(quest), nil
}

// AcceptQuest accepts a pending quest invitation
func (s *FriendQuestService) AcceptQuest(ctx context.Context, userID uint, questID uuid.UUID) (*dto.FriendQuestResponse, error) {
	quest, err := s.questRepo.GetByID(ctx, questID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendQuestNotFound
		}
		return nil, err
	}

	if quest.PartnerID != userID {
		return nil, ErrNotQuestParticipant
	}

	if quest.Status != model.FriendQuestPending {
		return nil, ErrQuestNotPending
	}

	quest.Status = model.FriendQuestActive
	if err := s.questRepo.Update(ctx, quest); err != nil {
		return nil, err
	}

	return s.toQuestResponse(quest), nil
}

// DeclineQuest declines a pending quest invitation
func (s *FriendQuestService) DeclineQuest(ctx context.Context, userID uint, questID uuid.UUID) error {
	quest, err := s.questRepo.GetByID(ctx, questID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFriendQuestNotFound
		}
		return err
	}

	if quest.PartnerID != userID {
		return ErrNotQuestParticipant
	}

	if quest.Status != model.FriendQuestPending {
		return ErrQuestNotPending
	}

	quest.Status = model.FriendQuestDeclined
	return s.questRepo.Update(ctx, quest)
}

// GetMyQuests returns user's quests with pagination
func (s *FriendQuestService) GetMyQuests(ctx context.Context, userID uint, filter dto.FriendQuestFilterRequest) ([]dto.FriendQuestResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 50 {
		filter.Limit = 10
	}

	quests, total, err := s.questRepo.GetUserQuests(ctx, userID, filter.Status, filter.Page, filter.Limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.FriendQuestResponse, len(quests))
	for i, q := range quests {
		result[i] = *s.toQuestResponse(&q)
	}

	return result, total, nil
}

// GetQuest returns a single quest details
func (s *FriendQuestService) GetQuest(ctx context.Context, userID uint, questID uuid.UUID) (*dto.FriendQuestResponse, error) {
	quest, err := s.questRepo.GetByID(ctx, questID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFriendQuestNotFound
		}
		return nil, err
	}

	if quest.RequesterID != userID && quest.PartnerID != userID {
		return nil, ErrNotQuestParticipant
	}

	return s.toQuestResponse(quest), nil
}

func (s *FriendQuestService) toQuestResponse(q *model.FriendQuest) *dto.FriendQuestResponse {
	resp := &dto.FriendQuestResponse{
		ID:                q.ID,
		Title:             q.Title,
		Description:       q.Description,
		QuestType:         string(q.QuestType),
		TargetValue:       q.TargetValue,
		RequesterProgress: q.RequesterProgress,
		PartnerProgress:   q.PartnerProgress,
		TotalProgress:     q.TotalProgress(),
		ProgressPercent:   q.ProgressPercent(),
		XPReward:          q.XPReward,
		CoinReward:        q.CoinReward,
		Status:            string(q.Status),
		StartsAt:          q.StartsAt,
		EndsAt:            q.EndsAt,
		CompletedAt:       q.CompletedAt,
		CreatedAt:         q.CreatedAt,
	}

	if q.Requester != nil {
		resp.Requester = dto.QuestUserInfo{
			ID:       q.Requester.ID,
			Username: q.Requester.Name,
			Avatar:   q.Requester.Avatar,
		}
	}

	if q.Partner != nil {
		resp.Partner = dto.QuestUserInfo{
			ID:       q.Partner.ID,
			Username: q.Partner.Name,
			Avatar:   q.Partner.Avatar,
		}
	}

	return resp
}
