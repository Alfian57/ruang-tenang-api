package service

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type ExpHistoryService struct {
	expHistoryRepo *repository.ExpHistoryRepository
}

func NewExpHistoryService(expHistoryRepo *repository.ExpHistoryRepository) *ExpHistoryService {
	return &ExpHistoryService{expHistoryRepo: expHistoryRepo}
}

func (s *ExpHistoryService) GetHistory(userID uint, filter *dto.ExpHistoryFilterRequest) ([]model.ExpHistory, int64, error) {
	repoFilter := repository.ExpHistoryFilter{
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
