package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedCrisisKeywords seeds the crisis detection keywords
func SeedCrisisKeywords(db *gorm.DB) error {
	keywords := []model.CrisisKeyword{
		// Self-harm keywords (Indonesian)
		{Keyword: "menyakiti diri sendiri", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "melukai diri", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "menyayat", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "menyilet", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "potong nadi", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "gantung diri", Category: "self_harm", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "lompat dari", Category: "self_harm", Severity: "high", Language: "id", IsActive: true, Notes: "Context-dependent"},

		// Self-harm keywords (English)
		{Keyword: "self harm", Category: "self_harm", Severity: "critical", Language: "en", IsActive: true},
		{Keyword: "hurt myself", Category: "self_harm", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "cutting", Category: "self_harm", Severity: "high", Language: "en", IsActive: true},

		// Suicide keywords (Indonesian)
		{Keyword: "bunuh diri", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "ingin mati", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "mau mati", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "mengakhiri hidup", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "mengakhiri semuanya", Category: "suicide", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak ingin hidup", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "nggak mau hidup", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "gak mau hidup lagi", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "lebih baik mati", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "overdose", Category: "suicide", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "minum obat banyak", Category: "suicide", Severity: "high", Language: "id", IsActive: true},

		// Suicide keywords (English)
		{Keyword: "suicide", Category: "suicide", Severity: "critical", Language: "en", IsActive: true},
		{Keyword: "kill myself", Category: "suicide", Severity: "critical", Language: "en", IsActive: true},
		{Keyword: "end my life", Category: "suicide", Severity: "critical", Language: "en", IsActive: true},

		// Severe depression indicators (Indonesian)
		{Keyword: "tidak ada harapan", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "nggak ada harapan", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "hidup tidak berarti", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "hidup nggak ada artinya", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "beban bagi semua orang", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "semua akan lebih baik tanpa aku", Category: "severe_depression", Severity: "critical", Language: "id", IsActive: true},
		{Keyword: "tidak ada yang peduli", Category: "severe_depression", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "merasa tidak berguna", Category: "severe_depression", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "tidak ada gunanya", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "sangat depresi", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak sanggup lagi", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak kuat lagi", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "lelah dengan hidup", Category: "severe_depression", Severity: "high", Language: "id", IsActive: true},

		// Severe depression indicators (English)
		{Keyword: "hopeless", Category: "severe_depression", Severity: "high", Language: "en", IsActive: true},
		{Keyword: "worthless", Category: "severe_depression", Severity: "medium", Language: "en", IsActive: true},
		{Keyword: "no point in living", Category: "severe_depression", Severity: "high", Language: "en", IsActive: true},

		// Emergency indicators (Indonesian)
		{Keyword: "darurat", Category: "emergency", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tolong", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "butuh bantuan sekarang", Category: "emergency", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak bisa menahan", Category: "emergency", Severity: "high", Language: "id", IsActive: true},
		{Keyword: "tidak bisa tidur", Category: "emergency", Severity: "medium", Language: "id", IsActive: true, Notes: "May indicate crisis"},
		{Keyword: "tidak bisa makan", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "panik", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},
		{Keyword: "serangan panik", Category: "emergency", Severity: "medium", Language: "id", IsActive: true},

		// Emergency indicators (English)
		{Keyword: "panic attack", Category: "emergency", Severity: "medium", Language: "en", IsActive: true},
	}

	for _, kw := range keywords {
		var existing model.CrisisKeyword
		if db.Where("keyword = ? AND language = ?", kw.Keyword, kw.Language).First(&existing).RowsAffected == 0 {
			if err := db.Create(&kw).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
