package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type CommunityProgressRepository struct {
	db *gorm.DB
}

func NewCommunityProgressRepository(db *gorm.DB) *CommunityProgressRepository {
	return &CommunityProgressRepository{db: db}
}

// GetCommunityStats retrieves the latest community stats (singleton row)
func (r *CommunityProgressRepository) GetCommunityStats(ctx context.Context) (*model.CommunityStats, error) {
	var stats model.CommunityStats
	err := r.db.WithContext(ctx).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		// Return empty stats if none exist
		return &model.CommunityStats{}, nil
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateCommunityStats updates or creates community stats
func (r *CommunityProgressRepository) UpdateCommunityStats(ctx context.Context, stats *model.CommunityStats) error {
	// Upsert - create if not exists, update if exists
	var existing model.CommunityStats
	err := r.db.WithContext(ctx).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.WithContext(ctx).Create(stats).Error
	}
	if err != nil {
		return err
	}
	stats.ID = existing.ID
	return r.db.WithContext(ctx).Save(stats).Error
}

// RecalculateCommunityStats calculates fresh stats from the database for current month
func (r *CommunityProgressRepository) RecalculateCommunityStats(ctx context.Context) (*model.CommunityStats, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	stats := &model.CommunityStats{
		Month: month,
		Year:  year,
	}

	// Total EXP earned this month
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("created_at >= ?", startOfMonth).
		Select("COALESCE(SUM(exp_earned), 0)").
		Scan(&stats.TotalXPEarned)

	// Active members this month (distinct users with activity)
	var activeMembers int64
	r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("created_at >= ?", startOfMonth).
		Distinct("user_id").
		Count(&activeMembers)
	stats.ActiveMembers = int(activeMembers)

	// Total achievements (badges earned) this month
	var totalAchievements int64
	r.db.WithContext(ctx).Model(&model.UserBadge{}).
		Where("earned_at >= ?", startOfMonth).
		Count(&totalAchievements)
	stats.TotalAchievements = int(totalAchievements)

	// New members this month
	var newMembers int64
	r.db.WithContext(ctx).Model(&model.User{}).
		Where("created_at >= ?", startOfMonth).
		Count(&newMembers)
	stats.NewMembers = int(newMembers)

	// Stories published this month
	var totalStories int64
	r.db.WithContext(ctx).Model(&model.InspiringStory{}).
		Where("created_at >= ? AND status = 'approved'", startOfMonth).
		Count(&totalStories)
	stats.TotalStoriesPublished = int(totalStories)

	// Articles published this month
	var totalArticles int64
	r.db.WithContext(ctx).Model(&model.Article{}).
		Where("created_at >= ?", startOfMonth).
		Count(&totalArticles)
	stats.TotalArticlesPublished = int(totalArticles)

	return stats, nil
}

// GetUsersInLevelRange gets users within a level range for hall of fame
func (r *CommunityProgressRepository) GetUsersInLevelRange(ctx context.Context, minLevel, maxLevel int, limit int) ([]model.User, error) {
	var users []model.User

	// Get min/max exp for the level range
	var minExp, maxExp int64
	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", minLevel).Select("min_exp").Scan(&minExp)

	if maxLevel >= 10 {
		maxExp = 999999999 // Very high number for max level
	} else {
		r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", maxLevel+1).Select("min_exp").Scan(&maxExp)
	}

	err := r.db.WithContext(ctx).Where("exp >= ? AND exp < ?", minExp, maxExp).
		Order("exp DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}

// GetTopUsersInLevel gets top users within a specific level
func (r *CommunityProgressRepository) GetTopUsersInLevel(ctx context.Context, level int, limit int) ([]model.User, error) {
	var users []model.User

	// Get exp range for this level
	var minExp int64
	var maxExp int64 = 999999999

	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", level).Select("min_exp").Scan(&minExp)
	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", level+1).Select("min_exp").Scan(&maxExp)

	if maxExp == 0 {
		maxExp = 999999999
	}

	err := r.db.WithContext(ctx).Where("exp >= ? AND exp < ?", minExp, maxExp).
		Order("exp DESC").
		Limit(limit).
		Find(&users).Error

	return users, err
}

// GetHallOfFame gets monthly hall of fame entries
func (r *CommunityProgressRepository) GetHallOfFame(ctx context.Context, month, year int, category string) ([]model.MonthlyHallOfFame, error) {
	var entries []model.MonthlyHallOfFame

	query := r.db.WithContext(ctx).Where("month = ? AND year = ?", month, year)
	if category != "" {
		query = query.Where("category = ?", category)
	}

	err := query.Order("rank ASC").Preload("User").Find(&entries).Error
	return entries, err
}

// CreateHallOfFameEntry creates a new hall of fame entry
func (r *CommunityProgressRepository) CreateHallOfFameEntry(ctx context.Context, entry *model.MonthlyHallOfFame) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// GetHallOfFameCategories returns available categories
func (r *CommunityProgressRepository) GetHallOfFameCategories(ctx context.Context) []string {
	return []string{
		"most_supportive",
		"most_consistent",
		"most_helpful_comments",
		"rising_star",
		"community_builder",
	}
}

// GetUserRankInLevel returns user's rank within their level
func (r *CommunityProgressRepository) GetUserRankInLevel(ctx context.Context, userID uint, level int) (int, int, error) {
	var minExp int64
	var maxExp int64 = 999999999

	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", level).Select("min_exp").Scan(&minExp)
	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", level+1).Select("min_exp").Scan(&maxExp)

	if maxExp == 0 {
		maxExp = 999999999
	}

	// Get user's exp
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return 0, 0, err
	}

	// Count users with more exp in the same level
	var rank int64
	r.db.WithContext(ctx).Model(&model.User{}).
		Where("exp >= ? AND exp < ? AND exp > ?", minExp, maxExp, user.Exp).
		Count(&rank)

	// Count total users in level
	var total int64
	r.db.WithContext(ctx).Model(&model.User{}).
		Where("exp >= ? AND exp < ?", minExp, maxExp).
		Count(&total)

	return int(rank) + 1, int(total), nil
}

// GetWeeklyProgress returns user's progress for the current week
func (r *CommunityProgressRepository) GetWeeklyProgress(ctx context.Context, userID uint) (int64, int, error) {
	weekStart := getStartOfWeek(time.Now())

	var expEarned int64
	r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, weekStart).
		Select("COALESCE(SUM(exp_earned), 0)").
		Scan(&expEarned)

	var activitiesCount int64
	r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, weekStart).
		Count(&activitiesCount)

	return expEarned, int(activitiesCount), nil
}

// GetMonthlyProgress returns user's progress for the current month
func (r *CommunityProgressRepository) GetMonthlyProgress(ctx context.Context, userID uint) (int64, int, error) {
	monthStart := getStartOfMonth(time.Now())

	var expEarned int64
	r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, monthStart).
		Select("COALESCE(SUM(exp_earned), 0)").
		Scan(&expEarned)

	var activitiesCount int64
	r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, monthStart).
		Count(&activitiesCount)

	return expEarned, int(activitiesCount), nil
}

// GetUserActivityTypes returns breakdown of user's activity types
func (r *CommunityProgressRepository) GetUserActivityTypes(ctx context.Context, userID uint, since time.Time) (map[string]int64, error) {
	type ActivityCount struct {
		ActivityType string
		Count        int64
	}

	var results []ActivityCount
	err := r.db.WithContext(ctx).Model(&model.ExpHistory{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Select("activity_type, COUNT(*) as count").
		Group("activity_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	activityMap := make(map[string]int64)
	for _, r := range results {
		activityMap[r.ActivityType] = r.Count
	}

	return activityMap, nil
}

// Helper functions
func getStartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-weekday+1, 0, 0, 0, 0, t.Location())
}

func getStartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
