package repositories

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// FindAll retrieves articles with optional filters
func (r *ArticleRepository) FindAll(categoryID uint, search string, page, limit int, status string, userID uint) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).Preload("Category").Preload("Author")

	if categoryID > 0 {
		query = query.Where("article_category_id = ?", categoryID)
	}

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
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
func (r *ArticleRepository) FindPublished(categoryID uint, search string, page, limit int) ([]models.Article, int64, error) {
	return r.FindAll(categoryID, search, page, limit, string(models.ArticleStatusPublished), 0)
}

// FindByUserID retrieves articles by user ID (for user's own articles)
func (r *ArticleRepository) FindByUserID(userID uint, page, limit int) ([]models.Article, int64, error) {
	var articles []models.Article
	var total int64

	query := r.db.Model(&models.Article{}).
		Preload("Category").
		Where("user_id = ?", userID)

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&articles).Error

	return articles, total, err
}

func (r *ArticleRepository) FindByID(id uint) (*models.Article, error) {
	var article models.Article
	err := r.db.Preload("Category").Preload("Author").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}

func (r *ArticleRepository) Update(article *models.Article) error {
	return r.db.Save(article).Error
}

func (r *ArticleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Article{}, id).Error
}

// UpdateStatus updates the status of an article
func (r *ArticleRepository) UpdateStatus(id uint, status models.ArticleStatus) error {
	return r.db.Model(&models.Article{}).Where("id = ?", id).Update("status", status).Error
}

// Category Repository
type ArticleCategoryRepository struct {
	db *gorm.DB
}

func NewArticleCategoryRepository(db *gorm.DB) *ArticleCategoryRepository {
	return &ArticleCategoryRepository{db: db}
}

func (r *ArticleCategoryRepository) FindAll() ([]models.ArticleCategory, error) {
	var categories []models.ArticleCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *ArticleCategoryRepository) FindByID(id uint) (*models.ArticleCategory, error) {
	var category models.ArticleCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ArticleCategoryRepository) Create(category *models.ArticleCategory) error {
	return r.db.Create(category).Error
}
