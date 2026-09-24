package presentation

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedGamification seeds the active gamification tables used by the dashboard.
func SeedGamification(db *gorm.DB) error {
	var users []model.User
	if err := db.Where("role = ? AND email IN ?", model.RoleUser, presentationAccountEmails).Order("id ASC").Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	var admin model.User
	if err := db.Where("email = ? AND role = ?", presentationAdminEmail, model.RoleAdmin).First(&admin).Error; err != nil {
		return err
	}

	fns := []func(*gorm.DB, []model.User, model.User) error{
		seedUserMapProgress,
		seedUserCombosAndBoosts,
		seedNotifications,
		seedRewardClaims,
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
			LastActivityType: "chat",
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
