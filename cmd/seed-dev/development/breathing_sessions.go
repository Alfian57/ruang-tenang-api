package development

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedBreathingSessions seeds breathing practice history for development
func SeedBreathingSessions(db *gorm.DB) error {
	var users []model.User
	if err := db.Where("role = ?", model.RoleMember).Order("id ASC").Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	// Get breathing techniques
	var techniques []model.BreathingTechnique
	if err := db.Where("is_system = ? AND is_active = ?", true, true).Find(&techniques).Error; err != nil {
		return err
	}
	if len(techniques) == 0 {
		return nil
	}

	moods := []string{"calm", "anxious", "stressed", "neutral", "happy"}

	type sessionData struct {
		UserIdx    int
		TechIdx    int
		Duration   int
		Target     int
		Cycles     int
		Completed  bool
		Percentage int
		XP         int
		DaysAgo    int
		HourOfDay  int
		MoodBefore string
		MoodAfter  string
	}

	sessions := []sessionData{
		// User 1 — more active practitioner
		{0, 0, 240, 240, 15, true, 100, 15, 10, 7, "anxious", "calm"},
		{0, 1, 300, 300, 10, true, 100, 20, 8, 22, "stressed", "calm"},
		{0, 0, 240, 240, 15, true, 100, 15, 6, 8, "neutral", "happy"},
		{0, 2, 180, 300, 9, false, 60, 8, 5, 12, "stressed", "neutral"},
		{0, 0, 240, 240, 15, true, 100, 15, 4, 7, "neutral", "calm"},
		{0, 3, 120, 180, 6, true, 100, 12, 3, 21, "anxious", "calm"},
		{0, 1, 300, 300, 10, true, 100, 20, 2, 22, "neutral", "happy"},
		{0, 0, 240, 240, 15, true, 100, 15, 1, 7, "calm", "happy"},
		{0, 4, 360, 360, 12, true, 100, 25, 0, 8, "neutral", "calm"},
		// User 2 — moderate
		{1, 0, 240, 240, 15, true, 100, 15, 9, 20, "stressed", "neutral"},
		{1, 1, 200, 300, 7, false, 67, 10, 7, 21, "anxious", "neutral"},
		{1, 0, 240, 240, 15, true, 100, 15, 5, 20, "neutral", "calm"},
		{1, 2, 300, 300, 15, true, 100, 20, 3, 19, "stressed", "calm"},
		{1, 0, 240, 240, 15, true, 100, 15, 1, 20, "neutral", "happy"},
		// User 3 — beginner
		{2, 0, 120, 240, 8, false, 50, 5, 4, 9, "neutral", "neutral"},
		{2, 0, 240, 240, 15, true, 100, 15, 2, 10, "anxious", "calm"},
	}

	// Check if already seeded
	var count int64
	db.Model(&model.BreathingSession{}).Where("user_id IN (?)", []uint{users[0].ID}).Count(&count)
	if count > 0 {
		return nil
	}

	for _, s := range sessions {
		if s.UserIdx >= len(users) || s.TechIdx >= len(techniques) {
			continue
		}
		user := users[s.UserIdx]
		technique := techniques[s.TechIdx]

		startedAt := time.Now().AddDate(0, 0, -s.DaysAgo).
			Truncate(24 * time.Hour).
			Add(time.Duration(s.HourOfDay) * time.Hour)
		endedAt := startedAt.Add(time.Duration(s.Duration) * time.Second)

		moodBefore := s.MoodBefore
		moodAfter := s.MoodAfter
		if moodBefore == "" {
			moodBefore = moods[s.DaysAgo%len(moods)]
		}
		if moodAfter == "" {
			moodAfter = "calm"
		}

		session := model.BreathingSession{
			ID:                    uuid.New(),
			UserID:                int(user.ID),
			TechniqueID:           technique.ID,
			DurationSeconds:       s.Duration,
			TargetDurationSeconds: s.Target,
			CyclesCompleted:       s.Cycles,
			Completed:             s.Completed,
			CompletedPercentage:   s.Percentage,
			StartedAt:             startedAt,
			EndedAt:               &endedAt,
			XPEarned:              s.XP,
			MoodBefore:            &moodBefore,
			MoodAfter:             &moodAfter,
		}

		if err := db.Create(&session).Error; err != nil {
			return err
		}
	}

	// Seed favorites for first user
	user1 := users[0]
	for i, tech := range techniques {
		if i >= 3 {
			break // Favorite just the first 3 techniques
		}
		fav := model.BreathingFavorite{
			UserID:      int(user1.ID),
			TechniqueID: tech.ID,
			SortOrder:   i,
		}
		if err := db.Create(&fav).Error; err != nil {
			return err
		}
	}

	return nil
}
