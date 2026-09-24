package infrastructure

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type SongCategoryRepository struct {
	db *gorm.DB
}

func NewSongCategoryRepository(db *gorm.DB) *SongCategoryRepository {
	return &SongCategoryRepository{db: db}
}

func (r *SongCategoryRepository) FindAll(ctx context.Context) ([]model.SongCategory, error) {
	var categories []model.SongCategory
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *SongCategoryRepository) FindPage(ctx context.Context, page, limit int) ([]model.SongCategory, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.SongCategory{})
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var categories []model.SongCategory
	err := query.Order("name ASC").Limit(limit).Offset((page - 1) * limit).Find(&categories).Error
	return categories, total, err
}

func (r *SongCategoryRepository) FindAllWithSongCount(ctx context.Context) ([]model.SongCategory, error) {
	var categories []model.SongCategory
	err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *SongCategoryRepository) FindByID(ctx context.Context, id uint) (*model.SongCategory, error) {
	var category model.SongCategory
	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *SongCategoryRepository) Create(ctx context.Context, category *model.SongCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *SongCategoryRepository) FindBySlug(ctx context.Context, slug string) (*model.SongCategory, error) {
	var category model.SongCategory
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// SongRepository
type SongRepository struct {
	db *gorm.DB
}

func NewSongRepository(db *gorm.DB) *SongRepository {
	return &SongRepository{db: db}
}

func (r *SongRepository) FindByCategoryID(ctx context.Context, categoryID uint) ([]model.Song, error) {
	var songs []model.Song
	err := r.db.WithContext(ctx).Where("song_category_id = ?", categoryID).Order("title ASC").Find(&songs).Error
	return songs, err
}

func (r *SongRepository) FindByID(ctx context.Context, id uint) (*model.Song, error) {
	var song model.Song
	err := r.db.WithContext(ctx).Preload("Category").First(&song, id).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) FindAll(ctx context.Context) ([]model.Song, error) {
	var songs []model.Song
	err := r.db.WithContext(ctx).Preload("Category").Order("title ASC").Find(&songs).Error
	return songs, err
}

func (r *SongRepository) Create(ctx context.Context, song *model.Song) error {
	return r.db.WithContext(ctx).Create(song).Error
}

func (r *SongRepository) FindBySlug(ctx context.Context, slug string) (*model.Song, error) {
	var song model.Song
	err := r.db.WithContext(ctx).Preload("Category").Where("slug = ?", slug).First(&song).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) CountByCategoryID(ctx context.Context, categoryID uint) int64 {
	var count int64
	r.db.WithContext(ctx).Model(&model.Song{}).Where("song_category_id = ?", categoryID).Count(&count)
	return count
}

func (r *SongRepository) Search(ctx context.Context, query string) ([]model.Song, error) {
	var songs []model.Song
	searchTerm := "%" + query + "%"
	queryBuilder := r.db.WithContext(ctx).Preload("Category").
		Joins("JOIN song_categories ON song_categories.id = songs.song_category_id")

	if r.db.Dialector.Name() == "sqlite" {
		queryBuilder = queryBuilder.Where("songs.title LIKE ?", searchTerm).
			Or("song_categories.name LIKE ?", searchTerm)
	} else {
		queryBuilder = queryBuilder.Where("songs.title ILIKE ?", searchTerm).
			Or("song_categories.name ILIKE ?", searchTerm)
	}

	err := queryBuilder.
		Limit(5).
		Find(&songs).Error
	return songs, err
}

func (r *SongRepository) SearchPage(ctx context.Context, search string, page, limit int) ([]model.Song, int64, error) {
	pattern := "%" + search + "%"
	operator := "ILIKE"
	if r.db.Dialector.Name() == "sqlite" {
		operator = "LIKE"
	}
	query := r.db.WithContext(ctx).Model(&model.Song{}).Joins("JOIN song_categories ON song_categories.id = songs.song_category_id").Where("songs.title "+operator+" ? OR song_categories.name "+operator+" ?", pattern, pattern)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var songs []model.Song
	err := query.Preload("Category").Order("songs.title ASC").Limit(limit).Offset((page - 1) * limit).Find(&songs).Error
	return songs, total, err
}
