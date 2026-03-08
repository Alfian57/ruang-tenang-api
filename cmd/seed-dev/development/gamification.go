package development

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedGamification seeds gamification-related tables so all dashboard widgets
// and gamification pages look populated for demo/judging:
//   - league_participants, user_map_progress, user_landmark_progress,
//     user_society_memberships, user_timed_challenges, user_combos,
//     xp_boosts, user_chests, user_spins, notifications
func SeedGamification(db *gorm.DB) error {
	var users []model.User
	if err := db.Where("role = ?", model.RoleMember).Order("id ASC").Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	var admin model.User
	if err := db.Where("role = ?", model.RoleAdmin).First(&admin).Error; err != nil {
		return err
	}

	fns := []func(*gorm.DB, []model.User, model.User) error{
		seedLeagueParticipants,
		seedUserMapProgress,
		seedUserSocietyMemberships,
		seedUserTimedChallenges,
		seedUserCombosAndBoosts,
		seedUserChests,
		seedUserSpins,
		seedNotifications,
		seedRewardClaims,
		seedFriendQuests,
		seedForumPostVotes,
		seedStoryCommentHearts,
		seedChatFolders,
	}
	for _, fn := range fns {
		if err := fn(db, users, admin); err != nil {
			return err
		}
	}
	return nil
}

func seedLeagueParticipants(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.LeagueParticipant{}).Count(&count)
	if count > 0 {
		return nil
	}

	var season model.LeagueSeason
	if err := db.Where("is_active = ?", true).First(&season).Error; err != nil {
		return nil // no active season, skip
	}

	// Get first (lowest) division
	var division model.LeagueDivision
	if err := db.Order("tier ASC").First(&division).Error; err != nil {
		return nil
	}

	weeklyXPs := []int64{480, 310, 175}
	for i, user := range users {
		idx := i
		if idx >= len(weeklyXPs) {
			idx = len(weeklyXPs) - 1
		}
		p := model.LeagueParticipant{
			SeasonID:   season.ID,
			UserID:     user.ID,
			DivisionID: division.ID,
			WeeklyXP:   weeklyXPs[idx],
			Rank:       i + 1,
		}
		if err := db.Create(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedUserMapProgress(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.UserMapProgress{}).Count(&count)
	if count > 0 {
		return nil
	}

	var regions []model.MapRegion
	if err := db.Order("display_order ASC").Find(&regions).Error; err != nil || len(regions) == 0 {
		return nil
	}

	now := time.Now()

	// First user unlocks first 3 regions, second user first 2, third user first 1
	unlockCounts := []int{3, 2, 1}

	for i, user := range users {
		idx := i
		if idx >= len(unlockCounts) {
			idx = len(unlockCounts) - 1
		}
		toUnlock := unlockCounts[idx]
		if toUnlock > len(regions) {
			toUnlock = len(regions)
		}

		for r := 0; r < toUnlock; r++ {
			unlockedAt := now.AddDate(0, 0, -(30 - r*5))
			progress := model.UserMapProgress{
				UserID:     user.ID,
				RegionID:   regions[r].ID,
				IsUnlocked: true,
				UnlockedAt: &unlockedAt,
			}
			if err := db.Create(&progress).Error; err != nil {
				return err
			}

			// Also unlock landmarks for this region
			var landmarks []model.MapLandmark
			db.Where("region_id = ?", regions[r].ID).Order("display_order ASC").Find(&landmarks)

			for l, lm := range landmarks {
				// Unlock 70% of landmarks for first user, 50% for second, 30% for third
				ratio := 0.7 - float64(i)*0.2
				if float64(l)/float64(len(landmarks)) >= ratio {
					break
				}
				lmAt := unlockedAt.Add(time.Duration(l+1) * 24 * time.Hour)
				lmProg := model.UserLandmarkProgress{
					UserID:        user.ID,
					LandmarkID:    lm.ID,
					IsUnlocked:    true,
					CurrentValue:  lm.UnlockValue,
					UnlockedAt:    &lmAt,
					RewardClaimed: true,
				}
				if err := db.Create(&lmProg).Error; err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func seedUserSocietyMemberships(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.UserSocietyMembership{}).Count(&count)
	if count > 0 {
		return nil
	}

	// First user has streak 7, find the appropriate society
	for _, user := range users {
		var society model.StreakSociety
		if err := db.Where("min_streak <= ?", user.CurrentStreak).
			Order("min_streak DESC").First(&society).Error; err != nil {
			continue // no matching society
		}
		membership := model.UserSocietyMembership{
			UserID:    user.ID,
			SocietyID: society.ID,
			JoinedAt:  time.Now().AddDate(0, 0, -user.CurrentStreak),
			IsActive:  true,
		}
		if err := db.Create(&membership).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedUserTimedChallenges(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.UserTimedChallenge{}).Count(&count)
	if count > 0 {
		return nil
	}

	var templates []model.TimedChallengeTemplate
	if err := db.Where("is_active = ?", true).Order("id ASC").Find(&templates).Error; err != nil || len(templates) == 0 {
		return nil
	}

	now := time.Now()

	// First user: 1 active + 1 completed challenge
	// Second user: 1 active challenge
	for i, user := range users {
		if i >= 2 {
			break
		}

		// Active challenge
		tmpl := templates[i%len(templates)]
		startedAt := now.Add(-12 * time.Hour)
		expiresAt := startedAt.Add(time.Duration(tmpl.DurationMinutes) * time.Minute)
		active := model.UserTimedChallenge{
			UserID:       user.ID,
			TemplateID:   tmpl.ID,
			CurrentValue: tmpl.TargetValue / 3, // 33% progress
			Status:       model.TimedChallengeActive,
			StartedAt:    startedAt,
			ExpiresAt:    expiresAt,
		}
		if err := db.Create(&active).Error; err != nil {
			return err
		}

		// First user also gets a completed challenge
		if i == 0 && len(templates) > 1 {
			tmpl2 := templates[1]
			cStarted := now.AddDate(0, 0, -3)
			cCompleted := cStarted.Add(time.Duration(tmpl2.DurationMinutes/2) * time.Minute)
			completed := model.UserTimedChallenge{
				UserID:       user.ID,
				TemplateID:   tmpl2.ID,
				CurrentValue: tmpl2.TargetValue,
				Status:       model.TimedChallengeCompleted,
				StartedAt:    cStarted,
				ExpiresAt:    cStarted.Add(time.Duration(tmpl2.DurationMinutes) * time.Minute),
				CompletedAt:  &cCompleted,
			}
			if err := db.Create(&completed).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedUserCombosAndBoosts(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.UserCombo{}).Count(&count)
	if count > 0 {
		return nil
	}

	now := time.Now()

	// First user has an active combo chain
	if len(users) > 0 {
		lastActivity := now.Add(-15 * time.Minute)
		sessionStart := now.Add(-2 * time.Hour)
		combo := model.UserCombo{
			UserID:           users[0].ID,
			ComboCount:       5,
			Multiplier:       1.5,
			LastActivityType: "breathing",
			LastActivityAt:   &lastActivity,
			SessionStartedAt: &sessionStart,
		}
		if err := db.Create(&combo).Error; err != nil {
			return err
		}

		// Active XP boost from the combo
		boost := model.XPBoost{
			UserID:      users[0].ID,
			Multiplier:  1.5,
			TriggerType: model.BoostTriggerActivityChain,
			StartedAt:   now.Add(-30 * time.Minute),
			ExpiresAt:   now.Add(30 * time.Minute),
			IsActive:    true,
		}
		if err := db.Create(&boost).Error; err != nil {
			return err
		}
	}

	return nil
}

func seedUserChests(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.UserChest{}).Count(&count)
	if count > 0 {
		return nil
	}

	chests := []struct {
		UserIdx            int
		Rarity             model.ChestRarity
		IsOpened           bool
		RewardType         model.ChestRewardType
		RewardValue        int
		RewardLabel        string
		TriggerType        string
		TriggerDescription string
		DaysAgo            int
	}{
		{0, model.ChestCommon, true, model.ChestRewardXP, 50, "+50 XP", "milestone", "Menyelesaikan 10 aktivitas", 20},
		{0, model.ChestRare, true, model.ChestRewardCoins, 25, "+25 Koin Emas", "milestone", "Streak 7 hari berturut-turut", 10},
		{0, model.ChestEpic, false, "", 0, "", "milestone", "Mencapai Level 3", 1},
		{1, model.ChestCommon, true, model.ChestRewardXP, 30, "+30 XP", "milestone", "Menyelesaikan 5 aktivitas", 15},
		{1, model.ChestCommon, false, "", 0, "", "milestone", "Streak 5 hari berturut-turut", 2},
	}

	now := time.Now()
	for _, c := range chests {
		if c.UserIdx >= len(users) {
			continue
		}
		chest := model.UserChest{
			UserID:             users[c.UserIdx].ID,
			Rarity:             c.Rarity,
			IsOpened:           c.IsOpened,
			RewardType:         c.RewardType,
			RewardValue:        c.RewardValue,
			RewardLabel:        c.RewardLabel,
			TriggerType:        c.TriggerType,
			TriggerDescription: c.TriggerDescription,
			CreatedAt:          now.AddDate(0, 0, -c.DaysAgo),
		}
		if c.IsOpened {
			openedAt := now.AddDate(0, 0, -c.DaysAgo+1)
			chest.OpenedAt = &openedAt
		}
		if err := db.Create(&chest).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedUserSpins(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.UserSpin{}).Count(&count)
	if count > 0 {
		return nil
	}

	var spinRewards []model.SpinReward
	if err := db.Where("is_active = ?", true).Find(&spinRewards).Error; err != nil || len(spinRewards) == 0 {
		return nil
	}

	now := time.Now()

	// Give first user 5 spin history entries, second user 3
	spinCounts := []int{5, 3, 1}
	for i, user := range users {
		idx := i
		if idx >= len(spinCounts) {
			idx = len(spinCounts) - 1
		}
		numSpins := spinCounts[idx]

		for d := 1; d <= numSpins; d++ {
			reward := spinRewards[d%len(spinRewards)]
			spinDate := now.AddDate(0, 0, -d).Truncate(24 * time.Hour)
			spin := model.UserSpin{
				UserID:    user.ID,
				RewardID:  reward.ID,
				SpinDate:  spinDate,
				CreatedAt: spinDate.Add(9 * time.Hour), // 9 AM
			}
			if err := db.Create(&spin).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedNotifications(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.Notification{}).Count(&count)
	if count > 0 {
		return nil
	}

	now := time.Now()
	notifications := []struct {
		UserIdx  int
		Type     model.NotificationType
		Title    string
		Message  string
		IsRead   bool
		HoursAgo int
	}{
		{0, model.NotificationTypeLevelUp, "Naik Level!", "Selamat! Kamu naik ke Level 3. Fitur baru telah terbuka untukmu.", false, 2},
		{0, model.NotificationTypeBadgeEarned, "Badge Baru!", "Kamu mendapatkan badge \"Streak 7 Hari\" — konsistensi adalah kunci kesehatan mental.", false, 6},
		{0, model.NotificationTypeHeart, "Ceritamu Diapresiasi", "Seseorang memberikan hati untuk ceritamu \"Belajar Mengelola Kecemasan\".", true, 24},
		{0, model.NotificationTypeStoryApproved, "Cerita Disetujui", "Ceritamu \"Menemukan Harapan di Tengah Kegelapan\" telah disetujui dan dipublikasikan.", true, 48},
		{1, model.NotificationTypeBadgeEarned, "Badge Baru!", "Kamu mendapatkan badge \"Aktivitas 10\" — terus jaga momentum positifmu!", false, 8},
		{1, model.NotificationTypeHeart, "Ceritamu Diapresiasi", "Seseorang memberikan hati untuk ceritamu tentang self-care.", true, 36},
		{2, model.NotificationTypeLevelUp, "Naik Level!", "Selamat! Kamu naik ke Level 2. Terus lanjutkan perjalananmu.", true, 72},
	}

	for _, n := range notifications {
		if n.UserIdx >= len(users) {
			continue
		}
		notif := model.Notification{
			ID:        uuid.New(),
			UserID:    users[n.UserIdx].ID,
			Type:      n.Type,
			Title:     n.Title,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: now.Add(-time.Duration(n.HoursAgo) * time.Hour),
		}
		if err := db.Create(&notif).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedRewardClaims(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.RewardClaim{}).Count(&count)
	if count > 0 {
		return nil
	}

	var rewards []model.Reward
	if err := db.Where("is_active = ?", true).Order("coin_cost ASC").Find(&rewards).Error; err != nil || len(rewards) == 0 {
		return nil
	}

	// First user claimed the cheapest reward
	if len(users) > 0 && len(rewards) > 0 {
		claim := model.RewardClaim{
			UserID:    users[0].ID,
			RewardID:  rewards[0].ID,
			CoinSpent: rewards[0].CoinCost,
			ClaimedAt: time.Now().AddDate(0, 0, -5),
		}
		if err := db.Create(&claim).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedFriendQuests(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.FriendQuest{}).Count(&count)
	if count > 0 {
		return nil
	}
	if len(users) < 2 {
		return nil
	}

	now := time.Now()
	startsAt := now.AddDate(0, 0, -2)
	endsAt := now.AddDate(0, 0, 5)

	quests := []model.FriendQuest{
		{
			RequesterID:       users[0].ID,
			PartnerID:         users[1].ID,
			Title:             "Tantangan Pernapasan Bersama",
			Description:       "Mari berlatih pernapasan bersama. Selesaikan 5 sesi pernapasan dalam seminggu untuk mendapatkan hadiah!",
			QuestType:         model.FQTypeBreathing,
			TargetValue:       5,
			RequesterProgress: 3,
			PartnerProgress:   2,
			XPReward:          100,
			CoinReward:        20,
			Status:            model.FriendQuestActive,
			StartsAt:          &startsAt,
			EndsAt:            &endsAt,
		},
	}

	// If 3+ users, add a completed quest
	if len(users) >= 3 {
		completedAt := now.AddDate(0, 0, -7)
		cStart := now.AddDate(0, 0, -14)
		cEnd := now.AddDate(0, 0, -7)
		quests = append(quests, model.FriendQuest{
			RequesterID:       users[1].ID,
			PartnerID:         users[2].ID,
			Title:             "Jurnal Refleksi Harian",
			Description:       "Tantang sahabatmu untuk menulis jurnal refleksi selama 3 hari berturut-turut.",
			QuestType:         model.FQTypeJournal,
			TargetValue:       3,
			RequesterProgress: 3,
			PartnerProgress:   3,
			XPReward:          75,
			CoinReward:        15,
			Status:            model.FriendQuestCompleted,
			StartsAt:          &cStart,
			EndsAt:            &cEnd,
			CompletedAt:       &completedAt,
		})
	}

	for _, q := range quests {
		if err := db.Create(&q).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedForumPostVotes(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.ForumPostVote{}).Count(&count)
	if count > 0 {
		return nil
	}

	var posts []model.ForumPost
	if err := db.Order("id ASC").Limit(10).Find(&posts).Error; err != nil || len(posts) == 0 {
		return nil
	}

	now := time.Now()

	for i, post := range posts {
		// Each post gets 1-2 upvotes from different users
		for j, user := range users {
			if user.ID == post.UserID {
				continue // don't self-vote
			}
			if j > 1 {
				break // max 2 votes per post
			}
			vote := model.ForumPostVote{
				PostID:    post.ID,
				UserID:    user.ID,
				VoteType:  model.VoteTypeUpvote,
				CreatedAt: now.AddDate(0, 0, -(10 - i)),
			}
			if err := db.Create(&vote).Error; err != nil {
				return err
			}

			// Update the post's vote count
			db.Model(&model.ForumPost{}).Where("id = ?", post.ID).
				UpdateColumn("upvotes_count", gorm.Expr("upvotes_count + 1"))
		}
	}
	return nil
}

func seedStoryCommentHearts(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.StoryCommentHeart{}).Count(&count)
	if count > 0 {
		return nil
	}

	var comments []model.StoryComment
	if err := db.Order("created_at DESC").Limit(5).Find(&comments).Error; err != nil || len(comments) == 0 {
		return nil
	}

	for _, comment := range comments {
		for _, user := range users {
			if user.ID == comment.UserID {
				continue
			}
			heart := model.StoryCommentHeart{
				CommentID: comment.ID,
				UserID:    user.ID,
				CreatedAt: time.Now().AddDate(0, 0, -1),
			}
			if err := db.Create(&heart).Error; err != nil {
				return err
			}
			db.Model(&model.StoryComment{}).Where("id = ?", comment.ID).
				UpdateColumn("heart_count", gorm.Expr("heart_count + 1"))
			break // one heart per comment is enough
		}
	}
	return nil
}

func seedChatFolders(db *gorm.DB, users []model.User, _ model.User) error {
	var count int64
	db.Model(&model.ChatFolder{}).Count(&count)
	if count > 0 {
		return nil
	}

	if len(users) == 0 {
		return nil
	}

	folders := []model.ChatFolder{
		{
			UserID:   users[0].ID,
			Name:     "Kecemasan & Stres",
			Color:    "#6366f1",
			Icon:     "brain",
			Position: 0,
		},
		{
			UserID:   users[0].ID,
			Name:     "Motivasi Harian",
			Color:    "#f59e0b",
			Icon:     "sun",
			Position: 1,
		},
		{
			UserID:   users[0].ID,
			Name:     "Tidur & Relaksasi",
			Color:    "#8b5cf6",
			Icon:     "moon",
			Position: 2,
		},
	}

	for _, f := range folders {
		if err := db.Create(&f).Error; err != nil {
			return err
		}
	}
	return nil
}
