package presentation

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// badgeImageURL returns the canonical URL for a level badge.
func badgeImageURL(level int) string {
	filename := fmt.Sprintf("%d.png", level)
	storagePath := filepath.Join("storage", "badge-image", filename)
	if _, err := os.Stat(storagePath); err == nil {
		return fmt.Sprintf("/storage/badge-image/%s", filename)
	}

	log.Printf("⚠️ Badge image not found for level %d, using fallback", level)
	return ""
}

// SeedLevelConfigs seeds the level configuration data
func SeedLevelConfigs(db *gorm.DB) error {
	type levelData struct {
		Level       int
		MinExp      int
		BadgeName   string
		TierName    string
		TierColor   string
		Description string
	}

	levels := []levelData{
		{Level: 1, MinExp: 0, BadgeName: "Newcomer", TierName: "Newcomer", TierColor: "gray", Description: "Welcome to Ruang Tenang! Start your journey."},
		{Level: 2, MinExp: 100, BadgeName: "Explorer", TierName: "Newcomer", TierColor: "gray", Description: "You're starting to explore."},
		{Level: 3, MinExp: 300, BadgeName: "Learner", TierName: "Explorer", TierColor: "blue", Description: "You are exploring and growing."},
		{Level: 4, MinExp: 600, BadgeName: "Intermediate", TierName: "Explorer", TierColor: "blue", Description: "Building momentum."},
		{Level: 5, MinExp: 1000, BadgeName: "Advanced", TierName: "Explorer", TierColor: "blue", Description: "You're making great progress!"},
		{Level: 6, MinExp: 1500, BadgeName: "Expert", TierName: "Trusted", TierColor: "green", Description: "A trusted member of the community."},
		{Level: 7, MinExp: 2000, BadgeName: "Master", TierName: "Trusted", TierColor: "green", Description: "Your dedication shows."},
		{Level: 8, MinExp: 3000, BadgeName: "Grandmaster", TierName: "Veteran", TierColor: "purple", Description: "A veteran with valuable experience."},
		{Level: 9, MinExp: 5000, BadgeName: "Legend", TierName: "Veteran", TierColor: "purple", Description: "You inspire others."},
		{Level: 10, MinExp: 10000, BadgeName: "Guardian", TierName: "Guardian", TierColor: "gold", Description: "A guardian of the community."},
	}

	for _, l := range levels {
		badgeIcon := badgeImageURL(l.Level)
		if badgeIcon == "" {
			log.Printf("⚠️ Skipping badge image for level %d (file not found)", l.Level)
		}

		config := model.LevelConfig{
			Level:       l.Level,
			MinExp:      l.MinExp,
			BadgeName:   l.BadgeName,
			BadgeIcon:   badgeIcon,
			TierName:    l.TierName,
			TierColor:   l.TierColor,
			Description: l.Description,
		}

		var existing model.LevelConfig
		if db.Where("level = ?", config.Level).First(&existing).RowsAffected == 0 {
			if err := db.Create(&config).Error; err != nil {
				return err
			}
			continue
		}

		updates := map[string]interface{}{
			"min_exp":     config.MinExp,
			"badge_name":  config.BadgeName,
			"tier_name":   config.TierName,
			"tier_color":  config.TierColor,
			"description": config.Description,
		}
		if config.BadgeIcon != "" {
			updates["badge_icon"] = config.BadgeIcon
		}

		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}
