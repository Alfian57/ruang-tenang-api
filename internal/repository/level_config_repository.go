package repository

import (
	"context"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type LevelConfigRepository struct {
	db *gorm.DB
}

func NewLevelConfigRepository(db *gorm.DB) *LevelConfigRepository {
	return &LevelConfigRepository{db: db}
}

func (r *LevelConfigRepository) GetAll(ctx context.Context) ([]model.LevelConfig, error) {
	var configs []model.LevelConfig
	err := r.db.WithContext(ctx).Order("level ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (r *LevelConfigRepository) GetByID(ctx context.Context, id uint) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.WithContext(ctx).First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *LevelConfigRepository) GetByLevel(ctx context.Context, level int) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.WithContext(ctx).Where("level = ?", level).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetLevelByExp returns the level config for a given exp amount
func (r *LevelConfigRepository) GetLevelByExp(ctx context.Context, exp int64) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.WithContext(ctx).Where("min_exp <= ?", exp).Order("min_exp DESC").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetNextLevel returns the next level config after current level
func (r *LevelConfigRepository) GetNextLevel(ctx context.Context, currentLevel int) (*model.LevelConfig, error) {
	var config model.LevelConfig
	err := r.db.WithContext(ctx).Where("level > ?", currentLevel).Order("level ASC").First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *LevelConfigRepository) Create(ctx context.Context, config *model.LevelConfig) error {
	return r.db.WithContext(ctx).Create(config).Error
}

func (r *LevelConfigRepository) Update(ctx context.Context, config *model.LevelConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *LevelConfigRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.LevelConfig{}, id).Error
}

func (r *LevelConfigRepository) ExistsByLevel(ctx context.Context, level int) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ?", level).Count(&count)
	return count > 0
}

func (r *LevelConfigRepository) ExistsByLevelExcept(ctx context.Context, level int, exceptID uint) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.LevelConfig{}).Where("level = ? AND id != ?", level, exceptID).Count(&count)
	return count > 0
}

func (r *LevelConfigRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.LevelConfig{}).Count(&count).Error
	return count, err
}
