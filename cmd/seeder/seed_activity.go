package main

import (
	"log"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedActivity(db *gorm.DB) {
	log.Println("📝 Seeding user activity and EXP history...")

	// Get member user
	var member models.User
	if err := db.Where("email = ?", "john@example.com").First(&member).Error; err != nil {
		log.Println("  ⚠️ User john@example.com not found, skipping activity seeding")
		return
	}

	// Define sample history to match EXP 850
	history := []struct {
		ActivityType string
		Points       int
		Description  string
		Date         time.Time
	}{
		{ActivityType: "daily_login", Points: 10, Description: "Login harian", Date: time.Now().Add(-5 * 24 * time.Hour)},
		{ActivityType: "daily_login", Points: 10, Description: "Login harian", Date: time.Now().Add(-4 * 24 * time.Hour)},
		{ActivityType: "daily_login", Points: 10, Description: "Login harian", Date: time.Now().Add(-3 * 24 * time.Hour)},
		{ActivityType: "daily_login", Points: 10, Description: "Login harian", Date: time.Now().Add(-2 * 24 * time.Hour)},
		{ActivityType: "daily_login", Points: 10, Description: "Login harian", Date: time.Now().Add(-1 * 24 * time.Hour)},

		{ActivityType: "create_forum", Points: 100, Description: "Membuat diskusi baru: Cara mengatasi burnout", Date: time.Now().Add(-5 * 24 * time.Hour)},
		{ActivityType: "create_forum", Points: 100, Description: "Membuat diskusi baru: Butuh saran meditasi", Date: time.Now().Add(-2 * 24 * time.Hour)},
		{ActivityType: "create_forum", Points: 100, Description: "Membuat diskusi baru: Pengalaman konseling", Date: time.Now().Add(-1 * 24 * time.Hour)},

		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Mengenal Kecemasan", Date: time.Now().Add(-5 * 24 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Teknik Pernapasan", Date: time.Now().Add(-4 * 24 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Panduan Meditasi", Date: time.Now().Add(-3 * 24 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Mengatasi Stres", Date: time.Now().Add(-2 * 24 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Pentingnya Tidur", Date: time.Now().Add(-1 * 24 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Menjaga Keseimbangan Hidup", Date: time.Now().Add(-6 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Makanan Sehat Mental", Date: time.Now().Add(-5 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Olahraga dan Mood", Date: time.Now().Add(-4 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Journaling", Date: time.Now().Add(-3 * time.Hour)},
		{ActivityType: "read_article", Points: 50, Description: "Membaca artikel: Bersyukur", Date: time.Now().Add(-2 * time.Hour)},
	}
	// Total points: 50 + 300 + 500 = 850. Correct.

	for _, h := range history {
		// Log History
		db.Create(&models.ExpHistory{
			UserID:       member.ID,
			ActivityType: h.ActivityType,
			Points:       h.Points,
			Description:  h.Description,
			CreatedAt:    h.Date,
		})

		// Update Daily Activity Stats (UserActivity)
		// Check if exists for that day
		dateOnly := h.Date
		var userActivity models.UserActivity
		if db.Where("user_id = ? AND activity_type = ? AND date(date) = date(?)", member.ID, h.ActivityType, dateOnly).First(&userActivity).RowsAffected == 0 {
			db.Create(&models.UserActivity{
				UserID:       member.ID,
				ActivityType: h.ActivityType,
				Date:         dateOnly,
				Count:        1,
			})
		} else {
			db.Model(&userActivity).Update("count", userActivity.Count+1)
		}
	}
	log.Println("  ✓ Seeded EXP history and activity for John Doe")
}
