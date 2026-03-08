package repository

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GuildRepository struct {
	db *gorm.DB
}

func NewGuildRepository(db *gorm.DB) *GuildRepository {
	return &GuildRepository{db: db}
}

// ==========================================
// Guild CRUD
// ==========================================

// Create creates a new guild
func (r *GuildRepository) Create(ctx context.Context, guild *model.Guild) error {
	return r.db.WithContext(ctx).Create(guild).Error
}

// GetByID retrieves a guild by ID
func (r *GuildRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Guild, error) {
	var guild model.Guild
	err := r.db.WithContext(ctx).
		Preload("Leader").
		First(&guild, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &guild, nil
}

// GetByInviteCode retrieves a guild by its invite code
func (r *GuildRepository) GetByInviteCode(ctx context.Context, code string) (*model.Guild, error) {
	var guild model.Guild
	err := r.db.WithContext(ctx).
		Preload("Leader").
		Where("invite_code = ?", code).First(&guild).Error
	if err != nil {
		return nil, err
	}
	return &guild, nil
}

// Update updates a guild
func (r *GuildRepository) Update(ctx context.Context, guild *model.Guild) error {
	return r.db.WithContext(ctx).Save(guild).Error
}

// Delete deletes a guild
func (r *GuildRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Guild{}, "id = ?", id).Error
}

// GetPublicGuilds retrieves all public guilds with pagination
func (r *GuildRepository) GetPublicGuilds(ctx context.Context, page, limit int) ([]model.Guild, int64, error) {
	var guilds []model.Guild
	var total int64

	r.db.WithContext(ctx).Model(&model.Guild{}).Where("is_public = ?", true).Count(&total)

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Where("is_public = ?", true).
		Preload("Leader").
		Order("total_xp DESC").
		Offset(offset).Limit(limit).
		Find(&guilds).Error

	return guilds, total, err
}

// GetGuildLeaderboard retrieves guilds ranked by XP
func (r *GuildRepository) GetGuildLeaderboard(ctx context.Context, limit int) ([]model.Guild, error) {
	var guilds []model.Guild
	err := r.db.WithContext(ctx).
		Preload("Leader").
		Order("total_xp DESC").
		Limit(limit).
		Find(&guilds).Error
	return guilds, err
}

// AddXP adds XP to a guild's total
func (r *GuildRepository) AddXP(ctx context.Context, guildID uuid.UUID, xp int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Guild{}).
		Where("id = ?", guildID).
		UpdateColumn("total_xp", gorm.Expr("total_xp + ?", xp)).Error
}

// ==========================================
// Guild Members
// ==========================================

// AddMember adds a user to a guild
func (r *GuildRepository) AddMember(ctx context.Context, member *model.GuildMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// RemoveMember removes a user from a guild
func (r *GuildRepository) RemoveMember(ctx context.Context, guildID uuid.UUID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		Delete(&model.GuildMember{}).Error
}

// GetMember retrieves a guild member
func (r *GuildRepository) GetMember(ctx context.Context, guildID uuid.UUID, userID uint) (*model.GuildMember, error) {
	var member model.GuildMember
	err := r.db.WithContext(ctx).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetMembers retrieves all members of a guild
func (r *GuildRepository) GetMembers(ctx context.Context, guildID uuid.UUID) ([]model.GuildMember, error) {
	var members []model.GuildMember
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Preload("User").
		Order("role ASC, xp_contributed DESC").
		Find(&members).Error
	return members, err
}

// GetMemberCount returns the number of members in a guild
func (r *GuildRepository) GetMemberCount(ctx context.Context, guildID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.GuildMember{}).
		Where("guild_id = ?", guildID).
		Count(&count).Error
	return int(count), err
}

// GetUserGuild retrieves the guild a user belongs to
func (r *GuildRepository) GetUserGuild(ctx context.Context, userID uint) (*model.GuildMember, error) {
	var member model.GuildMember
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Guild").
		Preload("Guild.Leader").
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// IsUserInGuild checks if a user is a member of any guild
func (r *GuildRepository) IsUserInGuild(ctx context.Context, userID uint) bool {
	var count int64
	r.db.WithContext(ctx).
		Model(&model.GuildMember{}).
		Where("user_id = ?", userID).
		Count(&count)
	return count > 0
}

// UpdateMemberRole updates a member's role
func (r *GuildRepository) UpdateMemberRole(ctx context.Context, guildID uuid.UUID, userID uint, role model.GuildMemberRole) error {
	return r.db.WithContext(ctx).
		Model(&model.GuildMember{}).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		Update("role", role).Error
}

// AddMemberXP adds XP contribution to a member
func (r *GuildRepository) AddMemberXP(ctx context.Context, guildID uuid.UUID, userID uint, xp int64) error {
	return r.db.WithContext(ctx).
		Model(&model.GuildMember{}).
		Where("guild_id = ? AND user_id = ?", guildID, userID).
		UpdateColumn("xp_contributed", gorm.Expr("xp_contributed + ?", xp)).Error
}

// ==========================================
// Guild Challenges
// ==========================================

// CreateChallenge creates a new guild challenge
func (r *GuildRepository) CreateChallenge(ctx context.Context, challenge *model.GuildChallenge) error {
	return r.db.WithContext(ctx).Create(challenge).Error
}

// GetChallenge retrieves a challenge by ID
func (r *GuildRepository) GetChallenge(ctx context.Context, id uuid.UUID) (*model.GuildChallenge, error) {
	var challenge model.GuildChallenge
	err := r.db.WithContext(ctx).First(&challenge, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &challenge, nil
}

// GetActiveChallenges retrieves active challenges for a guild
func (r *GuildRepository) GetActiveChallenges(ctx context.Context, guildID uuid.UUID) ([]model.GuildChallenge, error) {
	var challenges []model.GuildChallenge
	now := time.Now()
	err := r.db.WithContext(ctx).
		Where("guild_id = ? AND starts_at <= ? AND ends_at >= ? AND is_completed = ?", guildID, now, now, false).
		Order("ends_at ASC").
		Find(&challenges).Error
	return challenges, err
}

// GetChallengeHistory retrieves all challenges for a guild
func (r *GuildRepository) GetChallengeHistory(ctx context.Context, guildID uuid.UUID, page, limit int) ([]model.GuildChallenge, int64, error) {
	var challenges []model.GuildChallenge
	var total int64

	r.db.WithContext(ctx).Model(&model.GuildChallenge{}).Where("guild_id = ?", guildID).Count(&total)

	offset := (page - 1) * limit
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&challenges).Error

	return challenges, total, err
}

// UpdateChallengeProgress updates the current value of a challenge
func (r *GuildRepository) UpdateChallengeProgress(ctx context.Context, challengeID uuid.UUID, increment int) error {
	return r.db.WithContext(ctx).
		Model(&model.GuildChallenge{}).
		Where("id = ?", challengeID).
		UpdateColumn("current_value", gorm.Expr("current_value + ?", increment)).Error
}

// CompleteChallenge marks a challenge as completed
func (r *GuildRepository) CompleteChallenge(ctx context.Context, challengeID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.GuildChallenge{}).
		Where("id = ?", challengeID).
		Updates(map[string]interface{}{
			"is_completed": true,
			"completed_at": now,
		}).Error
}

// AddChallengeContribution adds a contribution record
func (r *GuildRepository) AddChallengeContribution(ctx context.Context, contribution *model.GuildChallengeContribution) error {
	return r.db.WithContext(ctx).Create(contribution).Error
}

// GetChallengeContributions retrieves contributions for a challenge
func (r *GuildRepository) GetChallengeContributions(ctx context.Context, challengeID uuid.UUID) ([]model.GuildChallengeContribution, error) {
	var contribs []model.GuildChallengeContribution
	err := r.db.WithContext(ctx).
		Where("challenge_id = ?", challengeID).
		Preload("User").
		Order("value DESC").
		Find(&contribs).Error
	return contribs, err
}

// GetTopContributors retrieves top contributors for a challenge
func (r *GuildRepository) GetTopContributors(ctx context.Context, challengeID uuid.UUID, limit int) ([]model.GuildChallengeContribution, error) {
	var contribs []model.GuildChallengeContribution

	// Aggregate contributions per user
	err := r.db.WithContext(ctx).
		Table("guild_challenge_contributions").
		Select("user_id, SUM(value) as value").
		Where("challenge_id = ?", challengeID).
		Group("user_id").
		Order("value DESC").
		Limit(limit).
		Find(&contribs).Error
	if err != nil {
		return nil, err
	}

	// Load user data
	for i := range contribs {
		var user model.User
		r.db.WithContext(ctx).First(&user, contribs[i].UserID)
		contribs[i].User = &user
	}

	return contribs, nil
}

// ==========================================
// Guild Activities
// ==========================================

// CreateActivity adds an activity log entry
func (r *GuildRepository) CreateActivity(ctx context.Context, activity *model.GuildActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// GetRecentActivities retrieves recent activities for a guild
func (r *GuildRepository) GetRecentActivities(ctx context.Context, guildID uuid.UUID, limit int) ([]model.GuildActivity, error) {
	var activities []model.GuildActivity
	err := r.db.WithContext(ctx).
		Where("guild_id = ?", guildID).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&activities).Error
	return activities, err
}
