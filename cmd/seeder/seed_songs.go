package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedSongs(db *gorm.DB) {
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
}
