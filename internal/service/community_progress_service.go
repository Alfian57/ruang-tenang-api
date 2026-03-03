package service

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/gorm"
)

type CommunityProgressService struct {
	communityRepo   *repository.CommunityProgressRepository
	levelConfigRepo *repository.LevelConfigRepository
	featureRepo     *repository.FeatureUnlockRepository
	badgeRepo       *repository.BadgeRepository
	userRepo        *repository.UserRepository
}

func NewCommunityProgressService(
	communityRepo *repository.CommunityProgressRepository,
	levelConfigRepo *repository.LevelConfigRepository,
	featureRepo *repository.FeatureUnlockRepository,
	badgeRepo *repository.BadgeRepository,
	userRepo *repository.UserRepository,
) *CommunityProgressService {
	return &CommunityProgressService{
		communityRepo:   communityRepo,
		levelConfigRepo: levelConfigRepo,
		featureRepo:     featureRepo,
		badgeRepo:       badgeRepo,
		userRepo:        userRepo,
	}
}

// ==========================================
// Community Stats
// ==========================================

// GetCommunityStats returns aggregated community statistics
func (s *CommunityProgressService) GetCommunityStats(ctx context.Context) (*dto.CommunityStatsResponse, error) {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	// Try to get cached stats first
	stats, err := s.communityRepo.GetCommunityStats(ctx)
	if err != nil {
		return nil, err
	}

	// If stats are for different month or don't exist, recalculate
	if stats.Month != month || stats.Year != year {
		stats, err = s.communityRepo.RecalculateCommunityStats(ctx)
		if err != nil {
			return nil, err
		}
		// Save the recalculated stats
		s.communityRepo.UpdateCommunityStats(ctx, stats)
	}

	// Calculate growth percentage (compare with last month)
	growthPercent := 0
	// Simple calculation - can be enhanced

	return &dto.CommunityStatsResponse{
		TotalXPEarned:          stats.TotalXPEarned,
		ActiveMembers:          stats.ActiveMembers,
		TotalAchievements:      stats.TotalAchievements,
		GrowthPercentage:       growthPercent,
		NewMembers:             stats.NewMembers,
		TotalStoriesPublished:  stats.TotalStoriesPublished,
		TotalArticlesPublished: stats.TotalArticlesPublished,
		Month:                  stats.Month,
		Year:                   stats.Year,
	}, nil
}

// ==========================================
// Level Hall of Fame
// ==========================================

// GetLevelHallOfFame returns top users within a level tier
func (s *CommunityProgressService) GetLevelHallOfFame(ctx context.Context, level int, limit int) (*dto.LevelHallOfFameResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	// Get level config for tier info
	levelConfig, err := s.levelConfigRepo.GetByLevel(ctx, level)
	if err != nil {
		return nil, err
	}

	// Get top users in this level
	users, err := s.communityRepo.GetTopUsersInLevel(ctx, level, limit)
	if err != nil {
		return nil, err
	}

	entries := make([]dto.HallOfFameEntry, len(users))
	for i, user := range users {
		entries[i] = dto.HallOfFameEntry{
			Rank:      i + 1,
			UserID:    user.ID,
			UserName:  user.Name,
			Avatar:    user.Avatar,
			MonthlyXP: int(user.Exp), // Using total exp as monthly for now
			TierName:  levelConfig.TierName,
			TierColor: levelConfig.TierColor,
		}
	}

	return &dto.LevelHallOfFameResponse{
		Level:         level,
		Month:         int(time.Now().Month()),
		Year:          time.Now().Year(),
		FeaturedUsers: entries,
		TotalMembers:  len(users),
	}, nil
}

// GetMonthlyHallOfFame returns monthly hall of fame by category
func (s *CommunityProgressService) GetMonthlyHallOfFame(ctx context.Context, month, year int, category string) ([]dto.HallOfFameEntry, error) {
	entries, err := s.communityRepo.GetHallOfFame(ctx, month, year, category)
	if err != nil {
		return nil, err
	}

	result := make([]dto.HallOfFameEntry, len(entries))
	for i, e := range entries {
		userName := ""
		avatar := ""
		if e.User != nil {
			userName = e.User.Name
			avatar = e.User.Avatar
		}
		result[i] = dto.HallOfFameEntry{
			Rank:      e.Rank,
			UserID:    e.UserID,
			UserName:  userName,
			Avatar:    avatar,
			MonthlyXP: e.MonthlyXP,
			Message:   e.Message,
		}
	}

	return result, nil
}

// GetAvailableHallOfFameCategories returns categories for hall of fame
func (s *CommunityProgressService) GetAvailableHallOfFameCategories(ctx context.Context) []string {
	return []string{
		"most_supportive",
		"most_consistent",
		"most_helpful_comments",
		"rising_star",
		"community_builder",
	}
}

// ==========================================
// Personal Journey
// ==========================================

// GetPersonalJourney returns user's personal progress journey
func (s *CommunityProgressService) GetPersonalJourney(ctx context.Context, userID uint) (*dto.PersonalJourneyResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get current level
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
	if err != nil {
		return nil, err
	}

	// Get next level
	nextLevel, _ := s.levelConfigRepo.GetNextLevel(ctx, currentLevel.Level)

	// Calculate progress to next level
	var expToNext int = 0
	var progressPercent float64 = 100
	if nextLevel != nil {
		expToNext = nextLevel.MinExp - int(user.Exp)
		levelRange := nextLevel.MinExp - currentLevel.MinExp
		userProgress := int(user.Exp) - currentLevel.MinExp
		if levelRange > 0 {
			progressPercent = float64(userProgress) / float64(levelRange) * 100
		}
	}

	// Get user's rank in their level
	rank, totalInLevel, _ := s.communityRepo.GetUserRankInLevel(ctx, userID, currentLevel.Level)

	// Get monthly progress
	monthlyXP, monthlyActivities, _ := s.communityRepo.GetMonthlyProgress(ctx, userID)

	// Get badge count
	badgeCount, _ := s.badgeRepo.GetUserBadgeCount(ctx, userID)

	// Calculate percentile
	var percentile float64 = 0
	if totalInLevel > 0 {
		percentile = float64(totalInLevel-rank) / float64(totalInLevel) * 100
	}

	return &dto.PersonalJourneyResponse{
		UserID:            userID,
		CurrentLevel:      currentLevel.Level,
		CurrentExp:        user.Exp,
		ExpToNextLevel:    expToNext,
		ProgressPercent:   progressPercent,
		TierName:          currentLevel.TierName,
		TierColor:         currentLevel.TierColor,
		BadgeName:         currentLevel.BadgeName,
		BadgeIcon:         currentLevel.BadgeIcon,
		MonthlyXP:         monthlyXP,
		MonthlyActivities: monthlyActivities,
		NewBadgesCount:    badgeCount,
		RankInLevel:       rank,
		TotalInLevel:      totalInLevel,
		Percentile:        percentile,
		CurrentStreak:     user.CurrentStreak,
		LongestStreak:     user.LongestStreak,
		TotalActivities:   user.TotalActivities,
		MemberSince:       user.CreatedAt,
	}, nil
}

// ==========================================
// Weekly & Monthly Progress
// ==========================================

// GetWeeklyProgress returns user's weekly progress breakdown
func (s *CommunityProgressService) GetWeeklyProgress(ctx context.Context, userID uint) (*dto.WeeklyProgressResponse, error) {
	weekStart := getStartOfWeek(time.Now())
	lastWeekStart := weekStart.AddDate(0, 0, -7)

	// This week stats
	thisWeekXP, thisWeekActivities, err := s.communityRepo.GetWeeklyProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Last week stats (simplified - would need separate query)
	lastWeekXP := int64(0)
	lastWeekActivities := 0

	// Calculate change
	xpChange := thisWeekXP - lastWeekXP
	var xpChangePercent float64 = 0
	if lastWeekXP > 0 {
		xpChangePercent = float64(xpChange) / float64(lastWeekXP) * 100
	}

	// Get most productive day (simplified)
	mostProductiveDay := "Monday"

	// Get badge count this week
	badgesThisWeek, err := s.badgeRepo.GetRecentlyEarnedBadges(ctx, userID, weekStart)
	if err != nil {
		return nil, err
	}
	badgesLastWeek, err := s.badgeRepo.GetRecentlyEarnedBadges(ctx, userID, lastWeekStart)
	if err != nil {
		return nil, err
	}

	return &dto.WeeklyProgressResponse{
		ThisWeek: dto.ProgressStats{
			XPEarned:        thisWeekXP,
			ActivitiesCount: thisWeekActivities,
			BadgesEarned:    len(badgesThisWeek),
		},
		LastWeek: dto.ProgressStats{
			XPEarned:        lastWeekXP,
			ActivitiesCount: lastWeekActivities,
			BadgesEarned:    len(badgesLastWeek),
		},
		XPChange:          xpChange,
		XPChangePercent:   xpChangePercent,
		StreakDays:        0, // Would need to calculate
		MostProductiveDay: mostProductiveDay,
	}, nil
}

// GetMonthlyProgress returns user's monthly progress breakdown
func (s *CommunityProgressService) GetMonthlyProgress(ctx context.Context, userID uint) (*dto.MonthlyProgressResponse, error) {
	monthStart := getStartOfMonth(time.Now())

	monthlyXP, activitiesCount, err := s.communityRepo.GetMonthlyProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get activity breakdown for top activity
	activityTypes, err := s.communityRepo.GetUserActivityTypes(ctx, userID, monthStart)
	if err != nil {
		return nil, err
	}

	// Find top activity
	topActivity := ""
	topActivityCount := 0
	for actType, count := range activityTypes {
		if int(count) > topActivityCount {
			topActivity = getActivityLabel(actType)
			topActivityCount = int(count)
		}
	}

	// Get badges earned this month
	badgesCount, err := s.badgeRepo.GetRecentlyEarnedBadges(ctx, userID, monthStart)
	if err != nil {
		return nil, err
	}

	// Calculate level progress
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
	if err != nil {
		return nil, err
	}
	nextLevel, err := s.levelConfigRepo.GetNextLevel(ctx, currentLevel.Level)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var levelProgress float64 = 100
	if nextLevel != nil {
		levelRange := nextLevel.MinExp - currentLevel.MinExp
		userProgress := int(user.Exp) - currentLevel.MinExp
		if levelRange > 0 {
			levelProgress = float64(userProgress) / float64(levelRange) * 100
		}
	}

	return &dto.MonthlyProgressResponse{
		TotalXP:          monthlyXP,
		LevelProgress:    levelProgress,
		BadgesEarned:     len(badgesCount),
		TopActivity:      topActivity,
		TopActivityCount: topActivityCount,
		DaysActive:       activitiesCount, // Simplified, would need distinct day count
	}, nil
}

// ==========================================
// All Time Stats
// ==========================================

// GetAllTimeStats returns user's all-time statistics
func (s *CommunityProgressService) GetAllTimeStats(ctx context.Context, userID uint) (*dto.AllTimeStatsResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get current level
	currentLevel, _ := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)

	// Get badges
	badgeCount, _ := s.badgeRepo.GetUserBadgeCount(ctx, userID)

	// Get stories shared
	storiesCount := 0 // Would need repository method

	// Get articles written
	articlesCount := 0 // Would need repository method

	return &dto.AllTimeStatsResponse{
		MemberSince:     user.CreatedAt,
		TotalXP:         user.Exp,
		CurrentLevel:    currentLevel.Level,
		TotalBadges:     badgeCount,
		LongestStreak:   user.LongestStreak,
		TotalActivities: user.TotalActivities,
		StoriesShared:   storiesCount,
		ArticlesWritten: articlesCount,
	}, nil
}

// ==========================================
// Level Up Celebration
// ==========================================

// GetLevelUpCelebration returns celebration data for a level up
func (s *CommunityProgressService) GetLevelUpCelebration(ctx context.Context, userID uint, newLevel int) (*dto.LevelUpCelebrationResponse, error) {
	// Get level config
	levelConfig, err := s.levelConfigRepo.GetByLevel(ctx, newLevel)
	if err != nil {
		return nil, err
	}

	// Get newly unlocked features
	newFeatures, _ := s.featureRepo.GetFeaturesByLevel(ctx, newLevel)

	featureResponses := make([]dto.FeatureUnlockResponse, len(newFeatures))
	for i, f := range newFeatures {
		featureResponses[i] = dto.FeatureUnlockResponse{
			ID:            f.ID,
			FeatureKey:    f.FeatureKey,
			FeatureName:   f.FeatureName,
			Description:   f.Description,
			Icon:          f.Icon,
			RequiredLevel: f.RequiredLevel,
			Category:      f.Category,
			IsUnlocked:    true,
		}
	}

	// Generate celebration message
	message := generateLevelUpMessage(newLevel, levelConfig.TierName)

	return &dto.LevelUpCelebrationResponse{
		NewLevel:         newLevel,
		BadgeName:        levelConfig.BadgeName,
		BadgeIcon:        levelConfig.BadgeIcon,
		TierName:         levelConfig.TierName,
		TierColor:        levelConfig.TierColor,
		UnlockedFeatures: featureResponses,
		CongratMessage:   message,
	}, nil
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

func getActivityLabel(activityType string) string {
	labels := map[string]string{
		"chat_ai":        "Chat dengan AI",
		"article_read":   "Membaca Artikel",
		"article_upload": "Mengunggah Artikel",
		"forum_post":     "Posting di Forum",
		"forum_comment":  "Komentar Forum",
		"mood_log":       "Log Mood",
		"music_listen":   "Mendengarkan Musik",
		"story_share":    "Berbagi Cerita",
		"story_heart":    "Memberi Heart",
		"story_comment":  "Komentar Cerita",
	}

	if label, ok := labels[activityType]; ok {
		return label
	}
	return activityType
}

func generateLevelUpMessage(level int, tierName string) string {
	messages := map[int]string{
		1:  "Selamat datang di perjalananmu! Langkah pertama adalah yang terpaling berani. 🌱",
		2:  "Kamu sudah membuat kemajuan! Terus melangkah, kamu tidak sendirian. 🌿",
		3:  "Luar biasa! Konsistensimu menginspirasi. Terus bertumbuh! 🌳",
		4:  "Kamu semakin kuat setiap harinya. Bangga padamu! 💪",
		5:  "Setengah jalan! Kamu sudah membuktikan ketekunanmu. ⭐",
		6:  "Prestasimu luar biasa! Terus berbagi dan mendukung sesama. 🌟",
		7:  "Kamu adalah teladan bagi komunitas. Terima kasih sudah ada di sini! 💫",
		8:  "Senior yang bijak! Pengalamanmu berharga bagi yang lain. 🏆",
		9:  "Hampir di puncak! Perjalananmu menginspirasi banyak orang. 👑",
		10: "Legendaris! Kamu telah mencapai puncak. Terima kasih telah menjadi bagian dari komunitas ini. 🌈",
	}

	if msg, ok := messages[level]; ok {
		return msg
	}

	return "Selamat naik level! Terus semangat dalam perjalananmu. 🎉"
}
