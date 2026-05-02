package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrWellnessProfileNotFound = errors.New("wellness profile not found")
	ErrWellnessPlanNotFound    = errors.New("wellness plan not found")
	ErrWellnessItemNotFound    = errors.New("wellness plan item not found")
	ErrWeeklyInsightNotFound   = errors.New("weekly insight not found")
)

type WeeklyAggregate struct {
	MoodCounts        map[string]int
	LatestMood        string
	JournalCount      int
	JournalWords      int
	BreathingCount    int
	BreathingMinutes  int
	ChatSessionCount  int
	ChatMessageCount  int
	TaskCompleted     int
	TaskClaimed       int
	RewardClaims      int
	UnlockedLandmarks int
	Streak            int
}

type WellnessRepository struct {
	db *gorm.DB
}

func NewWellnessRepository(db *gorm.DB) *WellnessRepository {
	return &WellnessRepository{db: db}
}

func (r *WellnessRepository) GetProfileByUserID(ctx context.Context, userID uint) (*model.UserWellnessProfile, error) {
	var profile model.UserWellnessProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWellnessProfileNotFound
		}
		return nil, err
	}
	return &profile, nil
}

func (r *WellnessRepository) UpsertProfile(ctx context.Context, profile *model.UserWellnessProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"initial_mood":            profile.InitialMood,
			"goals_json":              profile.GoalsJSON,
			"habits_json":             profile.HabitsJSON,
			"onboarding_completed_at": profile.OnboardingCompletedAt,
			"updated_at":              time.Now(),
		}),
	}).Create(profile).Error
}

func (r *WellnessRepository) MarkTourCompleted(ctx context.Context, userID uint, completedAt time.Time) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"tour_completed_at": completedAt,
			"updated_at":        time.Now(),
		}),
	}).Create(&model.UserWellnessProfile{
		UserID:          userID,
		GoalsJSON:       "[]",
		HabitsJSON:      "[]",
		TourCompletedAt: &completedAt,
	}).Error
}

func (r *WellnessRepository) ArchiveActivePlans(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&model.WellnessPlan{}).
		Where("user_id = ? AND status = ?", userID, model.WellnessPlanStatusActive).
		Updates(map[string]any{
			"status":     model.WellnessPlanStatusArchived,
			"updated_at": time.Now(),
		}).Error
}

func (r *WellnessRepository) CreatePlanWithItems(ctx context.Context, plan *model.WellnessPlan, items []model.WellnessPlanItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(plan).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].PlanID = plan.ID
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *WellnessRepository) GetActivePlan(ctx context.Context, userID uint) (*model.WellnessPlan, error) {
	var plan model.WellnessPlan
	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("day_number ASC") }).
		Where("user_id = ? AND status = ?", userID, model.WellnessPlanStatusActive).
		Order("created_at DESC").
		First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWellnessPlanNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func (r *WellnessRepository) GetPlanItemForUser(ctx context.Context, userID uint, itemID uuid.UUID) (*model.WellnessPlanItem, error) {
	var item model.WellnessPlanItem
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWellnessItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *WellnessRepository) SavePlanItem(ctx context.Context, item *model.WellnessPlanItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *WellnessRepository) GetWeeklyInsight(ctx context.Context, userID uint, weekStart time.Time) (*model.WeeklyInsightSnapshot, error) {
	var snapshot model.WeeklyInsightSnapshot
	err := r.db.WithContext(ctx).Where("user_id = ? AND week_start = ?", userID, weekStart).First(&snapshot).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrWeeklyInsightNotFound
		}
		return nil, err
	}
	return &snapshot, nil
}

func (r *WellnessRepository) UpsertWeeklyInsight(ctx context.Context, snapshot *model.WeeklyInsightSnapshot) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "week_start"}},
		DoUpdates: clause.Assignments(map[string]any{
			"week_end":              snapshot.WeekEnd,
			"mood_summary_json":     snapshot.MoodSummaryJSON,
			"activity_summary_json": snapshot.ActivitySummaryJSON,
			"insight_json":          snapshot.InsightJSON,
			"premium_sections_json": snapshot.PremiumSectionsJSON,
			"narrative":             snapshot.Narrative,
			"recommendations_json":  snapshot.RecommendationsJSON,
			"is_ai_enhanced":        snapshot.IsAIEnhanced,
			"updated_at":            time.Now(),
		}),
	}).Create(snapshot).Error
}

func (r *WellnessRepository) CreateNeedEvent(ctx context.Context, event *model.WellnessNeedEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *WellnessRepository) HasPremiumAccess(ctx context.Context, userID uint, at time.Time) (bool, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("id", "role", "is_premium", "premium_expires_at").First(&user, userID).Error; err != nil {
		return false, err
	}
	if user.Role == model.RoleMitra {
		return true, nil
	}
	if user.IsPremium && (user.PremiumExpiresAt == nil || user.PremiumExpiresAt.After(at)) {
		return true, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Table("organization_members AS om").
		Joins("JOIN b2b_seat_allocations AS bsa ON bsa.organization_member_id = om.id AND bsa.released_at IS NULL").
		Joins("JOIN b2b_subscriptions AS bs ON bs.id = bsa.subscription_id").
		Where("om.user_id = ? AND om.status = ?", userID, model.OrganizationMemberStatusActive).
		Where("bs.status = ? AND bs.starts_at <= ? AND bs.ends_at > ?", model.B2BSubscriptionStatusActive, at, at).
		Count(&count).Error
	return count > 0, err
}

func (r *WellnessRepository) AggregateWeekly(ctx context.Context, userID uint, start, end time.Time) (*WeeklyAggregate, error) {
	aggregate := &WeeklyAggregate{MoodCounts: map[string]int{}}

	var user model.User
	if err := r.db.WithContext(ctx).Select("current_streak").First(&user, userID).Error; err == nil {
		aggregate.Streak = user.CurrentStreak
	}

	var moodRows []struct {
		Mood  string
		Count int
	}
	if err := r.db.WithContext(ctx).Model(&model.UserMood{}).
		Select("mood, COUNT(*) AS count").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Group("mood").
		Scan(&moodRows).Error; err != nil {
		return nil, err
	}
	for _, row := range moodRows {
		aggregate.MoodCounts[row.Mood] = row.Count
	}

	var latestMood model.UserMood
	if err := r.db.WithContext(ctx).Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Order("created_at DESC").
		First(&latestMood).Error; err == nil {
		aggregate.LatestMood = string(latestMood.Mood)
	}

	var journalStats struct {
		Count int
		Words int
	}
	if err := r.db.WithContext(ctx).Model(&model.Journal{}).
		Select("COUNT(*) AS count, COALESCE(SUM(word_count), 0) AS words").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, start, end).
		Scan(&journalStats).Error; err != nil {
		return nil, err
	}
	aggregate.JournalCount = journalStats.Count
	aggregate.JournalWords = journalStats.Words

	var breathingStats struct {
		Count   int
		Minutes int
	}
	if err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select("COUNT(*) AS count, COALESCE(SUM(duration_seconds), 0) / 60 AS minutes").
		Where("user_id = ? AND started_at >= ? AND started_at < ? AND completed = ?", userID, start, end, true).
		Scan(&breathingStats).Error; err != nil {
		return nil, err
	}
	aggregate.BreathingCount = breathingStats.Count
	aggregate.BreathingMinutes = breathingStats.Minutes

	var chatSessions int64
	if err := r.db.WithContext(ctx).Model(&model.ChatSession{}).
		Where("user_id = ? AND updated_at >= ? AND updated_at < ? AND is_trash = ?", userID, start, end, false).
		Count(&chatSessions).Error; err != nil {
		return nil, err
	}
	aggregate.ChatSessionCount = int(chatSessions)

	var chatMessages int64
	if err := r.db.WithContext(ctx).Table("chat_messages AS cm").
		Joins("JOIN chat_sessions AS cs ON cs.id = cm.chat_session_id").
		Where("cs.user_id = ? AND cm.created_at >= ? AND cm.created_at < ? AND cm.role = ?", userID, start, end, model.ChatRoleUser).
		Count(&chatMessages).Error; err != nil {
		return nil, err
	}
	aggregate.ChatMessageCount = int(chatMessages)

	var taskRows []struct {
		IsCompleted bool
		IsClaimed   bool
		Count       int
	}
	if err := r.db.WithContext(ctx).Model(&model.DailyTask{}).
		Select("is_completed, is_claimed, COUNT(*) AS count").
		Where("user_id = ? AND task_date >= ? AND task_date < ?", userID, start, end).
		Group("is_completed, is_claimed").
		Scan(&taskRows).Error; err != nil {
		return nil, err
	}
	for _, row := range taskRows {
		if row.IsCompleted {
			aggregate.TaskCompleted += row.Count
		}
		if row.IsClaimed {
			aggregate.TaskClaimed += row.Count
		}
	}

	var rewardClaims int64
	if err := r.db.WithContext(ctx).Model(&model.RewardClaim{}).
		Where("user_id = ? AND claimed_at >= ? AND claimed_at < ?", userID, start, end).
		Count(&rewardClaims).Error; err == nil {
		aggregate.RewardClaims = int(rewardClaims)
	}

	var unlocked int64
	if err := r.db.WithContext(ctx).Model(&model.UserLandmarkProgress{}).
		Where("user_id = ? AND is_unlocked = ? AND unlocked_at >= ? AND unlocked_at < ?", userID, true, start, end).
		Count(&unlocked).Error; err == nil {
		aggregate.UnlockedLandmarks = int(unlocked)
	}

	return aggregate, nil
}

func (r *WellnessRepository) CountJourneySignals(ctx context.Context, userID uint, since time.Time) (map[string]int, error) {
	signals := map[string]int{
		"mood":      0,
		"journal":   0,
		"breathing": 0,
		"chat":      0,
		"reward":    0,
		"landmarks": 0,
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.UserMood{}).Where("user_id = ? AND created_at >= ?", userID, since).Count(&count).Error; err != nil {
		return nil, err
	}
	signals["mood"] = int(count)
	if err := r.db.WithContext(ctx).Model(&model.Journal{}).Where("user_id = ? AND created_at >= ?", userID, since).Count(&count).Error; err != nil {
		return nil, err
	}
	signals["journal"] = int(count)
	if err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).Where("user_id = ? AND started_at >= ? AND completed = ?", userID, since, true).Count(&count).Error; err != nil {
		return nil, err
	}
	signals["breathing"] = int(count)
	if err := r.db.WithContext(ctx).Model(&model.ChatSession{}).Where("user_id = ? AND updated_at >= ? AND is_trash = ?", userID, since, false).Count(&count).Error; err != nil {
		return nil, err
	}
	signals["chat"] = int(count)
	if err := r.db.WithContext(ctx).Model(&model.RewardClaim{}).Where("user_id = ? AND claimed_at >= ?", userID, since).Count(&count).Error; err == nil {
		signals["reward"] = int(count)
	}
	if err := r.db.WithContext(ctx).Model(&model.UserLandmarkProgress{}).Where("user_id = ? AND is_unlocked = ?", userID, true).Count(&count).Error; err == nil {
		signals["landmarks"] = int(count)
	}
	return signals, nil
}

func (r *WellnessRepository) GetUserStreak(ctx context.Context, userID uint) (int, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("current_streak").First(&user, userID).Error; err != nil {
		return 0, err
	}
	return user.CurrentStreak, nil
}
