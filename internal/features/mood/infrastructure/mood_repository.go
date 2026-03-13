package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type UserMoodRepository struct {
	db *gorm.DB
}

func NewUserMoodRepository(db *gorm.DB) *UserMoodRepository {
	return &UserMoodRepository{db: db}
}

func (r *UserMoodRepository) Create(ctx context.Context, mood *model.UserMood) error {
	return r.db.WithContext(ctx).Create(mood).Error
}

func (r *UserMoodRepository) FindByUserID(ctx context.Context, userID uint, startDate, endDate *time.Time, page, limit int) ([]model.UserMood, int64, error) {
	var moods []model.UserMood
	var total int64

	query := r.db.WithContext(ctx).Model(&model.UserMood{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}

	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&moods).Error

	return moods, total, err
}

func (r *UserMoodRepository) GetLatestByUserID(ctx context.Context, userID uint) (*model.UserMood, error) {
	var mood model.UserMood
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").First(&mood).Error
	if err != nil {
		return nil, err
	}
	return &mood, nil
}

func (r *UserMoodRepository) GetMoodStats(ctx context.Context, userID uint, days int) (map[string]int, error) {
	stats := make(map[string]int)

	// Use Asia/Jakarta timezone (UTC+7) for consistent date handling
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}

	startDate := time.Now().In(loc).AddDate(0, 0, -days)

	var results []struct {
		Mood  string
		Count int
	}

	err = r.db.WithContext(ctx).Model(&model.UserMood{}).
		Select("mood, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Group("mood").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	for _, r := range results {
		stats[r.Mood] = r.Count
	}

	return stats, nil
}

// FindTodayByUserID finds today's mood for a user
func (r *UserMoodRepository) FindTodayByUserID(ctx context.Context, userID uint) (*model.UserMood, error) {
	var mood model.UserMood

	// Use Asia/Jakarta timezone (UTC+7) for consistent date handling
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to fixed UTC+7 offset if timezone data not available
		loc = time.FixedZone("WIB", 7*60*60)
	}

	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err = r.db.WithContext(ctx).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startOfDay, endOfDay).
		First(&mood).Error
	if err != nil {
		return nil, err
	}
	return &mood, nil
}

// Update updates an existing mood record
func (r *UserMoodRepository) Update(ctx context.Context, mood *model.UserMood) error {
	return r.db.WithContext(ctx).Save(mood).Error
}
