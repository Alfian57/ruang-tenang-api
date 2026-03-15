package infrastructure

import (
	"context"
	"strings"

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

// GetAllLandmarks retrieves all landmarks for admin management
func (r *ProgressMapRepository) GetAllLandmarks(ctx context.Context) ([]model.MapLandmark, error) {
	var landmarks []model.MapLandmark
	err := r.db.WithContext(ctx).
		Preload("Region").
		Order("display_order ASC").
		Order("created_at ASC").
		Find(&landmarks).Error
	return landmarks, err
}

// UpdateLandmark updates an existing landmark
func (r *ProgressMapRepository) UpdateLandmark(ctx context.Context, landmark *model.MapLandmark) error {
	return r.db.WithContext(ctx).Save(landmark).Error
}

// DeactivateLandmark soft-deletes landmark by setting is_active to false
func (r *ProgressMapRepository) DeactivateLandmark(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.MapLandmark{}).
		Where("id = ?", id).
		Update("is_active", false).Error
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

// CountExpHistoryByTypes counts exp history rows for a user filtered by activity types.
func (r *ProgressMapRepository) CountExpHistoryByTypes(ctx context.Context, userID uint, activityTypes []string) (int, error) {
	if len(activityTypes) == 0 {
		return 0, nil
	}

	normalized := make([]string, 0, len(activityTypes))
	for _, t := range activityTypes {
		trimmed := strings.TrimSpace(strings.ToLower(t))
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	if len(normalized) == 0 {
		return 0, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ExpHistory{}).
		Where("user_id = ?", userID).
		Where("LOWER(activity_type) IN ?", normalized).
		Count(&count).Error
	return int(count), err
}

// CountUserMoods counts mood entries created by user.
func (r *ProgressMapRepository) CountUserMoods(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserMood{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

// CountUserJournals counts journal entries created by user.
func (r *ProgressMapRepository) CountUserJournals(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Journal{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

// CountUserForums counts forum topics created by user.
func (r *ProgressMapRepository) CountUserForums(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Forum{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

// CountUserStories counts inspiring stories created by user.
func (r *ProgressMapRepository) CountUserStories(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.InspiringStory{}).
		Where("author_id = ?", userID).
		Count(&count).Error
	return int(count), err
}
