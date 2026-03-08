package service

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrLeagueSeasonNotFound = errors.New("tidak ada season liga aktif")
	ErrAlreadyJoinedLeague  = errors.New("kamu sudah bergabung di liga minggu ini")
)

type WeeklyLeagueService struct {
	leagueRepo *repository.WeeklyLeagueRepository
	userRepo   *repository.UserRepository
}

func NewWeeklyLeagueService(
	leagueRepo *repository.WeeklyLeagueRepository,
	userRepo *repository.UserRepository,
) *WeeklyLeagueService {
	return &WeeklyLeagueService{
		leagueRepo: leagueRepo,
		userRepo:   userRepo,
	}
}

// GetOverview returns the current league overview for a user
func (s *WeeklyLeagueService) GetOverview(ctx context.Context, userID uint) (*dto.LeagueOverviewResponse, error) {
	season, err := s.leagueRepo.GetActiveSeason(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLeagueSeasonNotFound
		}
		return nil, err
	}

	participant, err := s.leagueRepo.GetParticipant(ctx, season.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Auto-join lowest division
			participant, err = s.autoJoin(ctx, season, userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	leaderboard, err := s.leagueRepo.GetLeaderboard(ctx, season.ID, participant.DivisionID, 30)
	if err != nil {
		return nil, err
	}

	lbEntries := make([]dto.LeagueParticipantResponse, len(leaderboard))
	for i, p := range leaderboard {
		username := ""
		avatar := ""
		if p.User != nil {
			username = p.User.Name
			avatar = p.User.Avatar
		}
		lbEntries[i] = dto.LeagueParticipantResponse{
			Rank:       i + 1,
			UserID:     p.UserID,
			Username:   username,
			Avatar:     avatar,
			WeeklyXP:   p.WeeklyXP,
			IsPromoted: p.IsPromoted,
			IsDemoted:  p.IsDemoted,
			IsMe:       p.UserID == userID,
		}
	}

	division := participant.Division
	if division == nil {
		d, _ := s.leagueRepo.GetDivisionByTier(ctx, 1)
		division = d
	}

	timeLeft := int(time.Until(season.EndsAt).Seconds())
	if timeLeft < 0 {
		timeLeft = 0
	}

	myRank := 0
	for _, e := range lbEntries {
		if e.IsMe {
			myRank = e.Rank
			break
		}
	}

	return &dto.LeagueOverviewResponse{
		Season: dto.LeagueSeasonResponse{
			ID:         season.ID,
			WeekNumber: season.WeekNumber,
			Year:       season.Year,
			StartsAt:   season.StartsAt,
			EndsAt:     season.EndsAt,
			IsActive:   season.IsActive,
		},
		Division: dto.LeagueDivisionResponse{
			ID:             division.ID,
			Name:           division.Name,
			Icon:           division.Icon,
			Tier:           division.Tier,
			Color:          division.Color,
			PromotionSlots: division.PromotionSlots,
			DemotionSlots:  division.DemotionSlots,
		},
		MyRank:       myRank,
		MyWeeklyXP:   participant.WeeklyXP,
		Leaderboard:  lbEntries,
		TimeLeftSecs: timeLeft,
	}, nil
}

// GetDivisions returns all league divisions
func (s *WeeklyLeagueService) GetDivisions(ctx context.Context) ([]dto.LeagueDivisionResponse, error) {
	divisions, err := s.leagueRepo.GetAllDivisions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.LeagueDivisionResponse, len(divisions))
	for i, d := range divisions {
		result[i] = dto.LeagueDivisionResponse{
			ID:             d.ID,
			Name:           d.Name,
			Icon:           d.Icon,
			Tier:           d.Tier,
			Color:          d.Color,
			PromotionSlots: d.PromotionSlots,
			DemotionSlots:  d.DemotionSlots,
		}
	}
	return result, nil
}

// AddXP adds weekly XP to a participant
func (s *WeeklyLeagueService) AddXP(ctx context.Context, userID uint, xp int64) error {
	season, err := s.leagueRepo.GetActiveSeason(ctx)
	if err != nil {
		return nil // No active season, silently skip
	}

	_, err = s.leagueRepo.GetParticipant(ctx, season.ID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_, err = s.autoJoin(ctx, season, userID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return s.leagueRepo.AddWeeklyXP(ctx, season.ID, userID, xp)
}

func (s *WeeklyLeagueService) autoJoin(ctx context.Context, season *model.LeagueSeason, userID uint) (*model.LeagueParticipant, error) {
	lowestDiv, err := s.leagueRepo.GetLowestDivision(ctx)
	if err != nil {
		return nil, err
	}

	p := &model.LeagueParticipant{
		SeasonID:   season.ID,
		UserID:     userID,
		DivisionID: lowestDiv.ID,
	}
	if err := s.leagueRepo.CreateParticipant(ctx, p); err != nil {
		return nil, err
	}
	p.Division = lowestDiv
	return p, nil
}
