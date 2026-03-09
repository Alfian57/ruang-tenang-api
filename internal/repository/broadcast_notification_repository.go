package repository

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type BroadcastNotificationRepository struct {
	db *gorm.DB
}

func NewBroadcastNotificationRepository(db *gorm.DB) *BroadcastNotificationRepository {
	return &BroadcastNotificationRepository{db: db}
}

func (r *BroadcastNotificationRepository) Create(ctx context.Context, b *model.BroadcastNotification) error {
	return r.db.WithContext(ctx).Create(b).Error
}

func (r *BroadcastNotificationRepository) GetByID(ctx context.Context, id string) (*model.BroadcastNotification, error) {
	var b model.BroadcastNotification
	err := r.db.WithContext(ctx).Preload("Creator").Where("id = ?", id).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *BroadcastNotificationRepository) GetAll(ctx context.Context, page, limit int, search string) ([]model.BroadcastNotification, int64, error) {
	var broadcasts []model.BroadcastNotification
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BroadcastNotification{})

	if search != "" {
		query = query.Where("title ILIKE ? OR body ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Creator").Order("created_at DESC").Offset(offset).Limit(limit).Find(&broadcasts).Error
	return broadcasts, total, err
}

func (r *BroadcastNotificationRepository) Update(ctx context.Context, b *model.BroadcastNotification) error {
	return r.db.WithContext(ctx).Save(b).Error
}

func (r *BroadcastNotificationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.BroadcastNotification{}).Error
}

func (r *BroadcastNotificationRepository) GetScheduledDue(ctx context.Context) ([]model.BroadcastNotification, error) {
	var broadcasts []model.BroadcastNotification
	err := r.db.WithContext(ctx).
		Where("status = ? AND scheduled_at <= NOW()", model.BroadcastStatusScheduled).
		Find(&broadcasts).Error
	return broadcasts, err
}
