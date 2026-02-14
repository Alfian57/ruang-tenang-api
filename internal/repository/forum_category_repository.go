package repository

import (
	"context"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type ForumCategoryRepository interface {
	Create(ctx context.Context, category *model.ForumCategory) error
	FindAll(ctx context.Context, ) ([]model.ForumCategory, error)
	FindByID(ctx context.Context, id uint) (*model.ForumCategory, error)
	Update(ctx context.Context, category *model.ForumCategory) error
	Delete(ctx context.Context, id uint) error
}

type forumCategoryRepository struct {
	db *gorm.DB
}

func NewForumCategoryRepository(db *gorm.DB) ForumCategoryRepository {
	return &forumCategoryRepository{db}
}

func (r *forumCategoryRepository) Create(ctx context.Context, category *model.ForumCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *forumCategoryRepository) FindAll(ctx context.Context) ([]model.ForumCategory, error) {
	var categories []model.ForumCategory
	err := r.db.WithContext(ctx).Find(&categories).Error
	return categories, err
}

func (r *forumCategoryRepository) FindByID(ctx context.Context, id uint) (*model.ForumCategory, error) {
	var category model.ForumCategory
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *forumCategoryRepository) Update(ctx context.Context, category *model.ForumCategory) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *forumCategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ForumCategory{}, id).Error
}
