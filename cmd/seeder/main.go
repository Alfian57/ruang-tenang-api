package main

import (
	"fmt"
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("🔥 Dropping all tables to clear data...")
	migrator := db.Migrator()
	if err := migrator.DropTable(
		&models.UserMood{},
		&models.ChatMessage{},
		&models.ChatSession{},
		&models.Song{},
		&models.SongCategory{},
		&models.Article{},
		&models.ArticleCategory{},
		&models.User{},
	); err != nil {
		log.Printf("⚠️ Failed to drop tables (might not exist): %v", err)
	}

	log.Println("🔄 Running migrations (AutoMigrate)...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.ArticleCategory{},
		&models.Article{},
		&models.SongCategory{},
		&models.Song{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.UserMood{},
	); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}

	log.Println("🌱 Starting database seeder...")

	// Seed Users
	log.Println("📝 Seeding users...")
	adminPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	memberPassword, err := utils.HashPassword("member123")
	if err != nil {
		log.Fatalf("Failed to hash member password: %v", err)
	}

	users := []models.User{
		{Name: "Admin", Email: "admin@ruangtenang.id", Password: adminPassword, Role: models.RoleAdmin, Exp: 1500},
		{Name: "John Doe", Email: "john@example.com", Password: memberPassword, Role: models.RoleMember, Exp: 850},
		{Name: "Jane Smith", Email: "jane@example.com", Password: memberPassword, Role: models.RoleMember, Exp: 1200},
		{Name: "Alice Wonderland", Email: "alice@example.com", Password: memberPassword, Role: models.RoleMember, Exp: 2300},
		{Name: "Bob Builder", Email: "bob@example.com", Password: memberPassword, Role: models.RoleMember, Exp: 450},
	}

	for _, user := range users {
		var existing models.User
		result := db.Where("email = ?", user.Email).First(&existing)

		if result.RowsAffected == 0 {
			// Create new user
			if err := db.Create(&user).Error; err != nil {
				log.Printf("  ❌ Failed to create user %s: %v", user.Email, err)
			} else {
				log.Printf("  ✓ Created user: %s", user.Email)
			}
		} else {
			// Update existing user's password and exp
			if err := db.Model(&existing).Updates(map[string]interface{}{
				"password": user.Password,
				"exp":      user.Exp,
			}).Error; err != nil {
				log.Printf("  ❌ Failed to update user %s: %v", user.Email, err)
			} else {
				log.Printf("  ✓ Updated user: %s", user.Email)
			}
		}
	}

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
		} else {
			// Update description if needed
			if existing.Description != cat.Description {
				existing.Description = cat.Description
				db.Save(&existing)
				log.Printf("  ✓ Updated article category description: %s", cat.Name)
			}
		}
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
		log.Printf("⚠️ Failed to find admin user for articles: %v", err)
		// Fallback to first available user or create one if desperate logic needed,
		// but since we just seeded it, it should be there.
		// Let's try to fetch ANY user
		db.First(&adminUser)
	}

	articles := []models.Article{
		{
			Title:             "Mengenal Kecemasan dan Cara Mengatasinya",
			Thumbnail:         "/images/dummy-article-1.png",
			Content:           "Kecemasan adalah respons alami tubuh terhadap stres. Ini adalah perasaan takut atau khawatir tentang apa yang akan datang.\n\nCara Mengatasi Kecemasan:\n1. Latihan pernapasan dalam\n2. Meditasi teratur\n3. Olahraga rutin\n4. Tidur yang cukup\n5. Mengurangi kafein",
			ArticleCategoryID: healthCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "5 Teknik Pernapasan untuk Menenangkan Pikiran",
			Thumbnail:         "/images/dummy-article-2.png",
			Content:           "Pernapasan yang tepat dapat membantu menenangkan sistem saraf.\n\n1. Teknik 4-7-8\nTarik napas selama 4 detik, tahan 7 detik, hembuskan 8 detik.\n\n2. Pernapasan Kotak\nTarik napas 4 detik, tahan 4 detik, hembuskan 4 detik, tahan 4 detik.",
			ArticleCategoryID: tipsCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Panduan Meditasi untuk Pemula",
			Thumbnail:         "/images/dummy-article-3.png",
			Content:           "Meditasi tidak harus rumit. Mulailah dengan 5 menit sehari.\n\nLangkah-langkah:\n1. Duduk dengan nyaman\n2. Tutup mata\n3. Fokus pada napas\n4. Biarkan pikiran mengalir\n5. Kembalikan fokus ke napas",
			ArticleCategoryID: meditasiCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Mengatasi Stres di Tempat Kerja",
			Thumbnail:         "/images/dummy-article-4.png",
			Content:           "Stres kerja adalah masalah umum yang dapat mempengaruhi kesehatan mental.\n\nTips Mengatasi:\n1. Buat batasan yang jelas antara kerja dan kehidupan pribadi\n2. Ambil break secara teratur\n3. Prioritaskan tugas dengan baik\n4. Jangan takut untuk meminta bantuan",
			ArticleCategoryID: tipsCategory.ID,
			UserID:            adminUser.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Pentingnya Tidur untuk Kesehatan Mental",
			Thumbnail:         "/images/dummy-article-5.png",
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

	// Seed Song Categories
	log.Println("📝 Seeding song categories...")
	songCategories := []models.SongCategory{
		{Name: "Alam", Thumbnail: "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=400&h=400&fit=crop"},
		{Name: "Piano", Thumbnail: "https://images.unsplash.com/photo-1520523839897-bd0b52f945a0?w=400&h=400&fit=crop"},
		{Name: "Hujan", Thumbnail: "https://images.unsplash.com/photo-1515694346937-94d85e41e6f0?w=400&h=400&fit=crop"},
		{Name: "Laut", Thumbnail: "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=400&h=400&fit=crop"},
		{Name: "Meditasi", Thumbnail: "https://images.unsplash.com/photo-1506126613408-eca07ce68773?w=400&h=400&fit=crop"},
	}

	for _, cat := range songCategories {
		var existing models.SongCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			db.Create(&cat)
			log.Printf("  ✓ Created song category: %s", cat.Name)
		} else {
			// Update existing with new thumbnail
			db.Model(&existing).Update("thumbnail", cat.Thumbnail)
			log.Printf("  ✓ Updated song category: %s", cat.Name)
		}
	}

	// Seed Songs
	log.Println("📝 Seeding songs...")
	var alamCategory models.SongCategory
	db.Where("name = ?", "Alam").First(&alamCategory)

	var pianoCategory models.SongCategory
	db.Where("name = ?", "Piano").First(&pianoCategory)

	var rainCategory models.SongCategory
	db.Where("name = ?", "Hujan").First(&rainCategory)

	songs := []models.Song{
		{Title: "Forest Birds Morning", FilePath: "/audio/forest-birds.mp3", Thumbnail: "https://images.unsplash.com/photo-1448375240586-882707db888b?w=400&h=400&fit=crop", SongCategoryID: alamCategory.ID},
		{Title: "River Stream", FilePath: "/audio/river-stream.mp3", Thumbnail: "https://images.unsplash.com/photo-1433086966358-54859d0ed716?w=400&h=400&fit=crop", SongCategoryID: alamCategory.ID},
		{Title: "Peaceful Piano", FilePath: "/audio/peaceful-piano.mp3", Thumbnail: "https://images.unsplash.com/photo-1552422535-c45813c61732?w=400&h=400&fit=crop", SongCategoryID: pianoCategory.ID},
		{Title: "Soft Piano Melody", FilePath: "/audio/soft-piano.mp3", Thumbnail: "https://images.unsplash.com/photo-1512733596533-7b00ccf8ebaf?w=400&h=400&fit=crop", SongCategoryID: pianoCategory.ID},
		{Title: "Gentle Rain", FilePath: "/audio/gentle-rain.mp3", Thumbnail: "https://images.unsplash.com/photo-1428592953211-077101b2021b?w=400&h=400&fit=crop", SongCategoryID: rainCategory.ID},
		{Title: "Thunderstorm", FilePath: "/audio/thunderstorm.mp3", Thumbnail: "https://images.unsplash.com/photo-1605727216801-e27ce1d0cc28?w=400&h=400&fit=crop", SongCategoryID: rainCategory.ID},
	}

	for _, song := range songs {
		var existing models.Song
		if db.Where("title = ?", song.Title).First(&existing).RowsAffected == 0 {
			db.Create(&song)
			log.Printf("  ✓ Created song: %s", song.Title)
		} else {
			// Update existing with new thumbnail
			db.Model(&existing).Update("thumbnail", song.Thumbnail)
			log.Printf("  ✓ Updated song: %s", song.Title)
		}
	}

	fmt.Println("\n✅ Database seeding completed!")
	fmt.Println("\n📋 Test Accounts:")
	fmt.Println("   Admin: admin@ruangtenang.id / admin123")
	fmt.Println("   Member: john@example.com / member123")
}
