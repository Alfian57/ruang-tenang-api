package development

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedSongs seeds test songs for development
func SeedSongs(db *gorm.DB) error {
	// First, update song categories with thumbnails
	categoryThumbnails := map[string]string{
		"Alam":     "cat-alam.jpg",
		"Piano":    "cat-piano.jpg",
		"Hujan":    "cat-hujan.jpg",
		"Laut":     "cat-laut.jpg",
		"Meditasi": "cat-meditasi.jpg",
	}

	for catName, imgName := range categoryThumbnails {
		var category model.SongCategory
		if db.Where("name = ?", catName).First(&category).RowsAffected > 0 {
			if url, ok := placeholderImages[imgName]; ok {
				thumbnail := getOrDownloadImage(url, imgName)
				if thumbnail != "" {
					db.Model(&category).Update("thumbnail", thumbnail)
				}
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

	// Note: For development, we use placeholder audio files
	// In production, you would have actual audio files
	songs := []struct {
		Title      string
		CategoryID uint
		Image      string
	}{
		{Title: "Forest Birds Morning", CategoryID: alamCat.ID, Image: "song-forest.jpg"},
		{Title: "River Stream", CategoryID: alamCat.ID, Image: "song-forest.jpg"},
		{Title: "Mountain Wind", CategoryID: alamCat.ID, Image: "song-forest.jpg"},
		{Title: "Peaceful Piano", CategoryID: pianoCat.ID, Image: "song-piano.jpg"},
		{Title: "Soft Piano Melody", CategoryID: pianoCat.ID, Image: "song-piano.jpg"},
		{Title: "Evening Piano", CategoryID: pianoCat.ID, Image: "song-piano.jpg"},
		{Title: "Gentle Rain", CategoryID: hujanCat.ID, Image: "song-rain.jpg"},
		{Title: "Thunderstorm Ambience", CategoryID: hujanCat.ID, Image: "song-rain.jpg"},
		{Title: "Rain on Window", CategoryID: hujanCat.ID, Image: "song-rain.jpg"},
		{Title: "Ocean Waves", CategoryID: lautCat.ID, Image: "song-ocean.jpg"},
		{Title: "Beach Sunset", CategoryID: lautCat.ID, Image: "song-ocean.jpg"},
		{Title: "Deep Sea", CategoryID: lautCat.ID, Image: "song-ocean.jpg"},
		{Title: "Zen Meditation", CategoryID: meditasiCat.ID, Image: "song-meditation.jpg"},
		{Title: "Tibetan Bowls", CategoryID: meditasiCat.ID, Image: "song-meditation.jpg"},
		{Title: "Om Chanting", CategoryID: meditasiCat.ID, Image: "song-meditation.jpg"},
	}

	for _, s := range songs {
		var existing model.Song
		if db.Where("title = ?", s.Title).First(&existing).RowsAffected > 0 {
			continue
		}

		// Get thumbnail
		thumbnail := ""
		if url, ok := placeholderImages[s.Image]; ok {
			thumbnail = getOrDownloadImage(url, s.Image)
		}

		// Get audio file
		audioPath := "/uploads/audio/placeholder.mp3"
		if url, ok := placeholderImages["song-placeholder.mp3"]; ok {
			downloadedPath := getOrDownloadAsset(url, "song-placeholder.mp3", "audio")
			if downloadedPath != "" {
				audioPath = downloadedPath
			}
		}

		song := model.Song{
			Title:          s.Title,
			FilePath:       audioPath,
			Thumbnail:      thumbnail,
			SongCategoryID: s.CategoryID,
		}

		if err := db.Create(&song).Error; err != nil {
			return err
		}
	}

	return nil
}
