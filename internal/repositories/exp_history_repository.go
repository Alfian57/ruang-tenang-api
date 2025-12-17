package repositories

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

type ExpHistoryRepository struct {
	db *gorm.DB
}

func NewExpHistoryRepository(db *gorm.DB) *ExpHistoryRepository {
	return &ExpHistoryRepository{db: db}
}

type ExpHistoryFilter struct {
	UserID       uint
	ActivityType string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	Limit        int
}

func (r *ExpHistoryRepository) GetByUserID(filter ExpHistoryFilter) ([]models.ExpHistory, int64, error) {
	var histories []models.ExpHistory
	var total int64

	query := r.db.Model(&models.ExpHistory{}).Where("user_id = ?", filter.UserID)

	if filter.ActivityType != "" {
		query = query.Where("activity_type = ?", filter.ActivityType)
	}

	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}

	if filter.EndDate != nil {
		// Add 1 day to include the end date fully
		endDate := filter.EndDate.Add(24 * time.Hour)
		query = query.Where("created_at < ?", endDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&histories).Error
	if err != nil {
		return nil, 0, err
	}

	return histories, total, nil
}

func (r *ExpHistoryRepository) Create(history *models.ExpHistory) error {
	return r.db.Create(history).Error
}

func (r *ExpHistoryRepository) GetTotalExpByUserID(userID uint) (int64, error) {
	var total int64
	err := r.db.Model(&models.ExpHistory{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(points), 0)").
		Scan(&total).Error
	return total, err
}

// GetActivityTypes returns distinct activity types for filter dropdown
func (r *ExpHistoryRepository) GetActivityTypes() ([]string, error) {
	var types []string
	err := r.db.Model(&models.ExpHistory{}).
		Distinct("activity_type").
		Pluck("activity_type", &types).Error
	return types, err
}
