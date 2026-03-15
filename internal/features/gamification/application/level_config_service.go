package application

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/serviceerror"

	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/internal/features/gamification/infrastructure"
)

type LevelConfigService struct {
	levelConfigRepo *infrastructure.LevelConfigRepository
	cacheService    *cache.CacheService
}

func NewLevelConfigService(levelConfigRepo *infrastructure.LevelConfigRepository, cacheService *cache.CacheService) *LevelConfigService {
	return &LevelConfigService{
		levelConfigRepo: levelConfigRepo,
		cacheService:    cacheService,
	}
}

func (s *LevelConfigService) GetAll(ctx context.Context) ([]model.LevelConfig, error) {
	// Check cache first
	if cached := s.cacheService.Get(cache.CacheKeyLevelConfigs); cached != nil {
		return cached.([]model.LevelConfig), nil
	}

	// Fetch from database
	configs, err := s.levelConfigRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Store in cache
	s.cacheService.SetWithTTL(cache.CacheKeyLevelConfigs, configs, s.cacheService.LevelConfigTTL)
	return configs, nil
}

func (s *LevelConfigService) GetByID(ctx context.Context, id uint) (*model.LevelConfig, error) {
	return s.levelConfigRepo.GetByID(ctx, id)
}

func (s *LevelConfigService) GetLevelByExp(ctx context.Context, exp int64) (*model.LevelConfig, error) {
	return s.levelConfigRepo.GetLevelByExp(ctx, exp)
}

func (s *LevelConfigService) GetNextLevel(ctx context.Context, currentLevel int) (*model.LevelConfig, error) {
	return s.levelConfigRepo.GetNextLevel(ctx, currentLevel)
}

func (s *LevelConfigService) Create(ctx context.Context, config *model.LevelConfig) error {
	if s.levelConfigRepo.ExistsByLevel(ctx, config.Level) {
		return serviceerror.ErrLevelExists
	}
	err := s.levelConfigRepo.Create(ctx, config)
	if err == nil {
		s.invalidateLevelCache(ctx)
	}
	return err
}

func (s *LevelConfigService) Update(ctx context.Context, id uint, config *model.LevelConfig) error {
	existing, err := s.levelConfigRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if config.Level != existing.Level && s.levelConfigRepo.ExistsByLevelExcept(ctx, config.Level, id) {
		return serviceerror.ErrLevelExists
	}

	existing.Level = config.Level
	existing.MinExp = config.MinExp
	existing.BadgeName = config.BadgeName
	existing.TaskDescription = config.TaskDescription
	if config.BadgeIcon != "" {
		existing.BadgeIcon = config.BadgeIcon
	}

	err = s.levelConfigRepo.Update(ctx, existing)
	if err == nil {
		s.invalidateLevelCache(ctx)
	}
	return err
}

func (s *LevelConfigService) Delete(ctx context.Context, id uint) error {
	err := s.levelConfigRepo.Delete(ctx, id)
	if err == nil {
		s.invalidateLevelCache(ctx)
	}
	return err
}

// invalidateLevelCache clears all level-related caches
func (s *LevelConfigService) invalidateLevelCache(ctx context.Context) {
	s.cacheService.Delete(cache.CacheKeyLevelConfigs)
	s.cacheService.DeletePrefix(cache.CacheKeyLevelByExp)
}

// GetUserLevelInfo returns level information for a user based on their exp
func (s *LevelConfigService) GetUserLevelInfo(ctx context.Context, exp int64) (*model.LevelConfig, *model.LevelConfig, error) {
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(ctx, exp)
	if err != nil {
		// Return default if no config found
		return &model.LevelConfig{
			Level:     1,
			MinExp:    0,
			BadgeName: "Pemula",
			BadgeIcon: "",
		}, nil, nil
	}

	nextLevel, _ := s.levelConfigRepo.GetNextLevel(ctx, currentLevel.Level)
	return currentLevel, nextLevel, nil
}
