package infrastructure

import (
	"context"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/timeutil"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BreathingRepository interface {
	// Techniques
	GetAllTechniques(ctx context.Context) ([]model.BreathingTechnique, error)
	GetSystemTechniques(ctx context.Context) ([]model.BreathingTechnique, error)
	GetUserTechniques(ctx context.Context, userID uint) ([]model.BreathingTechnique, error)
	GetTechniqueByID(ctx context.Context, id uuid.UUID) (*model.BreathingTechnique, error)
	GetTechniqueBySlug(ctx context.Context, slug string) (*model.BreathingTechnique, error)
	CreateTechnique(ctx context.Context, technique *model.BreathingTechnique) error
	UpdateTechnique(ctx context.Context, technique *model.BreathingTechnique) error
	DeleteTechnique(ctx context.Context, id uuid.UUID) error

	// Sessions
	CreateSession(ctx context.Context, session *model.BreathingSession) error
	UpdateSession(ctx context.Context, session *model.BreathingSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*model.BreathingSession, error)
	GetUserSessions(ctx context.Context, userID uint, startDate, endDate *time.Time, techniqueID *uuid.UUID, limit, offset int) ([]model.BreathingSession, int64, error)
	GetUserSessionsToday(ctx context.Context, userID uint) ([]model.BreathingSession, error)
	GetUserSessionsByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]model.BreathingSession, error)
	GetUserTotalSessionsCount(ctx context.Context, userID uint) (int64, error)
	GetUserTotalMinutes(ctx context.Context, userID uint) (int, error)
	GetUserTodayXP(ctx context.Context, userID uint) (int, error)
	GetMostUsedTechnique(ctx context.Context, userID uint) (*model.BreathingTechnique, int64, error)
	GetCompletionRate(ctx context.Context, userID uint) (float64, error)
	CountSessionsSince(ctx context.Context, userID uint, since time.Time) (int64, error)

	// Preferences
	GetPreferences(ctx context.Context, userID uint) (*model.BreathingPreference, error)
	CreatePreferences(ctx context.Context, pref *model.BreathingPreference) error
	UpdatePreferences(ctx context.Context, pref *model.BreathingPreference) error
	GetOrCreatePreferences(ctx context.Context, userID uint) (*model.BreathingPreference, error)

	// Favorites
	GetFavorites(ctx context.Context, userID uint) ([]model.BreathingFavorite, error)
	AddFavorite(ctx context.Context, favorite *model.BreathingFavorite) error
	RemoveFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) error
	IsFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) (bool, error)
	UpdateFavoriteOrder(ctx context.Context, userID uint, techniqueIDs []uuid.UUID) error

	// Calendar/Stats
	GetDailyStats(ctx context.Context, userID uint, date time.Time) (int, int, error) // sessions count, total minutes
	GetMonthlyCalendar(ctx context.Context, userID uint, year, month int) ([]DailyCalendarData, error)
	GetTechniqueUsageStats(ctx context.Context, userID uint) ([]TechniqueUsageData, error)
}

type DailyCalendarData struct {
	Date           time.Time
	SessionsCount  int
	TotalMinutes   int
	TechniquesUsed []string
}

type TechniqueUsageData struct {
	TechniqueID   uuid.UUID
	TechniqueName string
	Icon          string
	SessionsCount int64
	TotalMinutes  int
}

type breathingRepository struct {
	db *gorm.DB
}

func NewBreathingRepository(db *gorm.DB) BreathingRepository {
	return &breathingRepository{db: db}
}

// ==========================================
// Techniques Implementation
// ==========================================

func (r *breathingRepository) GetAllTechniques(ctx context.Context) ([]model.BreathingTechnique, error) {
	var techniques []model.BreathingTechnique
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("is_system DESC, name ASC").Find(&techniques).Error
	return techniques, err
}

func (r *breathingRepository) GetSystemTechniques(ctx context.Context) ([]model.BreathingTechnique, error) {
	var techniques []model.BreathingTechnique
	err := r.db.WithContext(ctx).Where("is_system = ? AND is_active = ?", true, true).Order("name ASC").Find(&techniques).Error
	return techniques, err
}

func (r *breathingRepository) GetUserTechniques(ctx context.Context, userID uint) ([]model.BreathingTechnique, error) {
	var techniques []model.BreathingTechnique
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Order("name ASC").Find(&techniques).Error
	return techniques, err
}

func (r *breathingRepository) GetTechniqueByID(ctx context.Context, id uuid.UUID) (*model.BreathingTechnique, error) {
	var technique model.BreathingTechnique
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&technique).Error
	if err != nil {
		return nil, err
	}
	return &technique, nil
}

func (r *breathingRepository) GetTechniqueBySlug(ctx context.Context, slug string) (*model.BreathingTechnique, error) {
	var technique model.BreathingTechnique
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&technique).Error
	if err != nil {
		return nil, err
	}
	return &technique, nil
}

func (r *breathingRepository) CreateTechnique(ctx context.Context, technique *model.BreathingTechnique) error {
	return r.db.WithContext(ctx).Create(technique).Error
}

func (r *breathingRepository) UpdateTechnique(ctx context.Context, technique *model.BreathingTechnique) error {
	return r.db.WithContext(ctx).Save(technique).Error
}

func (r *breathingRepository) DeleteTechnique(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.BreathingTechnique{}, "id = ?", id).Error
}

// ==========================================
// Sessions Implementation
// ==========================================

func (r *breathingRepository) CreateSession(ctx context.Context, session *model.BreathingSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *breathingRepository) UpdateSession(ctx context.Context, session *model.BreathingSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *breathingRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*model.BreathingSession, error) {
	var session model.BreathingSession
	err := r.db.WithContext(ctx).Preload("Technique").Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *breathingRepository) GetUserSessions(ctx context.Context, userID uint, startDate, endDate *time.Time, techniqueID *uuid.UUID, limit, offset int) ([]model.BreathingSession, int64, error) {
	var sessions []model.BreathingSession
	var total int64

	query := r.db.WithContext(ctx).Model(&model.BreathingSession{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("started_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("started_at <= ?", *endDate)
	}
	if techniqueID != nil {
		query = query.Where("technique_id = ?", *techniqueID)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Preload("Technique").
		Order("started_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error

	return sessions, total, err
}

func (r *breathingRepository) GetUserSessionsToday(ctx context.Context, userID uint) ([]model.BreathingSession, error) {
	var sessions []model.BreathingSession
	today := timeutil.Today()
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).Preload("Technique").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, today, tomorrow).
		Order("started_at DESC").
		Find(&sessions).Error

	return sessions, err
}

func (r *breathingRepository) GetUserSessionsByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]model.BreathingSession, error) {
	var sessions []model.BreathingSession
	err := r.db.WithContext(ctx).Preload("Technique").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, startDate, endDate).
		Order("started_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *breathingRepository) GetUserTotalSessionsCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Where("user_id = ? AND completed = ?", userID, true).
		Count(&count).Error
	return count, err
}

func (r *breathingRepository) GetUserTotalMinutes(ctx context.Context, userID uint) (int, error) {
	var result struct {
		TotalSeconds int
	}
	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select("COALESCE(SUM(duration_seconds), 0) as total_seconds").
		Where("user_id = ?", userID).
		Scan(&result).Error
	return result.TotalSeconds / 60, err
}

func (r *breathingRepository) GetUserTodayXP(ctx context.Context, userID uint) (int, error) {
	var result struct {
		TotalXP int
	}
	today := timeutil.Today()
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select("COALESCE(SUM(xp_earned), 0) as total_xp").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, today, tomorrow).
		Scan(&result).Error
	return result.TotalXP, err
}

func (r *breathingRepository) GetMostUsedTechnique(ctx context.Context, userID uint) (*model.BreathingTechnique, int64, error) {
	var result struct {
		TechniqueID uuid.UUID
		Count       int64
	}

	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select("technique_id, COUNT(*) as count").
		Where("user_id = ? AND completed = ?", userID, true).
		Group("technique_id").
		Order("count DESC").
		Limit(1).
		Scan(&result).Error

	if err != nil {
		return nil, 0, err
	}

	if result.TechniqueID == uuid.Nil {
		return nil, 0, nil
	}

	technique, err := r.GetTechniqueByID(ctx, result.TechniqueID)
	if err != nil {
		return nil, 0, err
	}

	return technique, result.Count, nil
}

func (r *breathingRepository) GetCompletionRate(ctx context.Context, userID uint) (float64, error) {
	var result struct {
		Total     int64
		Completed int64
	}

	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select("COUNT(*) as total, SUM(CASE WHEN completed THEN 1 ELSE 0 END) as completed").
		Where("user_id = ?", userID).
		Scan(&result).Error

	if err != nil {
		return 0, err
	}

	if result.Total == 0 {
		return 0, nil
	}

	return float64(result.Completed) / float64(result.Total) * 100, nil
}

func (r *breathingRepository) CountSessionsSince(ctx context.Context, userID uint, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Where("user_id = ? AND started_at >= ? AND completed = ?", userID, since, true).
		Count(&count).Error
	return count, err
}

// ==========================================
// Preferences Implementation
// ==========================================

func (r *breathingRepository) GetPreferences(ctx context.Context, userID uint) (*model.BreathingPreference, error) {
	var pref model.BreathingPreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *breathingRepository) CreatePreferences(ctx context.Context, pref *model.BreathingPreference) error {
	return r.db.WithContext(ctx).Create(pref).Error
}

func (r *breathingRepository) UpdatePreferences(ctx context.Context, pref *model.BreathingPreference) error {
	return r.db.WithContext(ctx).Save(pref).Error
}

func (r *breathingRepository) GetOrCreatePreferences(ctx context.Context, userID uint) (*model.BreathingPreference, error) {
	pref, err := r.GetPreferences(ctx, userID)
	if err == nil {
		return pref, nil
	}

	if err == gorm.ErrRecordNotFound {
		// Create default preferences
		newPref := &model.BreathingPreference{
			UserID:                 int(userID),
			DefaultDurationSeconds: 300, // 5 minutes
			VoiceGuidance:          "ask",
			BackgroundSound:        "ask",
			DefaultBackgroundSound: "none",
			HapticFeedback:         true,
			AnimationSpeed:         "normal",
			Theme:                  "default",
			ReminderEnabled:        false,
			ReminderDays:           "1,2,3,4,5,6,7", // all days
			TutorialCompleted:      false,
			CurrentStreak:          0,
			LongestStreak:          0,
			StreakFreezeAvailable:  true,
		}

		if err := r.CreatePreferences(ctx, newPref); err != nil {
			return nil, err
		}
		return newPref, nil
	}

	return nil, err
}

// ==========================================
// Favorites Implementation
// ==========================================

func (r *breathingRepository) GetFavorites(ctx context.Context, userID uint) ([]model.BreathingFavorite, error) {
	var favorites []model.BreathingFavorite
	err := r.db.WithContext(ctx).Preload("Technique").
		Where("user_id = ?", userID).
		Order("sort_order ASC").
		Find(&favorites).Error
	return favorites, err
}

func (r *breathingRepository) AddFavorite(ctx context.Context, favorite *model.BreathingFavorite) error {
	// Get current max sort_order
	var maxOrder int
	r.db.WithContext(ctx).Model(&model.BreathingFavorite{}).
		Where("user_id = ?", favorite.UserID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder)

	favorite.SortOrder = maxOrder + 1
	return r.db.WithContext(ctx).Create(favorite).Error
}

func (r *breathingRepository) RemoveFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND technique_id = ?", userID, techniqueID).
		Delete(&model.BreathingFavorite{}).Error
}

func (r *breathingRepository) IsFavorite(ctx context.Context, userID uint, techniqueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.BreathingFavorite{}).
		Where("user_id = ? AND technique_id = ?", userID, techniqueID).
		Count(&count).Error
	return count > 0, err
}

func (r *breathingRepository) UpdateFavoriteOrder(ctx context.Context, userID uint, techniqueIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, techID := range techniqueIDs {
			if err := tx.Model(&model.BreathingFavorite{}).
				Where("user_id = ? AND technique_id = ?", userID, techID).
				Update("sort_order", i+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ==========================================
// Calendar/Stats Implementation
// ==========================================

func (r *breathingRepository) GetDailyStats(ctx context.Context, userID uint, date time.Time) (int, int, error) {
	dayStart := timeutil.StartOfDay(date)
	dayEnd := dayStart.Add(24 * time.Hour)

	var result struct {
		Count        int
		TotalSeconds int
	}

	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select("COUNT(*) as count, COALESCE(SUM(duration_seconds), 0) as total_seconds").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, dayStart, dayEnd).
		Scan(&result).Error

	return result.Count, result.TotalSeconds / 60, err
}

func (r *breathingRepository) GetMonthlyCalendar(ctx context.Context, userID uint, year, month int) ([]DailyCalendarData, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	var rawData []struct {
		Date           string
		SessionsCount  int
		TotalSeconds   int
		TechniqueNames string
	}

	aggregateExpr := "STRING_AGG(DISTINCT bt.name, ',')"
	if r.db.Dialector.Name() == "sqlite" {
		aggregateExpr = "GROUP_CONCAT(DISTINCT bt.name)"
	}

	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select(`
			DATE(started_at) as date,
			COUNT(*) as sessions_count,
			COALESCE(SUM(duration_seconds), 0) as total_seconds,
			`+aggregateExpr+` as technique_names
		`).
		Joins("LEFT JOIN breathing_techniques bt ON bt.id = breathing_sessions.technique_id").
		Where("breathing_sessions.user_id = ? AND breathing_sessions.started_at >= ? AND breathing_sessions.started_at < ?", userID, startDate, endDate).
		Group("DATE(started_at)").
		Order("date ASC").
		Scan(&rawData).Error

	if err != nil {
		return nil, err
	}

	result := make([]DailyCalendarData, len(rawData))
	for i, d := range rawData {
		parsedDate, _ := time.ParseInLocation("2006-01-02", d.Date, time.Local)
		var techniques []string
		if d.TechniqueNames != "" {
			techniques = splitAndTrim(d.TechniqueNames)
		}
		result[i] = DailyCalendarData{
			Date:           parsedDate,
			SessionsCount:  d.SessionsCount,
			TotalMinutes:   d.TotalSeconds / 60,
			TechniquesUsed: techniques,
		}
	}

	return result, nil
}

func (r *breathingRepository) GetTechniqueUsageStats(ctx context.Context, userID uint) ([]TechniqueUsageData, error) {
	var result []TechniqueUsageData

	err := r.db.WithContext(ctx).Model(&model.BreathingSession{}).
		Select(`
			breathing_sessions.technique_id,
			bt.name as technique_name,
			bt.icon,
			COUNT(*) as sessions_count,
			COALESCE(SUM(breathing_sessions.duration_seconds), 0) / 60 as total_minutes
		`).
		Joins("LEFT JOIN breathing_techniques bt ON bt.id = breathing_sessions.technique_id").
		Where("breathing_sessions.user_id = ? AND breathing_sessions.completed = ?", userID, true).
		Group("breathing_sessions.technique_id, bt.name, bt.icon").
		Order("sessions_count DESC").
		Scan(&result).Error

	return result, err
}

// Helper function to split comma-separated strings
func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0)
	for _, p := range splitString(s, ',') {
		trimmed := trimSpace(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func splitString(s string, sep rune) []string {
	result := make([]string, 0)
	current := ""
	for _, c := range s {
		if c == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
