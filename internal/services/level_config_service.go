package services

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type LevelConfigService struct {
	levelConfigRepo *repositories.LevelConfigRepository
}

func NewLevelConfigService(levelConfigRepo *repositories.LevelConfigRepository) *LevelConfigService {
	return &LevelConfigService{levelConfigRepo: levelConfigRepo}
}

func (s *LevelConfigService) GetAll() ([]models.LevelConfig, error) {
	return s.levelConfigRepo.GetAll()
}

func (s *LevelConfigService) GetByID(id uint) (*models.LevelConfig, error) {
	return s.levelConfigRepo.GetByID(id)
}

func (s *LevelConfigService) GetLevelByExp(exp int64) (*models.LevelConfig, error) {
	return s.levelConfigRepo.GetLevelByExp(exp)
}

func (s *LevelConfigService) GetNextLevel(currentLevel int) (*models.LevelConfig, error) {
	return s.levelConfigRepo.GetNextLevel(currentLevel)
}

func (s *LevelConfigService) Create(config *models.LevelConfig) error {
	if s.levelConfigRepo.ExistsByLevel(config.Level) {
		return ErrLevelExists
	}
	return s.levelConfigRepo.Create(config)
}

func (s *LevelConfigService) Update(id uint, config *models.LevelConfig) error {
	existing, err := s.levelConfigRepo.GetByID(id)
	if err != nil {
		return err
	}

	if config.Level != existing.Level && s.levelConfigRepo.ExistsByLevelExcept(config.Level, id) {
		return ErrLevelExists
	}

	existing.Level = config.Level
	existing.MinExp = config.MinExp
	existing.BadgeName = config.BadgeName
	existing.BadgeIcon = config.BadgeIcon

	return s.levelConfigRepo.Update(existing)
}

func (s *LevelConfigService) Delete(id uint) error {
	return s.levelConfigRepo.Delete(id)
}

// GetUserLevelInfo returns level information for a user based on their exp
func (s *LevelConfigService) GetUserLevelInfo(exp int64) (*models.LevelConfig, *models.LevelConfig, error) {
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(exp)
	if err != nil {
		// Return default if no config found
		return &models.LevelConfig{
			Level:     1,
			MinExp:    0,
			BadgeName: "Pemula",
			BadgeIcon: "🌱",
		}, nil, nil
	}

	nextLevel, _ := s.levelConfigRepo.GetNextLevel(currentLevel.Level)
	return currentLevel, nextLevel, nil
}
