package presentation

import (
	"errors"
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

const incompetechLicenseURL = "https://creativecommons.org/licenses/by/4.0/"

var curatedPresentationSongTitles = []string{
	"Meditation Impromptu 01", "Meditation Impromptu 02", "Meditation Impromptu 03",
	"Friday Morning", "Starry", "Water Lily", "Ethereal Relaxation",
	"That Zen Moment", "Ever Mindful", "Organic Meditations Three",
}

// SeedSongs adds a small, curated music library. These are published
// compositions by Kevin MacLeod, not generated audio placeholders. Incompetech
// publishes the catalog under CC BY 4.0; each row carries its attribution and
// links so the clients can show the required credit alongside playback.
func SeedSongs(db *gorm.DB) error {
	catalog := []struct {
		Title     string
		Category  string
		ISRC      string
		Filename  string
		Thumbnail string
	}{
		{"Meditation Impromptu 01", "Piano", "USUAN1100163", "Meditation%20Impromptu%2001.mp3", "article-pause.webp"},
		{"Meditation Impromptu 02", "Piano", "USUAN1100162", "Meditation%20Impromptu%2002.mp3", "article-calm-start.webp"},
		{"Meditation Impromptu 03", "Piano", "USUAN1100161", "Meditation%20Impromptu%2003.mp3", "article-pause.webp"},
		{"Friday Morning", "Piano", "USUAN1100224", "Friday%20Morning.mp3", "article-calm-start.webp"},
		{"Starry", "Piano", "USUAN1100062", "Starry.mp3", "article-sleep-routine.webp"},
		{"Water Lily", "Piano", "USUAN1400035", "Water%20Lily.mp3", "story-community-support.webp"},
		{"Ethereal Relaxation", "Meditasi", "USUAN2100031", "Ethereal%20Relaxation.mp3", "story-community-support.webp"},
		{"That Zen Moment", "Meditasi", "USUAN2400001", "That%20Zen%20Moment.mp3", "article-pause.webp"},
		{"Ever Mindful", "Meditasi", "USUAN1700033", "Ever%20Mindful.mp3", "article-calm-start.webp"},
		{"Organic Meditations Three", "Meditasi", "USUAN1100759", "Organic%20Meditations%20Three.mp3", "article-sleep-routine.webp"},
	}

	// The former catalog used invented names and local generated-audio filenames.
	// Retire only those known seed records; a custom song with another title is
	// left intact. Soft deletion preserves old playlist references.
	legacyTitles := []string{
		"Forest Ambience", "Nature Soundscape", "Morning Chirps",
		"Peaceful Piano", "Soft Piano Melody", "Evening Piano",
		"Gentle Rain", "Thunderstorm Ambience", "Rain on Window",
		"Ocean Waves", "Beach Sunset", "Deep Sea",
		"Zen Meditation", "Tibetan Bowls", "Om Chanting",
		"Gentle ASMR", "Crisp Tingles", "Whispering Wind",
		"Classical Focus", "Study Mozart", "Relaxing Chopin",
		"Pure White Noise", "Pink Noise Sleep", "Brown Noise Calm",
	}
	if err := db.Where("title IN ? AND file_path LIKE ?", legacyTitles, "%gen_audio_%").Delete(&model.Song{}).Error; err != nil {
		return err
	}

	for _, track := range catalog {
		var category model.SongCategory
		if err := db.Where("name = ?", track.Category).First(&category).Error; err != nil {
			return fmt.Errorf("song category %q must be seeded before songs: %w", track.Category, err)
		}

		thumbnail := getSeedAsset(track.Thumbnail, "images")
		if thumbnail == "" {
			return fmt.Errorf("thumbnail not found for song %q (%s)", track.Title, track.Thumbnail)
		}

		sourceURL := fmt.Sprintf("https://incompetech.com/music/royalty-free/index.html?isrc=%s", track.ISRC)
		payload := model.Song{
			Title:          track.Title,
			FilePath:       "https://incompetech.com/music/royalty-free/mp3-royaltyfree/" + track.Filename,
			Attribution:    "Kevin MacLeod (incompetech.com)",
			SourceURL:      sourceURL,
			LicenseURL:     incompetechLicenseURL,
			Thumbnail:      thumbnail,
			SongCategoryID: category.ID,
		}

		var existing model.Song
		findResult := db.Where("source_url = ?", sourceURL).First(&existing)
		if findResult.Error == nil {
			if err := db.Model(&existing).Updates(map[string]any{
				"file_path":        payload.FilePath,
				"attribution":      payload.Attribution,
				"source_url":       payload.SourceURL,
				"license_url":      payload.LicenseURL,
				"thumbnail":        payload.Thumbnail,
				"song_category_id": payload.SongCategoryID,
			}).Error; err != nil {
				return err
			}
			continue
		}
		if !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			return findResult.Error
		}

		if err := db.Create(&payload).Error; err != nil {
			return err
		}
	}

	// Hide obsolete empty seed categories, while preserving any category still
	// used by another active song in an existing database.
	var categories []model.SongCategory
	if err := db.Find(&categories).Error; err != nil {
		return err
	}
	legacySeedCategories := []string{"Alam", "Hujan", "Laut", "White Noise", "ASMR", "Classical"}
	for _, category := range categories {
		if category.Name == "Piano" || category.Name == "Meditasi" {
			continue
		}
		isLegacySeedCategory := false
		for _, name := range legacySeedCategories {
			if category.Name == name {
				isLegacySeedCategory = true
				break
			}
		}
		if !isLegacySeedCategory {
			continue
		}
		var activeSongs int64
		if err := db.Model(&model.Song{}).Where("song_category_id = ?", category.ID).Count(&activeSongs).Error; err != nil {
			return err
		}
		if activeSongs == 0 {
			if err := db.Delete(&category).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
