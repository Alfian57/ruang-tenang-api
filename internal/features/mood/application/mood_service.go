package application

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/mood/infrastructure")

type MoodService struct {
	moodRepo *infrastructure.UserMoodRepository
}

func NewMoodService(moodRepo *infrastructure.UserMoodRepository) *MoodService {
	return &MoodService{moodRepo: moodRepo}
}

func (s *MoodService) RecordMood(ctx context.Context, userID uint, req *dto.CreateMoodRequest) (*dto.UserMoodDTO, error) {
	// Check if user already has a mood recorded for today
	existingMood, err := s.moodRepo.FindTodayByUserID(ctx, userID)

	if err == nil && existingMood != nil {
		// Update existing mood for today
		existingMood.Mood = model.MoodType(req.Mood)
		if err := s.moodRepo.Update(ctx, existingMood); err != nil {
			return nil, err
		}
		return &dto.UserMoodDTO{
			ID:        existingMood.ID,
			Mood:      string(existingMood.Mood),
			Emoji:     existingMood.GetMoodEmoji(),
			CreatedAt: existingMood.CreatedAt,
		}, nil
	}

	// Create new mood entry
	mood := &model.UserMood{
		UserID: userID,
		Mood:   model.MoodType(req.Mood),
	}

	if err := s.moodRepo.Create(ctx, mood); err != nil {
		return nil, err
	}

	return &dto.UserMoodDTO{
		ID:        mood.ID,
		Mood:      string(mood.Mood),
		Emoji:     mood.GetMoodEmoji(),
		CreatedAt: mood.CreatedAt,
	}, nil
}

func (s *MoodService) GetMoodHistory(ctx context.Context, userID uint, params *dto.MoodQueryParams) (*dto.MoodHistoryDTO, error) {
	var startDate, endDate *time.Time

	if params.StartDate != "" {
		t, err := time.Parse("2006-01-02", params.StartDate)
		if err == nil {
			startDate = &t
		}
	}

	if params.EndDate != "" {
		t, err := time.Parse("2006-01-02", params.EndDate)
		if err == nil {
			// Add 24 hours to include the end date
			t = t.Add(24 * time.Hour)
			endDate = &t
		}
	}

	moods, total, err := s.moodRepo.FindByUserID(ctx, userID, startDate, endDate, params.Page, params.Limit)
	if err != nil {
		return nil, err
	}

	var result []dto.UserMoodDTO
	for _, mood := range moods {
		result = append(result, dto.UserMoodDTO{
			ID:        mood.ID,
			Mood:      string(mood.Mood),
			Emoji:     mood.GetMoodEmoji(),
			CreatedAt: mood.CreatedAt,
		})
	}

	return &dto.MoodHistoryDTO{
		Moods:      result,
		TotalCount: total,
	}, nil
}

func (s *MoodService) GetLatestMood(ctx context.Context, userID uint) (*dto.UserMoodDTO, error) {
	mood, err := s.moodRepo.GetLatestByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserMoodDTO{
		ID:        mood.ID,
		Mood:      string(mood.Mood),
		Emoji:     mood.GetMoodEmoji(),
		CreatedAt: mood.CreatedAt,
	}, nil
}

func (s *MoodService) GetMoodStats(ctx context.Context, userID uint, days int) (map[string]int, error) {
	return s.moodRepo.GetMoodStats(ctx, userID, days)
}

func (s *MoodService) GetTodayMood(ctx context.Context, userID uint) (*dto.TodayMoodResponse, error) {
	mood, err := s.moodRepo.FindTodayByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		// No mood found for today
		return &dto.TodayMoodResponse{
			HasChecked: false,
			Mood:       nil,
		}, nil
	}

	return &dto.TodayMoodResponse{
		HasChecked: true,
		Mood: &dto.UserMoodDTO{
			ID:        mood.ID,
			Mood:      string(mood.Mood),
			Emoji:     mood.GetMoodEmoji(),
			CreatedAt: mood.CreatedAt,
		},
	}, nil
}
