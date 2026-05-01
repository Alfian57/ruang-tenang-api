package presentation

import (
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedSongs seeds test songs for development
func SeedSongs(db *gorm.DB) error {
	// First, update song categories with thumbnails
	categoryThumbnails := map[string]string{
		"Alam":        "category-alam.jpg",
		"Piano":       "category-piano.jpg",
		"Hujan":       "category-hujan.jpg",
		"Laut":        "category-laut.jpg",
		"Meditasi":    "category-meditasi.jpg",
		"ASMR":        "category-asmr.jpg",
		"Classical":   "category-classical.jpg",
		"White Noise": "category-white-noise.jpg",
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
	var alamCat, pianoCat, hujanCat, lautCat, meditasiCat, asmrCat, classicalCat, whiteNoiseCat model.SongCategory
	db.Where("name = ?", "Alam").First(&alamCat)
	db.Where("name = ?", "Piano").First(&pianoCat)
	db.Where("name = ?", "Hujan").First(&hujanCat)
	db.Where("name = ?", "Laut").First(&lautCat)
	db.Where("name = ?", "Meditasi").First(&meditasiCat)
	db.Where("name = ?", "ASMR").First(&asmrCat)
	db.Where("name = ?", "Classical").First(&classicalCat)
	db.Where("name = ?", "White Noise").First(&whiteNoiseCat)

	songs := []struct {
		Title      string
		CategoryID uint
		Image      string
		Audio      string
	}{
		{Title: "Forest Ambience", CategoryID: alamCat.ID, Image: "song-alam-1.png", Audio: "gen_audio_alam_1.wav"},
		{Title: "Nature Soundscape", CategoryID: alamCat.ID, Image: "song-alam-2.png", Audio: "gen_audio_alam_2.wav"},
		{Title: "Morning Chirps", CategoryID: alamCat.ID, Image: "song-alam-3.png", Audio: "gen_audio_alam_3.wav"},

		{Title: "Peaceful Piano", CategoryID: pianoCat.ID, Image: "song-piano-1.png", Audio: "gen_audio_piano_1.wav"},
		{Title: "Soft Piano Melody", CategoryID: pianoCat.ID, Image: "song-piano-2.png", Audio: "gen_audio_piano_2.wav"},
		{Title: "Evening Piano", CategoryID: pianoCat.ID, Image: "song-piano-3.png", Audio: "gen_audio_piano_3.wav"},

		{Title: "Gentle Rain", CategoryID: hujanCat.ID, Image: "song-hujan-1.png", Audio: "gen_audio_hujan_1.wav"},
		{Title: "Thunderstorm Ambience", CategoryID: hujanCat.ID, Image: "song-hujan-2.png", Audio: "gen_audio_hujan_2.wav"},
		{Title: "Rain on Window", CategoryID: hujanCat.ID, Image: "song-hujan-3.png", Audio: "gen_audio_hujan_3.wav"},

		{Title: "Ocean Waves", CategoryID: lautCat.ID, Image: "song-laut-1.png", Audio: "gen_audio_laut_1.wav"},
		{Title: "Beach Sunset", CategoryID: lautCat.ID, Image: "song-laut-2.png", Audio: "gen_audio_laut_2.wav"},
		{Title: "Deep Sea", CategoryID: lautCat.ID, Image: "song-laut-3.png", Audio: "gen_audio_laut_3.wav"},

		{Title: "Zen Meditation", CategoryID: meditasiCat.ID, Image: "song-meditasi-1.png", Audio: "gen_audio_meditasi_1.wav"},
		{Title: "Tibetan Bowls", CategoryID: meditasiCat.ID, Image: "song-meditasi-2.png", Audio: "gen_audio_meditasi_2.wav"},
		{Title: "Om Chanting", CategoryID: meditasiCat.ID, Image: "song-meditasi-3.png", Audio: "gen_audio_meditasi_3.wav"},

		{Title: "Gentle ASMR", CategoryID: asmrCat.ID, Image: "song-asmr-1.png", Audio: "gen_audio_asmr_1.wav"},
		{Title: "Crisp Tingles", CategoryID: asmrCat.ID, Image: "song-asmr-2.png", Audio: "gen_audio_asmr_2.wav"},
		{Title: "Whispering Wind", CategoryID: asmrCat.ID, Image: "song-asmr-3.png", Audio: "gen_audio_asmr_3.wav"},

		{Title: "Classical Focus", CategoryID: classicalCat.ID, Image: "song-classical-1.png", Audio: "gen_audio_classical_1.wav"},
		{Title: "Study Mozart", CategoryID: classicalCat.ID, Image: "song-classical-2.png", Audio: "gen_audio_classical_2.wav"},
		{Title: "Relaxing Chopin", CategoryID: classicalCat.ID, Image: "song-classical-3.png", Audio: "gen_audio_classical_3.wav"},

		{Title: "Pure White Noise", CategoryID: whiteNoiseCat.ID, Image: "song-white-noise-1.png", Audio: "gen_audio_white_noise_1.wav"},
		{Title: "Pink Noise Sleep", CategoryID: whiteNoiseCat.ID, Image: "song-white-noise-2.png", Audio: "gen_audio_white_noise_2.wav"},
		{Title: "Brown Noise Calm", CategoryID: whiteNoiseCat.ID, Image: "song-white-noise-3.png", Audio: "gen_audio_white_noise_3.wav"},
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
