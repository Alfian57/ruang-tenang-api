package service

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
)

type LevelConfigService struct {
	levelConfigRepo *repository.LevelConfigRepository
	cacheService    *CacheService
}

func NewLevelConfigService(levelConfigRepo *repository.LevelConfigRepository, cacheService *CacheService) *LevelConfigService {
	return &LevelConfigService{
		levelConfigRepo: levelConfigRepo,
		cacheService:    cacheService,
	}
}

func (s *LevelConfigService) GetAll() ([]model.LevelConfig, error) {
	// Check cache first
	if cached := s.cacheService.Get(CacheKeyLevelConfigs); cached != nil {
		return cached.([]model.LevelConfig), nil
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

func (s *LevelConfigService) GetByID(id uint) (*model.LevelConfig, error) {
	return s.levelConfigRepo.GetByID(id)
}

func (s *LevelConfigService) GetLevelByExp(exp int64) (*model.LevelConfig, error) {
	return s.levelConfigRepo.GetLevelByExp(exp)
}

func (s *LevelConfigService) GetNextLevel(currentLevel int) (*model.LevelConfig, error) {
	return s.levelConfigRepo.GetNextLevel(currentLevel)
}

func (s *LevelConfigService) Create(config *model.LevelConfig) error {
	if s.levelConfigRepo.ExistsByLevel(config.Level) {
		return ErrLevelExists
	}
	err := s.levelConfigRepo.Create(config)
	if err == nil {
		s.invalidateLevelCache()
	}
	return err
}

func (s *LevelConfigService) Update(id uint, config *model.LevelConfig) error {
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
func (s *LevelConfigService) GetUserLevelInfo(exp int64) (*model.LevelConfig, *model.LevelConfig, error) {
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(exp)
	if err != nil {
		// Return default if no config found
		return &model.LevelConfig{
			Level:     1,
			MinExp:    0,
			BadgeName: "Pemula",
			BadgeIcon: "🌱",
		}, nil, nil
	}

	nextLevel, _ := s.levelConfigRepo.GetNextLevel(currentLevel.Level)
	return currentLevel, nextLevel, nil
}
