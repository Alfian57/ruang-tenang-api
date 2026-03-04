package production

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// copyBadgeImage copies a badge image from the badge-image directory to uploads/images
func copyBadgeImage(level int) string {
	srcPath := filepath.Join("storage", "badge-image", fmt.Sprintf("%d.png", level))

	// Check if source file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		log.Printf("⚠️ Badge image not found: %s", srcPath)
		return ""
	}

	// Create uploads directory
	uploadDir := filepath.Join("uploads", "images")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("⚠️ Failed to create upload dir: %v", err)
		return ""
	}

	// Use a deterministic filename for seeder (so re-running doesn't duplicate)
	filename := fmt.Sprintf("badge_level_%d.png", level)
	dstPath := filepath.Join(uploadDir, filename)

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		log.Printf("⚠️ Failed to open badge image: %v", err)
		return ""
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Printf("⚠️ Failed to create destination file: %v", err)
		return ""
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Printf("⚠️ Failed to copy badge image: %v", err)
		return ""
	}

	return fmt.Sprintf("/uploads/images/%s", filename)
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
		badgeIcon := copyBadgeImage(l.Level)
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
		}
	}
	return nil
}
