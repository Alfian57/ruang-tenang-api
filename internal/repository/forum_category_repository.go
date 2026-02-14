package repository

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type ForumCategoryRepository interface {
	Create(category *model.ForumCategory) error
	FindAll() ([]model.ForumCategory, error)
	FindByID(id uint) (*model.ForumCategory, error)
	Update(category *model.ForumCategory) error
	Delete(id uint) error
}

type forumCategoryRepository struct {
	db *gorm.DB
}

func NewForumCategoryRepository(db *gorm.DB) ForumCategoryRepository {
	return &forumCategoryRepository{db}
}

func (r *forumCategoryRepository) Create(category *model.ForumCategory) error {
	return r.db.Create(category).Error
}

func (r *forumCategoryRepository) FindAll() ([]model.ForumCategory, error) {
	var categories []model.ForumCategory
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *forumCategoryRepository) FindByID(id uint) (*model.ForumCategory, error) {
	var category model.ForumCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *forumCategoryRepository) Update(category *model.ForumCategory) error {
	return r.db.Save(category).Error
}

func (r *forumCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&model.ForumCategory{}, id).Error
}
