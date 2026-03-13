package infrastructure

import (
	"context"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FeatureUnlockRepository struct {
	db *gorm.DB
}

func NewFeatureUnlockRepository(db *gorm.DB) *FeatureUnlockRepository {
	return &FeatureUnlockRepository{db: db}
}

// ==========================================
// Feature Definitions
// ==========================================

// GetAllFeatureDefinitions retrieves all feature definitions
func (r *FeatureUnlockRepository) GetAllFeatureDefinitions(ctx context.Context) ([]model.FeatureDefinition, error) {
	var features []model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("is_active = ?", true).
		Order("required_level ASC, feature_key ASC").
		Find(&features).Error
	return features, err
}

// GetFeaturesByLevel retrieves features unlocked at a specific level
func (r *FeatureUnlockRepository) GetFeaturesByLevel(ctx context.Context, level int) ([]model.FeatureDefinition, error) {
	var features []model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("required_level = ? AND is_active = ?", level, true).
		Order("feature_key ASC").
		Find(&features).Error
	return features, err
}

// GetFeaturesUpToLevel retrieves all features unlocked up to and including a level
func (r *FeatureUnlockRepository) GetFeaturesUpToLevel(ctx context.Context, level int) ([]model.FeatureDefinition, error) {
	var features []model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("required_level <= ? AND is_active = ?", level, true).
		Order("required_level ASC, feature_key ASC").
		Find(&features).Error
	return features, err
}

// GetFeatureByKey retrieves a feature definition by its key
func (r *FeatureUnlockRepository) GetFeatureByKey(ctx context.Context, key string) (*model.FeatureDefinition, error) {
	var feature model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("feature_key = ?", key).First(&feature).Error
	if err != nil {
		return nil, err
	}
	return &feature, nil
}

// GetFeatureByID retrieves a feature definition by ID
func (r *FeatureUnlockRepository) GetFeatureByID(ctx context.Context, id uuid.UUID) (*model.FeatureDefinition, error) {
	var feature model.FeatureDefinition
	err := r.db.WithContext(ctx).First(&feature, id).Error
	if err != nil {
		return nil, err
	}
	return &feature, nil
}

// CreateFeatureDefinition creates a new feature definition
func (r *FeatureUnlockRepository) CreateFeatureDefinition(ctx context.Context, feature *model.FeatureDefinition) error {
	return r.db.WithContext(ctx).Create(feature).Error
}

// UpdateFeatureDefinition updates a feature definition
func (r *FeatureUnlockRepository) UpdateFeatureDefinition(ctx context.Context, feature *model.FeatureDefinition) error {
	return r.db.WithContext(ctx).Save(feature).Error
}

// ==========================================
// User Feature Unlocks
// ==========================================

// GetUserUnlockedFeatures retrieves all features unlocked by a user
func (r *FeatureUnlockRepository) GetUserUnlockedFeatures(ctx context.Context, userID uint) ([]model.UserFeatureUnlock, error) {
	var unlocks []model.UserFeatureUnlock
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Preload("Feature").
		Find(&unlocks).Error
	return unlocks, err
}

// IsFeatureUnlocked checks if a user has unlocked a specific feature
func (r *FeatureUnlockRepository) IsFeatureUnlocked(ctx context.Context, userID uint, featureID uuid.UUID) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.UserFeatureUnlock{}).
		Where("user_id = ? AND feature_id = ?", userID, featureID).
		Count(&count)
	return count > 0
}

// IsFeatureUnlockedByKey checks if a user has unlocked a feature by its key
func (r *FeatureUnlockRepository) IsFeatureUnlockedByKey(ctx context.Context, userID uint, featureKey string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&model.UserFeatureUnlock{}).
		Joins("JOIN feature_definitions ON feature_definitions.id = user_feature_unlocks.feature_id").
		Where("user_feature_unlocks.user_id = ? AND feature_definitions.feature_key = ?", userID, featureKey).
		Count(&count)
	return count > 0
}

// UnlockFeature unlocks a feature for a user
func (r *FeatureUnlockRepository) UnlockFeature(ctx context.Context, userID uint, featureID uuid.UUID) error {
	unlock := &model.UserFeatureUnlock{
		UserID:    userID,
		FeatureID: featureID,
	}
	return r.db.WithContext(ctx).Create(unlock).Error
}

// UnlockFeaturesForLevel unlocks all features for a given level
func (r *FeatureUnlockRepository) UnlockFeaturesForLevel(ctx context.Context, userID uint, level int) ([]model.FeatureDefinition, error) {
	// Get features for this level
	features, err := r.GetFeaturesByLevel(ctx, level)
	if err != nil {
		return nil, err
	}

	var newlyUnlocked []model.FeatureDefinition

	for _, feature := range features {
		// Check if already unlocked
		if !r.IsFeatureUnlocked(ctx, userID, feature.ID) {
			if err := r.UnlockFeature(ctx, userID, feature.ID); err != nil {
				continue // Skip on error but continue with others
			}
			newlyUnlocked = append(newlyUnlocked, feature)
		}
	}

	return newlyUnlocked, nil
}

// GetNewlyAvailableFeatures gets features that are newly available at a level
func (r *FeatureUnlockRepository) GetNewlyAvailableFeatures(ctx context.Context, userID uint, oldLevel, newLevel int) ([]model.FeatureDefinition, error) {
	var features []model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("required_level > ? AND required_level <= ? AND is_active = ?", oldLevel, newLevel, true).
		Order("required_level ASC").
		Find(&features).Error
	return features, err
}

// GetUserFeatureStats returns stats about user's feature unlocks
func (r *FeatureUnlockRepository) GetUserFeatureStats(ctx context.Context, userID uint) (int, int, error) {
	var totalFeatures int64
	r.db.WithContext(ctx).Model(&model.FeatureDefinition{}).Where("is_active = ?", true).Count(&totalFeatures)

	var unlockedFeatures int64
	r.db.WithContext(ctx).Model(&model.UserFeatureUnlock{}).Where("user_id = ?", userID).Count(&unlockedFeatures)

	return int(unlockedFeatures), int(totalFeatures), nil
}

// GetUpcomingFeatures returns features user will unlock at next levels
func (r *FeatureUnlockRepository) GetUpcomingFeatures(ctx context.Context, currentLevel int, limit int) ([]model.FeatureDefinition, error) {
	var features []model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("required_level > ? AND is_active = ?", currentLevel, true).
		Order("required_level ASC").
		Limit(limit).
		Find(&features).Error
	return features, err
}

// GetFeaturesByCategory retrieves features by category
func (r *FeatureUnlockRepository) GetFeaturesByCategory(ctx context.Context, category string) ([]model.FeatureDefinition, error) {
	var features []model.FeatureDefinition
	err := r.db.WithContext(ctx).Where("category = ? AND is_active = ?", category, true).
		Order("required_level ASC").
		Find(&features).Error
	return features, err
}
