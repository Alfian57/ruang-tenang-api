package application

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/internal/features/gamification/infrastructure")

type ExpHistoryService struct {
	expHistoryRepo *infrastructure.ExpHistoryRepository
}

func NewExpHistoryService(expHistoryRepo *infrastructure.ExpHistoryRepository) *ExpHistoryService {
	return &ExpHistoryService{expHistoryRepo: expHistoryRepo}
}

func (s *ExpHistoryService) GetHistory(ctx context.Context, userID uint, filter *dto.ExpHistoryFilterRequest) ([]model.ExpHistory, int64, error) {
	repoFilter := infrastructure.ExpHistoryFilter{
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

	return s.expHistoryRepo.GetByUserID(ctx, repoFilter)
}

func (s *ExpHistoryService) GetActivityTypes(ctx context.Context) ([]string, error) {
	return s.expHistoryRepo.GetActivityTypes(ctx)
}
