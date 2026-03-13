package application

import (
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"context"
	"errors"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/streak_society/infrastructure")

var (
	ErrSocietyNotFound      = errors.New("streak society tidak ditemukan")
	ErrStreakTooLow         = errors.New("streak kamu belum cukup untuk bergabung")
	ErrAlreadySocietyMember = errors.New("kamu sudah menjadi anggota society ini")
)

type StreakSocietyService struct {
	societyRepo *infrastructure.StreakSocietyRepository
	userRepo    *authinfra.UserRepository
}

func NewStreakSocietyService(
	societyRepo *infrastructure.StreakSocietyRepository,
	userRepo *authinfra.UserRepository,
) *StreakSocietyService {
	return &StreakSocietyService{
		societyRepo: societyRepo,
		userRepo:    userRepo,
	}
}

// GetOverview returns all societies and the user's current status
func (s *StreakSocietyService) GetOverview(ctx context.Context, userID uint) (*dto.StreakSocietyOverviewResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	allSocieties, err := s.societyRepo.GetAllSocieties(ctx)
	if err != nil {
		return nil, err
	}

	memberships, err := s.societyRepo.GetUserActiveMemberships(ctx, userID)
	if err != nil {
		return nil, err
	}

	membershipMap := make(map[int]bool)
	for _, m := range memberships {
		membershipMap[m.SocietyID] = true
	}

	result := make([]dto.StreakSocietyResponse, len(allSocieties))
	var currentSociety *dto.StreakSocietyResponse

	for i, soc := range allSocieties {
		memberCount, _ := s.societyRepo.CountMembers(ctx, soc.ID)
		isMember := membershipMap[soc.ID]

		resp := dto.StreakSocietyResponse{
			ID:            soc.ID,
			Name:          soc.Name,
			Icon:          soc.Icon,
			MinStreak:     soc.MinStreak,
			BorderColor:   soc.BorderColor,
			BadgeGlow:     soc.BadgeGlow,
			ExclusiveChat: soc.ExclusiveChat,
			MemberCount:   memberCount,
			IsMember:      isMember,
		}

		result[i] = resp

		if isMember && (currentSociety == nil || soc.MinStreak > currentSociety.MinStreak) {
			currentSociety = &result[i]
		}
	}

	return &dto.StreakSocietyOverviewResponse{
		CurrentStreak:  user.CurrentStreak,
		CurrentSociety: currentSociety,
		AllSocieties:   result,
	}, nil
}

// JoinSociety automatically joins the highest eligible society
func (s *StreakSocietyService) JoinSociety(ctx context.Context, userID uint) (*dto.StreakSocietyResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	society, err := s.societyRepo.GetSocietyByMinStreak(ctx, user.CurrentStreak)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStreakTooLow
		}
		return nil, err
	}

	// Check if already a member
	_, err = s.societyRepo.GetUserMembership(ctx, userID, society.ID)
	if err == nil {
		return nil, ErrAlreadySocietyMember
	}

	membership := &model.UserSocietyMembership{
		UserID:    userID,
		SocietyID: society.ID,
		IsActive:  true,
	}

	if err := s.societyRepo.CreateMembership(ctx, membership); err != nil {
		return nil, err
	}

	memberCount, _ := s.societyRepo.CountMembers(ctx, society.ID)

	return &dto.StreakSocietyResponse{
		ID:            society.ID,
		Name:          society.Name,
		Icon:          society.Icon,
		MinStreak:     society.MinStreak,
		BorderColor:   society.BorderColor,
		BadgeGlow:     society.BadgeGlow,
		ExclusiveChat: society.ExclusiveChat,
		MemberCount:   memberCount,
		IsMember:      true,
	}, nil
}

// GetSocietyMembers returns members of a society
func (s *StreakSocietyService) GetSocietyMembers(ctx context.Context, societyID int, page, limit int) ([]dto.SocietyMemberResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	members, total, err := s.societyRepo.GetSocietyMembers(ctx, societyID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.SocietyMemberResponse, len(members))
	for i, m := range members {
		username := ""
		avatar := ""
		streak := 0
		if m.User != nil {
			username = m.User.Name
			avatar = m.User.Avatar
			streak = m.User.CurrentStreak
		}
		result[i] = dto.SocietyMemberResponse{
			UserID:   m.UserID,
			Username: username,
			Avatar:   avatar,
			Streak:   streak,
			JoinedAt: m.JoinedAt,
		}
	}

	return result, total, nil
}
