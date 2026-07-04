package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// FindAll retrieves articles with optional filters
func (r *ArticleRepository) FindAll(ctx context.Context, categoryID uint, search string, page, limit int, status string, userID uint) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Article{}).Preload("Category").Preload("Author")

	if categoryID > 0 {
		query = query.Where("article_category_id = ?", categoryID)
	}

	if search != "" {
		if r.db.Dialector.Name() == "sqlite" {
			query = query.Where("title LIKE ?", "%"+search+"%")
		} else {
			query = query.Where("title ILIKE ?", "%"+search+"%")
		}
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// Get total count
	query.Count(&total)

	// Pagination
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&articles).Error

	return articles, total, err
}

// FindPublished retrieves only published articles for public view
func (r *ArticleRepository) FindPublished(ctx context.Context, categoryID uint, search string, page, limit int) ([]model.Article, int64, error) {
	return r.FindAll(ctx, categoryID, search, page, limit, string(model.ArticleStatusPublished), 0)
}

// FindByUserID retrieves articles by user ID (for user's own articles)
func (r *ArticleRepository) FindByUserID(ctx context.Context, userID uint, page, limit int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Article{}).
		Preload("Category").
		Where("user_id = ?", userID)

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&articles).Error

	return articles, total, err
}

func (r *ArticleRepository) FindByID(ctx context.Context, id uint) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) FindBySlug(ctx context.Context, slug string) (*model.Article, error) {
	var article model.Article
	err := r.db.WithContext(ctx).Preload("Category").Preload("Author").Where("slug = ?", slug).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) Create(ctx context.Context, article *model.Article) error {
	return r.db.WithContext(ctx).Create(article).Error
}

func (r *ArticleRepository) Update(ctx context.Context, article *model.Article) error {
	return r.db.WithContext(ctx).Save(article).Error
}

func (r *ArticleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Article{}, id).Error
}

// UpdateStatus updates the status of an article
func (r *ArticleRepository) UpdateStatus(ctx context.Context, id uint, status model.ArticleStatus) error {
	return r.db.WithContext(ctx).Model(&model.Article{}).Where("id = ?", id).Update("status", status).Error
}

// FindUpdatedSince retrieves articles updated since the given time (for incremental cache sync)
func (r *ArticleRepository) FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Article, error) {
	var articles []model.Article
	err := r.db.WithContext(ctx).Model(&model.Article{}).
		Preload("Category").
		Where("updated_at > ?", since).
		Find(&articles).Error
	return articles, err
}

// Category Repository
type ArticleCategoryRepository struct {
	db *gorm.DB
}

func NewArticleCategoryRepository(db *gorm.DB) *ArticleCategoryRepository {
	return &ArticleCategoryRepository{db: db}
}

func (r *ArticleCategoryRepository) FindAll(ctx context.Context) ([]model.ArticleCategory, error) {
	var categories []model.ArticleCategory
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *ArticleCategoryRepository) FindByID(ctx context.Context, id uint) (*model.ArticleCategory, error) {
	var category model.ArticleCategory
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ArticleCategoryRepository) Create(ctx context.Context, category *model.ArticleCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}
