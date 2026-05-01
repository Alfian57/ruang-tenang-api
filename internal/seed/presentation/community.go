package presentation

import (
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedCommunityData seeds community data (exp history, badges, features, stats, hall of fame)
// so the /dashboard/community page looks populated for judges.
func SeedCommunityData(db *gorm.DB) error {
	// Get test member users
	var users []model.User
	if err := db.Where("role = ?", model.RoleUser).Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	// 1. Update user gamification stats
	if err := seedUserGamificationStats(db, users); err != nil {
		return err
	}

	// 2. Seed exp histories
	if err := seedExpHistories(db, users); err != nil {
		return err
	}

	// 3. Seed user badges
	if err := seedUserBadges(db, users); err != nil {
		return err
	}

	// 4. Seed user feature unlocks
	if err := seedUserFeatureUnlocks(db, users); err != nil {
		return err
	}

	// 5. Seed community stats
	if err := seedCommunityStats(db); err != nil {
		return err
	}

	// 6. Seed monthly hall of fame
	if err := seedMonthlyHallOfFame(db, users); err != nil {
		return err
	}

	return nil
}

func seedUserGamificationStats(db *gorm.DB, users []model.User) error {
	// Give each user progressively higher stats
	statsData := []struct {
		Exp             int64
		GoldCoins       int64
		CurrentStreak   int
		LongestStreak   int
		TotalActivities int
		LoginStreak     int
	}{
		{Exp: 1500, GoldCoins: 320, CurrentStreak: 7, LongestStreak: 14, TotalActivities: 45, LoginStreak: 7},
		{Exp: 850, GoldCoins: 180, CurrentStreak: 5, LongestStreak: 10, TotalActivities: 30, LoginStreak: 5},
		{Exp: 550, GoldCoins: 100, CurrentStreak: 3, LongestStreak: 7, TotalActivities: 18, LoginStreak: 3},
	}

	now := time.Now()
	for i, user := range users {
		idx := i
		if idx >= len(statsData) {
			idx = len(statsData) - 1
		}
		s := statsData[idx]

		if err := db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"exp":                s.Exp,
			"gold_coins":         s.GoldCoins,
			"current_streak":     s.CurrentStreak,
			"longest_streak":     s.LongestStreak,
			"total_activities":   s.TotalActivities,
			"login_streak":       s.LoginStreak,
			"last_activity_date": now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedExpHistories(db *gorm.DB, users []model.User) error {
	activityTypes := []struct {
		Type        string
		Description string
		MinPoints   int
		MaxPoints   int
	}{
		{"chat_ai", "Melakukan chat dengan AI", 5, 15},
		{"breathing", "Latihan pernapasan", 10, 20},
		{"forum_comment", "Berkomentar di forum", 5, 10},
		{"upload_article", "Mengunggah artikel", 15, 25},
		{"story_approved", "Story disetujui moderator", 20, 30},
		{"heart_received", "Menerima hati pada story", 3, 5},
		{"post_upvote_given", "Menerima upvote", 2, 5},
	}

	for _, user := range users {
		// Check if already seeded
		var count int64
		db.Model(&model.ExpHistory{}).Where("user_id = ?", user.ID).Count(&count)
		if count > 0 {
			continue
		}

		// Create exp history for the last 30 days
		for day := 0; day < 30; day++ {
			// 2-4 activities per day
			numActivities := rand.Intn(3) + 2
			for a := 0; a < numActivities; a++ {
				act := activityTypes[rand.Intn(len(activityTypes))]
				points := rand.Intn(act.MaxPoints-act.MinPoints+1) + act.MinPoints

				baseDate := time.Now().UTC().AddDate(0, 0, -day).Truncate(24 * time.Hour)
				randomHour := time.Duration(rand.Intn(14)+7) * time.Hour // 7am-9pm
				actTime := baseDate.Add(randomHour)

				history := model.ExpHistory{
					UserID:       user.ID,
					ActivityType: act.Type,
					Points:       points,
					Description:  act.Description,
					CreatedAt:    actTime,
				}
				if err := db.Create(&history).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func seedUserBadges(db *gorm.DB, users []model.User) error {
	// Badge keys to assign per user index (first user gets more badges)
	badgeAssignments := [][]string{
		// User 1 (most active): engagement + contribution + milestone badges
		{"streak_7", "streak_14", "activities_10", "activities_50", "xp_1000", "first_article", "first_story"},
		// User 2: engagement + some milestones
		{"streak_7", "activities_10", "activities_50", "first_story"},
		// User 3: starter badges
		{"streak_7", "activities_10"},
	}

	for i, user := range users {
		// Check if already seeded
		var count int64
		db.Model(&model.UserBadge{}).Where("user_id = ?", user.ID).Count(&count)
		if count > 0 {
			continue
		}

		idx := i
		if idx >= len(badgeAssignments) {
			idx = len(badgeAssignments) - 1
		}

		for j, badgeKey := range badgeAssignments[idx] {
			var badge model.BadgeDefinition
			if err := db.Where("badge_key = ?", badgeKey).First(&badge).Error; err != nil {
				continue // badge definition doesn't exist, skip
			}

			earnedAt := time.Now().AddDate(0, 0, -(30 - j*3)) // staggered earning dates

			userBadge := model.UserBadge{
				UserID:      user.ID,
				BadgeID:     badge.ID,
				EarnedAt:    earnedAt,
				IsShowcased: j < 3, // showcase the first 3 badges
			}
			if err := db.Create(&userBadge).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedUserFeatureUnlocks(db *gorm.DB, users []model.User) error {
	for _, user := range users {
		// Check if already seeded
		var count int64
		db.Model(&model.UserFeatureUnlock{}).Where("user_id = ?", user.ID).Count(&count)
		if count > 0 {
			continue
		}

		// Determine user level from their exp
		var levelConfig model.LevelConfig
		db.Where("min_exp <= ?", user.Exp).Order("level DESC").First(&levelConfig)
		userLevel := levelConfig.Level
		if userLevel == 0 {
			userLevel = 1
		}

		// Unlock all features up to their level
		var features []model.FeatureDefinition
		if err := db.Where("required_level <= ? AND is_active = ?", userLevel, true).Find(&features).Error; err != nil {
			return err
		}

		for j, feature := range features {
			unlockedAt := time.Now().AddDate(0, 0, -(30 - j))
			unlock := model.UserFeatureUnlock{
				UserID:     user.ID,
				FeatureID:  feature.ID,
				UnlockedAt: unlockedAt,
			}
			if err := db.Create(&unlock).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedCommunityStats(db *gorm.DB) error {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	// Check if already seeded for this month
	var existing model.CommunityStats
	if db.Where("month = ? AND year = ?", month, year).First(&existing).RowsAffected > 0 {
		return nil
	}

	stats := model.CommunityStats{
		Month:                  month,
		Year:                   year,
		TotalXPEarned:          4850,
		ActiveMembers:          3,
		TotalAchievements:      13,
		NewMembers:             2,
		TotalStoriesPublished:  5,
		TotalArticlesPublished: 8,
	}

	return db.Create(&stats).Error
}

func seedMonthlyHallOfFame(db *gorm.DB, users []model.User) error {
	now := time.Now()
	month := int(now.Month())
	year := now.Year()

	// Check if already seeded
	var count int64
	db.Model(&model.MonthlyHallOfFame{}).Where("month = ? AND year = ?", month, year).Count(&count)
	if count > 0 {
		return nil
	}

	messages := []string{
		"Terus semangat menjalani hidup!",
		"Setiap langkah kecil adalah kemajuan.",
		"Bersama kita bisa lebih kuat.",
	}

	// Create hall of fame entries for up to 3 users
	for i, user := range users {
		if i >= 3 {
			break
		}

		// Determine user level
		var levelConfig model.LevelConfig
		db.Where("min_exp <= ?", user.Exp).Order("level DESC").First(&levelConfig)
		level := levelConfig.Level
		if level == 0 {
			level = 1
		}

		entry := model.MonthlyHallOfFame{
			UserID:    user.ID,
			Level:     level,
			Month:     month,
			Year:      year,
			Rank:      i + 1,
			MonthlyXP: int(user.Exp) - (i * 200),
			Message:   messages[i],
		}

		if err := db.Create(&entry).Error; err != nil {
			return err
		}
	}

	return nil
}
