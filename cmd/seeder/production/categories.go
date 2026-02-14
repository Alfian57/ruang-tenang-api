package production

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedArticleCategories seeds the article categories
func SeedArticleCategories(db *gorm.DB) error {
	categories := []model.ArticleCategory{
		{Name: "Kesehatan Mental", Description: "Artikel seputar kesehatan mental dan psikologi"},
		{Name: "Tips & Trik", Description: "Tips praktis untuk kehidupan sehari-hari yang lebih tenang"},
		{Name: "Meditasi", Description: "Panduan dan informasi mengenai teknik meditasi"},
		{Name: "Motivasi", Description: "Inspirasi untuk tetap semangat dan positif"},
		{Name: "Mindfulness", Description: "Praktik kesadaran penuh untuk ketenangan pikiran"},
		{Name: "Relaksasi", Description: "Teknik dan tips untuk relaksasi"},
		{Name: "Self-Care", Description: "Panduan merawat diri sendiri"},
		{Name: "Hubungan", Description: "Artikel tentang hubungan interpersonal"},
	}

	for _, cat := range categories {
		var existing model.ArticleCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// SeedSongCategories seeds the song/music categories
func SeedSongCategories(db *gorm.DB) error {
	categories := []model.SongCategory{
		{Name: "Alam"},
		{Name: "Piano"},
		{Name: "Hujan"},
		{Name: "Laut"},
		{Name: "Meditasi"},
		{Name: "White Noise"},
		{Name: "ASMR"},
		{Name: "Classical"},
	}

	for _, cat := range categories {
		var existing model.SongCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// SeedForumCategories seeds the forum categories
func SeedForumCategories(db *gorm.DB) error {
	categories := []model.ForumCategory{
		{Name: "Diskusi Umum"},
		{Name: "Curhat & Keluh Kesah"},
		{Name: "Dukungan Emosional"},
		{Name: "Tips Mengelola Stres"},
		{Name: "Kisah Inspiratif"},
		{Name: "Kesehatan Mental di Tempat Kerja"},
		{Name: "Kesehatan Mental di Sekolah"},
		{Name: "Pertanyaan & Jawaban"},
	}

	for _, cat := range categories {
		var existing model.ForumCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
