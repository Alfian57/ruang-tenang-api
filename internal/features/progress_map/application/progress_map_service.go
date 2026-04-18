package application

import (
	"context"
	"errors"
	"strings"
	"time"

	authinfra "github.com/Alfian57/ruang-tenang-api/internal/features/auth/infrastructure"
	gamificationinfra "github.com/Alfian57/ruang-tenang-api/internal/features/gamification/infrastructure"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/infrastructure"
)

var (
	ErrRegionNotFound       = errors.New("region tidak ditemukan")
	ErrLandmarkNotFound     = errors.New("landmark tidak ditemukan")
	ErrLandmarkNotUnlocked  = errors.New("landmark belum terbuka")
	ErrRewardAlreadyClaimed = errors.New("reward sudah diklaim")
	ErrInvalidUnlockType    = errors.New("unlock_type tidak valid")
)

type ProgressMapService struct {
	mapRepo         *infrastructure.ProgressMapRepository
	userRepo        *authinfra.UserRepository
	levelConfigRepo *gamificationinfra.LevelConfigRepository
}

func NewProgressMapService(
	mapRepo *infrastructure.ProgressMapRepository,
	userRepo *authinfra.UserRepository,
	levelConfigRepo *gamificationinfra.LevelConfigRepository,
) *ProgressMapService {
	return &ProgressMapService{
		mapRepo:         mapRepo,
		userRepo:        userRepo,
		levelConfigRepo: levelConfigRepo,
	}
}

// ==========================================
// Map Viewing
// ==========================================

// GetFullMap retrieves the complete map with user progress
func (s *ProgressMapService) GetFullMap(ctx context.Context, userID uint) (*dto.FullMapResponse, error) {
	regions, err := s.mapRepo.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user progress for regions and landmarks
	regionProgress, _ := s.mapRepo.GetUserRegionProgress(ctx, userID)
	landmarkProgress, _ := s.mapRepo.GetUserLandmarkProgress(ctx, userID)

	// Build lookup maps
	regionUnlockMap := make(map[uuid.UUID]*model.UserMapProgress)
	for i := range regionProgress {
		regionUnlockMap[regionProgress[i].RegionID] = &regionProgress[i]
	}

	landmarkUnlockMap := make(map[uuid.UUID]*model.UserLandmarkProgress)
	for i := range landmarkProgress {
		landmarkUnlockMap[landmarkProgress[i].LandmarkID] = &landmarkProgress[i]
	}

	totalRegions := 0
	unlockedRegions := 0
	totalLandmarks := 0
	unlockedLandmarks := 0

	regionResponses := make([]dto.MapRegionResponse, len(regions))
	for i, region := range regions {
		totalRegions++

		isRegionUnlocked := s.checkRegionUnlock(ctx, region, user)

		// Auto-unlock if conditions met
		if isRegionUnlocked {
			if _, exists := regionUnlockMap[region.ID]; !exists {
				now := time.Now()
				progress := &model.UserMapProgress{
					UserID:     userID,
					RegionID:   region.ID,
					IsUnlocked: true,
					UnlockedAt: &now,
				}
				_ = s.mapRepo.UpsertRegionProgress(ctx, progress)
				regionUnlockMap[region.ID] = progress
			}
			unlockedRegions++
		}

		var unlockTime *time.Time
		if rp, exists := regionUnlockMap[region.ID]; exists && rp.IsUnlocked {
			unlockTime = rp.UnlockedAt
			isRegionUnlocked = true
		}

		// Process landmarks
		regionUnlockedLandmarks := 0
		landmarkResponses := make([]dto.MapLandmarkResponse, len(region.Landmarks))
		for j, landmark := range region.Landmarks {
			totalLandmarks++

			isLandmarkUnlocked := false
			currentValue := 0
			var landmarkUnlockTime *time.Time
			rewardClaimed := false

			if lp, exists := landmarkUnlockMap[landmark.ID]; exists {
				isLandmarkUnlocked = lp.IsUnlocked
				currentValue = lp.CurrentValue
				landmarkUnlockTime = lp.UnlockedAt
				rewardClaimed = lp.RewardClaimed
			}

			// Check if landmark should be unlocked (only if region is unlocked)
			if isRegionUnlocked && !isLandmarkUnlocked {
				currentValue = s.getCurrentValueForLandmark(ctx, landmark, user)
				if currentValue >= landmark.UnlockValue {
					isLandmarkUnlocked = true
					now := time.Now()
					landmarkUnlockTime = &now
					lp := &model.UserLandmarkProgress{
						UserID:       userID,
						LandmarkID:   landmark.ID,
						IsUnlocked:   true,
						CurrentValue: currentValue,
						UnlockedAt:   &now,
					}
					_ = s.mapRepo.UpsertLandmarkProgress(ctx, lp)
				} else {
					// Update current value
					lp := &model.UserLandmarkProgress{
						UserID:       userID,
						LandmarkID:   landmark.ID,
						IsUnlocked:   false,
						CurrentValue: currentValue,
					}
					_ = s.mapRepo.UpsertLandmarkProgress(ctx, lp)
				}
			}

			if isLandmarkUnlocked {
				unlockedLandmarks++
				regionUnlockedLandmarks++
			}

			progressPercent := float64(0)
			if landmark.UnlockValue > 0 {
				progressPercent = float64(currentValue) / float64(landmark.UnlockValue) * 100
				if progressPercent > 100 {
					progressPercent = 100
				}
			}

			landmarkResponses[j] = dto.MapLandmarkResponse{
				ID:              landmark.ID,
				LandmarkKey:     landmark.LandmarkKey,
				Name:            landmark.Name,
				Description:     landmark.Description,
				Icon:            landmark.Icon,
				UnlockType:      string(landmark.UnlockType),
				UnlockActivity:  landmark.UnlockActivity,
				UnlockValue:     landmark.UnlockValue,
				PositionX:       landmark.PositionX,
				PositionY:       landmark.PositionY,
				XPReward:        landmark.XPReward,
				CoinReward:      landmark.CoinReward,
				IsUnlocked:      isLandmarkUnlocked,
				CurrentValue:    currentValue,
				ProgressPercent: progressPercent,
				UnlockedAt:      landmarkUnlockTime,
				RewardClaimed:   rewardClaimed,
			}
		}

		regionResponses[i] = dto.MapRegionResponse{
			ID:                region.ID,
			RegionKey:         region.RegionKey,
			Name:              region.Name,
			Description:       region.Description,
			Icon:              region.Icon,
			Image:             region.Image,
			UnlockType:        string(region.UnlockType),
			UnlockValue:       region.UnlockValue,
			PositionX:         region.PositionX,
			PositionY:         region.PositionY,
			DisplayOrder:      region.DisplayOrder,
			IsUnlocked:        isRegionUnlocked,
			UnlockedAt:        unlockTime,
			Landmarks:         landmarkResponses,
			TotalLandmarks:    len(region.Landmarks),
			UnlockedLandmarks: regionUnlockedLandmarks,
		}
	}

	overallProgress := float64(0)
	totalItems := totalRegions + totalLandmarks
	if totalItems > 0 {
		overallProgress = float64(unlockedRegions+unlockedLandmarks) / float64(totalItems) * 100
	}

	return &dto.FullMapResponse{
		Regions:           regionResponses,
		TotalRegions:      totalRegions,
		UnlockedRegions:   unlockedRegions,
		TotalLandmarks:    totalLandmarks,
		UnlockedLandmarks: unlockedLandmarks,
		OverallProgress:   overallProgress,
	}, nil
}

// GetRegionDetail retrieves a single region with user progress
func (s *ProgressMapService) GetRegionDetail(ctx context.Context, regionKey string, userID uint) (*dto.MapRegionResponse, error) {
	region, err := s.mapRepo.GetRegionByKey(ctx, regionKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRegionNotFound
		}
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	isUnlocked := s.checkRegionUnlock(ctx, *region, user)
	landmarkProgress, _ := s.mapRepo.GetUserLandmarkProgress(ctx, userID)

	landmarkUnlockMap := make(map[uuid.UUID]*model.UserLandmarkProgress)
	for i := range landmarkProgress {
		landmarkUnlockMap[landmarkProgress[i].LandmarkID] = &landmarkProgress[i]
	}

	unlockedLandmarks := 0
	landmarkResponses := make([]dto.MapLandmarkResponse, len(region.Landmarks))
	for j, landmark := range region.Landmarks {
		isLandmarkUnlocked := false
		currentValue := 0
		var landmarkUnlockTime *time.Time
		rewardClaimed := false

		if lp, exists := landmarkUnlockMap[landmark.ID]; exists {
			isLandmarkUnlocked = lp.IsUnlocked
			currentValue = lp.CurrentValue
			landmarkUnlockTime = lp.UnlockedAt
			rewardClaimed = lp.RewardClaimed
		}

		if isUnlocked && !isLandmarkUnlocked {
			currentValue = s.getCurrentValueForLandmark(ctx, landmark, user)
		}

		if isLandmarkUnlocked {
			unlockedLandmarks++
		}

		progressPercent := float64(0)
		if landmark.UnlockValue > 0 {
			progressPercent = float64(currentValue) / float64(landmark.UnlockValue) * 100
			if progressPercent > 100 {
				progressPercent = 100
			}
		}

		landmarkResponses[j] = dto.MapLandmarkResponse{
			ID:              landmark.ID,
			LandmarkKey:     landmark.LandmarkKey,
			Name:            landmark.Name,
			Description:     landmark.Description,
			Icon:            landmark.Icon,
			UnlockType:      string(landmark.UnlockType),
			UnlockActivity:  landmark.UnlockActivity,
			UnlockValue:     landmark.UnlockValue,
			PositionX:       landmark.PositionX,
			PositionY:       landmark.PositionY,
			XPReward:        landmark.XPReward,
			CoinReward:      landmark.CoinReward,
			IsUnlocked:      isLandmarkUnlocked,
			CurrentValue:    currentValue,
			ProgressPercent: progressPercent,
			UnlockedAt:      landmarkUnlockTime,
			RewardClaimed:   rewardClaimed,
		}
	}

	var unlockTime *time.Time
	rp, err := s.mapRepo.GetUserRegionProgressByID(ctx, userID, region.ID)
	if err == nil && rp.IsUnlocked {
		unlockTime = rp.UnlockedAt
	}

	return &dto.MapRegionResponse{
		ID:                region.ID,
		RegionKey:         region.RegionKey,
		Name:              region.Name,
		Description:       region.Description,
		Icon:              region.Icon,
		Image:             region.Image,
		UnlockType:        string(region.UnlockType),
		UnlockValue:       region.UnlockValue,
		PositionX:         region.PositionX,
		PositionY:         region.PositionY,
		DisplayOrder:      region.DisplayOrder,
		IsUnlocked:        isUnlocked,
		UnlockedAt:        unlockTime,
		Landmarks:         landmarkResponses,
		TotalLandmarks:    len(region.Landmarks),
		UnlockedLandmarks: unlockedLandmarks,
	}, nil
}

// GetProgressSummary returns a summary of the user's map progress
func (s *ProgressMapService) GetProgressSummary(ctx context.Context, userID uint) (*dto.MapProgressSummary, error) {
	regions, err := s.mapRepo.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	totalRegions := len(regions)
	totalLandmarks := 0
	for _, r := range regions {
		totalLandmarks += len(r.Landmarks)
	}

	unlockedRegions, _ := s.mapRepo.CountUnlockedRegions(ctx, userID)
	unlockedLandmarks, _ := s.mapRepo.CountUnlockedLandmarks(ctx, userID)

	overallProgress := float64(0)
	totalItems := totalRegions + totalLandmarks
	if totalItems > 0 {
		overallProgress = float64(unlockedRegions+unlockedLandmarks) / float64(totalItems) * 100
	}

	latestName, _ := s.mapRepo.GetLatestUnlock(ctx, userID)

	return &dto.MapProgressSummary{
		UnlockedRegions:   unlockedRegions,
		TotalRegions:      totalRegions,
		UnlockedLandmarks: unlockedLandmarks,
		TotalLandmarks:    totalLandmarks,
		OverallProgress:   overallProgress,
		LatestUnlock:      latestName,
	}, nil
}

// ClaimLandmarkReward claims the XP/coin reward for an unlocked landmark
func (s *ProgressMapService) ClaimLandmarkReward(ctx context.Context, userID uint, landmarkID uuid.UUID) error {
	landmark, err := s.mapRepo.GetLandmarkByID(ctx, landmarkID)
	if err != nil {
		return ErrLandmarkNotFound
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	region, err := s.mapRepo.GetRegionByID(ctx, landmark.RegionID)
	if err != nil {
		return ErrRegionNotFound
	}

	currentValue := s.getCurrentValueForLandmark(ctx, *landmark, user)
	isRegionUnlocked := s.checkRegionUnlock(ctx, *region, user)
	isCriteriaMet := isRegionUnlocked && currentValue >= landmark.UnlockValue

	progress, err := s.syncLandmarkProgressForClaim(ctx, userID, landmarkID, currentValue, isCriteriaMet)
	if err != nil {
		return err
	}

	if progress.RewardClaimed {
		return ErrRewardAlreadyClaimed
	}

	err = s.mapRepo.ClaimLandmarkReward(ctx, userID, landmarkID, landmark.XPReward, landmark.CoinReward)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRewardAlreadyClaimed
	}
	return err
}

func (s *ProgressMapService) syncLandmarkProgressForClaim(ctx context.Context, userID uint, landmarkID uuid.UUID, currentValue int, isCriteriaMet bool) (*model.UserLandmarkProgress, error) {
	progress, err := s.mapRepo.GetUserLandmarkProgressByID(ctx, userID, landmarkID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		progress = &model.UserLandmarkProgress{
			UserID:     userID,
			LandmarkID: landmarkID,
		}
	}

	progress.CurrentValue = currentValue
	if isCriteriaMet {
		progress.IsUnlocked = true
		if progress.UnlockedAt == nil {
			now := time.Now()
			progress.UnlockedAt = &now
		}
	} else {
		progress.IsUnlocked = false
		progress.UnlockedAt = nil
	}

	if err := s.mapRepo.UpsertLandmarkProgress(ctx, progress); err != nil {
		return nil, err
	}

	if !progress.IsUnlocked {
		return nil, ErrLandmarkNotUnlocked
	}

	return progress, nil
}

// ==========================================
// Admin Landmark Management
// ==========================================

func (s *ProgressMapService) AdminGetAllLandmarks(ctx context.Context) ([]model.MapLandmark, error) {
	return s.mapRepo.GetAllLandmarks(ctx)
}

func (s *ProgressMapService) AdminCreateLandmark(ctx context.Context, landmark *model.MapLandmark) error {
	if !isValidMapUnlockType(landmark.UnlockType) {
		return ErrInvalidUnlockType
	}

	if _, err := s.mapRepo.GetRegionByID(ctx, landmark.RegionID); err != nil {
		return ErrRegionNotFound
	}

	return s.mapRepo.CreateLandmark(ctx, landmark)
}

func (s *ProgressMapService) AdminUpdateLandmark(ctx context.Context, landmarkID uuid.UUID, payload *model.MapLandmark) error {
	existing, err := s.mapRepo.GetLandmarkByID(ctx, landmarkID)
	if err != nil {
		return ErrLandmarkNotFound
	}

	if !isValidMapUnlockType(payload.UnlockType) {
		return ErrInvalidUnlockType
	}

	if _, err := s.mapRepo.GetRegionByID(ctx, payload.RegionID); err != nil {
		return ErrRegionNotFound
	}

	existing.RegionID = payload.RegionID
	existing.LandmarkKey = payload.LandmarkKey
	existing.Name = payload.Name
	existing.Description = payload.Description
	existing.Icon = payload.Icon
	existing.UnlockType = payload.UnlockType
	existing.UnlockActivity = payload.UnlockActivity
	existing.UnlockValue = payload.UnlockValue
	existing.PositionX = payload.PositionX
	existing.PositionY = payload.PositionY
	existing.XPReward = payload.XPReward
	existing.CoinReward = payload.CoinReward
	existing.DisplayOrder = payload.DisplayOrder
	existing.IsActive = payload.IsActive

	return s.mapRepo.UpdateLandmark(ctx, existing)
}

func (s *ProgressMapService) AdminDeleteLandmark(ctx context.Context, landmarkID uuid.UUID) error {
	if _, err := s.mapRepo.GetLandmarkByID(ctx, landmarkID); err != nil {
		return ErrLandmarkNotFound
	}

	return s.mapRepo.DeactivateLandmark(ctx, landmarkID)
}

// ==========================================
// Helpers
// ==========================================

func (s *ProgressMapService) checkRegionUnlock(ctx context.Context, region model.MapRegion, user *model.User) bool {
	switch region.UnlockType {
	case model.MapUnlockLevel:
		levelConfig, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
		if err != nil {
			return region.UnlockValue <= 1
		}
		return levelConfig.Level >= region.UnlockValue
	case model.MapUnlockXP:
		return user.Exp >= int64(region.UnlockValue)
	case model.MapUnlockStreak:
		return user.CurrentStreak >= region.UnlockValue || user.LongestStreak >= region.UnlockValue
	default:
		return region.UnlockValue <= 1
	}
}

func (s *ProgressMapService) getCurrentValueForLandmark(ctx context.Context, landmark model.MapLandmark, user *model.User) int {
	switch landmark.UnlockType {
	case model.MapUnlockLevel:
		levelConfig, err := s.levelConfigRepo.GetLevelByExp(ctx, user.Exp)
		if err != nil {
			return 1
		}
		return levelConfig.Level
	case model.MapUnlockXP:
		return int(user.Exp)
	case model.MapUnlockStreak:
		if user.LongestStreak > user.CurrentStreak {
			return user.LongestStreak
		}
		return user.CurrentStreak
	case model.MapUnlockActivityCount:
		activity := strings.TrimSpace(strings.ToLower(landmark.UnlockActivity))
		switch activity {
		case "":
			return user.TotalActivities
		case "login":
			if user.LoginStreak > 0 {
				return user.LoginStreak
			}
			if user.LastLoginDate != nil {
				return 1
			}
			return 0
		case "mood":
			count, err := s.mapRepo.CountUserMoods(ctx, user.ID)
			if err != nil {
				return user.TotalActivities
			}
			return count
		case "journal":
			count, err := s.mapRepo.CountUserJournals(ctx, user.ID)
			if err != nil {
				return user.TotalActivities
			}
			return count
		case "forum":
			count, err := s.mapRepo.CountUserForums(ctx, user.ID)
			if err != nil {
				return user.TotalActivities
			}
			return count
		case "story":
			count, err := s.mapRepo.CountUserStories(ctx, user.ID)
			if err != nil {
				return user.TotalActivities
			}
			return count
		default:
			aliases := mapUnlockActivityToExpHistoryTypes(activity)
			count, err := s.mapRepo.CountExpHistoryByTypes(ctx, user.ID, aliases)
			if err != nil {
				return user.TotalActivities
			}
			if count > 0 {
				return count
			}
			return user.TotalActivities
		}
	default:
		return 0
	}
}

func mapUnlockActivityToExpHistoryTypes(activity string) []string {
	switch activity {
	case "chat":
		return []string{"chat", "chat_ai"}
	case "breathing":
		return []string{"breathing"}
	case "article":
		return []string{"article", "read_article", "article_read", "upload_article"}
	case "write_article":
		return []string{"write_article", "upload_article"}
	case "forum":
		return []string{"forum", "forum_comment", "forum_post"}
	case "story":
		return []string{"story", "story_approved"}
	case "mood":
		return []string{"mood", "mood_checkin"}
	case "journal":
		return []string{"journal", "journal_write", "journal_entry"}
	case "login":
		return []string{"login", "daily_login"}
	default:
		return []string{activity}
	}
}

func isValidMapUnlockType(value model.MapUnlockType) bool {
	switch value {
	case model.MapUnlockLevel, model.MapUnlockActivityCount, model.MapUnlockStreak, model.MapUnlockBadge, model.MapUnlockXP:
		return true
	default:
		return false
	}
}
