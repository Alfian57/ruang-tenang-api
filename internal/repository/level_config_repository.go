package repository

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type LevelConfigRepository struct {
	db *gorm.DB
}

func NewLevelConfigRepository(db *gorm.DB) *LevelConfigRepository {
	return &LevelConfigRepository{db: db}
}

func (r *LevelConfigRepository) GetAll() ([]model.LevelConfig, error) {
	var configs []model.LevelConfig
	err := r.db.Order("level ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *LevelConfigRepository) GetByID(id uint) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *LevelConfigRepository) GetByLevel(level int) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.Where("level = ?", level).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetLevelByExp returns the level config for a given exp amount
func (r *LevelConfigRepository) GetLevelByExp(exp int64) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.Where("min_exp <= ?", exp).Order("min_exp DESC").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetNextLevel returns the next level config after current level
func (r *LevelConfigRepository) GetNextLevel(currentLevel int) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.Where("level > ?", currentLevel).Order("level ASC").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *LevelConfigRepository) Create(config *model.LevelConfig) error {
	return r.db.Create(config).Error
}

func (r *LevelConfigRepository) Update(config *model.LevelConfig) error {
	return r.db.Save(config).Error
}

func (r *LevelConfigRepository) Delete(id uint) error {
	return r.db.Delete(&model.LevelConfig{}, id).Error
}

func (r *LevelConfigRepository) ExistsByLevel(level int) bool {
	var count int64
	r.db.Model(&model.LevelConfig{}).Where("level = ?", level).Count(&count)
	return count > 0
}

func (r *LevelConfigRepository) ExistsByLevelExcept(level int, exceptID uint) bool {
	var count int64
	r.db.Model(&model.LevelConfig{}).Where("level = ? AND id != ?", level, exceptID).Count(&count)
	return count > 0
}

func (r *LevelConfigRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.LevelConfig{}).Count(&count).Error
	return count, err
}
