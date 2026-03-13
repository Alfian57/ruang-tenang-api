package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type WeeklyLeagueRepository struct {
	db *gorm.DB
}

func NewWeeklyLeagueRepository(db *gorm.DB) *WeeklyLeagueRepository {
	return &WeeklyLeagueRepository{db: db}
}

// === Divisions ===

func (r *WeeklyLeagueRepository) GetAllDivisions(ctx context.Context) ([]model.LeagueDivision, error) {
	var divisions []model.LeagueDivision
	err := r.db.WithContext(ctx).Order("tier ASC").Find(&divisions).Error
	return divisions, err
}

func (r *WeeklyLeagueRepository) GetDivisionByTier(ctx context.Context, tier int) (*model.LeagueDivision, error) {
	var div model.LeagueDivision
	err := r.db.WithContext(ctx).Where("tier = ?", tier).First(&div).Error
	return &div, err
}

func (r *WeeklyLeagueRepository) GetLowestDivision(ctx context.Context) (*model.LeagueDivision, error) {
	var div model.LeagueDivision
	err := r.db.WithContext(ctx).Order("tier ASC").First(&div).Error
	return &div, err
}

// === Seasons ===

func (r *WeeklyLeagueRepository) GetActiveSeason(ctx context.Context) (*model.LeagueSeason, error) {
	var season model.LeagueSeason
	err := r.db.WithContext(ctx).Where("is_active = true").First(&season).Error
	return &season, err
}

func (r *WeeklyLeagueRepository) CreateSeason(ctx context.Context, season *model.LeagueSeason) error {
	return r.db.WithContext(ctx).Create(season).Error
}

func (r *WeeklyLeagueRepository) UpdateSeason(ctx context.Context, season *model.LeagueSeason) error {
	return r.db.WithContext(ctx).Save(season).Error
}

// === Participants ===

func (r *WeeklyLeagueRepository) GetParticipant(ctx context.Context, seasonID uuid.UUID, userID uint) (*model.LeagueParticipant, error) {
	var p model.LeagueParticipant
	err := r.db.WithContext(ctx).
		Preload("Division").
		Where("season_id = ? AND user_id = ?", seasonID, userID).
		First(&p).Error
	return &p, err
}

func (r *WeeklyLeagueRepository) CreateParticipant(ctx context.Context, p *model.LeagueParticipant) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *WeeklyLeagueRepository) AddWeeklyXP(ctx context.Context, seasonID uuid.UUID, userID uint, xp int64) error {
	return r.db.WithContext(ctx).
		Model(&model.LeagueParticipant{}).
		Where("season_id = ? AND user_id = ?", seasonID, userID).
		UpdateColumn("weekly_xp", gorm.Expr("weekly_xp + ?", xp)).Error
}

func (r *WeeklyLeagueRepository) GetLeaderboard(ctx context.Context, seasonID uuid.UUID, divisionID int, limit int) ([]model.LeagueParticipant, error) {
	var participants []model.LeagueParticipant
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("season_id = ? AND division_id = ?", seasonID, divisionID).
		Order("weekly_xp DESC").
		Limit(limit).
		Find(&participants).Error
	return participants, err
}

func (r *WeeklyLeagueRepository) GetParticipantsByDivision(ctx context.Context, seasonID uuid.UUID, divisionID int) ([]model.LeagueParticipant, error) {
	var participants []model.LeagueParticipant
	err := r.db.WithContext(ctx).
		Where("season_id = ? AND division_id = ?", seasonID, divisionID).
		Order("weekly_xp DESC").
		Find(&participants).Error
	return participants, err
}

func (r *WeeklyLeagueRepository) UpdateParticipant(ctx context.Context, p *model.LeagueParticipant) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *WeeklyLeagueRepository) CountParticipantsInDivision(ctx context.Context, seasonID uuid.UUID, divisionID int) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.LeagueParticipant{}).
		Where("season_id = ? AND division_id = ?", seasonID, divisionID).
		Count(&count).Error
	return count, err
}
