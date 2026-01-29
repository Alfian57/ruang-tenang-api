package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedCrisisKeywords(db *gorm.DB) {
	log.Println("🚨 Seeding crisis keywords...")

	keywords := []models.CrisisKeyword{
		// Suicide-related (Critical)
		{Keyword: "bunuh diri", Category: "suicide", Severity: "critical", Language: "id", IsActive: true, Notes: "Direct suicide term"},
		{Keyword: "ingin mati", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "mau mati", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "mengakhiri hidup", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "mengakhiri hidupku", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "tidak ingin hidup", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "lebih baik mati", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "menyerah hidup", Category: "suicide", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "suicide", Category: "suicide", Severity: "critical", Language: "en", IsActive: true},
		{Keyword: "kill myself", Category: "suicide", Severity: "critical", Language: "en", IsActive: true},
		{Keyword: "end my life", Category: "suicide", Severity: "critical", Language: "en", IsActive: true},

		// Self-harm (High)
		{Keyword: "menyakiti diri", Category: "self_harm", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "melukai diri", Category: "self_harm", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "menyakiti diri sendiri", Category: "self_harm", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "self harm", Category: "self_harm", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "hurt myself", Category: "self_harm", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "cutting", Category: "self_harm", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "potong nadi", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "gantung diri", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "lompat dari", Category: "self_harm", Severity: "high", Language: "id", IsActive: true, Notes: "Context-dependent"},

		// Severe Depression (Medium-High)
		{Keyword: "tidak ada harapan", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak ada gunanya", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "merasa tidak berguna", Category: "severe_depression", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "tidak ada yang peduli", Category: "severe_depression", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "tidak ada artinya hidup", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "hopeless", Category: "severe_depression", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "worthless", Category: "severe_depression", Severity: "medium", Language: "en", IsActive: true},
		{Keyword: "no point in living", Category: "severe_depression", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "sangat depresi", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak sanggup lagi", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak kuat lagi", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "lelah dengan hidup", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},

		// Emergency indicators (Medium)
		{Keyword: "tidak bisa tidur", Category: "emergency", Severity: "medium", Language: "id", IsActive: true, Notes: "May indicate crisis"},
		{Keyword: "tidak bisa makan", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "panik", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "serangan panik", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "panic attack", Category: "emergency", Severity: "medium", Language: "en", IsActive: true},
	}

	for _, keyword := range keywords {
		if err := db.Create(&keyword).Error; err != nil {
			log.Printf("   ⚠️ Failed to create keyword '%s': %v", keyword.Keyword, err)
		}
	}

	log.Println("   ✅ Crisis keywords seeded successfully")
}
