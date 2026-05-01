package presentation

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedDemoProfile tunes seeded data into a curated state for product demos.
func SeedDemoProfile(db *gorm.DB) error {
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	var demoUser model.User
	if err := db.Where("email = ?", "gading@gmail.com").First(&demoUser).Error; err != nil {
		return nil
	}

	var backupUser model.User
	_ = db.Where("email = ?", "dery@gmail.com").First(&backupUser).Error

	var quotaUser model.User
	if err := db.Where("email = ?", "andhika@gmail.com").First(&quotaUser).Error; err == nil {
		if err := upsertFeatureUsage(db, quotaUser.ID, model.FeatureKeyChatAIMessages, startOfToday, 24); err != nil {
			return err
		}
	}

	if err := db.Model(&model.User{}).Where("id = ?", demoUser.ID).Updates(map[string]interface{}{
		"is_premium":         true,
		"premium_since":      now.AddDate(0, 0, -14),
		"premium_expires_at": now.AddDate(0, 0, 30),
		"gold_coins":         1200,
	}).Error; err != nil {
		return err
	}

	if backupUser.ID != 0 {
		if err := db.Model(&model.User{}).Where("id = ?", backupUser.ID).Updates(map[string]interface{}{
			"is_premium":         false,
			"premium_since":      nil,
			"premium_expires_at": nil,
			"gold_coins":         950,
		}).Error; err != nil {
			return err
		}
	}

	var organization model.Organization
	if err := db.Where("code = ?", "demo-kampus-nusantara").First(&organization).Error; err == nil {
		if err := db.Model(&organization).Updates(map[string]interface{}{
			"name":                     "Demo Kampus Ruang Tenang Live",
			"contact_email":            "demo-live@kampus-demo.ac.id",
			"requires_member_approval": true,
		}).Error; err != nil {
			return err
		}

		var activeSub model.B2BSubscription
		if err := db.Where("organization_id = ? AND status = ?", organization.ID, model.B2BSubscriptionStatusActive).First(&activeSub).Error; err == nil {
			if err := db.Model(&activeSub).Updates(map[string]interface{}{
				"used_seats":    1,
				"metadata_json": `{"source":"presentation-seeder","scenario":"live_showcase"}`,
			}).Error; err != nil {
				return err
			}

			if err := upsertReminderJob(db, model.B2BReminderJob{
				OrganizationID: organization.ID,
				SubscriptionID: uintPtr(activeSub.ID),
				JobType:        model.B2BReminderJobTypeSeatThreshold,
				Status:         model.B2BReminderJobStatusPending,
				DueAt:          startOfToday.Add(10 * time.Hour),
				PayloadJSON:    `{"threshold":0.8,"used_seats":1,"contracted_seats":120,"demo":true}`,
				AttemptCount:   0,
			}); err != nil {
				return err
			}
		}

		if err := upsertPricingRecommendation(db, model.B2BPricingRecommendation{
			OrganizationID:          organization.ID,
			GeneratedForDate:        startOfToday,
			RecommendedBillingCycle: model.B2BBillingCycleYearly,
			RecommendedSeats:        160,
			EstimatedMonthlyCost:    8200000,
			EstimatedYearlySaving:   9800000,
			ConfidenceScore:         91.25,
			ReasonsJSON:             `["Aktivitas mingguan stabil tinggi","Penggunaan fitur sosial meningkat","Potensi onboarding batch baru pada bulan depan"]`,
		}); err != nil {
			return err
		}
	}

	var sessions []model.ChatSession
	if err := db.Where("user_id = ?", demoUser.ID).Order("created_at ASC").Find(&sessions).Error; err != nil {
		return err
	}
	if len(sessions) > 0 {
		firstID := sessions[0].ID
		if err := db.Model(&model.ChatSession{}).Where("id = ?", firstID).Updates(map[string]interface{}{
			"title":                       "Demo Live: Refleksi Harian",
			"session_intent":              model.ChatSessionIntentReflection,
			"enable_mood_context":         true,
			"enable_journal_context":      true,
			"enable_daily_task_context":   true,
			"enable_xp_level_context":     true,
			"enable_breathing_context":    true,
			"enable_playlist_context":     true,
			"enable_rewards_context":      true,
			"enable_progress_map_context": true,
			"enable_social_context":       true,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
