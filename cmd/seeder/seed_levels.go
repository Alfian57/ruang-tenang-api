package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedLevels(db *gorm.DB) {
	log.Println("📝 Seeding level configs...")
	levelConfigs := []models.LevelConfig{
		{Level: 1, MinExp: 0, BadgeName: "Beginner", BadgeIcon: "🌱"},
		{Level: 2, MinExp: 100, BadgeName: "Explorer", BadgeIcon: "🌿"},
		{Level: 3, MinExp: 300, BadgeName: "Learner", BadgeIcon: "📚"},
		{Level: 4, MinExp: 600, BadgeName: "Intermediate", BadgeIcon: "🌳"},
		{Level: 5, MinExp: 1000, BadgeName: "Advanced", BadgeIcon: "🏆"},
		{Level: 6, MinExp: 1500, BadgeName: "Expert", BadgeIcon: "💎"},
		{Level: 7, MinExp: 2000, BadgeName: "Master", BadgeIcon: "⭐"},
		{Level: 8, MinExp: 3000, BadgeName: "Grandmaster", BadgeIcon: "👑"},
	}

	for _, lc := range levelConfigs {
		var existing models.LevelConfig
		if db.Where("level = ?", lc.Level).First(&existing).RowsAffected == 0 {
			db.Create(&lc)
			log.Printf("  ✓ Created level config: Level %d - %s (%s)", lc.Level, lc.BadgeName, lc.BadgeIcon)
		}
	}
}
