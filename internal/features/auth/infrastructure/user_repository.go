package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (r *UserRepository) ExistsByEmailExcept(ctx context.Context, email string, exceptID uint) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.User{}).Where("email = ? AND id != ?", email, exceptID).Count(&count)
	return count > 0
}

func (r *UserRepository) GetTopUsers(ctx context.Context, limit int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).Order("exp desc").Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (r *UserRepository) FindByResetToken(ctx context.Context, token string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateResetToken(ctx context.Context, email string, token string, expiry time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": expiry,
	}).Error
}

func (r *UserRepository) ClearResetToken(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"reset_token":        nil,
		"reset_token_expiry": nil,
	}).Error
}

func (r *UserRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return &user, err
}

func (r *UserRepository) AddExp(ctx context.Context, userID uint, amount int64) error {
	if amount <= 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("exp", gorm.Expr("exp + ?", amount)).Error
}

func (r *UserRepository) AddGoldCoins(ctx context.Context, userID uint, amount int64) error {
	if amount <= 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		UpdateColumn("gold_coins", gorm.Expr("gold_coins + ?", amount)).Error
}

func (r *UserRepository) UpdateField(ctx context.Context, userID uint, field string, value interface{}) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update(field, value).Error
}
