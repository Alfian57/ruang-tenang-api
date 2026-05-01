package presentation

import (
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedPremiumCatalog seeds premium plans and topup packages used by billing flows.
func SeedPremiumCatalog(db *gorm.DB) error {
	premiumPlans := []model.PremiumPlan{
		{Code: "premium_monthly", Name: "Premium Monthly", Description: "Akses premium 30 hari untuk personal use.", Price: 29900, DurationDays: 30, IsActive: true},
		{Code: "premium_quarterly", Name: "Premium Quarterly", Description: "Akses premium 90 hari dengan harga lebih hemat.", Price: 79900, DurationDays: 90, IsActive: true},
		{Code: "premium_yearly", Name: "Premium Yearly", Description: "Akses premium 365 hari untuk komitmen jangka panjang.", Price: 279000, DurationDays: 365, IsActive: true},
	}

	for _, plan := range premiumPlans {
		var existing model.PremiumPlan
		err := db.Where("code = ?", plan.Code).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&plan).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}

		if err := db.Model(&existing).Updates(map[string]interface{}{
			"name":          plan.Name,
			"description":   plan.Description,
			"price":         plan.Price,
			"duration_days": plan.DurationDays,
			"is_active":     plan.IsActive,
		}).Error; err != nil {
			return err
		}
	}

	topupPackages := []model.TopupPackage{
		{Code: "topup_100", Name: "Topup 100 Koin", Coins: 100, BonusCoins: 0, Price: 12000, IsActive: true},
		{Code: "topup_250", Name: "Topup 250 Koin", Coins: 250, BonusCoins: 15, Price: 29000, IsActive: true},
		{Code: "topup_500", Name: "Topup 500 Koin", Coins: 500, BonusCoins: 50, Price: 55000, IsActive: true},
		{Code: "topup_1000", Name: "Topup 1000 Koin", Coins: 1000, BonusCoins: 150, Price: 99000, IsActive: true},
	}

	for _, pkg := range topupPackages {
		var existing model.TopupPackage
		err := db.Where("code = ?", pkg.Code).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&pkg).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}

		if err := db.Model(&existing).Updates(map[string]interface{}{
			"name":        pkg.Name,
			"coins":       pkg.Coins,
			"bonus_coins": pkg.BonusCoins,
			"price":       pkg.Price,
			"is_active":   pkg.IsActive,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
