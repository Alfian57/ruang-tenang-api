package repositories

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *UserRepository) ExistsByEmail(email string) bool {
	var count int64
	r.db.Model(&models.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

func (r *UserRepository) ExistsByEmailExcept(email string, exceptID uint) bool {
	var count int64
	r.db.Model(&models.User{}).Where("email = ? AND id != ?", email, exceptID).Count(&count)
	return count > 0
}

func (r *UserRepository) GetTopUsers(limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.Order("exp desc").Limit(limit).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (r *UserRepository) FindByResetToken(token string) (*models.User, error) {
	var user models.User
	err := r.db.Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateResetToken(email string, token string, expiry time.Time) error {
	return r.db.Model(&models.User{}).Where("email = ?", email).Updates(map[string]interface{}{
		"reset_token":        token,
		"reset_token_expiry": expiry,
	}).Error
}

func (r *UserRepository) ClearResetToken(userID uint) error {
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"reset_token":        nil,
		"reset_token_expiry": nil,
	}).Error
}
