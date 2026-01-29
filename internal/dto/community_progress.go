package dto

import (
	"time"

	"github.com/google/uuid"
)

// ==========================================
// Community Progress Board DTOs
// ==========================================

// CommunityStatsResponse represents community-wide statistics
type CommunityStatsResponse struct {
	TotalXPEarned          int64 `json:"total_xp_earned"`
	ActiveMembers          int   `json:"active_members"`
	TotalAchievements      int   `json:"total_achievements"`
	GrowthPercentage       int   `json:"growth_percentage"`
	NewMembers             int   `json:"new_members"`
	TotalStoriesPublished  int   `json:"total_stories_published"`
	TotalArticlesPublished int   `json:"total_articles_published"`
	Month                  int   `json:"month"`
	Year                   int   `json:"year"`
}

// PersonalJourneyResponse represents a user's personal progress
type PersonalJourneyResponse struct {
	UserID            uint      `json:"user_id"`
	CurrentLevel      int       `json:"current_level"`
	CurrentExp        int64     `json:"current_exp"`
	ExpToNextLevel    int       `json:"exp_to_next_level"`
	ProgressPercent   float64   `json:"progress_percent"`
	TierName          string    `json:"tier_name"`
	TierColor         string    `json:"tier_color"`
	BadgeName         string    `json:"badge_name"`
	BadgeIcon         string    `json:"badge_icon"`
	MonthlyXP         int64     `json:"monthly_xp"`
	MonthlyActivities int       `json:"monthly_activities"`
	NewBadgesCount    int       `json:"new_badges_count"`
	RankInLevel       int       `json:"rank_in_level"`
	TotalInLevel      int       `json:"total_in_level"`
	Percentile        float64   `json:"percentile"`
	CurrentStreak     int       `json:"current_streak"`
	LongestStreak     int       `json:"longest_streak"`
	TotalActivities   int       `json:"total_activities"`
	MemberSince       time.Time `json:"member_since"`
}

// HallOfFameEntry represents a featured user in the hall of fame
type HallOfFameEntry struct {
	Rank      int    `json:"rank"`
	UserID    uint   `json:"user_id"`
	UserName  string `json:"user_name"`
	Avatar    string `json:"avatar"`
	MonthlyXP int    `json:"monthly_xp"`
	Message   string `json:"message,omitempty"`
	TierName  string `json:"tier_name"`
	TierColor string `json:"tier_color"`
}

// LevelHallOfFameResponse represents the hall of fame for a specific level
type LevelHallOfFameResponse struct {
	Level         int               `json:"level"`
	Month         int               `json:"month"`
	Year          int               `json:"year"`
	FeaturedUsers []HallOfFameEntry `json:"featured_users"`
	TotalMembers  int               `json:"total_members"`
}

// UpdateHallOfFameMessageRequest for updating inspiring message
type UpdateHallOfFameMessageRequest struct {
	Message string `json:"message" binding:"max=300"`
}

// LevelMembersResponse for listing all members at a level
type LevelMembersResponse struct {
	Level   int                  `json:"level"`
	Total   int                  `json:"total"`
	Members []LevelMemberSummary `json:"members"`
}

// LevelMemberSummary for member list without ranking
type LevelMemberSummary struct {
	UserID    uint   `json:"user_id"`
	UserName  string `json:"user_name"`
	Avatar    string `json:"avatar"`
	TotalExp  int64  `json:"total_exp"`
	TierName  string `json:"tier_name"`
	TierColor string `json:"tier_color"`
}

// ==========================================
// Virtual Rewards DTOs
// ==========================================

// FeatureUnlockResponse represents an unlockable feature
type FeatureUnlockResponse struct {
	ID            uuid.UUID  `json:"id"`
	FeatureKey    string     `json:"feature_key"`
	FeatureName   string     `json:"feature_name"`
	Description   string     `json:"description"`
	Icon          string     `json:"icon"`
	RequiredLevel int        `json:"required_level"`
	Category      string     `json:"category"`
	IsUnlocked    bool       `json:"is_unlocked"`
	UnlockedAt    *time.Time `json:"unlocked_at,omitempty"`
}

// FeatureUnlocksListResponse represents all features grouped by level
type FeatureUnlocksListResponse struct {
	UserLevel       int                             `json:"user_level"`
	UnlockedCount   int                             `json:"unlocked_count"`
	TotalCount      int                             `json:"total_count"`
	Features        []FeatureUnlockResponse         `json:"features"`
	FeaturesByLevel map[int][]FeatureUnlockResponse `json:"features_by_level"`
}

// CheckFeatureAccessResponse for checking if user has access to a feature
type CheckFeatureAccessResponse struct {
	FeatureKey    string `json:"feature_key"`
	HasAccess     bool   `json:"has_access"`
	RequiredLevel int    `json:"required_level"`
	UserLevel     int    `json:"user_level"`
	Message       string `json:"message,omitempty"`
}

// ==========================================
// Badge DTOs
// ==========================================

// BadgeResponse represents a badge
type BadgeResponse struct {
	ID               uuid.UUID  `json:"id"`
	BadgeKey         string     `json:"badge_key"`
	BadgeName        string     `json:"badge_name"`
	Description      string     `json:"description"`
	Icon             string     `json:"icon"`
	Category         string     `json:"category"`
	RequirementType  string     `json:"requirement_type"`
	RequirementValue int        `json:"requirement_value"`
	IsEarned         bool       `json:"is_earned"`
	IsShowcased      bool       `json:"is_showcased"`
	EarnedAt         *time.Time `json:"earned_at,omitempty"`
}

// UserBadgesResponse represents a user's badges
type UserBadgesResponse struct {
	TotalBadges      int                        `json:"total_badges"`
	EarnedBadges     int                        `json:"earned_badges"`
	ShowcasedBadges  []BadgeResponse            `json:"showcased_badges"`
	AllBadges        []BadgeResponse            `json:"all_badges"`
	BadgesByCategory map[string][]BadgeResponse `json:"badges_by_category"`
}

// UpdateShowcasedBadgesRequest for updating showcased badges
type UpdateShowcasedBadgesRequest struct {
	BadgeIDs []uuid.UUID `json:"badge_ids" binding:"required,max=5"`
}

// ==========================================
// Personal Progress DTOs
// ==========================================

// WeeklyProgressResponse compares this week to last week
type WeeklyProgressResponse struct {
	ThisWeek          ProgressStats `json:"this_week"`
	LastWeek          ProgressStats `json:"last_week"`
	XPChange          int64         `json:"xp_change"`
	XPChangePercent   float64       `json:"xp_change_percent"`
	StreakDays        int           `json:"streak_days"`
	MostProductiveDay string        `json:"most_productive_day"`
}

// ProgressStats represents stats for a period
type ProgressStats struct {
	XPEarned        int64 `json:"xp_earned"`
	ActivitiesCount int   `json:"activities_count"`
	BadgesEarned    int   `json:"badges_earned"`
}

// MonthlyProgressResponse represents monthly overview
type MonthlyProgressResponse struct {
	TotalXP          int64   `json:"total_xp"`
	LevelProgress    float64 `json:"level_progress_percent"`
	BadgesEarned     int     `json:"badges_earned"`
	TopActivity      string  `json:"top_activity"`
	TopActivityCount int     `json:"top_activity_count"`
	DaysActive       int     `json:"days_active"`
}

// AllTimeStatsResponse represents all-time user statistics
type AllTimeStatsResponse struct {
	MemberSince     time.Time `json:"member_since"`
	TotalXP         int64     `json:"total_xp"`
	CurrentLevel    int       `json:"current_level"`
	TotalBadges     int       `json:"total_badges"`
	LongestStreak   int       `json:"longest_streak"`
	TotalActivities int       `json:"total_activities"`
	StoriesShared   int       `json:"stories_shared"`
	ArticlesWritten int       `json:"articles_written"`
}

// XPHistoryChartItem represents a single XP entry for charts
type XPHistoryChartItem struct {
	Date string `json:"date"`
	XP   int64  `json:"xp"`
}

// ActivityBreakdownItem represents activity type breakdown
type ActivityBreakdownItem struct {
	ActivityType string  `json:"activity_type"`
	Count        int     `json:"count"`
	TotalXP      int64   `json:"total_xp"`
	Percentage   float64 `json:"percentage"`
}

// PersonalInsightsResponse represents AI-generated insights
type PersonalInsightsResponse struct {
	MostActiveDay    string   `json:"most_active_day"`
	FavoriteActivity string   `json:"favorite_activity"`
	GrowthPercent    float64  `json:"growth_percent"`
	DaysToNextLevel  int      `json:"days_to_next_level"`
	Insights         []string `json:"insights"`
}

// ==========================================
// Level Up Celebration DTO
// ==========================================

// LevelUpCelebrationResponse for level up events
type LevelUpCelebrationResponse struct {
	NewLevel         int                     `json:"new_level"`
	BadgeName        string                  `json:"badge_name"`
	BadgeIcon        string                  `json:"badge_icon"`
	TierName         string                  `json:"tier_name"`
	TierColor        string                  `json:"tier_color"`
	UnlockedFeatures []FeatureUnlockResponse `json:"unlocked_features"`
	NewBadges        []BadgeResponse         `json:"new_badges,omitempty"`
	CongratMessage   string                  `json:"congrats_message"`
}

// ==========================================
// Profile Customization DTOs
// ==========================================

// UpdateProfileCustomizationRequest for updating profile appearance
type UpdateProfileCustomizationRequest struct {
	ProfileTheme      string `json:"profile_theme" binding:"omitempty,max=50"`
	ProfileBanner     string `json:"profile_banner" binding:"omitempty,max=500"`
	AvatarBorderColor string `json:"avatar_border_color" binding:"omitempty,max=20"`
	Tagline           string `json:"tagline" binding:"omitempty,max=200"`
	Bio               string `json:"bio" binding:"omitempty,max=1000"`
}

// UserProfileResponse extended profile with gamification
type UserProfileResponse struct {
	ID                uint            `json:"id"`
	Name              string          `json:"name"`
	Email             string          `json:"email,omitempty"`
	Avatar            string          `json:"avatar"`
	Bio               string          `json:"bio"`
	Tagline           string          `json:"tagline"`
	ProfileTheme      string          `json:"profile_theme"`
	ProfileBanner     string          `json:"profile_banner"`
	AvatarBorderColor string          `json:"avatar_border_color"`
	Level             int             `json:"level"`
	Exp               int64           `json:"exp"`
	TierName          string          `json:"tier_name"`
	TierColor         string          `json:"tier_color"`
	BadgeName         string          `json:"badge_name"`
	BadgeIcon         string          `json:"badge_icon"`
	CurrentStreak     int             `json:"current_streak"`
	TotalBadges       int             `json:"total_badges"`
	ShowcasedBadges   []BadgeResponse `json:"showcased_badges"`
	MemberSince       time.Time       `json:"member_since"`
	Role              string          `json:"role"`
}
