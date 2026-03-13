package application

import (
	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/feature_unlock/infrastructure")

type FeatureUnlockService struct {
	featureRepo     *infrastructure.FeatureUnlockRepository
	levelConfigRepo LevelConfigRepository
	userRepo        *authinfra.UserRepository
}

func NewFeatureUnlockService(
	featureRepo *infrastructure.FeatureUnlockRepository,
	levelConfigRepo LevelConfigRepository,
	userRepo *authinfra.UserRepository,
) *FeatureUnlockService {
	return &FeatureUnlockService{
		featureRepo:     featureRepo,
		levelConfigRepo: levelConfigRepo,
		userRepo:        userRepo,
	}
}

// ==========================================
// Feature Definitions
// ==========================================

// GetAllFeatures returns all feature definitions grouped by level
func (s *FeatureUnlockService) GetAllFeatures(ctx context.Context) ([]dto.FeaturesByLevelResponse, error) {
	features, err := s.featureRepo.GetAllFeatureDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	// Group features by level
	levelMap := make(map[int][]model.FeatureDefinition)
	for _, f := range features {
		levelMap[f.RequiredLevel] = append(levelMap[f.RequiredLevel], f)
	}

	// Get level configs for tier info
	levelConfigs, _ := s.levelConfigRepo.GetAll(ctx)
	configMap := make(map[int]model.LevelConfig)
	for _, lc := range levelConfigs {
		configMap[lc.Level] = lc
	}

	var result []dto.FeaturesByLevelResponse
	for level := 1; level <= 10; level++ {
		if levelFeatures, ok := levelMap[level]; ok {
			config := configMap[level]

			featureResponses := make([]dto.FeatureUnlockResponse, len(levelFeatures))
			for i, f := range levelFeatures {
				featureResponses[i] = s.toFeatureResponse(ctx, f)
			}

			result = append(result, dto.FeaturesByLevelResponse{
				Level:     level,
				TierName:  config.TierName,
				TierColor: config.TierColor,
				Features:  featureResponses,
			})
		}
	}

	return result, nil
}

// GetFeaturesByCategory returns features grouped by category
func (s *FeatureUnlockService) GetFeaturesByCategory(ctx context.Context, category string) ([]dto.FeatureUnlockResponse, error) {
	features, err := s.featureRepo.GetFeaturesByCategory(ctx, category)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FeatureUnlockResponse, len(features))
	for i, f := range features {
		result[i] = s.toFeatureResponse(ctx, f)
	}

	return result, nil
}

// ==========================================
// User Feature Unlocks
// ==========================================

// GetUserFeatures returns user's feature unlock status
func (s *FeatureUnlockService) GetUserFeatures(ctx context.Context, userID uint) (*dto.UserFeaturesResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user's current level
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
	if err != nil {
		return nil, err
	}

	// Get all features
	allFeatures, err := s.featureRepo.GetAllFeatureDefinitions(ctx)
	if err != nil {
		return nil, err
	}

	// Get user's unlocked features
	userUnlocks, _ := s.featureRepo.GetUserUnlockedFeatures(ctx, userID)
	unlockedMap := make(map[uuid.UUID]bool)
	for _, u := range userUnlocks {
		unlockedMap[u.FeatureID] = true
	}

	var unlocked []dto.FeatureUnlockResponse
	var locked []dto.LockedFeatureResponse

	for _, f := range allFeatures {
		if unlockedMap[f.ID] {
			unlocked = append(unlocked, s.toFeatureResponse(ctx, f))
		} else {
			locked = append(locked, dto.LockedFeatureResponse{
				ID:            f.ID,
				FeatureKey:    f.FeatureKey,
				FeatureName:   f.FeatureName,
				Description:   f.Description,
				Icon:          f.Icon,
				Category:      f.Category,
				RequiredLevel: f.RequiredLevel,
				LevelsAway:    f.RequiredLevel - currentLevel.Level,
			})
		}
	}

	return &dto.UserFeaturesResponse{
		CurrentLevel:     currentLevel.Level,
		TotalUnlocked:    len(unlocked),
		TotalFeatures:    len(allFeatures),
		UnlockedFeatures: unlocked,
		LockedFeatures:   locked,
	}, nil
}

// CheckFeatureAccess checks if user has access to a specific feature
func (s *FeatureUnlockService) CheckFeatureAccess(ctx context.Context, userID uint, featureKey string) (*dto.FeatureAccessResponse, error) {
	// Check if feature exists
	feature, err := s.featureRepo.GetFeatureByKey(ctx, featureKey)
	if err != nil {
		return &dto.FeatureAccessResponse{
			HasAccess: false,
			Reason:    "Fitur tidak ditemukan",
		}, nil
	}

	// Get user's current level
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentLevel, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
	if err != nil {
		return nil, err
	}

	// Check if user has unlocked the feature (either by level or explicit unlock)
	isUnlocked := s.featureRepo.IsFeatureUnlockedByKey(ctx, userID, featureKey)

	// Also check if user's level is sufficient
	hasLevelAccess := currentLevel.Level >= feature.RequiredLevel

	if isUnlocked || hasLevelAccess {
		// If level is sufficient but not explicitly unlocked, unlock it
		if hasLevelAccess && !isUnlocked {
			s.featureRepo.UnlockFeature(ctx, userID, feature.ID)
		}

		return &dto.FeatureAccessResponse{
			HasAccess:   true,
			FeatureKey:  featureKey,
			FeatureName: feature.FeatureName,
		}, nil
	}

	return &dto.FeatureAccessResponse{
		HasAccess:     false,
		FeatureKey:    featureKey,
		FeatureName:   feature.FeatureName,
		RequiredLevel: feature.RequiredLevel,
		CurrentLevel:  currentLevel.Level,
		LevelsAway:    feature.RequiredLevel - currentLevel.Level,
		Reason:        "Level kamu belum cukup untuk mengakses fitur ini",
	}, nil
}

// UnlockFeaturesOnLevelUp unlocks all features for a user when they level up
func (s *FeatureUnlockService) UnlockFeaturesOnLevelUp(ctx context.Context, userID uint, newLevel int) ([]dto.FeatureUnlockResponse, error) {
	newFeatures, err := s.featureRepo.UnlockFeaturesForLevel(ctx, userID, newLevel)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FeatureUnlockResponse, len(newFeatures))
	for i, f := range newFeatures {
		result[i] = s.toFeatureResponse(ctx, f)
	}

	return result, nil
}

// GetUpcomingFeatures returns features user will unlock in upcoming levels
func (s *FeatureUnlockService) GetUpcomingFeatures(ctx context.Context, userID uint, limit int) ([]dto.LockedFeatureResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentLevel, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
	if err != nil {
		return nil, err
	}

	features, err := s.featureRepo.GetUpcomingFeatures(ctx, currentLevel.Level, limit)
	if err != nil {
		return nil, err
	}

	result := make([]dto.LockedFeatureResponse, len(features))
	for i, f := range features {
		result[i] = dto.LockedFeatureResponse{
			ID:            f.ID,
			FeatureKey:    f.FeatureKey,
			FeatureName:   f.FeatureName,
			Description:   f.Description,
			Icon:          f.Icon,
			Category:      f.Category,
			RequiredLevel: f.RequiredLevel,
			LevelsAway:    f.RequiredLevel - currentLevel.Level,
		}
	}

	return result, nil
}

// GetFeatureCategories returns all feature categories with counts
func (s *FeatureUnlockService) GetFeatureCategories(ctx context.Context) []dto.FeatureCategoryInfo {
	categories := []dto.FeatureCategoryInfo{
		{
			Key:         "profile",
			Name:        "Profile & Personalization",
			Description: "Fitur untuk mempersonalisasi profil kamu",
			Icon:        "user",
		},
		{
			Key:         "community",
			Name:        "Komunitas",
			Description: "Fitur untuk berinteraksi dengan komunitas",
			Icon:        "users",
		},
		{
			Key:         "content",
			Name:        "Konten",
			Description: "Fitur untuk membuat dan mengelola konten",
			Icon:        "edit",
		},
		{
			Key:         "ai",
			Name:        "AI & Assistant",
			Description: "Fitur AI untuk membantu perjalananmu",
			Icon:        "bot",
		},
		{
			Key:         "special",
			Name:        "Spesial",
			Description: "Fitur eksklusif untuk member senior",
			Icon:        "crown",
		},
	}

	return categories
}

// Helper function to convert model to DTO
func (s *FeatureUnlockService) toFeatureResponse(ctx context.Context, f model.FeatureDefinition) dto.FeatureUnlockResponse {
	return dto.FeatureUnlockResponse{
		ID:            f.ID,
		FeatureKey:    f.FeatureKey,
		FeatureName:   f.FeatureName,
		Description:   f.Description,
		Icon:          f.Icon,
		RequiredLevel: f.RequiredLevel,
		Category:      f.Category,
		IsUnlocked:    true,
	}
}
