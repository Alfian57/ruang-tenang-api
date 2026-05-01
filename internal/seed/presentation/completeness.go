package presentation

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedPresentationCompleteness fills supporting tables that are easy to miss
// but make admin, notification, moderation, and daily-engagement pages useful.
func SeedPresentationCompleteness(db *gorm.DB) error {
	steps := []func(*gorm.DB) error{
		seedDailyTasks,
		seedUserActivities,
		seedJournalSettingsAndAccessLogs,
		seedBreathingPreferences,
		seedBroadcastsAndPushSubscriptions,
		seedModerationFixtures,
	}

	for _, step := range steps {
		if err := step(db); err != nil {
			return err
		}
	}

	return nil
}

func presentationUsers(db *gorm.DB) ([]model.User, error) {
	var users []model.User
	err := db.Where("role IN ?", []model.UserRole{model.RoleUser, model.RoleMitra}).
		Order("id ASC").
		Find(&users).Error
	return users, err
}

func presentationAdmin(db *gorm.DB) (*model.User, error) {
	var admin model.User
	if err := db.Where("role = ?", model.RoleAdmin).Order("id ASC").First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}

func seedDailyTasks(db *gorm.DB) error {
	users, err := presentationUsers(db)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	today := startOfDay(time.Now().UTC())
	configs := model.GetDailyTaskConfigsByTier(true)

	for userIdx, user := range users {
		for configIdx, config := range configs {
			if config.PremiumOnly && !user.IsPremium {
				continue
			}

			isCompleted := (userIdx+configIdx)%3 != 1
			currentCount := config.TargetCount
			var completedAt *time.Time
			var claimedAt *time.Time
			isClaimed := false

			if isCompleted {
				doneAt := today.Add(time.Duration(8+configIdx) * time.Hour)
				completedAt = &doneAt
				isClaimed = configIdx%2 == 0
				if isClaimed {
					claimTime := doneAt.Add(15 * time.Minute)
					claimedAt = &claimTime
				}
			} else if config.TargetCount > 1 {
				currentCount = config.TargetCount - 1
			} else {
				currentCount = 0
			}

			updates := map[string]interface{}{
				"target_count":  config.TargetCount,
				"current_count": currentCount,
				"is_completed":  isCompleted,
				"is_claimed":    isClaimed,
				"xp_reward":     config.XPReward,
				"coin_reward":   config.CoinReward,
				"completed_at":  completedAt,
				"claimed_at":    claimedAt,
			}

			var existing model.DailyTask
			err := db.Where("user_id = ? AND task_type = ? AND task_date = ?", user.ID, config.Type, today).
				First(&existing).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					task := model.DailyTask{
						UserID:       user.ID,
						TaskType:     config.Type,
						TaskDate:     today,
						TargetCount:  config.TargetCount,
						CurrentCount: currentCount,
						IsCompleted:  isCompleted,
						IsClaimed:    isClaimed,
						XPReward:     config.XPReward,
						CoinReward:   config.CoinReward,
						CompletedAt:  completedAt,
						ClaimedAt:    claimedAt,
					}
					if err := db.Create(&task).Error; err != nil {
						return err
					}
					continue
				}
				return err
			}

			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func seedUserActivities(db *gorm.DB) error {
	users, err := presentationUsers(db)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	activityTypes := []string{"login", "mood", "chat", "article", "journal", "breathing", "forum", "music"}
	today := startOfDay(time.Now().UTC())

	for userIdx, user := range users {
		for dayOffset := 0; dayOffset < 10; dayOffset++ {
			date := today.AddDate(0, 0, -dayOffset)
			for typeIdx, activityType := range activityTypes {
				count := ((userIdx + 1) * (typeIdx + 2) * (dayOffset + 1)) % 5
				if count == 0 && dayOffset < 4 {
					count = 1
				}

				var existing model.UserActivity
				err := db.Where("user_id = ? AND activity_type = ? AND date = ?", user.ID, activityType, date).
					First(&existing).Error
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						activity := model.UserActivity{
							UserID:       user.ID,
							ActivityType: activityType,
							Date:         date,
							Count:        count,
						}
						if err := db.Create(&activity).Error; err != nil {
							return err
						}
						continue
					}
					return err
				}

				if err := db.Model(&existing).Update("count", count).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func seedJournalSettingsAndAccessLogs(db *gorm.DB) error {
	users, err := presentationUsers(db)
	if err != nil {
		return err
	}

	for idx, user := range users {
		settings := model.JournalSettings{
			UserID:              user.ID,
			AllowAIAccess:       idx%2 == 0,
			AIContextDays:       14,
			AIContextMaxEntries: 6,
			DefaultShareWithAI:  idx%2 == 0,
			IsBlocked:           false,
		}
		if err := db.Where("user_id = ?", user.ID).
			Assign(settings).
			FirstOrCreate(&model.JournalSettings{}).Error; err != nil {
			return err
		}

		var journal model.Journal
		if err := db.Where("user_id = ?", user.ID).Order("created_at DESC").First(&journal).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}

		var session model.ChatSession
		sessionID := (*uint)(nil)
		if err := db.Where("user_id = ?", user.ID).Order("updated_at DESC").First(&session).Error; err == nil {
			sessionID = &session.ID
		}

		var existing model.JournalAIAccessLog
		err := db.Where("user_id = ? AND journal_id = ? AND context_type = ?", user.ID, journal.ID, "summary").
			First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logEntry := model.JournalAIAccessLog{
					UserID:        user.ID,
					JournalID:     journal.ID,
					ChatSessionID: sessionID,
					AccessedAt:    time.Now().UTC().Add(-2 * time.Hour),
					ContextType:   "summary",
				}
				if err := db.Create(&logEntry).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	return nil
}

func seedBreathingPreferences(db *gorm.DB) error {
	users, err := presentationUsers(db)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	var techniques []model.BreathingTechnique
	if err := db.Where("is_system = ? AND is_active = ?", true, true).Order("created_at ASC").Find(&techniques).Error; err != nil {
		return err
	}
	if len(techniques) == 0 {
		return nil
	}

	reminderTimes := []string{"07:30", "12:15", "20:00"}
	now := time.Now().UTC()
	today := startOfDay(now)

	for idx, user := range users {
		technique := techniques[idx%len(techniques)]
		reminder := reminderTimes[idx%len(reminderTimes)]
		lastPractice := today.AddDate(0, 0, -idx%4)
		dailyXPDate := today

		preference := model.BreathingPreference{
			UserID:                 int(user.ID),
			DefaultDurationSeconds: 240 + ((idx % 3) * 60),
			DefaultTechniqueID:     &technique.ID,
			VoiceGuidance:          "enabled",
			BackgroundSound:        "enabled",
			DefaultBackgroundSound: "rain",
			HapticFeedback:         true,
			AnimationSpeed:         "normal",
			Theme:                  "calm",
			ReminderEnabled:        idx%2 == 0,
			ReminderTime:           &reminder,
			ReminderDays:           "12345",
			TutorialCompleted:      true,
			CurrentStreak:          2 + idx,
			LongestStreak:          4 + idx,
			LastPracticeDate:       &lastPractice,
			StreakFreezeAvailable:  idx%2 == 0,
			DailyXPEarned:          15 + idx*5,
			DailyXPDate:            &dailyXPDate,
		}

		if err := db.Where("user_id = ?", user.ID).
			Assign(preference).
			FirstOrCreate(&model.BreathingPreference{}).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedBroadcastsAndPushSubscriptions(db *gorm.DB) error {
	admin, err := presentationAdmin(db)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	scheduledAt := now.Add(24 * time.Hour)
	sentAt := now.Add(-6 * time.Hour)

	broadcasts := []model.BroadcastNotification{
		{
			ID:          uuid.New(),
			Title:       "Pengingat check-in mood sore",
			Body:        "Luangkan 1 menit untuk mencatat mood dan lihat pola emosimu hari ini.",
			Icon:        "/logo.webp",
			URL:         "/dashboard/mood-tracker",
			Status:      model.BroadcastStatusScheduled,
			ScheduledAt: &scheduledAt,
			CreatedBy:   admin.ID,
		},
		{
			ID:        uuid.New(),
			Title:     "Materi presentasi siap dicoba",
			Body:      "Dashboard demo sudah berisi forum, artikel, musik, dan insight komunitas.",
			Icon:      "/logo.webp",
			URL:       "/dashboard",
			Status:    model.BroadcastStatusSent,
			SentAt:    &sentAt,
			SentCount: 24,
			CreatedBy: admin.ID,
		},
		{
			ID:        uuid.New(),
			Title:     "Draft kampanye premium",
			Body:      "Contoh draft broadcast untuk admin sebelum dikirim ke pengguna premium.",
			Icon:      "/logo.webp",
			URL:       "/dashboard/billing",
			Status:    model.BroadcastStatusDraft,
			CreatedBy: admin.ID,
		},
	}

	for _, broadcast := range broadcasts {
		var existing model.BroadcastNotification
		err := db.Where("title = ?", broadcast.Title).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&broadcast).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	users, err := presentationUsers(db)
	if err != nil {
		return err
	}
	for _, user := range users {
		endpoint := fmt.Sprintf("https://push.ruang-tenang.local/subscriptions/%d", user.ID)
		subscription := model.PushSubscription{
			ID:       uuid.New(),
			UserID:   user.ID,
			Endpoint: endpoint,
			P256dh:   fmt.Sprintf("demo-p256dh-key-%d", user.ID),
			Auth:     fmt.Sprintf("demo-auth-token-%d", user.ID),
		}

		var existing model.PushSubscription
		err := db.Where("endpoint = ?", endpoint).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&subscription).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	return nil
}

func seedModerationFixtures(db *gorm.DB) error {
	admin, err := presentationAdmin(db)
	if err != nil {
		return err
	}

	users, err := presentationUsers(db)
	if err != nil {
		return err
	}
	if len(users) < 2 {
		return nil
	}

	reporter := users[0]
	reported := users[1]
	now := time.Now().UTC()

	if err := seedUserBlock(db, reporter, reported); err != nil {
		return err
	}

	reportID, err := seedUserReport(db, reporter, reported, *admin, now)
	if err != nil {
		return err
	}

	if err := seedForumPostReport(db, reporter, *admin, now); err != nil {
		return err
	}

	if err := seedContentFlags(db, *admin, now); err != nil {
		return err
	}

	if err := seedUserStrike(db, reported, *admin, reportID, now); err != nil {
		return err
	}

	if err := seedModeratorActions(db, *admin, reported.ID, now); err != nil {
		return err
	}

	return seedAppealsIfTableExists(db, reported, *admin, now)
}

func seedUserBlock(db *gorm.DB, blocker, blocked model.User) error {
	block := model.UserBlock{
		BlockerID: blocker.ID,
		BlockedID: blocked.ID,
		Reason:    "Contoh relasi block untuk halaman moderasi.",
	}
	return db.Where("blocker_id = ? AND blocked_id = ?", blocker.ID, blocked.ID).
		FirstOrCreate(&block).Error
}

func seedUserReport(db *gorm.DB, reporter, reported, admin model.User, now time.Time) (*uint, error) {
	var post model.ForumPost
	var contentID *uint
	if err := db.Order("created_at DESC").First(&post).Error; err == nil {
		contentID = &post.ID
	}

	reportedUserID := reported.ID
	handledAt := now.Add(-90 * time.Minute)
	report := model.UserReport{
		ReporterID:        reporter.ID,
		ReportType:        model.ReportTypeForumPost,
		ReportedContentID: contentID,
		ReportedUserID:    &reportedUserID,
		Reason:            model.ReportReasonHarassment,
		Description:       "Contoh laporan untuk simulasi antrean moderasi saat presentasi.",
		Status:            model.ReportStatusReviewing,
		HandledByID:       &admin.ID,
		HandledAt:         &handledAt,
		ActionTaken:       model.ActionTakenUserWarned,
		ModeratorNotes:    "Dibuat oleh seeder presentasi.",
	}

	var existing model.UserReport
	err := db.Where("reporter_id = ? AND report_type = ? AND reported_user_id = ?", reporter.ID, report.ReportType, reported.ID).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&report).Error; err != nil {
				return nil, err
			}
			return &report.ID, nil
		}
		return nil, err
	}

	return &existing.ID, nil
}

func seedForumPostReport(db *gorm.DB, reporter, admin model.User, now time.Time) error {
	var post model.ForumPost
	if err := db.Order("created_at DESC").First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	reviewedAt := now.Add(-45 * time.Minute)
	report := model.ForumPostReport{
		PostID:         post.ID,
		ReporterID:     reporter.ID,
		Reason:         model.PostReportReasonHarassment,
		Description:    "Balasan ini perlu ditinjau karena berpotensi tidak suportif.",
		Status:         model.PostReportStatusReviewed,
		ReviewedBy:     &admin.ID,
		ReviewedAt:     &reviewedAt,
		ModeratorNotes: "Sudah ditinjau untuk demo moderation queue.",
	}

	var existing model.ForumPostReport
	err := db.Where("post_id = ? AND reporter_id = ?", post.ID, reporter.ID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&report).Error
		}
		return err
	}

	return nil
}

func seedContentFlags(db *gorm.DB, admin model.User, now time.Time) error {
	confidence := 0.86
	resolvedAt := now.Add(-30 * time.Minute)

	var article model.Article
	if err := db.Order("created_at DESC").First(&article).Error; err == nil {
		flag := model.ContentFlag{
			ContentType:     "article",
			ContentID:       article.ID,
			FlagType:        model.ContentFlagTypeAIModeration,
			FlagCategory:    model.ContentFlagCategoryTrauma,
			Severity:        model.FlagSeverityMedium,
			AIConfidence:    &confidence,
			AIReason:        "Artikel membahas pengalaman stres dan perlu label konteks.",
			FlaggedByID:     &admin.ID,
			IsResolved:      true,
			ResolvedByID:    &admin.ID,
			ResolvedAt:      &resolvedAt,
			ResolutionNotes: "Trigger warning sudah diverifikasi untuk demo.",
		}
		if err := upsertContentFlag(db, flag); err != nil {
			return err
		}
	}

	var post model.ForumPost
	if err := db.Order("created_at DESC").First(&post).Error; err == nil {
		flag := model.ContentFlag{
			ContentType:  "forum_post",
			ContentID:    post.ID,
			FlagType:     model.ContentFlagTypeManual,
			FlagCategory: model.ContentFlagCategoryAbuse,
			Severity:     model.FlagSeverityLow,
			FlaggedByID:  &admin.ID,
			IsResolved:   false,
			AIReason:     "Contoh flag manual yang masih menunggu keputusan moderator.",
		}
		if err := upsertContentFlag(db, flag); err != nil {
			return err
		}
	}

	return nil
}

func upsertContentFlag(db *gorm.DB, flag model.ContentFlag) error {
	var existing model.ContentFlag
	err := db.Where("content_type = ? AND content_id = ? AND flag_category = ?", flag.ContentType, flag.ContentID, flag.FlagCategory).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&flag).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"flag_type":        flag.FlagType,
		"severity":         flag.Severity,
		"ai_confidence":    flag.AIConfidence,
		"ai_reason":        flag.AIReason,
		"flagged_by_id":    flag.FlaggedByID,
		"is_resolved":      flag.IsResolved,
		"resolved_by_id":   flag.ResolvedByID,
		"resolved_at":      flag.ResolvedAt,
		"resolution_notes": flag.ResolutionNotes,
	}).Error
}

func seedUserStrike(db *gorm.DB, user, admin model.User, reportID *uint, now time.Time) error {
	expiresAt := now.AddDate(0, 1, 0)
	strike := model.UserStrike{
		UserID:     user.ID,
		ReportID:   reportID,
		Reason:     "Peringatan demo akibat komentar yang kurang suportif.",
		Severity:   model.StrikeSeverityWarning,
		IssuedByID: admin.ID,
		ExpiresAt:  &expiresAt,
		IsActive:   true,
		Notes:      "Seeded untuk memperlihatkan riwayat enforcement.",
	}

	var existing model.UserStrike
	err := db.Where("user_id = ? AND reason = ?", user.ID, strike.Reason).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&strike).Error
		}
		return err
	}

	return nil
}

func seedModeratorActions(db *gorm.DB, admin model.User, targetUserID uint, now time.Time) error {
	actions := []model.ModeratorAction{
		{
			ModeratorID:   admin.ID,
			ActionType:    model.ModeratorActionReportResolved,
			TargetType:    "report",
			TargetID:      targetUserID,
			PreviousState: `{"status":"pending"}`,
			NewState:      `{"status":"reviewing","action":"user_warned"}`,
			Reason:        "Menandai laporan sebagai sedang ditindaklanjuti.",
			Notes:         "Seeder presentasi.",
			CreatedAt:     now.Add(-75 * time.Minute),
		},
		{
			ModeratorID:   admin.ID,
			ActionType:    model.ModeratorActionUserWarned,
			TargetType:    "user",
			TargetID:      targetUserID,
			PreviousState: `{"strikes":0}`,
			NewState:      `{"strikes":1}`,
			Reason:        "Memberikan peringatan ringan.",
			Notes:         "Seeder presentasi.",
			CreatedAt:     now.Add(-65 * time.Minute),
		},
	}

	for _, action := range actions {
		var existing model.ModeratorAction
		err := db.Where("moderator_id = ? AND action_type = ? AND target_type = ? AND target_id = ?", action.ModeratorID, action.ActionType, action.TargetType, action.TargetID).
			First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&action).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	return nil
}

func seedAppealsIfTableExists(db *gorm.DB, user, admin model.User, now time.Time) error {
	if !db.Migrator().HasTable(&model.Appeal{}) {
		return nil
	}

	appeals := []model.Appeal{
		{
			UserID:   user.ID,
			Reason:   "Saya ingin meminta peninjauan ulang peringatan akun.",
			Evidence: "Percakapan lanjutan menunjukkan konteks yang lebih lengkap.",
			Status:   model.AppealStatusPending,
		},
		{
			UserID:        user.ID,
			Reason:        "Contoh appeal yang sudah selesai ditinjau.",
			Evidence:      "Riwayat perilaku setelah peringatan sudah membaik.",
			Status:        model.AppealStatusApproved,
			ReviewerNotes: "Disetujui untuk kebutuhan demo alur appeal.",
			ReviewerID:    &admin.ID,
			UpdatedAt:     now.Add(-20 * time.Minute),
		},
	}

	for _, appeal := range appeals {
		var existing model.Appeal
		err := db.Where("user_id = ? AND reason = ?", appeal.UserID, appeal.Reason).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&appeal).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	return nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
