package repository

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type SongCategoryRepository struct {
	db *gorm.DB
}

func NewSongCategoryRepository(db *gorm.DB) *SongCategoryRepository {
	return &SongCategoryRepository{db: db}
}

func (r *SongCategoryRepository) FindAll() ([]model.SongCategory, error) {
	var categories []model.SongCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *SongCategoryRepository) FindAllWithSongCount() ([]model.SongCategory, error) {
	var categories []model.SongCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *SongCategoryRepository) FindByID(id uint) (*model.SongCategory, error) {
	var category model.SongCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *SongCategoryRepository) Create(category *model.SongCategory) error {
	return r.db.Create(category).Error
}

// SongRepository
type SongRepository struct {
	db *gorm.DB
}

func NewSongRepository(db *gorm.DB) *SongRepository {
	return &SongRepository{db: db}
}

func (r *SongRepository) FindByCategoryID(categoryID uint) ([]model.Song, error) {
	var songs []model.Song
	err := r.db.Where("song_category_id = ?", categoryID).Order("title ASC").Find(&songs).Error
	return songs, err
}

func (r *SongRepository) FindByID(id uint) (*model.Song, error) {
	var song model.Song
	err := r.db.Preload("Category").First(&song, id).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) FindAll() ([]model.Song, error) {
	var songs []model.Song
	err := r.db.Preload("Category").Order("title ASC").Find(&songs).Error
	return songs, err
}

func (r *SongRepository) Create(song *model.Song) error {
	return r.db.Create(song).Error
}

func (r *SongRepository) CountByCategoryID(categoryID uint) int64 {
	var count int64
	r.db.Model(&model.Song{}).Where("song_category_id = ?", categoryID).Count(&count)
	return count
}

func (r *SongRepository) Search(query string) ([]model.Song, error) {
	var songs []model.Song
	err := r.db.Preload("Category").
		Where("title ILIKE ?", "%"+query+"%").
		Or("song_categories.name ILIKE ?", "%"+query+"%").
		Joins("JOIN song_categories ON song_categories.id = songs.song_category_id").
		Limit(5).
		Find(&songs).Error
	return songs, err
}
