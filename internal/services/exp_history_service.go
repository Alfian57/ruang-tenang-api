package services

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type ExpHistoryService struct {
	expHistoryRepo *repositories.ExpHistoryRepository
}

func NewExpHistoryService(expHistoryRepo *repositories.ExpHistoryRepository) *ExpHistoryService {
	return &ExpHistoryService{expHistoryRepo: expHistoryRepo}
}

func (s *ExpHistoryService) GetHistory(userID uint, filter *dto.ExpHistoryFilterRequest) ([]models.ExpHistory, int64, error) {
	repoFilter := repositories.ExpHistoryFilter{
		UserID:       userID,
		ActivityType: filter.ActivityType,
		Page:         filter.Page,
		Limit:        filter.Limit,
	}

	// Parse dates if provided
	if filter.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", filter.StartDate)
		if err == nil {
			repoFilter.StartDate = &startDate
		}
	}

	if filter.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", filter.EndDate)
		if err == nil {
			repoFilter.EndDate = &endDate
		}
	}

	return s.expHistoryRepo.GetByUserID(repoFilter)
}

func (s *ExpHistoryService) GetActivityTypes() ([]string, error) {
	return s.expHistoryRepo.GetActivityTypes()
}
