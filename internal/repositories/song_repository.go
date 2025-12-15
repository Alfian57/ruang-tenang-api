package repositories

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

type SongCategoryRepository struct {
	db *gorm.DB
}

func NewSongCategoryRepository(db *gorm.DB) *SongCategoryRepository {
	return &SongCategoryRepository{db: db}
}

func (r *SongCategoryRepository) FindAll() ([]models.SongCategory, error) {
	var categories []models.SongCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *SongCategoryRepository) FindAllWithSongCount() ([]models.SongCategory, error) {
	var categories []models.SongCategory
	err := r.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (r *SongCategoryRepository) FindByID(id uint) (*models.SongCategory, error) {
	var category models.SongCategory
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *SongCategoryRepository) Create(category *models.SongCategory) error {
	return r.db.Create(category).Error
}

// SongRepository
type SongRepository struct {
	db *gorm.DB
}

func NewSongRepository(db *gorm.DB) *SongRepository {
	return &SongRepository{db: db}
}

func (r *SongRepository) FindByCategoryID(categoryID uint) ([]models.Song, error) {
	var songs []models.Song
	err := r.db.Where("song_category_id = ?", categoryID).Order("title ASC").Find(&songs).Error
	return songs, err
}

func (r *SongRepository) FindByID(id uint) (*models.Song, error) {
	var song models.Song
	err := r.db.Preload("Category").First(&song, id).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SongRepository) FindAll() ([]models.Song, error) {
	var songs []models.Song
	err := r.db.Preload("Category").Order("title ASC").Find(&songs).Error
	return songs, err
}

func (r *SongRepository) Create(song *models.Song) error {
	return r.db.Create(song).Error
}

func (r *SongRepository) CountByCategoryID(categoryID uint) int64 {
	var count int64
	r.db.Model(&models.Song{}).Where("song_category_id = ?", categoryID).Count(&count)
	return count
}

func (r *SongRepository) Search(query string) ([]models.Song, error) {
	var songs []models.Song
	err := r.db.Preload("Category").
		Where("title ILIKE ?", "%"+query+"%").
		Or("song_categories.name ILIKE ?", "%"+query+"%").
		Joins("JOIN song_categories ON song_categories.id = songs.song_category_id").
		Limit(5).
		Find(&songs).Error
	return songs, err
}
