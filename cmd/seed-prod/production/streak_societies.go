package production

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

func SeedStreakSocieties(db *gorm.DB) error {
	societies := []model.StreakSociety{
		{Name: "Pemula Konsisten", Icon: "🌱", MinStreak: 3, BorderColor: "#4CAF50", BadgeGlow: false, ExclusiveChat: false, DisplayOrder: 1},
		{Name: "Pejuang Seminggu", Icon: "🔥", MinStreak: 7, BorderColor: "#FF9800", BadgeGlow: false, ExclusiveChat: false, DisplayOrder: 2},
		{Name: "Ksatria Dua Minggu", Icon: "⚔️", MinStreak: 14, BorderColor: "#2196F3", BadgeGlow: false, ExclusiveChat: true, DisplayOrder: 3},
		{Name: "Penjaga Sebulan", Icon: "🛡️", MinStreak: 30, BorderColor: "#9C27B0", BadgeGlow: true, ExclusiveChat: true, DisplayOrder: 4},
		{Name: "Master Dua Bulan", Icon: "👑", MinStreak: 60, BorderColor: "#FFD700", BadgeGlow: true, ExclusiveChat: true, DisplayOrder: 5},
		{Name: "Legenda Seratus", Icon: "🏆", MinStreak: 100, BorderColor: "#FF4081", BadgeGlow: true, ExclusiveChat: true, DisplayOrder: 6},
	}

	for _, s := range societies {
		var existing model.StreakSociety
		if db.Where("name = ?", s.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&s).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
