package services

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type MoodService struct {
	moodRepo *repositories.UserMoodRepository
}

func NewMoodService(moodRepo *repositories.UserMoodRepository) *MoodService {
	return &MoodService{moodRepo: moodRepo}
}

func (s *MoodService) RecordMood(userID uint, req *dto.CreateMoodRequest) (*dto.UserMoodDTO, error) {
	mood := &models.UserMood{
		UserID: userID,
		Mood:   models.MoodType(req.Mood),
	}

	if err := s.moodRepo.Create(mood); err != nil {
		return nil, err
	}

	return &dto.UserMoodDTO{
		ID:        mood.ID,
		Mood:      string(mood.Mood),
		Emoji:     mood.GetMoodEmoji(),
		CreatedAt: mood.CreatedAt,
	}, nil
}

func (s *MoodService) GetMoodHistory(userID uint, params *dto.MoodQueryParams) (*dto.MoodHistoryDTO, error) {
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

	moods, total, err := s.moodRepo.FindByUserID(userID, startDate, endDate, params.Page, params.Limit)
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

func (s *MoodService) GetLatestMood(userID uint) (*dto.UserMoodDTO, error) {
	mood, err := s.moodRepo.GetLatestByUserID(userID)
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

func (s *MoodService) GetMoodStats(userID uint, days int) (map[string]int, error) {
	return s.moodRepo.GetMoodStats(userID, days)
}
