package repositories

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

type ForumCategoryRepository interface {
	Create(category *models.ForumCategory) error
	FindAll() ([]models.ForumCategory, error)
	FindByID(id uint) (*models.ForumCategory, error)
	Update(category *models.ForumCategory) error
	Delete(id uint) error
}

type forumCategoryRepository struct {
	db *gorm.DB
}

func NewForumCategoryRepository(db *gorm.DB) ForumCategoryRepository {
	return &forumCategoryRepository{db}
}

func (r *forumCategoryRepository) Create(category *models.ForumCategory) error {
	return r.db.Create(category).Error
}

func (r *forumCategoryRepository) FindAll() ([]models.ForumCategory, error) {
	var categories []models.ForumCategory
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *forumCategoryRepository) FindByID(id uint) (*models.ForumCategory, error) {
	var category models.ForumCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *forumCategoryRepository) Update(category *models.ForumCategory) error {
	return r.db.Save(category).Error
}

func (r *forumCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.ForumCategory{}, id).Error
}
