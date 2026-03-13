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

// SeedRewards seeds the reward catalog for the gold coin shop
func SeedRewards(db *gorm.DB) error {
	rewards := []model.Reward{
		{Name: "XP Boost 2x (24 Jam)", Description: "Dapatkan XP double selama 24 jam untuk semua aktivitas Anda di platform.", CoinCost: 250, Stock: -1, IsActive: true, RewardType: model.RewardTypeXPBoost, RewardValue: "2x_24h", Image: copyRewardImage("reward_xp.png")},
		{Name: "Tema Dashboard: Ocean Calm", Description: "Tema dashboard bernuansa biru laut yang menenangkan. Rasakan ketenangan samudra di setiap halaman.", CoinCost: 150, Stock: -1, IsActive: true, RewardType: model.RewardTypeTheme, RewardValue: "ocean_calm", Image: copyRewardImage("reward_ocean.png")},
		{Name: "Tema Dashboard: Forest Zen", Description: "Tema dashboard bernuansa hijau alam yang segar. Bawa suasana hutan yang damai ke dashboard Anda.", CoinCost: 150, Stock: -1, IsActive: true, RewardType: model.RewardTypeTheme, RewardValue: "forest_zen", Image: copyRewardImage("reward_forest.png")},
		{Name: "Tema Dashboard: Sunset Warmth", Description: "Tema dashboard bernuansa hangat matahari terbenam. Nikmati kehangatan senja di setiap kunjungan.", CoinCost: 200, Stock: -1, IsActive: true, RewardType: model.RewardTypeTheme, RewardValue: "sunset_warmth", Image: copyRewardImage("reward_sunset.png")},
	}

	// Deactivate old rewards that are not in the new list
	newNames := make([]string, len(rewards))
	for i, r := range rewards {
		newNames[i] = r.Name
	}
	db.Model(&model.Reward{}).Where("name NOT IN ?", newNames).Update("is_active", false)

	for _, reward := range rewards {
		var existing model.Reward
		if db.Where("name = ?", reward.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&reward).Error; err != nil {
				return err
			}
		} else {
			// Update existing reward to match new values
			db.Model(&existing).Updates(map[string]interface{}{
				"description":  reward.Description,
				"coin_cost":    reward.CoinCost,
				"stock":        reward.Stock,
				"is_active":    reward.IsActive,
				"reward_type":  reward.RewardType,
				"reward_value": reward.RewardValue,
				"image":        reward.Image,
			})
		}
	}
	return nil
}

// copyRewardImage returns a reward image URL
func copyRewardImage(filename string) string {
	paths := []string{
		filepath.Join("storage", "images", filename),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			uploadDir := filepath.Join("uploads", "images")
			os.MkdirAll(uploadDir, 0755)

			dstFilename := filename
			dstPath := filepath.Join(uploadDir, dstFilename)

			// Copy file
			if src, err := os.Open(p); err == nil {
				if dst, err := os.Create(dstPath); err == nil {
					io.Copy(dst, src)
					dst.Close()
					src.Close()
					return fmt.Sprintf("/uploads/images/%s", dstFilename)
				}
				src.Close()
			}
		}
	}
	log.Printf("⚠️ Reward image not found for %s", filename)
	return ""
}
