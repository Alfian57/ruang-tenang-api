package application

import (
	"github.com/Alfian57/ruang-tenang-api/internal/shared/serviceerror"
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/badge/infrastructure")

type BadgeService struct {
	badgeRepo       *infrastructure.BadgeRepository
	userRepo        *authinfra.UserRepository
	levelConfigRepo LevelConfigRepository
}

func NewBadgeService(
	badgeRepo *infrastructure.BadgeRepository,
	userRepo *authinfra.UserRepository,
	levelConfigRepo LevelConfigRepository,
) *BadgeService {
	return &BadgeService{
		badgeRepo:       badgeRepo,
		userRepo:        userRepo,
		levelConfigRepo: levelConfigRepo,
	}
}

// ==========================================
// Badge Definitions
// ==========================================

// GetAllBadges returns all badge definitions
func (s *BadgeService) GetAllBadges(ctx context.Context) ([]dto.BadgeDefinitionResponse, error) {
	badges, err := s.badgeRepo.GetAllBadgeDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BadgeDefinitionResponse, len(badges))
	for i, b := range badges {
		result[i] = s.toBadgeDefinitionResponse(ctx, b)
	}

	return result, nil
}

// GetBadgesByCategory returns badges by category
func (s *BadgeService) GetBadgesByCategory(ctx context.Context, category string) ([]dto.BadgeDefinitionResponse, error) {
	badges, err := s.badgeRepo.GetBadgesByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BadgeDefinitionResponse, len(badges))
	for i, b := range badges {
		result[i] = s.toBadgeDefinitionResponse(ctx, b)
	}

	return result, nil
}

// GetBadgeCategories returns all badge categories
func (s *BadgeService) GetBadgeCategories(ctx context.Context) []dto.BadgeCategoryInfo {
	return []dto.BadgeCategoryInfo{
		{
			Key:         "streak",
			Name:        "Streak",
			Description: "Badge untuk konsistensi aktivitas harian",
			Icon:        "flame",
		},
		{
			Key:         "activity",
			Name:        "Aktivitas",
			Description: "Badge untuk berbagai aktivitas di platform",
			Icon:        "activity",
		},
		{
			Key:         "contribution",
			Name:        "Kontribusi",
			Description: "Badge untuk kontribusi pada komunitas",
			Icon:        "heart",
		},
		{
			Key:         "special",
			Name:        "Spesial",
			Description: "Badge eksklusif untuk pencapaian khusus",
			Icon:        "award",
		},
		{
			Key:         "level",
			Name:        "Level",
			Description: "Badge untuk mencapai level tertentu",
			Icon:        "trending-up",
		},
	}
}

// ==========================================
// User Badges
// ==========================================

// GetUserBadges returns all badges earned by a user
func (s *BadgeService) GetUserBadges(ctx context.Context, userID uint) (*dto.UserBadgesResponse, error) {
	userBadges, err := s.badgeRepo.GetUserBadges(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get all badge definitions for progress
	allBadges, _ := s.badgeRepo.GetAllBadgeDefinitions(ctx)

	earnedBadges := make([]dto.BadgeResponse, len(userBadges))
	for i, ub := range userBadges {
		earnedBadges[i] = dto.BadgeResponse{
			ID:               ub.Badge.ID,
			BadgeKey:         ub.Badge.BadgeKey,
			BadgeName:        ub.Badge.BadgeName,
			Description:      ub.Badge.Description,
			Icon:             ub.Badge.Icon,
			Category:         ub.Badge.Category,
			RequirementType:  string(ub.Badge.RequirementType),
			RequirementValue: ub.Badge.RequirementValue,
			IsEarned:         true,
			IsShowcased:      ub.IsShowcased,
			EarnedAt:         &ub.EarnedAt,
		}
	}

	// Build all badges map with earned status
	allBadgeResponses := make([]dto.BadgeResponse, len(allBadges))
	earnedMap := make(map[string]bool)
	for _, ub := range userBadges {
		earnedMap[ub.Badge.BadgeKey] = true
	}

	for i, b := range allBadges {
		allBadgeResponses[i] = dto.BadgeResponse{
			ID:               b.ID,
			BadgeKey:         b.BadgeKey,
			BadgeName:        b.BadgeName,
			Description:      b.Description,
			Icon:             b.Icon,
			Category:         b.Category,
			RequirementType:  string(b.RequirementType),
			RequirementValue: b.RequirementValue,
			IsEarned:         earnedMap[b.BadgeKey],
		}
	}

	// Group badges by category
	badgesByCategory := make(map[string][]dto.BadgeResponse)
	for _, b := range allBadgeResponses {
		badgesByCategory[b.Category] = append(badgesByCategory[b.Category], b)
	}

	// Get showcased badges
	var showcasedBadges []dto.BadgeResponse
	for _, b := range earnedBadges {
		if b.IsShowcased {
			showcasedBadges = append(showcasedBadges, b)
		}
	}

	return &dto.UserBadgesResponse{
		TotalBadges:      len(allBadges),
		EarnedBadges:     len(userBadges),
		ShowcasedBadges:  showcasedBadges,
		AllBadges:        allBadgeResponses,
		BadgesByCategory: badgesByCategory,
	}, nil
}

// GetBadgeProgress returns progress towards all badges for a user
func (s *BadgeService) GetBadgeProgress(ctx context.Context, userID uint) ([]dto.BadgeProgressResponse, error) {
	progress, err := s.badgeRepo.GetBadgeProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BadgeProgressResponse, len(progress))
	for i, p := range progress {
		var progressPercent float64 = 0
		if p.TargetValue > 0 {
			progressPercent = float64(p.CurrentValue) / float64(p.TargetValue) * 100
			if progressPercent > 100 {
				progressPercent = 100
			}
		}

		result[i] = dto.BadgeProgressResponse{
			BadgeID:         p.Badge.ID,
			BadgeKey:        p.Badge.BadgeKey,
			BadgeName:       p.Badge.BadgeName,
			Description:     p.Badge.Description,
			Icon:            p.Badge.Icon,
			Category:        p.Badge.Category,
			Earned:          p.Earned,
			CurrentValue:    p.CurrentValue,
			TargetValue:     p.TargetValue,
			ProgressPercent: progressPercent,
		}
	}

	return result, nil
}

// GetRecentlyEarnedBadges returns badges earned within a time period
func (s *BadgeService) GetRecentlyEarnedBadges(ctx context.Context, userID uint, days int) ([]dto.BadgeResponse, error) {
	since := time.Now().AddDate(0, 0, -days)

	userBadges, err := s.badgeRepo.GetRecentlyEarnedBadges(ctx, userID, since)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BadgeResponse, len(userBadges))
	for i, ub := range userBadges {
		result[i] = dto.BadgeResponse{
			ID:               ub.Badge.ID,
			BadgeKey:         ub.Badge.BadgeKey,
			BadgeName:        ub.Badge.BadgeName,
			Description:      ub.Badge.Description,
			Icon:             ub.Badge.Icon,
			Category:         ub.Badge.Category,
			RequirementType:  string(ub.Badge.RequirementType),
			RequirementValue: ub.Badge.RequirementValue,
			IsEarned:         true,
			IsShowcased:      ub.IsShowcased,
			EarnedAt:         &ub.EarnedAt,
		}
	}

	return result, nil
}

// GetDisplayBadges returns badges for display on profile (limited)
func (s *BadgeService) GetDisplayBadges(ctx context.Context, userID uint, limit int) ([]dto.BadgeResponse, error) {
	userBadges, err := s.badgeRepo.GetDisplayBadges(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.BadgeResponse, len(userBadges))
	for i, ub := range userBadges {
		result[i] = dto.BadgeResponse{
			ID:               ub.Badge.ID,
			BadgeKey:         ub.Badge.BadgeKey,
			BadgeName:        ub.Badge.BadgeName,
			Description:      ub.Badge.Description,
			Icon:             ub.Badge.Icon,
			Category:         ub.Badge.Category,
			RequirementType:  string(ub.Badge.RequirementType),
			RequirementValue: ub.Badge.RequirementValue,
			IsEarned:         true,
			IsShowcased:      ub.IsShowcased,
			EarnedAt:         &ub.EarnedAt,
		}
	}

	return result, nil
}

// ==========================================
// Badge Checking & Awarding
// ==========================================

// CheckAndAwardBadges checks if user qualifies for any new badges
func (s *BadgeService) CheckAndAwardBadges(ctx context.Context, userID uint) ([]dto.BadgeResponse, error) {
	var newBadges []dto.BadgeResponse

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check streak badges
	streakBadges := s.checkStreakBadges(ctx, userID, user.CurrentStreak)
	newBadges = append(newBadges, streakBadges...)

	// Check level badges
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
	if currentLevel != nil {
		levelBadges := s.checkLevelBadges(ctx, userID, currentLevel.Level)
		newBadges = append(newBadges, levelBadges...)
	}

	// Check activity badges
	activityBadges := s.checkActivityBadges(ctx, userID)
	newBadges = append(newBadges, activityBadges...)

	// Check contribution badges
	contributionBadges := s.checkContributionBadges(ctx, userID)
	newBadges = append(newBadges, contributionBadges...)

	return newBadges, nil
}

// checkStreakBadges checks and awards streak-related badges
func (s *BadgeService) checkStreakBadges(ctx context.Context, userID uint, currentStreak int) []dto.BadgeResponse {
	var newBadges []dto.BadgeResponse

	streakBadgeKeys := map[int]string{
		3:   "streak_3_days",
		7:   "streak_7_days",
		14:  "streak_14_days",
		30:  "streak_30_days",
		60:  "streak_60_days",
		90:  "streak_90_days",
		180: "streak_180_days",
		365: "streak_365_days",
	}

	for streakDays, badgeKey := range streakBadgeKeys {
		if currentStreak >= streakDays && !s.badgeRepo.HasBadgeByKey(ctx, userID, badgeKey) {
			badge, err := s.badgeRepo.GetBadgeByKey(ctx, badgeKey)
			if err == nil {
				if err := s.badgeRepo.AwardBadge(ctx, userID, badge.ID); err == nil {
					now := time.Now()
					newBadges = append(newBadges, dto.BadgeResponse{
						ID:               badge.ID,
						BadgeKey:         badge.BadgeKey,
						BadgeName:        badge.BadgeName,
						Description:      badge.Description,
						Icon:             badge.Icon,
						Category:         badge.Category,
						RequirementType:  string(badge.RequirementType),
						RequirementValue: badge.RequirementValue,
						IsEarned:         true,
						EarnedAt:         &now,
					})
				}
			}
		}
	}

	return newBadges
}

// checkLevelBadges checks and awards level-related badges
func (s *BadgeService) checkLevelBadges(ctx context.Context, userID uint, currentLevel int) []dto.BadgeResponse {
	var newBadges []dto.BadgeResponse

	levelBadgeKeys := map[int]string{
		2:  "level_2",
		3:  "level_3",
		5:  "level_5",
		7:  "level_7",
		10: "level_10",
	}

	for level, badgeKey := range levelBadgeKeys {
		if currentLevel >= level && !s.badgeRepo.HasBadgeByKey(ctx, userID, badgeKey) {
			badge, err := s.badgeRepo.GetBadgeByKey(ctx, badgeKey)
			if err == nil {
				if err := s.badgeRepo.AwardBadge(ctx, userID, badge.ID); err == nil {
					now := time.Now()
					newBadges = append(newBadges, dto.BadgeResponse{
						ID:               badge.ID,
						BadgeKey:         badge.BadgeKey,
						BadgeName:        badge.BadgeName,
						Description:      badge.Description,
						Icon:             badge.Icon,
						Category:         badge.Category,
						RequirementType:  string(badge.RequirementType),
						RequirementValue: badge.RequirementValue,
						IsEarned:         true,
						EarnedAt:         &now,
					})
				}
			}
		}
	}

	return newBadges
}

// checkActivityBadges checks and awards activity-related badges
func (s *BadgeService) checkActivityBadges(ctx context.Context, userID uint) []dto.BadgeResponse {
	var newBadges []dto.BadgeResponse

	// First chat badge
	if !s.badgeRepo.HasBadgeByKey(ctx, userID, "first_chat") {
		// Check if user has any chat activity
		badge, err := s.badgeRepo.GetBadgeByKey(ctx, "first_chat")
		if err == nil {
			// Award would be triggered by actual chat activity
			_ = badge
		}
	}

	// First article badge
	if !s.badgeRepo.HasBadgeByKey(ctx, userID, "first_article") {
		badge, err := s.badgeRepo.GetBadgeByKey(ctx, "first_article")
		if err == nil {
			_ = badge
		}
	}

	// First mood log badge
	if !s.badgeRepo.HasBadgeByKey(ctx, userID, "first_mood_log") {
		badge, err := s.badgeRepo.GetBadgeByKey(ctx, "first_mood_log")
		if err == nil {
			_ = badge
		}
	}

	return newBadges
}

// checkContributionBadges checks and awards contribution-related badges
func (s *BadgeService) checkContributionBadges(ctx context.Context, userID uint) []dto.BadgeResponse {
	var newBadges []dto.BadgeResponse

	// Story-related badges would be checked here
	// Comment-related badges would be checked here
	// Heart-related badges would be checked here

	return newBadges
}

// AwardBadge awards a specific badge to a user
func (s *BadgeService) AwardBadge(ctx context.Context, userID uint, badgeKey string) (*dto.BadgeResponse, error) {
	badge, err := s.badgeRepo.GetBadgeByKey(ctx, badgeKey)
	if err != nil {
		return nil, ErrBadgeNotFound
	}

	if s.badgeRepo.HasBadge(ctx, userID, badge.ID) {
		return nil, ErrBadgeAlreadyEarned
	}

	if err := s.badgeRepo.AwardBadge(ctx, userID, badge.ID); err != nil {
		return nil, err
	}

	now := time.Now()
	return &dto.BadgeResponse{
		ID:               badge.ID,
		BadgeKey:         badge.BadgeKey,
		BadgeName:        badge.BadgeName,
		Description:      badge.Description,
		Icon:             badge.Icon,
		Category:         badge.Category,
		RequirementType:  string(badge.RequirementType),
		RequirementValue: badge.RequirementValue,
		IsEarned:         true,
		EarnedAt:         &now,
	}, nil
}

// AwardBadgeByID awards a badge by ID
func (s *BadgeService) AwardBadgeByID(ctx context.Context, userID uint, badgeID uuid.UUID) (*dto.BadgeResponse, error) {
	badge, err := s.badgeRepo.GetBadgeByID(ctx, badgeID)
	if err != nil {
		return nil, ErrBadgeNotFound
	}

	if s.badgeRepo.HasBadge(ctx, userID, badge.ID) {
		return nil, ErrBadgeAlreadyEarned
	}

	if err := s.badgeRepo.AwardBadge(ctx, userID, badge.ID); err != nil {
		return nil, err
	}

	now := time.Now()
	return &dto.BadgeResponse{
		ID:               badge.ID,
		BadgeKey:         badge.BadgeKey,
		BadgeName:        badge.BadgeName,
		Description:      badge.Description,
		Icon:             badge.Icon,
		Category:         badge.Category,
		RequirementType:  string(badge.RequirementType),
		RequirementValue: badge.RequirementValue,
		IsEarned:         true,
		EarnedAt:         &now,
	}, nil
}

// Helper function
func (s *BadgeService) toBadgeDefinitionResponse(ctx context.Context, b model.BadgeDefinition) dto.BadgeDefinitionResponse {
	return dto.BadgeDefinitionResponse{
		ID:               b.ID,
		BadgeKey:         b.BadgeKey,
		BadgeName:        b.BadgeName,
		Description:      b.Description,
		Icon:             b.Icon,
		Category:         b.Category,
		RequirementType:  b.RequirementType,
		RequirementValue: b.RequirementValue,
	}
}

// Custom errors
var (
	ErrBadgeNotFound      = &serviceerror.ServiceError{Code: "BADGE_NOT_FOUND", Message: "Badge tidak ditemukan"}
	ErrBadgeAlreadyEarned = &serviceerror.ServiceError{Code: "BADGE_ALREADY_EARNED", Message: "Badge sudah dimiliki"}
)
