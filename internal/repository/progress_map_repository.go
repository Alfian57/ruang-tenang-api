package repository

import (
	"context"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProgressMapRepository struct {
	db *gorm.DB
}

func NewProgressMapRepository(db *gorm.DB) *ProgressMapRepository {
	return &ProgressMapRepository{db: db}
}

// ==========================================
// Map Regions
// ==========================================

// GetAllRegions retrieves all active regions ordered by display_order
func (r *ProgressMapRepository) GetAllRegions(ctx context.Context) ([]model.MapRegion, error) {
	var regions []model.MapRegion
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Landmarks", "is_active = ?", true).
		Order("display_order ASC").
		Find(&regions).Error
	return regions, err
}

// GetRegionByID retrieves a region by ID
func (r *ProgressMapRepository) GetRegionByID(ctx context.Context, id uuid.UUID) (*model.MapRegion, error) {
	var region model.MapRegion
	err := r.db.WithContext(ctx).
		Preload("Landmarks", "is_active = ?", true).
		First(&region, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}

// GetRegionByKey retrieves a region by its key
func (r *ProgressMapRepository) GetRegionByKey(ctx context.Context, key string) (*model.MapRegion, error) {
	var region model.MapRegion
	err := r.db.WithContext(ctx).
		Preload("Landmarks", "is_active = ?", true).
		Where("region_key = ?", key).First(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}

// CreateRegion creates a new map region
func (r *ProgressMapRepository) CreateRegion(ctx context.Context, region *model.MapRegion) error {
	return r.db.WithContext(ctx).Create(region).Error
}

// UpdateRegion updates a map region
func (r *ProgressMapRepository) UpdateRegion(ctx context.Context, region *model.MapRegion) error {
	return r.db.WithContext(ctx).Save(region).Error
}

// ==========================================
// Map Landmarks
// ==========================================

// GetLandmarkByID retrieves a landmark by ID
func (r *ProgressMapRepository) GetLandmarkByID(ctx context.Context, id uuid.UUID) (*model.MapLandmark, error) {
	var landmark model.MapLandmark
	err := r.db.WithContext(ctx).First(&landmark, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &landmark, nil
}

// GetLandmarksByRegion retrieves all landmarks for a region
func (r *ProgressMapRepository) GetLandmarksByRegion(ctx context.Context, regionID uuid.UUID) ([]model.MapLandmark, error) {
	var landmarks []model.MapLandmark
	err := r.db.WithContext(ctx).
		Where("region_id = ? AND is_active = ?", regionID, true).
		Order("display_order ASC").
		Find(&landmarks).Error
	return landmarks, err
}

// CreateLandmark creates a new landmark
func (r *ProgressMapRepository) CreateLandmark(ctx context.Context, landmark *model.MapLandmark) error {
	return r.db.WithContext(ctx).Create(landmark).Error
}

// ==========================================
// User Map Progress
// ==========================================

// GetUserRegionProgress retrieves user's progress for all regions
func (r *ProgressMapRepository) GetUserRegionProgress(ctx context.Context, userID uint) ([]model.UserMapProgress, error) {
	var progress []model.UserMapProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&progress).Error
	return progress, err
}

// GetUserRegionProgressByID retrieves user's progress for a specific region
func (r *ProgressMapRepository) GetUserRegionProgressByID(ctx context.Context, userID uint, regionID uuid.UUID) (*model.UserMapProgress, error) {
	var progress model.UserMapProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND region_id = ?", userID, regionID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// UpsertRegionProgress creates or updates user's region progress
func (r *ProgressMapRepository) UpsertRegionProgress(ctx context.Context, progress *model.UserMapProgress) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND region_id = ?", progress.UserID, progress.RegionID).
		Assign(model.UserMapProgress{
			IsUnlocked: progress.IsUnlocked,
			UnlockedAt: progress.UnlockedAt,
		}).
		FirstOrCreate(progress).Error
}

// ==========================================
// User Landmark Progress
// ==========================================

// GetUserLandmarkProgress retrieves user's progress for all landmarks
func (r *ProgressMapRepository) GetUserLandmarkProgress(ctx context.Context, userID uint) ([]model.UserLandmarkProgress, error) {
	var progress []model.UserLandmarkProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&progress).Error
	return progress, err
}

// GetUserLandmarkProgressByID retrieves user's progress for a specific landmark
func (r *ProgressMapRepository) GetUserLandmarkProgressByID(ctx context.Context, userID uint, landmarkID uuid.UUID) (*model.UserLandmarkProgress, error) {
	var progress model.UserLandmarkProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND landmark_id = ?", userID, landmarkID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// UpsertLandmarkProgress creates or updates user's landmark progress
func (r *ProgressMapRepository) UpsertLandmarkProgress(ctx context.Context, progress *model.UserLandmarkProgress) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND landmark_id = ?", progress.UserID, progress.LandmarkID).
		Assign(model.UserLandmarkProgress{
			IsUnlocked:    progress.IsUnlocked,
			UnlockedAt:    progress.UnlockedAt,
			CurrentValue:  progress.CurrentValue,
			RewardClaimed: progress.RewardClaimed,
		}).
		FirstOrCreate(progress).Error
}

// ClaimLandmarkReward marks a landmark reward as claimed
func (r *ProgressMapRepository) ClaimLandmarkReward(ctx context.Context, userID uint, landmarkID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.UserLandmarkProgress{}).
		Where("user_id = ? AND landmark_id = ? AND is_unlocked = ?", userID, landmarkID, true).
		Update("reward_claimed", true).Error
}

// CountUnlockedRegions counts how many regions a user has unlocked
func (r *ProgressMapRepository) CountUnlockedRegions(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserMapProgress{}).
		Where("user_id = ? AND is_unlocked = ?", userID, true).
		Count(&count).Error
	return int(count), err
}

// CountUnlockedLandmarks counts how many landmarks a user has unlocked
func (r *ProgressMapRepository) CountUnlockedLandmarks(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserLandmarkProgress{}).
		Where("user_id = ? AND is_unlocked = ?", userID, true).
		Count(&count).Error
	return int(count), err
}

// GetLatestUnlock retrieves the name of the latest unlocked region or landmark
func (r *ProgressMapRepository) GetLatestUnlock(ctx context.Context, userID uint) (string, error) {
	// Check latest region unlock
	var regionProgress model.UserMapProgress
	regionErr := r.db.WithContext(ctx).
		Preload("Region").
		Where("user_id = ? AND is_unlocked = ?", userID, true).
		Order("unlocked_at DESC").
		First(&regionProgress).Error

	// Check latest landmark unlock
	var landmarkProgress model.UserLandmarkProgress
	landmarkErr := r.db.WithContext(ctx).
		Preload("Landmark").
		Where("user_id = ? AND is_unlocked = ?", userID, true).
		Order("unlocked_at DESC").
		First(&landmarkProgress).Error

	if regionErr != nil && landmarkErr != nil {
		return "", nil
	}
	if regionErr != nil && landmarkErr == nil {
		return landmarkProgress.Landmark.Name, nil
	}
	if landmarkErr != nil && regionErr == nil {
		return regionProgress.Region.Name, nil
	}

	if regionProgress.UnlockedAt != nil && landmarkProgress.UnlockedAt != nil {
		if regionProgress.UnlockedAt.After(*landmarkProgress.UnlockedAt) {
			return regionProgress.Region.Name, nil
		}
		return landmarkProgress.Landmark.Name, nil
	}
	return "", nil
}
