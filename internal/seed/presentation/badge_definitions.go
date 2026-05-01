package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedBadgeDefinitions seeds the badge definitions
func SeedBadgeDefinitions(db *gorm.DB) error {
	badges := []model.BadgeDefinition{
		// Streak badges
		{BadgeKey: "streak_7", BadgeName: "7-Day Streak", Description: "Active for 7 consecutive days", Icon: "🔥", Category: "engagement", RequirementType: model.BadgeRequirementStreak, RequirementValue: 7, DisplayOrder: 1},
		{BadgeKey: "streak_14", BadgeName: "14-Day Streak", Description: "Active for 14 consecutive days", Icon: "🔥", Category: "engagement", RequirementType: model.BadgeRequirementStreak, RequirementValue: 14, DisplayOrder: 2},
		{BadgeKey: "streak_30", BadgeName: "30-Day Streak", Description: "Active for 30 consecutive days", Icon: "🔥", Category: "engagement", RequirementType: model.BadgeRequirementStreak, RequirementValue: 30, DisplayOrder: 3},
		{BadgeKey: "streak_60", BadgeName: "60-Day Streak", Description: "Active for 60 consecutive days", Icon: "🔥", Category: "engagement", RequirementType: model.BadgeRequirementStreak, RequirementValue: 60, DisplayOrder: 4},
		{BadgeKey: "streak_100", BadgeName: "100-Day Streak", Description: "Active for 100 consecutive days", Icon: "🔥", Category: "engagement", RequirementType: model.BadgeRequirementStreak, RequirementValue: 100, DisplayOrder: 5},

		// Activity count badges
		{BadgeKey: "activities_10", BadgeName: "Getting Started", Description: "Completed 10 activities", Icon: "🌟", Category: "engagement", RequirementType: model.BadgeRequirementActivityCount, RequirementValue: 10, DisplayOrder: 10},
		{BadgeKey: "activities_50", BadgeName: "Regular User", Description: "Completed 50 activities", Icon: "⭐", Category: "engagement", RequirementType: model.BadgeRequirementActivityCount, RequirementValue: 50, DisplayOrder: 11},
		{BadgeKey: "activities_100", BadgeName: "Active Member", Description: "Completed 100 activities", Icon: "💫", Category: "engagement", RequirementType: model.BadgeRequirementActivityCount, RequirementValue: 100, DisplayOrder: 12},
		{BadgeKey: "activities_500", BadgeName: "Power User", Description: "Completed 500 activities", Icon: "✨", Category: "engagement", RequirementType: model.BadgeRequirementActivityCount, RequirementValue: 500, DisplayOrder: 13},

		// Contribution badges
		{BadgeKey: "first_article", BadgeName: "First Article", Description: "Published your first article", Icon: "📝", Category: "contribution", RequirementType: model.BadgeRequirementManual, RequirementValue: 1, DisplayOrder: 20},
		{BadgeKey: "articles_5", BadgeName: "Article Writer", Description: "Published 5 articles", Icon: "✍️", Category: "contribution", RequirementType: model.BadgeRequirementActivityCount, RequirementValue: 5, DisplayOrder: 21},
		{BadgeKey: "helpful_commenter", BadgeName: "Helpful Commenter", Description: "Received 10 hearts on comments", Icon: "💬", Category: "contribution", RequirementType: model.BadgeRequirementManual, RequirementValue: 10, DisplayOrder: 22},
		{BadgeKey: "top_contributor", BadgeName: "Top Contributor", Description: "Made significant contributions to community", Icon: "🏆", Category: "contribution", RequirementType: model.BadgeRequirementManual, RequirementValue: 0, DisplayOrder: 23},

		// Level badges
		{BadgeKey: "level_5", BadgeName: "Level 5", Description: "Reached Level 5", Icon: "🌱", Category: "milestone", RequirementType: model.BadgeRequirementLevel, RequirementValue: 5, DisplayOrder: 30},
		{BadgeKey: "level_10", BadgeName: "Level 10", Description: "Reached Level 10", Icon: "🌳", Category: "milestone", RequirementType: model.BadgeRequirementLevel, RequirementValue: 10, DisplayOrder: 31},

		// XP badges
		{BadgeKey: "xp_1000", BadgeName: "1K XP", Description: "Earned 1,000 XP", Icon: "💎", Category: "milestone", RequirementType: model.BadgeRequirementXP, RequirementValue: 1000, DisplayOrder: 40},
		{BadgeKey: "xp_5000", BadgeName: "5K XP", Description: "Earned 5,000 XP", Icon: "💎", Category: "milestone", RequirementType: model.BadgeRequirementXP, RequirementValue: 5000, DisplayOrder: 41},
		{BadgeKey: "xp_10000", BadgeName: "10K XP", Description: "Earned 10,000 XP", Icon: "💎", Category: "milestone", RequirementType: model.BadgeRequirementXP, RequirementValue: 10000, DisplayOrder: 42},

		// Special badges
		{BadgeKey: "beta_tester", BadgeName: "Beta Tester", Description: "Helped test beta features", Icon: "🧪", Category: "special", RequirementType: model.BadgeRequirementManual, RequirementValue: 0, DisplayOrder: 50},
		{BadgeKey: "community_mentor", BadgeName: "Community Mentor", Description: "Recognized as a community mentor", Icon: "🎓", Category: "special", RequirementType: model.BadgeRequirementLevel, RequirementValue: 9, DisplayOrder: 51},
		{BadgeKey: "guardian", BadgeName: "Guardian", Description: "Reached Guardian status", Icon: "👑", Category: "special", RequirementType: model.BadgeRequirementLevel, RequirementValue: 10, DisplayOrder: 52},

		// Story badges
		{BadgeKey: "first_story", BadgeName: "Storyteller", Description: "Published your first inspiring story", Icon: "📖", Category: "story", RequirementType: model.BadgeRequirementStory, RequirementValue: 1, DisplayOrder: 60},
		{BadgeKey: "stories_3", BadgeName: "Inspiring Storyteller", Description: "Published 3 inspiring stories", Icon: "🌟", Category: "story", RequirementType: model.BadgeRequirementStory, RequirementValue: 3, DisplayOrder: 61},
		{BadgeKey: "story_100_hearts", BadgeName: "Beloved Story", Description: "A story received 100 hearts", Icon: "💚", Category: "story", RequirementType: model.BadgeRequirementManual, RequirementValue: 100, DisplayOrder: 62},
	}

	for _, badge := range badges {
		var existing model.BadgeDefinition
		if db.Where("badge_key = ?", badge.BadgeKey).First(&existing).RowsAffected == 0 {
			if err := db.Create(&badge).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
