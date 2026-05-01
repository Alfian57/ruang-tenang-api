package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

func SeedLeagueDivisions(db *gorm.DB) error {
	divisions := []model.LeagueDivision{
		{Name: "Perunggu III", Icon: "🥉", Tier: 1, Color: "#CD7F32", MinRank: 0, PromotionSlots: 15, DemotionSlots: 0},
		{Name: "Perunggu II", Icon: "🥉", Tier: 2, Color: "#CD7F32", MinRank: 0, PromotionSlots: 12, DemotionSlots: 5},
		{Name: "Perunggu I", Icon: "🥉", Tier: 3, Color: "#CD7F32", MinRank: 0, PromotionSlots: 10, DemotionSlots: 5},
		{Name: "Perak III", Icon: "🥈", Tier: 4, Color: "#C0C0C0", MinRank: 0, PromotionSlots: 10, DemotionSlots: 5},
		{Name: "Perak II", Icon: "🥈", Tier: 5, Color: "#C0C0C0", MinRank: 0, PromotionSlots: 10, DemotionSlots: 5},
		{Name: "Perak I", Icon: "🥈", Tier: 6, Color: "#C0C0C0", MinRank: 0, PromotionSlots: 8, DemotionSlots: 5},
		{Name: "Emas III", Icon: "🥇", Tier: 7, Color: "#FFD700", MinRank: 0, PromotionSlots: 8, DemotionSlots: 5},
		{Name: "Emas II", Icon: "🥇", Tier: 8, Color: "#FFD700", MinRank: 0, PromotionSlots: 8, DemotionSlots: 5},
		{Name: "Emas I", Icon: "🥇", Tier: 9, Color: "#FFD700", MinRank: 0, PromotionSlots: 5, DemotionSlots: 5},
		{Name: "Berlian", Icon: "💎", Tier: 10, Color: "#B9F2FF", MinRank: 0, PromotionSlots: 0, DemotionSlots: 5},
	}

	for _, d := range divisions {
		var existing model.LeagueDivision
		if db.Where("name = ?", d.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&d).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
