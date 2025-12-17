package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedArticles(db *gorm.DB) {
	// Seed Article Categories
	log.Println("📝 Seeding article categories...")
	articleCategories := []models.ArticleCategory{
		{Name: "Kesehatan Mental", Description: "Artikel seputar kesehatan mental dan psikologi"},
		{Name: "Tips & Trik", Description: "Tips praktis untuk kehidupan sehari-hari yang lebih tenang"},
		{Name: "Meditasi", Description: "Panduan dan informasi mengenai teknik meditasi"},
		{Name: "Motivasi", Description: "Inspirasi untuk tetap semangat dan positif"},
		{Name: "Mindfulness", Description: "Praktik kesadaran penuh untuk ketenangan pikiran"},
	}

	for _, cat := range articleCategories {
		var existing models.ArticleCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			db.Create(&cat)
			log.Printf("  ✓ Created article category: %s", cat.Name)
		}
	}

	// Copy article images to uploads
	log.Println("📷 Copying article images to uploads...")
	articleImages := []string{
		copyAsset("assets/images/article-1.jpg", "images"),
		copyAsset("assets/images/article-2.jpg", "images"),
		copyAsset("assets/images/article-3.jpg", "images"),
		copyAsset("assets/images/article-4.jpg", "images"),
		copyAsset("assets/images/article-5.jpg", "images"),
	}

	// Seed Articles
	log.Println("📝 Seeding articles...")
	var healthCategory models.ArticleCategory
	db.Where("name = ?", "Kesehatan Mental").First(&healthCategory)

	var tipsCategory models.ArticleCategory
	db.Where("name = ?", "Tips & Trik").First(&tipsCategory)

	var meditasiCategory models.ArticleCategory
	db.Where("name = ?", "Meditasi").First(&meditasiCategory)

	var adminUser models.User
	if err := db.Where("email = ?", "admin@ruangtenang.id").First(&adminUser).Error; err != nil {
		db.First(&adminUser) // Fallback to first user
	}

	articles := []models.Article{
		{
			Title:             "Mengenal Kecemasan dan Cara Mengatasinya",
			Thumbnail:         articleImages[0],
			Content:           "Kecemasan adalah respons alami tubuh terhadap stres. Ini adalah perasaan takut atau khawatir tentang apa yang akan datang.\n\nCara Mengatasi Kecemasan:\n1. Latihan pernapasan dalam\n2. Meditasi teratur\n3. Olahraga rutin\n4. Tidur yang cukup\n5. Mengurangi kafein",
			ArticleCategoryID: healthCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "5 Teknik Pernapasan untuk Menenangkan Pikiran",
			Thumbnail:         articleImages[1],
			Content:           "Pernapasan yang tepat dapat membantu menenangkan sistem saraf.\n\n1. Teknik 4-7-8\nTarik napas selama 4 detik, tahan 7 detik, hembuskan 8 detik.\n\n2. Pernapasan Kotak\nTarik napas 4 detik, tahan 4 detik, hembuskan 4 detik, tahan 4 detik.",
			ArticleCategoryID: tipsCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Panduan Meditasi untuk Pemula",
			Thumbnail:         articleImages[2],
			Content:           "Meditasi tidak harus rumit. Mulailah dengan 5 menit sehari.\n\nLangkah-langkah:\n1. Duduk dengan nyaman\n2. Tutup mata\n3. Fokus pada napas\n4. Biarkan pikiran mengalir\n5. Kembalikan fokus ke napas",
			ArticleCategoryID: meditasiCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Mengatasi Stres di Tempat Kerja",
			Thumbnail:         articleImages[3],
			Content:           "Stres kerja adalah masalah umum yang dapat mempengaruhi kesehatan mental.\n\nTips Mengatasi:\n1. Buat batasan yang jelas antara kerja dan kehidupan pribadi\n2. Ambil break secara teratur\n3. Prioritaskan tugas dengan baik\n4. Jangan takut untuk meminta bantuan",
			ArticleCategoryID: tipsCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Pentingnya Tidur untuk Kesehatan Mental",
			Thumbnail:         articleImages[4],
			Content:           "Tidur yang cukup sangat penting untuk menjaga kesehatan mental.\n\nManfaat Tidur yang Cukup:\n1. Meningkatkan konsentrasi\n2. Memperbaiki mood\n3. Mengurangi risiko depresi\n4. Meningkatkan daya ingat",
			ArticleCategoryID: healthCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
	}

	for _, article := range articles {
		var existing models.Article
		if db.Where("title = ?", article.Title).First(&existing).RowsAffected == 0 {
			db.Create(&article)
			log.Printf("  ✓ Created article: %s", article.Title)
		}
	}
}
