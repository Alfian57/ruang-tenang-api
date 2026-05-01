package presentation

import (
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedB2BPlans seeds reusable B2B plans for organization subscription flows.
func SeedB2BPlans(db *gorm.DB) error {
	plans := []model.B2BPlan{
		{
			Code:             "b2b-campus-monthly",
			Name:             "B2B Campus Monthly",
			Description:      "Langganan bulanan untuk kampus/sekolah dengan skala kecil-menengah.",
			BillingCycle:     model.B2BBillingCycleMonthly,
			BasePricePerSeat: 59000,
			MinSeats:         25,
			MaxSeats:         5000,
			FeaturesJSON:     `{"sso":true,"seat_management":true,"analytics":true,"reminders":true}`,
			IsActive:         true,
		},
		{
			Code:             "b2b-campus-yearly",
			Name:             "B2B Campus Yearly",
			Description:      "Langganan tahunan dengan diskon untuk institusi pendidikan.",
			BillingCycle:     model.B2BBillingCycleYearly,
			BasePricePerSeat: 52000,
			MinSeats:         25,
			MaxSeats:         5000,
			FeaturesJSON:     `{"sso":true,"seat_management":true,"analytics":true,"reminders":true,"priority_support":true}`,
			IsActive:         true,
		},
		{
			Code:             "b2b-enterprise-monthly",
			Name:             "B2B Enterprise Monthly",
			Description:      "Paket enterprise bulanan untuk organisasi besar.",
			BillingCycle:     model.B2BBillingCycleMonthly,
			BasePricePerSeat: 79000,
			MinSeats:         100,
			MaxSeats:         20000,
			FeaturesJSON:     `{"sso":true,"seat_management":true,"analytics":true,"reminders":true,"advanced_audit":true,"custom_onboarding":true}`,
			IsActive:         true,
		},
		{
			Code:             "b2b-enterprise-yearly",
			Name:             "B2B Enterprise Yearly",
			Description:      "Paket enterprise tahunan dengan efisiensi biaya tinggi.",
			BillingCycle:     model.B2BBillingCycleYearly,
			BasePricePerSeat: 71000,
			MinSeats:         100,
			MaxSeats:         20000,
			FeaturesJSON:     `{"sso":true,"seat_management":true,"analytics":true,"reminders":true,"advanced_audit":true,"custom_onboarding":true,"dedicated_support":true}`,
			IsActive:         true,
		},
	}

	for _, plan := range plans {
		var existing model.B2BPlan
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
			"name":                plan.Name,
			"description":         plan.Description,
			"billing_cycle":       plan.BillingCycle,
			"base_price_per_seat": plan.BasePricePerSeat,
			"min_seats":           plan.MinSeats,
			"max_seats":           plan.MaxSeats,
			"features_json":       plan.FeaturesJSON,
			"is_active":           plan.IsActive,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
