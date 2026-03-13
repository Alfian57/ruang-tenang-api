package application

import (
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	gamificationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/infrastructure"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/guild/infrastructure")

var (
	ErrAlreadyInGuild      = errors.New("kamu sudah tergabung dalam guild")
	ErrGuildFull           = errors.New("guild sudah penuh")
	ErrGuildNotFound       = errors.New("guild tidak ditemukan")
	ErrNotGuildMember      = errors.New("kamu bukan anggota guild ini")
	ErrNotGuildLeader      = errors.New("hanya leader yang bisa melakukan aksi ini")
	ErrNotGuildAdmin       = errors.New("hanya leader atau admin guild yang bisa melakukan aksi ini")
	ErrCannotKickLeader    = errors.New("tidak bisa mengeluarkan leader dari guild")
	ErrCannotLeaveAsLeader = errors.New("leader harus mentransfer kepemimpinan sebelum keluar")
	ErrInvalidInviteCode   = errors.New("kode undangan tidak valid")
	ErrChallengeNotFound   = errors.New("challenge tidak ditemukan")
	ErrMaxActiveChallenges = errors.New("guild sudah memiliki maksimal 3 challenge aktif")
)

type GuildService struct {
	guildRepo       *infrastructure.GuildRepository
	userRepo        *authinfra.UserRepository
	levelConfigRepo *gamificationinfra.LevelConfigRepository
}

func NewGuildService(
	guildRepo *infrastructure.GuildRepository,
	userRepo *authinfra.UserRepository,
	levelConfigRepo *gamificationinfra.LevelConfigRepository,
) *GuildService {
	return &GuildService{
		guildRepo:       guildRepo,
		userRepo:        userRepo,
		levelConfigRepo: levelConfigRepo,
	}
}

// ==========================================
// Guild CRUD
// ==========================================

// CreateGuild creates a new guild with the user as leader
func (s *GuildService) CreateGuild(ctx context.Context, userID uint, req dto.CreateGuildRequest) (*dto.GuildResponse, error) {
	if s.guildRepo.IsUserInGuild(ctx, userID) {
		return nil, ErrAlreadyInGuild
	}

	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("gagal membuat kode undangan: %w", err)
	}

	guild := &model.Guild{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		LeaderID:    userID,
		MaxMembers:  10,
		IsPublic:    req.IsPublic,
		InviteCode:  inviteCode,
	}

	if guild.Icon == "" {
		guild.Icon = "shield"
	}

	if err := s.guildRepo.Create(ctx, guild); err != nil {
		return nil, err
	}

	// Add creator as leader member
	member := &model.GuildMember{
		GuildID: guild.ID,
		UserID:  userID,
		Role:    model.GuildRoleLeader,
	}
	if err := s.guildRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	// Log activity
	s.logActivity(ctx, guild.ID, &userID, model.GuildActivityCreated, "Guild dibuat")

	return s.toGuildResponse(ctx, guild)
}

// GetGuild retrieves guild details
func (s *GuildService) GetGuild(ctx context.Context, guildID uuid.UUID, currentUserID uint) (*dto.GuildDetailResponse, error) {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrGuildNotFound
		}
		return nil, err
	}

	return s.toGuildDetailResponse(ctx, guild, currentUserID)
}

// UpdateGuild updates guild info (leader/admin only)
func (s *GuildService) UpdateGuild(ctx context.Context, guildID uuid.UUID, userID uint, req dto.UpdateGuildRequest) (*dto.GuildResponse, error) {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return nil, ErrGuildNotFound
	}

	if err := s.checkAdminAccess(ctx, guildID, userID); err != nil {
		return nil, err
	}

	if req.Name != nil {
		guild.Name = *req.Name
	}
	if req.Description != nil {
		guild.Description = *req.Description
	}
	if req.Icon != nil {
		guild.Icon = *req.Icon
	}
	if req.Banner != nil {
		guild.Banner = *req.Banner
	}
	if req.IsPublic != nil {
		guild.IsPublic = *req.IsPublic
	}

	if err := s.guildRepo.Update(ctx, guild); err != nil {
		return nil, err
	}

	return s.toGuildResponse(ctx, guild)
}

// DeleteGuild deletes a guild (leader only)
func (s *GuildService) DeleteGuild(ctx context.Context, guildID uuid.UUID, userID uint) error {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return ErrGuildNotFound
	}

	if guild.LeaderID != userID {
		return ErrNotGuildLeader
	}

	return s.guildRepo.Delete(ctx, guildID)
}

// GetPublicGuilds retrieves public guilds
func (s *GuildService) GetPublicGuilds(ctx context.Context, page, limit int) ([]dto.GuildResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 20 {
		limit = 10
	}

	guilds, total, err := s.guildRepo.GetPublicGuilds(ctx, page, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.GuildResponse, len(guilds))
	for i, g := range guilds {
		resp, err := s.toGuildResponse(ctx, &g)
		if err != nil {
			continue
		}
		result[i] = *resp
	}

	return result, total, nil
}

// GetGuildLeaderboard retrieves the top guilds by XP
func (s *GuildService) GetGuildLeaderboard(ctx context.Context, limit int) ([]dto.GuildLeaderboardEntry, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	guilds, err := s.guildRepo.GetGuildLeaderboard(ctx, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.GuildLeaderboardEntry, len(guilds))
	for i, g := range guilds {
		memberCount, _ := s.guildRepo.GetMemberCount(ctx, g.ID)
		result[i] = dto.GuildLeaderboardEntry{
			ID:          g.ID,
			Name:        g.Name,
			Icon:        g.Icon,
			TotalXP:     g.TotalXP,
			Level:       g.Level,
			MemberCount: memberCount,
			Rank:        i + 1,
		}
	}

	return result, nil
}

// ==========================================
// Guild Membership
// ==========================================

// JoinGuild joins a public guild
func (s *GuildService) JoinGuild(ctx context.Context, guildID uuid.UUID, userID uint) error {
	if s.guildRepo.IsUserInGuild(ctx, userID) {
		return ErrAlreadyInGuild
	}

	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return ErrGuildNotFound
	}

	if !guild.IsPublic {
		return ErrGuildNotFound
	}

	memberCount, _ := s.guildRepo.GetMemberCount(ctx, guildID)
	if memberCount >= guild.MaxMembers {
		return ErrGuildFull
	}

	member := &model.GuildMember{
		GuildID: guildID,
		UserID:  userID,
		Role:    model.GuildRoleMember,
	}

	if err := s.guildRepo.AddMember(ctx, member); err != nil {
		return err
	}

	s.logActivity(ctx, guildID, &userID, model.GuildActivityMemberJoined, "Bergabung ke guild")
	return nil
}

// JoinByInviteCode joins a guild via invite code
func (s *GuildService) JoinByInviteCode(ctx context.Context, code string, userID uint) (*dto.GuildResponse, error) {
	if s.guildRepo.IsUserInGuild(ctx, userID) {
		return nil, ErrAlreadyInGuild
	}

	guild, err := s.guildRepo.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, ErrInvalidInviteCode
	}

	memberCount, _ := s.guildRepo.GetMemberCount(ctx, guild.ID)
	if memberCount >= guild.MaxMembers {
		return nil, ErrGuildFull
	}

	member := &model.GuildMember{
		GuildID: guild.ID,
		UserID:  userID,
		Role:    model.GuildRoleMember,
	}

	if err := s.guildRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	s.logActivity(ctx, guild.ID, &userID, model.GuildActivityMemberJoined, "Bergabung via kode undangan")
	return s.toGuildResponse(ctx, guild)
}

// LeaveGuild removes the user from their guild
func (s *GuildService) LeaveGuild(ctx context.Context, guildID uuid.UUID, userID uint) error {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return ErrGuildNotFound
	}

	if guild.LeaderID == userID {
		return ErrCannotLeaveAsLeader
	}

	if err := s.guildRepo.RemoveMember(ctx, guildID, userID); err != nil {
		return err
	}

	s.logActivity(ctx, guildID, &userID, model.GuildActivityMemberLeft, "Meninggalkan guild")
	return nil
}

// KickMember removes a member from the guild (leader/admin only)
func (s *GuildService) KickMember(ctx context.Context, guildID uuid.UUID, targetUserID uint, kickerUserID uint) error {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return ErrGuildNotFound
	}

	if targetUserID == guild.LeaderID {
		return ErrCannotKickLeader
	}

	if err := s.checkAdminAccess(ctx, guildID, kickerUserID); err != nil {
		return err
	}

	if err := s.guildRepo.RemoveMember(ctx, guildID, targetUserID); err != nil {
		return err
	}

	s.logActivity(ctx, guildID, &kickerUserID, model.GuildActivityMemberKicked, fmt.Sprintf("Mengeluarkan anggota (ID: %d)", targetUserID))
	return nil
}

// PromoteMember promotes a member to admin (leader only)
func (s *GuildService) PromoteMember(ctx context.Context, guildID uuid.UUID, targetUserID uint, promoterUserID uint) error {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return ErrGuildNotFound
	}

	if guild.LeaderID != promoterUserID {
		return ErrNotGuildLeader
	}

	if err := s.guildRepo.UpdateMemberRole(ctx, guildID, targetUserID, model.GuildRoleAdmin); err != nil {
		return err
	}

	s.logActivity(ctx, guildID, &promoterUserID, model.GuildActivityMemberPromoted, fmt.Sprintf("Mempromosikan anggota (ID: %d) menjadi admin", targetUserID))
	return nil
}

// TransferLeadership transfers guild leadership to another member
func (s *GuildService) TransferLeadership(ctx context.Context, guildID uuid.UUID, newLeaderID uint, currentLeaderID uint) error {
	guild, err := s.guildRepo.GetByID(ctx, guildID)
	if err != nil {
		return ErrGuildNotFound
	}

	if guild.LeaderID != currentLeaderID {
		return ErrNotGuildLeader
	}

	// Verify target is a member
	_, err = s.guildRepo.GetMember(ctx, guildID, newLeaderID)
	if err != nil {
		return ErrNotGuildMember
	}

	// Update roles
	if err := s.guildRepo.UpdateMemberRole(ctx, guildID, newLeaderID, model.GuildRoleLeader); err != nil {
		return err
	}
	if err := s.guildRepo.UpdateMemberRole(ctx, guildID, currentLeaderID, model.GuildRoleMember); err != nil {
		return err
	}

	// Update guild leader
	guild.LeaderID = newLeaderID
	return s.guildRepo.Update(ctx, guild)
}

// GetMyGuild retrieves the current user's guild info
func (s *GuildService) GetMyGuild(ctx context.Context, userID uint) (*dto.MyGuildResponse, error) {
	membership, err := s.guildRepo.GetUserGuild(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &dto.MyGuildResponse{IsMember: false}, nil
		}
		return nil, err
	}

	guildResp, err := s.toGuildResponse(ctx, membership.Guild)
	if err != nil {
		return nil, err
	}

	return &dto.MyGuildResponse{
		Guild:         guildResp,
		MemberRole:    string(membership.Role),
		XPContributed: membership.XPContributed,
		IsMember:      true,
	}, nil
}

// ==========================================
// Guild Challenges
// ==========================================

// CreateChallenge creates a new guild challenge (leader/admin only)
func (s *GuildService) CreateChallenge(ctx context.Context, guildID uuid.UUID, userID uint, req dto.CreateGuildChallengeRequest) (*dto.GuildChallengeResponse, error) {
	if err := s.checkAdminAccess(ctx, guildID, userID); err != nil {
		return nil, err
	}

	// Max 3 active challenges at once
	active, _ := s.guildRepo.GetActiveChallenges(ctx, guildID)
	if len(active) >= 3 {
		return nil, ErrMaxActiveChallenges
	}

	now := time.Now()
	challenge := &model.GuildChallenge{
		GuildID:       guildID,
		Title:         req.Title,
		Description:   req.Description,
		ChallengeType: model.GuildChallengeType(req.ChallengeType),
		TargetValue:   req.TargetValue,
		XPReward:      req.XPReward,
		CoinReward:    req.CoinReward,
		StartsAt:      now,
		EndsAt:        now.Add(time.Duration(req.DurationDays) * 24 * time.Hour),
	}

	if err := s.guildRepo.CreateChallenge(ctx, challenge); err != nil {
		return nil, err
	}

	s.logActivity(ctx, guildID, &userID, model.GuildActivityChallengeCreated, fmt.Sprintf("Challenge baru: %s", req.Title))

	return s.toChallengeResponse(ctx, challenge), nil
}

// GetActiveChallenges retrieves active guild challenges
func (s *GuildService) GetActiveChallenges(ctx context.Context, guildID uuid.UUID) ([]dto.GuildChallengeResponse, error) {
	challenges, err := s.guildRepo.GetActiveChallenges(ctx, guildID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.GuildChallengeResponse, len(challenges))
	for i, c := range challenges {
		resp := s.toChallengeResponse(ctx, &c)

		// Get top contributors
		contribs, _ := s.guildRepo.GetTopContributors(ctx, c.ID, 5)
		resp.TopContributors = make([]dto.GuildChallengeContributorDTO, len(contribs))
		for j, contrib := range contribs {
			resp.TopContributors[j] = dto.GuildChallengeContributorDTO{
				UserID: contrib.UserID,
				Value:  contrib.Value,
			}
			if contrib.User != nil {
				resp.TopContributors[j].Username = contrib.User.Username
				resp.TopContributors[j].Name = contrib.User.Name
				resp.TopContributors[j].Avatar = contrib.User.Avatar
			}
		}

		result[i] = *resp
	}

	return result, nil
}

// GetChallengeHistory retrieves challenge history for a guild
func (s *GuildService) GetChallengeHistory(ctx context.Context, guildID uuid.UUID, page, limit int) ([]dto.GuildChallengeResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 20 {
		limit = 10
	}

	challenges, total, err := s.guildRepo.GetChallengeHistory(ctx, guildID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.GuildChallengeResponse, len(challenges))
	for i, c := range challenges {
		result[i] = *s.toChallengeResponse(ctx, &c)
	}

	return result, total, nil
}

// GetRecentActivities retrieves recent guild activities
func (s *GuildService) GetRecentActivities(ctx context.Context, guildID uuid.UUID, limit int) ([]dto.GuildActivityResponse, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}

	activities, err := s.guildRepo.GetRecentActivities(ctx, guildID, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.GuildActivityResponse, len(activities))
	for i, a := range activities {
		result[i] = dto.GuildActivityResponse{
			ID:           a.ID,
			ActivityType: string(a.ActivityType),
			Description:  a.Description,
			CreatedAt:    a.CreatedAt,
		}
		if a.User != nil {
			result[i].Username = a.User.Username
			result[i].Avatar = a.User.Avatar
		}
	}

	return result, nil
}

// ==========================================
// Helpers
// ==========================================

func (s *GuildService) checkAdminAccess(ctx context.Context, guildID uuid.UUID, userID uint) error {
	member, err := s.guildRepo.GetMember(ctx, guildID, userID)
	if err != nil {
		return ErrNotGuildMember
	}
	if member.Role != model.GuildRoleLeader && member.Role != model.GuildRoleAdmin {
		return ErrNotGuildAdmin
	}
	return nil
}

func (s *GuildService) logActivity(ctx context.Context, guildID uuid.UUID, userID *uint, actType model.GuildActivityType, desc string) {
	activity := &model.GuildActivity{
		GuildID:      guildID,
		UserID:       userID,
		ActivityType: actType,
		Description:  desc,
	}
	_ = s.guildRepo.CreateActivity(ctx, activity)
}

func (s *GuildService) toGuildResponse(ctx context.Context, guild *model.Guild) (*dto.GuildResponse, error) {
	memberCount, _ := s.guildRepo.GetMemberCount(ctx, guild.ID)

	leaderName := ""
	if guild.Leader != nil {
		leaderName = guild.Leader.Name
	}

	return &dto.GuildResponse{
		ID:          guild.ID,
		Name:        guild.Name,
		Description: guild.Description,
		Icon:        guild.Icon,
		Banner:      guild.Banner,
		LeaderID:    guild.LeaderID,
		LeaderName:  leaderName,
		MaxMembers:  guild.MaxMembers,
		MemberCount: memberCount,
		TotalXP:     guild.TotalXP,
		Level:       guild.Level,
		IsPublic:    guild.IsPublic,
		InviteCode:  guild.InviteCode,
		CreatedAt:   guild.CreatedAt,
	}, nil
}

func (s *GuildService) toGuildDetailResponse(ctx context.Context, guild *model.Guild, currentUserID uint) (*dto.GuildDetailResponse, error) {
	guildResp, err := s.toGuildResponse(ctx, guild)
	if err != nil {
		return nil, err
	}

	// Get members
	members, _ := s.guildRepo.GetMembers(ctx, guild.ID)
	memberResponses := make([]dto.GuildMemberResponse, len(members))
	for i, m := range members {
		userLevel := 1
		if m.User != nil {
			levelConfig, _ := s.levelConfigRepo.GetLevelByExp(ctx, m.User.Exp)
			if levelConfig != nil {
				userLevel = levelConfig.Level
			}
		}

		memberResponses[i] = dto.GuildMemberResponse{
			ID:            m.ID,
			UserID:        m.UserID,
			Role:          string(m.Role),
			XPContributed: m.XPContributed,
			UserLevel:     userLevel,
			JoinedAt:      m.JoinedAt,
		}
		if m.User != nil {
			memberResponses[i].Username = m.User.Username
			memberResponses[i].Name = m.User.Name
			memberResponses[i].Avatar = m.User.Avatar
		}
	}

	// Get active challenges
	challenges, _ := s.GetActiveChallenges(ctx, guild.ID)

	// Get recent activities
	activities, _ := s.GetRecentActivities(ctx, guild.ID, 10)

	// Check current user membership
	isCurrentUserGuild := false
	currentUserRole := ""
	currentMember, err := s.guildRepo.GetMember(ctx, guild.ID, currentUserID)
	if err == nil {
		isCurrentUserGuild = true
		currentUserRole = string(currentMember.Role)
	}

	return &dto.GuildDetailResponse{
		GuildResponse:      *guildResp,
		Members:            memberResponses,
		ActiveChallenges:   challenges,
		RecentActivities:   activities,
		IsCurrentUserGuild: isCurrentUserGuild,
		CurrentUserRole:    currentUserRole,
	}, nil
}

func (s *GuildService) toChallengeResponse(ctx context.Context, c *model.GuildChallenge) *dto.GuildChallengeResponse {
	return &dto.GuildChallengeResponse{
		ID:              c.ID,
		Title:           c.Title,
		Description:     c.Description,
		ChallengeType:   string(c.ChallengeType),
		TargetValue:     c.TargetValue,
		CurrentValue:    c.CurrentValue,
		ProgressPercent: c.ProgressPercent(),
		XPReward:        c.XPReward,
		CoinReward:      c.CoinReward,
		StartsAt:        c.StartsAt,
		EndsAt:          c.EndsAt,
		IsCompleted:     c.IsCompleted,
		IsExpired:       c.IsExpired(),
		IsActive:        c.IsActive(),
		CreatedAt:       c.CreatedAt,
	}
}

func generateInviteCode() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
