package development

import (
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

	now := time.Now()

	// Create one guild led by the first user
	guild := model.Guild{
		Name:        "Ruang Semangat",
		Description: "Komunitas pendukung untuk saling memotivasi dalam menjaga kesehatan mental. Kami percaya bahwa setiap langkah kecil menuju kebaikan diri adalah sebuah pencapaian yang layak dirayakan.",
		Icon:        "shield",
		LeaderID:    users[0].ID,
		MaxMembers:  10,
		TotalXP:     2400,
		Level:       2,
		IsPublic:    true,
		InviteCode:  "SEMANGAT2025",
		CreatedAt:   now.AddDate(0, 0, -30),
	}
	if err := db.Create(&guild).Error; err != nil {
		return err
	}

	// Add members
	roles := []model.GuildMemberRole{model.GuildRoleLeader, model.GuildRoleMember, model.GuildRoleMember}
	xpContributions := []int64{1200, 800, 400}
	for i, user := range users {
		if i >= 3 {
			break
		}
		member := model.GuildMember{
			GuildID:       guild.ID,
			UserID:        user.ID,
			Role:          roles[i],
			XPContributed: xpContributions[i],
			JoinedAt:      now.AddDate(0, 0, -(30 - i*3)),
		}
		if err := db.Create(&member).Error; err != nil {
			return err
		}
	}

	// Active challenge
	challengeStart := now.AddDate(0, 0, -3)
	challengeEnd := now.AddDate(0, 0, 4)
	challenge := model.GuildChallenge{
		GuildID:       guild.ID,
		Title:         "Minggu Mindfulness",
		Description:   "Tantangan guild minggu ini: seluruh anggota bersama-sama mengumpulkan 500 XP dari latihan pernapasan. Pernapasan mindful membantu menenangkan pikiran dan mengurangi stres.",
		ChallengeType: model.GuildChallengeBreathing,
		TargetValue:   500,
		CurrentValue:  280,
		XPReward:      200,
		CoinReward:    50,
		StartsAt:      challengeStart,
		EndsAt:        challengeEnd,
	}
	if err := db.Create(&challenge).Error; err != nil {
		return err
	}

	// Contributions
	contribs := []struct {
		UserIdx int
		Value   int
	}{
		{0, 150},
		{1, 90},
		{2, 40},
	}
	for _, c := range contribs {
		if c.UserIdx >= len(users) {
			continue
		}
		contrib := model.GuildChallengeContribution{
			ChallengeID:   challenge.ID,
			UserID:        users[c.UserIdx].ID,
			Value:         c.Value,
			ContributedAt: now.Add(-time.Duration(c.UserIdx*8) * time.Hour),
		}
		if err := db.Create(&contrib).Error; err != nil {
			return err
		}
	}

	// Activity log
	activities := []model.GuildActivity{
		{
			GuildID:      guild.ID,
			UserID:       &users[0].ID,
			ActivityType: model.GuildActivityCreated,
			Description:  "Guild \"Ruang Semangat\" telah dibentuk",
			CreatedAt:    now.AddDate(0, 0, -30),
		},
		{
			GuildID:      guild.ID,
			UserID:       &users[1].ID,
			ActivityType: model.GuildActivityMemberJoined,
			Description:  users[1].Name + " bergabung ke guild",
			CreatedAt:    now.AddDate(0, 0, -27),
		},
		{
			GuildID:      guild.ID,
			UserID:       &users[0].ID,
			ActivityType: model.GuildActivityChallengeCreated,
			Description:  "Tantangan baru dibuat: Minggu Mindfulness",
			CreatedAt:    challengeStart,
		},
		{
			GuildID:      guild.ID,
			UserID:       &users[0].ID,
			ActivityType: model.GuildActivityXPContributed,
			Description:  users[0].Name + " berkontribusi 150 XP pada tantangan Minggu Mindfulness",
			CreatedAt:    now.Add(-16 * time.Hour),
		},
	}
	if len(users) >= 3 {
		activities = append(activities, model.GuildActivity{
			GuildID:      guild.ID,
			UserID:       &users[2].ID,
			ActivityType: model.GuildActivityMemberJoined,
			Description:  users[2].Name + " bergabung ke guild",
			CreatedAt:    now.AddDate(0, 0, -24),
		})
	}

	for _, a := range activities {
		if err := db.Create(&a).Error; err != nil {
			return err
		}
	}

	return nil
}
