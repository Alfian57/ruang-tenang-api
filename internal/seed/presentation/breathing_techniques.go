package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedBreathingTechniques seeds the predefined breathing techniques
func SeedBreathingTechniques(db *gorm.DB) error {
	boxSlug := "box-breathing"
	fourSevenEightSlug := "4-7-8-breathing"
	coherentSlug := "coherent-breathing"
	energizingSlug := "energizing-breath"
	deepCalmSlug := "deep-calm-breathing"

	boxDesc := "Teknik pernapasan kotak yang digunakan oleh Navy SEALs untuk tetap tenang dan fokus dalam situasi stres tinggi."
	boxBenefits := "Mengurangi stres, menenangkan sistem saraf, meningkatkan fokus dan konsentrasi, membantu mengendalikan respons fight-or-flight."
	boxBestFor := "Saat merasa cemas sebelum presentasi, menghadapi deadline, atau saat serangan panik ringan."
	boxOrigin := "Digunakan oleh Navy SEALs dan first responders"

	fourSevenEightDesc := "Teknik relaksasi yang dikembangkan Dr. Andrew Weil, sangat efektif untuk membantu tidur dan meredakan kecemasan."
	fourSevenEightBenefits := "Membantu tidur lebih cepat, meredakan kecemasan, memperlambat detak jantung, mengaktifkan sistem saraf parasimpatis."
	fourSevenEightBestFor := "Sebelum tidur, saat insomnia, pikiran yang terus berputar, atau butuh relaksasi mendalam."
	fourSevenEightOrigin := "Dikembangkan oleh Dr. Andrew Weil berdasarkan pranayama yoga"

	coherentDesc := "Pernapasan seimbang dengan ritme yang konsisten untuk mencapai keseimbangan sistem saraf dan ketenangan pikiran."
	coherentBenefits := "Menyeimbangkan sistem saraf, meningkatkan heart rate variability (HRV), meningkatkan konsentrasi dan clarity mental."
	coherentBestFor := "Meditasi, belajar, mental reset, atau saat butuh ketenangan tanpa mengantuk."
	coherentOrigin := "Berdasarkan penelitian heart rate variability"

	energizingDesc := "Teknik pernapasan cepat yang meningkatkan energi dan kesadaran tanpa kafein."
	energizingBenefits := "Meningkatkan energi dan alertness, melawan kelelahan, meningkatkan sirkulasi darah, membangunkan tubuh dan pikiran."
	energizingBestFor := "Pagi hari setelah bangun, afternoon slump, sebelum olahraga, atau saat merasa lesu."
	energizingOrigin := "Adaptasi dari teknik Kapalabhati pranayama"

	deepCalmDesc := "Pernapasan dalam dengan exhale panjang untuk menenangkan emosi intens dan overwhelm."
	deepCalmBenefits := "Regulasi emosi yang kuat, menenangkan setelah konflik, mengurangi overwhelm, menurunkan tekanan darah."
	deepCalmBestFor := "Setelah pertengkaran, saat merasa overwhelmed, emosi intens, atau butuh grounding cepat."
	deepCalmOrigin := "Berdasarkan teknik pernapasan diafragma klinis"

	techniques := []model.BreathingTechnique{
		{
			Name:               "Box Breathing",
			Slug:               &boxSlug,
			Description:        &boxDesc,
			Benefits:           &boxBenefits,
			BestFor:            &boxBestFor,
			InhaleDuration:     4,
			InhaleHoldDuration: 4,
			ExhaleDuration:     4,
			ExhaleHoldDuration: 4,
			Icon:               "⬜",
			Color:              "#3B82F6",
			AnimationType:      "square",
			Difficulty:         "easy",
			Category:           "stress",
			Origin:             &boxOrigin,
			IsSystem:           true,
			IsActive:           true,
		},
		{
			Name:               "4-7-8 Breathing",
			Slug:               &fourSevenEightSlug,
			Description:        &fourSevenEightDesc,
			Benefits:           &fourSevenEightBenefits,
			BestFor:            &fourSevenEightBestFor,
			InhaleDuration:     4,
			InhaleHoldDuration: 7,
			ExhaleDuration:     8,
			ExhaleHoldDuration: 0,
			Icon:               "🌙",
			Color:              "#8B5CF6",
			AnimationType:      "circle",
			Difficulty:         "intermediate",
			Category:           "sleep",
			Origin:             &fourSevenEightOrigin,
			IsSystem:           true,
			IsActive:           true,
		},
		{
			Name:               "Coherent Breathing",
			Slug:               &coherentSlug,
			Description:        &coherentDesc,
			Benefits:           &coherentBenefits,
			BestFor:            &coherentBestFor,
			InhaleDuration:     5,
			InhaleHoldDuration: 0,
			ExhaleDuration:     5,
			ExhaleHoldDuration: 0,
			Icon:               "♾️",
			Color:              "#10B981",
			AnimationType:      "wave",
			Difficulty:         "easy",
			Category:           "focus",
			Origin:             &coherentOrigin,
			IsSystem:           true,
			IsActive:           true,
		},
		{
			Name:               "Energizing Breath",
			Slug:               &energizingSlug,
			Description:        &energizingDesc,
			Benefits:           &energizingBenefits,
			BestFor:            &energizingBestFor,
			InhaleDuration:     2,
			InhaleHoldDuration: 0,
			ExhaleDuration:     4,
			ExhaleHoldDuration: 0,
			Icon:               "⚡",
			Color:              "#F59E0B",
			AnimationType:      "circle",
			Difficulty:         "easy",
			Category:           "energy",
			Origin:             &energizingOrigin,
			IsSystem:           true,
			IsActive:           true,
		},
		{
			Name:               "Deep Calm Breathing",
			Slug:               &deepCalmSlug,
			Description:        &deepCalmDesc,
			Benefits:           &deepCalmBenefits,
			BestFor:            &deepCalmBestFor,
			InhaleDuration:     6,
			InhaleHoldDuration: 2,
			ExhaleDuration:     8,
			ExhaleHoldDuration: 0,
			Icon:               "🌊",
			Color:              "#06B6D4",
			AnimationType:      "wave",
			Difficulty:         "intermediate",
			Category:           "stress",
			Origin:             &deepCalmOrigin,
			IsSystem:           true,
			IsActive:           true,
		},
	}

	for _, tech := range techniques {
		var existing model.BreathingTechnique
		if db.Where("slug = ?", *tech.Slug).First(&existing).RowsAffected == 0 {
			if err := db.Create(&tech).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
