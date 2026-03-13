package infrastructure

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
)

type TimedChallengeRepository struct {
	db *gorm.DB
}

func NewTimedChallengeRepository(db *gorm.DB) *TimedChallengeRepository {
	return &TimedChallengeRepository{db: db}
}

// === Templates ===

func (r *TimedChallengeRepository) GetActiveTemplates(ctx context.Context) ([]model.TimedChallengeTemplate, error) {
	var templates []model.TimedChallengeTemplate
	err := r.db.WithContext(ctx).Where("is_active = true").Order("id ASC").Find(&templates).Error
	return templates, err
}

func (r *TimedChallengeRepository) GetTemplateByID(ctx context.Context, id int) (*model.TimedChallengeTemplate, error) {
	var t model.TimedChallengeTemplate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	return &t, err
}

// === User Challenges ===

func (r *TimedChallengeRepository) Create(ctx context.Context, challenge *model.UserTimedChallenge) error {
	return r.db.WithContext(ctx).Create(challenge).Error
}

func (r *TimedChallengeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.UserTimedChallenge, error) {
	var c model.UserTimedChallenge
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("id = ?", id).First(&c).Error
	return &c, err
}

func (r *TimedChallengeRepository) GetUserChallenges(ctx context.Context, userID uint, status string, page, limit int) ([]model.UserTimedChallenge, int64, error) {
	query := r.db.WithContext(ctx).
		Preload("Template").
		Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Model(&model.UserTimedChallenge{}).Count(&total)

	var challenges []model.UserTimedChallenge
	err := query.Order("started_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&challenges).Error

	return challenges, total, err
}

func (r *TimedChallengeRepository) GetActiveChallenge(ctx context.Context, userID uint) (*model.UserTimedChallenge, error) {
	var c model.UserTimedChallenge
	err := r.db.WithContext(ctx).
		Preload("Template").
		Where("user_id = ? AND status = 'active' AND expires_at > ?", userID, time.Now()).
		First(&c).Error
	return &c, err
}

func (r *TimedChallengeRepository) Update(ctx context.Context, challenge *model.UserTimedChallenge) error {
	return r.db.WithContext(ctx).Save(challenge).Error
}

func (r *TimedChallengeRepository) ExpireOverdueChallenges(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&model.UserTimedChallenge{}).
		Where("status = 'active' AND expires_at <= ?", time.Now()).
		UpdateColumn("status", "expired").Error
}

func (r *TimedChallengeRepository) UpdateProgress(ctx context.Context, challengeID uuid.UUID, value int) error {
	return r.db.WithContext(ctx).
		Model(&model.UserTimedChallenge{}).
		Where("id = ?", challengeID).
		UpdateColumn("current_value", gorm.Expr("current_value + ?", value)).Error
}
