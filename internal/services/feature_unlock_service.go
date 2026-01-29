package services

import (
	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
	"github.com/google/uuid"
)

type FeatureUnlockService struct {
	featureRepo     *repositories.FeatureUnlockRepository
	levelConfigRepo *repositories.LevelConfigRepository
	userRepo        *repositories.UserRepository
}

func NewFeatureUnlockService(
	featureRepo *repositories.FeatureUnlockRepository,
	levelConfigRepo *repositories.LevelConfigRepository,
	userRepo *repositories.UserRepository,
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
func (s *FeatureUnlockService) GetAllFeatures() ([]dto.FeaturesByLevelResponse, error) {
	features, err := s.featureRepo.GetAllFeatureDefinitions()
	if err != nil {
		return nil, err
	}

	// Group features by level
	levelMap := make(map[int][]models.FeatureDefinition)
	for _, f := range features {
		levelMap[f.RequiredLevel] = append(levelMap[f.RequiredLevel], f)
	}

	// Get level configs for tier info
	levelConfigs, _ := s.levelConfigRepo.GetAll()
	configMap := make(map[int]models.LevelConfig)
	for _, lc := range levelConfigs {
		configMap[lc.Level] = lc
	}

	var result []dto.FeaturesByLevelResponse
	for level := 1; level <= 10; level++ {
		if levelFeatures, ok := levelMap[level]; ok {
			config := configMap[level]

			featureResponses := make([]dto.FeatureUnlockResponse, len(levelFeatures))
			for i, f := range levelFeatures {
				featureResponses[i] = s.toFeatureResponse(f)
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
func (s *FeatureUnlockService) GetFeaturesByCategory(category string) ([]dto.FeatureUnlockResponse, error) {
	features, err := s.featureRepo.GetFeaturesByCategory(category)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FeatureUnlockResponse, len(features))
	for i, f := range features {
		result[i] = s.toFeatureResponse(f)
	}

	return result, nil
}

// ==========================================
// User Feature Unlocks
// ==========================================

// GetUserFeatures returns user's feature unlock status
func (s *FeatureUnlockService) GetUserFeatures(userID uint) (*dto.UserFeaturesResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Get user's current level
	currentLevel, err := s.levelConfigRepo.GetLevelByExp(user.Exp)
	if err != nil {
		return nil, err
	}

	// Get all features
	allFeatures, err := s.featureRepo.GetAllFeatureDefinitions()
	if err != nil {
		return nil, err
	}

	// Get user's unlocked features
	userUnlocks, _ := s.featureRepo.GetUserUnlockedFeatures(userID)
	unlockedMap := make(map[uuid.UUID]bool)
	for _, u := range userUnlocks {
		unlockedMap[u.FeatureID] = true
	}

	var unlocked []dto.FeatureUnlockResponse
	var locked []dto.LockedFeatureResponse

	for _, f := range allFeatures {
		if unlockedMap[f.ID] {
			unlocked = append(unlocked, s.toFeatureResponse(f))
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
func (s *FeatureUnlockService) CheckFeatureAccess(userID uint, featureKey string) (*dto.FeatureAccessResponse, error) {
	// Check if feature exists
	feature, err := s.featureRepo.GetFeatureByKey(featureKey)
	if err != nil {
		return &dto.FeatureAccessResponse{
			HasAccess: false,
			Reason:    "Fitur tidak ditemukan",
		}, nil
	}

	// Get user's current level
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	currentLevel, err := s.levelConfigRepo.GetLevelByExp(user.Exp)
	if err != nil {
		return nil, err
	}

	// Check if user has unlocked the feature (either by level or explicit unlock)
	isUnlocked := s.featureRepo.IsFeatureUnlockedByKey(userID, featureKey)

	// Also check if user's level is sufficient
	hasLevelAccess := currentLevel.Level >= feature.RequiredLevel

	if isUnlocked || hasLevelAccess {
		// If level is sufficient but not explicitly unlocked, unlock it
		if hasLevelAccess && !isUnlocked {
			s.featureRepo.UnlockFeature(userID, feature.ID)
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
func (s *FeatureUnlockService) UnlockFeaturesOnLevelUp(userID uint, newLevel int) ([]dto.FeatureUnlockResponse, error) {
	newFeatures, err := s.featureRepo.UnlockFeaturesForLevel(userID, newLevel)
	if err != nil {
		return nil, err
	}

	result := make([]dto.FeatureUnlockResponse, len(newFeatures))
	for i, f := range newFeatures {
		result[i] = s.toFeatureResponse(f)
	}

	return result, nil
}

// GetUpcomingFeatures returns features user will unlock in upcoming levels
func (s *FeatureUnlockService) GetUpcomingFeatures(userID uint, limit int) ([]dto.LockedFeatureResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	currentLevel, err := s.levelConfigRepo.GetLevelByExp(user.Exp)
	if err != nil {
		return nil, err
	}

	features, err := s.featureRepo.GetUpcomingFeatures(currentLevel.Level, limit)
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
func (s *FeatureUnlockService) GetFeatureCategories() ([]dto.FeatureCategoryInfo, error) {
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

	return categories, nil
}

// Helper function to convert model to DTO
func (s *FeatureUnlockService) toFeatureResponse(f models.FeatureDefinition) dto.FeatureUnlockResponse {
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
