package services

import (
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

type LevelConfigService struct {
	levelConfigRepo *repositories.LevelConfigRepository
	cacheService    *CacheService
}

func NewLevelConfigService(levelConfigRepo *repositories.LevelConfigRepository, cacheService *CacheService) *LevelConfigService {
	return &LevelConfigService{
		levelConfigRepo: levelConfigRepo,
		cacheService:    cacheService,
	}
}

func (s *LevelConfigService) GetAll() ([]models.LevelConfig, error) {
	// Check cache first
	if cached := s.cacheService.Get(CacheKeyLevelConfigs); cached != nil {
		return cached.([]models.LevelConfig), nil
	}

	// Fetch from database
	configs, err := s.levelConfigRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cacheService.SetWithTTL(CacheKeyLevelConfigs, configs, s.cacheService.LevelConfigTTL)
	return configs, nil
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
	err := s.levelConfigRepo.Create(config)
	if err == nil {
		s.invalidateLevelCache()
	}
	return err
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

	err = s.levelConfigRepo.Update(existing)
	if err == nil {
		s.invalidateLevelCache()
	}
	return err
}

func (s *LevelConfigService) Delete(id uint) error {
	err := s.levelConfigRepo.Delete(id)
	if err == nil {
		s.invalidateLevelCache()
	}
	return err
}

// invalidateLevelCache clears all level-related caches
func (s *LevelConfigService) invalidateLevelCache() {
	s.cacheService.Delete(CacheKeyLevelConfigs)
	s.cacheService.DeletePrefix(CacheKeyLevelByExp)
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
