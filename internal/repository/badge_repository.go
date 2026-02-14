package repository

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BadgeRepository struct {
	db *gorm.DB
}

func NewBadgeRepository(db *gorm.DB) *BadgeRepository {
	return &BadgeRepository{db: db}
}

// ==========================================
// Badge Definitions
// ==========================================

// GetAllBadgeDefinitions retrieves all badge definitions
func (r *BadgeRepository) GetAllBadgeDefinitions() ([]model.BadgeDefinition, error) {
	var badges []model.BadgeDefinition
	err := r.db.Where("is_active = ?", true).
		Order("category ASC, badge_key ASC").
		Find(&badges).Error
	return badges, err
}

// GetBadgesByCategory retrieves badges by category
func (r *BadgeRepository) GetBadgesByCategory(category string) ([]model.BadgeDefinition, error) {
	var badges []model.BadgeDefinition
	err := r.db.Where("category = ? AND is_active = ?", category, true).
		Order("badge_key ASC").
		Find(&badges).Error
	return badges, err
}

// GetBadgeByKey retrieves a badge definition by its key
func (r *BadgeRepository) GetBadgeByKey(key string) (*model.BadgeDefinition, error) {
	var badge model.BadgeDefinition
	err := r.db.Where("badge_key = ?", key).First(&badge).Error
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

// GetBadgeByID retrieves a badge definition by ID
func (r *BadgeRepository) GetBadgeByID(id uuid.UUID) (*model.BadgeDefinition, error) {
	var badge model.BadgeDefinition
	err := r.db.First(&badge, id).Error
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

// GetBadgesByRequirementType gets badges by their requirement type
func (r *BadgeRepository) GetBadgesByRequirementType(reqType string) ([]model.BadgeDefinition, error) {
	var badges []model.BadgeDefinition
	err := r.db.Where("requirement_type = ? AND is_active = ?", reqType, true).
		Find(&badges).Error
	return badges, err
}

// CreateBadgeDefinition creates a new badge definition
func (r *BadgeRepository) CreateBadgeDefinition(badge *model.BadgeDefinition) error {
	return r.db.Create(badge).Error
}

// UpdateBadgeDefinition updates a badge definition
func (r *BadgeRepository) UpdateBadgeDefinition(badge *model.BadgeDefinition) error {
	return r.db.Save(badge).Error
}

// ==========================================
// User Badges
// ==========================================

// GetUserBadges retrieves all badges earned by a user
func (r *BadgeRepository) GetUserBadges(userID uint) ([]model.UserBadge, error) {
	var badges []model.UserBadge
	err := r.db.Where("user_id = ?", userID).
		Preload("Badge").
		Order("earned_at DESC").
		Find(&badges).Error
	return badges, err
}

// GetUserBadgesByCategory retrieves user badges by category
func (r *BadgeRepository) GetUserBadgesByCategory(userID uint, category string) ([]model.UserBadge, error) {
	var badges []model.UserBadge
	err := r.db.Where("user_id = ?", userID).
		Preload("Badge", "category = ?", category).
		Find(&badges).Error
	return badges, err
}

// HasBadge checks if a user has earned a specific badge
func (r *BadgeRepository) HasBadge(userID uint, badgeID uuid.UUID) bool {
	var count int64
	r.db.Model(&model.UserBadge{}).
		Where("user_id = ? AND badge_id = ?", userID, badgeID).
		Count(&count)
	return count > 0
}

// HasBadgeByKey checks if a user has earned a badge by its key
func (r *BadgeRepository) HasBadgeByKey(userID uint, badgeKey string) bool {
	var count int64
	r.db.Model(&model.UserBadge{}).
		Joins("JOIN badge_definitions ON badge_definitions.id = user_badges.badge_id").
		Where("user_badges.user_id = ? AND badge_definitions.badge_key = ?", userID, badgeKey).
		Count(&count)
	return count > 0
}

// AwardBadge awards a badge to a user
func (r *BadgeRepository) AwardBadge(userID uint, badgeID uuid.UUID) error {
	userBadge := &model.UserBadge{
		UserID:   userID,
		BadgeID:  badgeID,
		EarnedAt: time.Now(),
	}
	return r.db.Create(userBadge).Error
}

// AwardBadgeByKey awards a badge to a user by badge key
func (r *BadgeRepository) AwardBadgeByKey(userID uint, badgeKey string) error {
	badge, err := r.GetBadgeByKey(badgeKey)
	if err != nil {
		return err
	}

	if r.HasBadge(userID, badge.ID) {
		return nil // Already has badge
	}

	return r.AwardBadge(userID, badge.ID)
}

// GetUserBadgeCount returns the number of badges a user has earned
func (r *BadgeRepository) GetUserBadgeCount(userID uint) (int, error) {
	var count int64
	err := r.db.Model(&model.UserBadge{}).Where("user_id = ?", userID).Count(&count).Error
	return int(count), err
}

// GetBadgeProgress returns progress towards badges for a user
func (r *BadgeRepository) GetBadgeProgress(userID uint) ([]BadgeProgressInfo, error) {
	// Get all active badges
	badges, err := r.GetAllBadgeDefinitions()
	if err != nil {
		return nil, err
	}

	// Get user's earned badges
	earnedMap := make(map[uuid.UUID]bool)
	userBadges, _ := r.GetUserBadges(userID)
	for _, ub := range userBadges {
		earnedMap[ub.BadgeID] = true
	}

	var progress []BadgeProgressInfo
	for _, badge := range badges {
		info := BadgeProgressInfo{
			Badge:  badge,
			Earned: earnedMap[badge.ID],
		}

		if !info.Earned {
			// Calculate progress based on requirement type
			info.CurrentValue, info.TargetValue = r.calculateProgress(userID, badge)
		}

		progress = append(progress, info)
	}

	return progress, nil
}

// BadgeProgressInfo holds badge progress information
type BadgeProgressInfo struct {
	Badge        model.BadgeDefinition
	Earned       bool
	CurrentValue int
	TargetValue  int
}

// calculateProgress calculates progress towards a badge
func (r *BadgeRepository) calculateProgress(userID uint, badge model.BadgeDefinition) (int, int) {
	switch badge.RequirementType {
	case model.BadgeRequirementStreak:
		var user model.User
		r.db.First(&user, userID)
		return user.CurrentStreak, badge.RequirementValue

	case model.BadgeRequirementActivityCount:
		var count int64
		r.db.Model(&model.ExpHistory{}).
			Where("user_id = ?", userID).
			Count(&count)
		return int(count), badge.RequirementValue

	case model.BadgeRequirementLevel:
		var user model.User
		r.db.First(&user, userID)
		var config model.LevelConfig
		r.db.Where("min_exp <= ?", user.Exp).Order("min_exp DESC").First(&config)
		return config.Level, badge.RequirementValue

	case model.BadgeRequirementStory:
		var count int64
		r.db.Model(&model.InspiringStory{}).
			Where("author_id = ? AND status = 'approved'", userID).
			Count(&count)
		return int(count), badge.RequirementValue

	case model.BadgeRequirementXP:
		var user model.User
		r.db.First(&user, userID)
		return int(user.Exp), badge.RequirementValue

	default:
		return 0, badge.RequirementValue
	}
}

// GetRecentlyEarnedBadges gets badges earned within a time period
func (r *BadgeRepository) GetRecentlyEarnedBadges(userID uint, since time.Time) ([]model.UserBadge, error) {
	var badges []model.UserBadge
	err := r.db.Where("user_id = ? AND earned_at >= ?", userID, since).
		Preload("Badge").
		Order("earned_at DESC").
		Find(&badges).Error
	return badges, err
}

// GetUsersWithBadge gets all users who have earned a specific badge
func (r *BadgeRepository) GetUsersWithBadge(badgeID uuid.UUID, limit int) ([]model.UserBadge, error) {
	var userBadges []model.UserBadge
	err := r.db.Where("badge_id = ?", badgeID).
		Preload("User").
		Order("earned_at DESC").
		Limit(limit).
		Find(&userBadges).Error
	return userBadges, err
}

// GetBadgeCategoryStats returns stats about badges by category
func (r *BadgeRepository) GetBadgeCategoryStats(userID uint) (map[string]CategoryBadgeStats, error) {
	stats := make(map[string]CategoryBadgeStats)

	categories := []string{"streak", "activity", "contribution", "special", "level"}

	for _, cat := range categories {
		var total int64
		r.db.Model(&model.BadgeDefinition{}).
			Where("category = ? AND is_active = ?", cat, true).
			Count(&total)

		var earned int64
		r.db.Model(&model.UserBadge{}).
			Joins("JOIN badge_definitions ON badge_definitions.id = user_badges.badge_id").
			Where("user_badges.user_id = ? AND badge_definitions.category = ?", userID, cat).
			Count(&earned)

		stats[cat] = CategoryBadgeStats{
			Total:  int(total),
			Earned: int(earned),
		}
	}

	return stats, nil
}

// CategoryBadgeStats holds category-level badge stats
type CategoryBadgeStats struct {
	Total  int
	Earned int
}

// GetDisplayBadges gets badges to display on user profile (limited)
func (r *BadgeRepository) GetDisplayBadges(userID uint, limit int) ([]model.UserBadge, error) {
	var badges []model.UserBadge
	err := r.db.Where("user_id = ?", userID).
		Preload("Badge").
		Order("earned_at DESC").
		Limit(limit).
		Find(&badges).Error
	return badges, err
}
