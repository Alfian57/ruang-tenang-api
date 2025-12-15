package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/config"
	"github.com/Alfian57/ruang-tenang-api/internal/database"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
)

// copyFile copies a file from src to dst, creating directories as needed
func copyFile(src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyAsset copies an asset from assets/ to uploads/ and returns the URL path
func copyAsset(assetPath, subDir string) string {
	// Generate unique filename with timestamp
	ext := filepath.Ext(assetPath)
	baseName := filepath.Base(assetPath)
	baseName = baseName[:len(baseName)-len(ext)]
	timestamp := time.Now().UnixNano()
	newFileName := fmt.Sprintf("%s_%d%s", baseName, timestamp, ext)

	dstPath := filepath.Join("uploads", subDir, newFileName)
	if err := copyFile(assetPath, dstPath); err != nil {
		log.Printf("  ⚠️ Failed to copy %s: %v", assetPath, err)
		return ""
	}

	// Return URL path (relative to server root)
	return fmt.Sprintf("/uploads/%s/%s", subDir, newFileName)
}

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
		{Name: "Alfian Gading Saputra", Email: "alfian@gmail.com", Password: memberPassword, Role: models.RoleMember, Exp: 1200},
		{Name: "Dery Wahyu", Email: "dery@gmail.com", Password: memberPassword, Role: models.RoleMember, Exp: 2300},
		{Name: "Andhika Khusna", Email: "andhika@gmail.com", Password: memberPassword, Role: models.RoleMember, Exp: 450},
	}

	for _, user := range users {
		var existing models.User
		result := db.Where("email = ?", user.Email).First(&existing)

		if result.RowsAffected == 0 {
			if err := db.Create(&user).Error; err != nil {
				log.Printf("  ❌ Failed to create user %s: %v", user.Email, err)
			} else {
				log.Printf("  ✓ Created user: %s", user.Email)
			}
		} else {
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
		db.First(&adminUser)
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

	// Copy song category images to uploads
	log.Println("📷 Copying song category images to uploads...")
	categoryImages := map[string]string{
		"Alam":     copyAsset("assets/images/category-alam.jpg", "images"),
		"Piano":    copyAsset("assets/images/category-piano.jpg", "images"),
		"Hujan":    copyAsset("assets/images/category-hujan.jpg", "images"),
		"Laut":     copyAsset("assets/images/category-laut.jpg", "images"),
		"Meditasi": copyAsset("assets/images/category-meditasi.jpg", "images"),
	}

	// Seed Song Categories
	log.Println("📝 Seeding song categories...")
	for name, thumbnail := range categoryImages {
		var existing models.SongCategory
		if db.Where("name = ?", name).First(&existing).RowsAffected == 0 {
			cat := models.SongCategory{Name: name, Thumbnail: thumbnail}
			db.Create(&cat)
			log.Printf("  ✓ Created song category: %s", name)
		} else {
			db.Model(&existing).Update("thumbnail", thumbnail)
			log.Printf("  ✓ Updated song category: %s", name)
		}
	}

	// Copy song assets to uploads
	log.Println("📷 Copying song assets to uploads...")
	songAssets := []struct {
		title    string
		image    string
		audio    string
		category string
	}{
		{"Forest Birds Morning", "assets/images/song-forest.jpg", "assets/audio/song-1.mp3", "Alam"},
		{"River Stream", "assets/images/song-river.jpg", "assets/audio/song-2.mp3", "Alam"},
		{"Peaceful Piano", "assets/images/song-piano.jpg", "assets/audio/song-3.mp3", "Piano"},
		{"Soft Piano Melody", "assets/images/song-soft-piano.jpg", "assets/audio/song-4.mp3", "Piano"},
		{"Gentle Rain", "assets/images/song-rain.jpg", "assets/audio/song-5.mp3", "Hujan"},
		{"Thunderstorm", "assets/images/song-thunder.jpg", "assets/audio/song-6.mp3", "Hujan"},
	}

	// Seed Songs
	log.Println("📝 Seeding songs...")
	for _, sa := range songAssets {
		var category models.SongCategory
		db.Where("name = ?", sa.category).First(&category)

		thumbnail := copyAsset(sa.image, "images")
		filePath := copyAsset(sa.audio, "audio")

		var existing models.Song
		if db.Where("title = ?", sa.title).First(&existing).RowsAffected == 0 {
			song := models.Song{
				Title:          sa.title,
				FilePath:       filePath,
				Thumbnail:      thumbnail,
				SongCategoryID: category.ID,
			}
			db.Create(&song)
			log.Printf("  ✓ Created song: %s", sa.title)
		} else {
			db.Model(&existing).Updates(map[string]interface{}{
				"thumbnail": thumbnail,
				"file_path": filePath,
			})
			log.Printf("  ✓ Updated song: %s", sa.title)
		}
	}

	fmt.Println("\n✅ Database seeding completed!")
	fmt.Println("\n📋 Test Accounts:")
	fmt.Println("   Admin: admin@ruangtenang.id / admin123")
	fmt.Println("   Member: john@example.com / member123")
}
