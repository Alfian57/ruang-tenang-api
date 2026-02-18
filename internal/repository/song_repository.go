package repository

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
	err := r.db.WithContext(ctx).Preload("Category").
		Where("title ILIKE ?", "%"+query+"%").
		Or("song_categories.name ILIKE ?", "%"+query+"%").
		Joins("JOIN song_categories ON song_categories.id = songs.song_category_id").
		Limit(5).
		Find(&songs).Error
	return songs, err
}
