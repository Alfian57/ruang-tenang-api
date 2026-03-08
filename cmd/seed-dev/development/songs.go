package development

import (
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedSongs seeds test songs for development
func SeedSongs(db *gorm.DB) error {
	// First, update song categories with thumbnails
	categoryThumbnails := map[string]string{
		"Alam":     "category-alam.jpg",
		"Piano":    "category-piano.jpg",
		"Hujan":    "category-hujan.jpg",
		"Laut":     "category-laut.jpg",
		"Meditasi": "category-meditasi.jpg",
	}

	for catName, imgName := range categoryThumbnails {
		var category model.SongCategory
		if db.Where("name = ?", catName).First(&category).RowsAffected > 0 {
			thumbnail := getSeedAsset(imgName, "images")
			if thumbnail != "" {
				db.Model(&category).Update("thumbnail", thumbnail)
			}
		}
	}

	// Get categories
	var alamCat, pianoCat, hujanCat, lautCat, meditasiCat model.SongCategory
	db.Where("name = ?", "Alam").First(&alamCat)
	db.Where("name = ?", "Piano").First(&pianoCat)
	db.Where("name = ?", "Hujan").First(&hujanCat)
	db.Where("name = ?", "Laut").First(&lautCat)
	db.Where("name = ?", "Meditasi").First(&meditasiCat)

	songs := []struct {
		Title      string
		CategoryID uint
		Image      string
		Audio      string
	}{
		{Title: "Forest Birds Morning", CategoryID: alamCat.ID, Image: "song-forest.jpg", Audio: "song-1.mp3"},
		{Title: "River Stream", CategoryID: alamCat.ID, Image: "song-river.jpg", Audio: "song-2.mp3"},
		{Title: "Mountain Wind", CategoryID: alamCat.ID, Image: "song-forest.jpg", Audio: "song-3.mp3"},
		{Title: "Peaceful Piano", CategoryID: pianoCat.ID, Image: "song-piano.jpg", Audio: "song-4.mp3"},
		{Title: "Soft Piano Melody", CategoryID: pianoCat.ID, Image: "song-soft-piano.jpg", Audio: "song-5.mp3"},
		{Title: "Evening Piano", CategoryID: pianoCat.ID, Image: "song-piano.jpg", Audio: "song-6.mp3"},
		{Title: "Gentle Rain", CategoryID: hujanCat.ID, Image: "song-rain.jpg", Audio: "song-1.mp3"},
		{Title: "Thunderstorm Ambience", CategoryID: hujanCat.ID, Image: "song-thunder.jpg", Audio: "song-2.mp3"},
		{Title: "Rain on Window", CategoryID: hujanCat.ID, Image: "song-rain.jpg", Audio: "song-3.mp3"},
		{Title: "Ocean Waves", CategoryID: lautCat.ID, Image: "category-laut.jpg", Audio: "song-4.mp3"},
		{Title: "Beach Sunset", CategoryID: lautCat.ID, Image: "category-laut.jpg", Audio: "song-5.mp3"},
		{Title: "Deep Sea", CategoryID: lautCat.ID, Image: "category-laut.jpg", Audio: "song-6.mp3"},
		{Title: "Zen Meditation", CategoryID: meditasiCat.ID, Image: "category-meditasi.jpg", Audio: "song-1.mp3"},
		{Title: "Tibetan Bowls", CategoryID: meditasiCat.ID, Image: "category-meditasi.jpg", Audio: "song-2.mp3"},
		{Title: "Om Chanting", CategoryID: meditasiCat.ID, Image: "category-meditasi.jpg", Audio: "song-3.mp3"},
	}

	for _, s := range songs {
		thumbnail := getSeedAsset(s.Image, "images")
		if thumbnail == "" {
			return fmt.Errorf("thumbnail not found for song %q (%s)", s.Title, s.Image)
		}

		audioPath := getSeedAudio(s.Audio)
		if audioPath == "" {
			return fmt.Errorf("audio not found for song %q (%s)", s.Title, s.Audio)
		}

		payload := model.Song{
			Title:          s.Title,
			FilePath:       audioPath,
			Thumbnail:      thumbnail,
			SongCategoryID: s.CategoryID,
		}

		var existing model.Song
		if db.Where("title = ?", s.Title).First(&existing).RowsAffected > 0 {
			existing.FilePath = payload.FilePath
			existing.Thumbnail = payload.Thumbnail
			existing.SongCategoryID = payload.SongCategoryID

			if err := db.Save(&existing).Error; err != nil {
				return err
			}

			continue
		}

		if err := db.Create(&payload).Error; err != nil {
			return err
		}
	}

	return nil
}
