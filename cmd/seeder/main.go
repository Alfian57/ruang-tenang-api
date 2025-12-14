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

	log.Println("🌱 Starting database seeder...")

	// Seed Users
	log.Println("📝 Seeding users...")
	adminPassword, _ := utils.HashPassword("admin123")
	memberPassword, _ := utils.HashPassword("member123")

	users := []models.User{
		{Name: "Admin", Email: "admin@ruangtenang.id", Password: adminPassword, Role: models.RoleAdmin},
		{Name: "John Doe", Email: "john@example.com", Password: memberPassword, Role: models.RoleMember},
		{Name: "Jane Smith", Email: "jane@example.com", Password: memberPassword, Role: models.RoleMember},
	}

	for _, user := range users {
		var existing models.User
		if db.Where("email = ?", user.Email).First(&existing).RowsAffected == 0 {
			db.Create(&user)
			log.Printf("  ✓ Created user: %s", user.Email)
		}
	}

	// Seed Article Categories
	log.Println("📝 Seeding article categories...")
	articleCategories := []models.ArticleCategory{
		{Name: "Kesehatan Mental"},
		{Name: "Tips & Trik"},
		{Name: "Meditasi"},
		{Name: "Motivasi"},
		{Name: "Mindfulness"},
	}

	for _, cat := range articleCategories {
		var existing models.ArticleCategory
		if db.Where("name = ?", cat.Name).First(&existing).RowsAffected == 0 {
			db.Create(&cat)
			log.Printf("  ✓ Created article category: %s", cat.Name)
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

	articles := []models.Article{
		{
			Title:             "Mengenal Kecemasan dan Cara Mengatasinya",
			Thumbnail:         "/images/dummy-article-1.png",
			Content:           `<h2>Apa itu Kecemasan?</h2><p>Kecemasan adalah respons alami tubuh terhadap stres. Ini adalah perasaan takut atau khawatir tentang apa yang akan datang.</p><h2>Cara Mengatasi Kecemasan</h2><p>1. Latihan pernapasan dalam</p><p>2. Meditasi teratur</p><p>3. Olahraga rutin</p><p>4. Tidur yang cukup</p><p>5. Mengurangi kafein</p>`,
			ArticleCategoryID: healthCategory.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "5 Teknik Pernapasan untuk Menenangkan Pikiran",
			Thumbnail:         "/images/dummy-article-2.png",
			Content:           `<h2>Teknik Pernapasan</h2><p>Pernapasan yang tepat dapat membantu menenangkan sistem saraf.</p><h3>1. Teknik 4-7-8</h3><p>Tarik napas selama 4 detik, tahan 7 detik, hembuskan 8 detik.</p><h3>2. Pernapasan Kotak</h3><p>Tarik napas 4 detik, tahan 4 detik, hembuskan 4 detik, tahan 4 detik.</p>`,
			ArticleCategoryID: tipsCategory.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Panduan Meditasi untuk Pemula",
			Thumbnail:         "/images/dummy-article-3.png",
			Content:           `<h2>Memulai Meditasi</h2><p>Meditasi tidak harus rumit. Mulailah dengan 5 menit sehari.</p><h3>Langkah-langkah:</h3><p>1. Duduk dengan nyaman</p><p>2. Tutup mata</p><p>3. Fokus pada napas</p><p>4. Biarkan pikiran mengalir</p><p>5. Kembalikan fokus ke napas</p>`,
			ArticleCategoryID: meditasiCategory.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Mengatasi Stres di Tempat Kerja",
			Thumbnail:         "/images/dummy-article-4.png",
			Content:           `<h2>Stres di Tempat Kerja</h2><p>Stres kerja adalah masalah umum yang dapat mempengaruhi kesehatan mental.</p><h3>Tips Mengatasi:</h3><p>1. Buat batasan yang jelas antara kerja dan kehidupan pribadi</p><p>2. Ambil break secara teratur</p><p>3. Prioritaskan tugas dengan baik</p><p>4. Jangan takut untuk meminta bantuan</p>`,
			ArticleCategoryID: tipsCategory.ID,
			Status:            models.ArticleStatusPublished,
		},
		{
			Title:             "Pentingnya Tidur untuk Kesehatan Mental",
			Thumbnail:         "/images/dummy-article-5.png",
			Content:           `<h2>Tidur dan Kesehatan Mental</h2><p>Tidur yang cukup sangat penting untuk menjaga kesehatan mental.</p><h3>Manfaat Tidur yang Cukup:</h3><p>1. Meningkatkan konsentrasi</p><p>2. Memperbaiki mood</p><p>3. Mengurangi risiko depresi</p><p>4. Meningkatkan daya ingat</p>`,
			ArticleCategoryID: healthCategory.ID,
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
