package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

func SeedSpinRewards(db *gorm.DB) error {
	rewards := []model.SpinReward{
		{Name: "+10 XP", Icon: "⭐", RewardType: model.SpinRewardXP, RewardValue: 10, Weight: 250, Rarity: "common", IsActive: true},
		{Name: "+25 XP", Icon: "🌟", RewardType: model.SpinRewardXP, RewardValue: 25, Weight: 150, Rarity: "common", IsActive: true},
		{Name: "+50 XP", Icon: "💫", RewardType: model.SpinRewardXP, RewardValue: 50, Weight: 80, Rarity: "rare", IsActive: true},
		{Name: "+5 Koin", Icon: "🪙", RewardType: model.SpinRewardCoins, RewardValue: 5, Weight: 200, Rarity: "common", IsActive: true},
		{Name: "+15 Koin", Icon: "💰", RewardType: model.SpinRewardCoins, RewardValue: 15, Weight: 100, Rarity: "rare", IsActive: true},
		{Name: "+50 Koin", Icon: "💎", RewardType: model.SpinRewardCoins, RewardValue: 50, Weight: 30, Rarity: "legendary", IsActive: true},
		{Name: "Streak Freeze", Icon: "🧊", RewardType: model.SpinRewardStreakFreeze, RewardValue: 1, Weight: 50, Rarity: "epic", IsActive: true},
		{Name: "XP Boost 2x", Icon: "🚀", RewardType: model.SpinRewardXPBoost, RewardValue: 2, Weight: 40, Rarity: "epic", IsActive: true},
		{Name: "Coba Lagi", Icon: "😅", RewardType: model.SpinRewardNothing, RewardValue: 0, Weight: 100, Rarity: "common", IsActive: true},
	}

	for _, r := range rewards {
		var existing model.SpinReward
		if db.Where("name = ?", r.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&r).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
