package production

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedLeagueSeasons seeds the current active league season.
// Required for the weekly league feature to function.
func SeedLeagueSeasons(db *gorm.DB) error {
	now := time.Now().UTC()
	year, week := now.ISOWeek()

	// Calculate week boundaries (Monday to Sunday)
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -int(weekday-time.Monday)).Truncate(24 * time.Hour)
	sunday := monday.AddDate(0, 0, 7).Add(-time.Second)

	var existing model.LeagueSeason
	if db.Where("week_number = ? AND year = ?", week, year).First(&existing).RowsAffected > 0 {
		return nil
	}

	// Deactivate any previously active seasons
	db.Model(&model.LeagueSeason{}).Where("is_active = ?", true).Update("is_active", false)

	season := model.LeagueSeason{
		WeekNumber: week,
		Year:       year,
		StartsAt:   monday,
		EndsAt:     sunday,
		IsActive:   true,
	}

	return db.Create(&season).Error
}
