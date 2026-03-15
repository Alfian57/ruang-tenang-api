package development

import (
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedGuilds seeds guild-related tables so the guild feature looks populated:
// guilds, guild_members, guild_challenges, guild_challenge_contributions, guild_activities
func SeedGuilds(db *gorm.DB) error {
	var count int64
	db.Model(&model.Guild{}).Count(&count)
	if count > 0 {
		return nil
	}

	var users []model.User
	if err := db.Where("role = ?", model.RoleMember).Order("id ASC").Find(&users).Error; err != nil {
		return err
	}
	if len(users) < 2 {
		return nil
	}

	half := len(users) / 2
	guildGroups := []struct {
		Name        string
		Description string
		Icon        string
		TotalXP     int64
		Level       int
		InviteCode  string
		Users       []model.User
	}{
		{
			Name:        "Ruang Semangat",
			Description: "Komunitas pendukung untuk saling memotivasi dalam menjaga kesehatan mental. Kami percaya bahwa setiap langkah kecil menuju kebaikan diri adalah sebuah pencapaian yang layak dirayakan.",
			Icon:        "guild-icon-lotus.png",
			TotalXP:     2400,
			Level:       2,
			InviteCode:  "SEMANGAT2025",
			Users:       users[:half],
		},
		{
			Name:        "Harmoni Bintang",
			Description: "Tempat tenang untuk saling berbagi cerita dan menemukan ketenangan batin. Di bawah cahaya bintang, kita semua terkoneksi dengan damai dan saling mendengarkan.",
			Icon:        "guild-icon-star.png",
			TotalXP:     5600,
			Level:       4,
			InviteCode:  "BINTANG2025",
			Users:       users[half:],
		},
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	weekEnd := weekStart.Add(7 * 24 * time.Hour)

	for _, gGroup := range guildGroups {
		if len(gGroup.Users) == 0 {
			continue
		}

		icon := getSeedAsset(gGroup.Icon, "images")

		guild := model.Guild{
			Name:        gGroup.Name,
			Description: gGroup.Description,
			Icon:        icon,
			LeaderID:    gGroup.Users[0].ID,
			MaxMembers:  10,
			TotalXP:     gGroup.TotalXP,
			Level:       gGroup.Level,
			IsPublic:    true,
			InviteCode:  gGroup.InviteCode,
			CreatedAt:   now.AddDate(0, 0, -30),
		}
		if err := db.Create(&guild).Error; err != nil {
			return err
		}

		// Add members
		for i, user := range gGroup.Users {
			if i >= 10 {
				break
			}
			role := model.GuildRoleMember
			if i == 0 {
				role = model.GuildRoleLeader
			} else if i == 1 {
				role = model.GuildRoleAdmin
			}

			xpContrib := int64(1000 - (i * 100))
			if xpContrib < 0 {
				xpContrib = 10
			}

			member := model.GuildMember{
				GuildID:       guild.ID,
				UserID:        user.ID,
				Role:          role,
				XPContributed: xpContrib,
				JoinedAt:      now.AddDate(0, 0, -(30 - i*2)),
			}
			if err := db.Create(&member).Error; err != nil {
				return err
			}
		}

		// ==========================================
		// Daily Tasks (duration <= 1 day)
		// ==========================================

		dailyTasks := []model.GuildChallenge{
			{
				GuildID:       guild.ID,
				Title:         "Baca 3 Artikel Bersama",
				Description:   "Seluruh anggota guild membaca total 3 artikel hari ini. Membaca artikel tentang kesehatan mental meningkatkan kesadaran diri.",
				ChallengeType: model.GuildChallengeTask,
				TargetValue:   3,
				CurrentValue:  1,
				XPReward:      30,
				CoinReward:    5,
				StartsAt:      now.Truncate(24 * time.Hour),
				EndsAt:        now.Truncate(24 * time.Hour).Add(24 * time.Hour),
			},
			{
				GuildID:       guild.ID,
				Title:         "Chat AI 5 Kali",
				Description:   "Seluruh anggota guild mengirim total 5 pesan ke AI companion hari ini. Bercerita kepada AI membantu meluapkan perasaan.",
				ChallengeType: model.GuildChallengeChat,
				TargetValue:   5,
				CurrentValue:  2,
				XPReward:      40,
				CoinReward:    8,
				StartsAt:      now.Truncate(24 * time.Hour),
				EndsAt:        now.Truncate(24 * time.Hour).Add(24 * time.Hour),
			},
			{
				GuildID:       guild.ID,
				Title:         "Tulis 2 Jurnal Bersama",
				Description:   "Seluruh anggota guild menulis total 2 jurnal hari ini. Menulis jurnal membantu refleksi diri dan mengelola emosi.",
				ChallengeType: model.GuildChallengeJournal,
				TargetValue:   2,
				CurrentValue:  0,
				XPReward:      50,
				CoinReward:    10,
				StartsAt:      now.Truncate(24 * time.Hour),
				EndsAt:        now.Truncate(24 * time.Hour).Add(24 * time.Hour),
			},
		}

		for i := range dailyTasks {
			if err := db.Create(&dailyTasks[i]).Error; err != nil {
				return err
			}
		}

		// ==========================================
		// Weekly Tasks (duration > 1 day, <= 7 days)
		// ==========================================

		weeklyTasks := []model.GuildChallenge{
			{
				GuildID:       guild.ID,
				Title:         "Minggu Mindfulness",
				Description:   "Seluruh anggota guild bersama-sama mengumpulkan 500 XP dari latihan pernapasan minggu ini. Pernapasan mindful membantu menenangkan pikiran.",
				ChallengeType: model.GuildChallengeBreathing,
				TargetValue:   500,
				CurrentValue:  280,
				XPReward:      200,
				CoinReward:    50,
				StartsAt:      weekStart,
				EndsAt:        weekEnd,
			},
			{
				GuildID:       guild.ID,
				Title:         "Kumpulkan 1000 XP Guild",
				Description:   "Seluruh anggota guild bersama-sama mengumpulkan total 1000 XP dari berbagai aktivitas selama satu minggu.",
				ChallengeType: model.GuildChallengeXP,
				TargetValue:   1000,
				CurrentValue:  450,
				XPReward:      300,
				CoinReward:    60,
				StartsAt:      weekStart,
				EndsAt:        weekEnd,
			},
			{
				GuildID:       guild.ID,
				Title:         "Tulis 10 Jurnal Bersama",
				Description:   "Seluruh anggota guild bersama-sama menulis total 10 jurnal selama satu minggu. Refleksi diri melalui tulisan meningkatkan kesejahteraan emosional.",
				ChallengeType: model.GuildChallengeJournal,
				TargetValue:   10,
				CurrentValue:  4,
				XPReward:      250,
				CoinReward:    40,
				StartsAt:      weekStart,
				EndsAt:        weekEnd,
			},
		}

		for i := range weeklyTasks {
			if err := db.Create(&weeklyTasks[i]).Error; err != nil {
				return err
			}
		}

		// Add contributions for the first weekly task (Minggu Mindfulness)
		contribs := []struct {
			UserIdx int
			Value   int
		}{
			{0, 150},
			{1, 90},
			{2, 40},
		}
		for _, c := range contribs {
			if c.UserIdx >= len(gGroup.Users) {
				continue
			}
			contrib := model.GuildChallengeContribution{
				ChallengeID:   weeklyTasks[0].ID,
				UserID:        gGroup.Users[c.UserIdx].ID,
				Value:         c.Value,
				ContributedAt: now.Add(-time.Duration(c.UserIdx*8) * time.Hour),
			}
			if err := db.Create(&contrib).Error; err != nil {
				return err
			}
		}

		// ==========================================
		// Activity log (with proper names and role labels)
		// ==========================================

		if len(gGroup.Users) >= 2 {
			activities := []model.GuildActivity{
				{
					GuildID:      guild.ID,
					UserID:       &gGroup.Users[0].ID,
					ActivityType: model.GuildActivityCreated,
					Description:  fmt.Sprintf("Guild \"%s\" telah dibentuk oleh %s", guild.Name, gGroup.Users[0].Name),
					CreatedAt:    now.AddDate(0, 0, -30),
				},
				{
					GuildID:      guild.ID,
					UserID:       &gGroup.Users[1].ID,
					ActivityType: model.GuildActivityMemberJoined,
					Description:  fmt.Sprintf("%s bergabung ke guild", gGroup.Users[1].Name),
					CreatedAt:    now.AddDate(0, 0, -27),
				},
				{
					GuildID:      guild.ID,
					UserID:       &gGroup.Users[0].ID,
					ActivityType: model.GuildActivityMemberPromoted,
					Description:  fmt.Sprintf("%s mempromosikan %s menjadi Wakil Ketua", gGroup.Users[0].Name, gGroup.Users[1].Name),
					CreatedAt:    now.AddDate(0, 0, -26),
				},
				{
					GuildID:      guild.ID,
					UserID:       &gGroup.Users[0].ID,
					ActivityType: model.GuildActivityChallengeCreated,
					Description:  "Tugas baru dibuat: Minggu Mindfulness",
					CreatedAt:    weekStart,
				},
				{
					GuildID:      guild.ID,
					UserID:       &gGroup.Users[0].ID,
					ActivityType: model.GuildActivityXPContributed,
					Description:  fmt.Sprintf("%s berkontribusi 150 XP pada tugas Minggu Mindfulness", gGroup.Users[0].Name),
					CreatedAt:    now.Add(-16 * time.Hour),
				},
			}
			if len(gGroup.Users) >= 3 {
				activities = append(activities, model.GuildActivity{
					GuildID:      guild.ID,
					UserID:       &gGroup.Users[2].ID,
					ActivityType: model.GuildActivityMemberJoined,
					Description:  fmt.Sprintf("%s bergabung ke guild", gGroup.Users[2].Name),
					CreatedAt:    now.AddDate(0, 0, -24),
				})
			}

			for i := range activities {
				if err := db.Create(&activities[i]).Error; err != nil {
					return err
				}
			}
		}
	}

	return nil
}
