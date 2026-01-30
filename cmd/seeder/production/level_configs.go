package production

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

// SeedLevelConfigs seeds the level configuration data
func SeedLevelConfigs(db *gorm.DB) error {
	configs := []models.LevelConfig{
		{Level: 1, MinExp: 0, BadgeName: "Newcomer", BadgeIcon: "🌱", TierName: "Newcomer", TierColor: "gray", Description: "Welcome to Ruang Tenang! Start your journey."},
		{Level: 2, MinExp: 100, BadgeName: "Explorer", BadgeIcon: "🌿", TierName: "Newcomer", TierColor: "gray", Description: "You're starting to explore."},
		{Level: 3, MinExp: 300, BadgeName: "Learner", BadgeIcon: "📚", TierName: "Explorer", TierColor: "blue", Description: "You are exploring and growing."},
		{Level: 4, MinExp: 600, BadgeName: "Intermediate", BadgeIcon: "🌳", TierName: "Explorer", TierColor: "blue", Description: "Building momentum."},
		{Level: 5, MinExp: 1000, BadgeName: "Advanced", BadgeIcon: "🏆", TierName: "Explorer", TierColor: "blue", Description: "You're making great progress!"},
		{Level: 6, MinExp: 1500, BadgeName: "Expert", BadgeIcon: "💎", TierName: "Trusted", TierColor: "green", Description: "A trusted member of the community."},
		{Level: 7, MinExp: 2000, BadgeName: "Master", BadgeIcon: "⭐", TierName: "Trusted", TierColor: "green", Description: "Your dedication shows."},
		{Level: 8, MinExp: 3000, BadgeName: "Grandmaster", BadgeIcon: "👑", TierName: "Veteran", TierColor: "purple", Description: "A veteran with valuable experience."},
		{Level: 9, MinExp: 5000, BadgeName: "Legend", BadgeIcon: "🔥", TierName: "Veteran", TierColor: "purple", Description: "You inspire others."},
		{Level: 10, MinExp: 10000, BadgeName: "Guardian", BadgeIcon: "🛡️", TierName: "Guardian", TierColor: "gold", Description: "A guardian of the community."},
	}

	for _, config := range configs {
		var existing models.LevelConfig
		if db.Where("level = ?", config.Level).First(&existing).RowsAffected == 0 {
			if err := db.Create(&config).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
