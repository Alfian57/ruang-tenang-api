package infrastructure

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushSubscriptionRepository struct {
	db *gorm.DB
}

func NewPushSubscriptionRepository(db *gorm.DB) *PushSubscriptionRepository {
	return &PushSubscriptionRepository{db: db}
}

func (r *PushSubscriptionRepository) Upsert(ctx context.Context, sub *model.PushSubscription) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "endpoint"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "p256dh", "auth"}),
		}).
		Create(sub).Error
}

func (r *PushSubscriptionRepository) DeleteByEndpoint(ctx context.Context, endpoint string, userID uint) error {
	return r.db.WithContext(ctx).
		Where("endpoint = ? AND user_id = ?", endpoint, userID).
		Delete(&model.PushSubscription{}).Error
}

func (r *PushSubscriptionRepository) GetByUserID(ctx context.Context, userID uint) ([]model.PushSubscription, error) {
	var subs []model.PushSubscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func (r *PushSubscriptionRepository) DeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.PushSubscription{}).Error
}

func (r *PushSubscriptionRepository) GetAll(ctx context.Context) ([]model.PushSubscription, error) {
	var subs []model.PushSubscription
	err := r.db.WithContext(ctx).Find(&subs).Error
	return subs, err
}
