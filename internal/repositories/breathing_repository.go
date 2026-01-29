package repositories

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BreathingRepository interface {
	// Techniques
	GetAllTechniques() ([]models.BreathingTechnique, error)
	GetSystemTechniques() ([]models.BreathingTechnique, error)
	GetUserTechniques(userID uint) ([]models.BreathingTechnique, error)
	GetTechniqueByID(id uuid.UUID) (*models.BreathingTechnique, error)
	GetTechniqueBySlug(slug string) (*models.BreathingTechnique, error)
	CreateTechnique(technique *models.BreathingTechnique) error
	UpdateTechnique(technique *models.BreathingTechnique) error
	DeleteTechnique(id uuid.UUID) error

	// Sessions
	CreateSession(session *models.BreathingSession) error
	UpdateSession(session *models.BreathingSession) error
	GetSessionByID(id uuid.UUID) (*models.BreathingSession, error)
	GetUserSessions(userID uint, startDate, endDate *time.Time, techniqueID *uuid.UUID, limit, offset int) ([]models.BreathingSession, int64, error)
	GetUserSessionsToday(userID uint) ([]models.BreathingSession, error)
	GetUserSessionsByDateRange(userID uint, startDate, endDate time.Time) ([]models.BreathingSession, error)
	GetUserTotalSessionsCount(userID uint) (int64, error)
	GetUserTotalMinutes(userID uint) (int, error)
	GetUserTodayXP(userID uint) (int, error)
	GetMostUsedTechnique(userID uint) (*models.BreathingTechnique, int64, error)
	GetCompletionRate(userID uint) (float64, error)
	CountSessionsSince(userID uint, since time.Time) (int64, error)

	// Preferences
	GetPreferences(userID uint) (*models.BreathingPreference, error)
	CreatePreferences(pref *models.BreathingPreference) error
	UpdatePreferences(pref *models.BreathingPreference) error
	GetOrCreatePreferences(userID uint) (*models.BreathingPreference, error)

	// Favorites
	GetFavorites(userID uint) ([]models.BreathingFavorite, error)
	AddFavorite(favorite *models.BreathingFavorite) error
	RemoveFavorite(userID uint, techniqueID uuid.UUID) error
	IsFavorite(userID uint, techniqueID uuid.UUID) (bool, error)
	UpdateFavoriteOrder(userID uint, techniqueIDs []uuid.UUID) error

	// Calendar/Stats
	GetDailyStats(userID uint, date time.Time) (int, int, error) // sessions count, total minutes
	GetMonthlyCalendar(userID uint, year, month int) ([]DailyCalendarData, error)
	GetTechniqueUsageStats(userID uint) ([]TechniqueUsageData, error)
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

func (r *breathingRepository) GetAllTechniques() ([]models.BreathingTechnique, error) {
	var techniques []models.BreathingTechnique
	err := r.db.Where("is_active = ?", true).Order("is_system DESC, name ASC").Find(&techniques).Error
	return techniques, err
}

func (r *breathingRepository) GetSystemTechniques() ([]models.BreathingTechnique, error) {
	var techniques []models.BreathingTechnique
	err := r.db.Where("is_system = ? AND is_active = ?", true, true).Order("name ASC").Find(&techniques).Error
	return techniques, err
}

func (r *breathingRepository) GetUserTechniques(userID uint) ([]models.BreathingTechnique, error) {
	var techniques []models.BreathingTechnique
	err := r.db.Where("user_id = ? AND is_active = ?", userID, true).Order("name ASC").Find(&techniques).Error
	return techniques, err
}

func (r *breathingRepository) GetTechniqueByID(id uuid.UUID) (*models.BreathingTechnique, error) {
	var technique models.BreathingTechnique
	err := r.db.Where("id = ?", id).First(&technique).Error
	if err != nil {
		return nil, err
	}
	return &technique, nil
}

func (r *breathingRepository) GetTechniqueBySlug(slug string) (*models.BreathingTechnique, error) {
	var technique models.BreathingTechnique
	err := r.db.Where("slug = ?", slug).First(&technique).Error
	if err != nil {
		return nil, err
	}
	return &technique, nil
}

func (r *breathingRepository) CreateTechnique(technique *models.BreathingTechnique) error {
	return r.db.Create(technique).Error
}

func (r *breathingRepository) UpdateTechnique(technique *models.BreathingTechnique) error {
	return r.db.Save(technique).Error
}

func (r *breathingRepository) DeleteTechnique(id uuid.UUID) error {
	return r.db.Delete(&models.BreathingTechnique{}, "id = ?", id).Error
}

// ==========================================
// Sessions Implementation
// ==========================================

func (r *breathingRepository) CreateSession(session *models.BreathingSession) error {
	return r.db.Create(session).Error
}

func (r *breathingRepository) UpdateSession(session *models.BreathingSession) error {
	return r.db.Save(session).Error
}

func (r *breathingRepository) GetSessionByID(id uuid.UUID) (*models.BreathingSession, error) {
	var session models.BreathingSession
	err := r.db.Preload("Technique").Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *breathingRepository) GetUserSessions(userID uint, startDate, endDate *time.Time, techniqueID *uuid.UUID, limit, offset int) ([]models.BreathingSession, int64, error) {
	var sessions []models.BreathingSession
	var total int64

	query := r.db.Model(&models.BreathingSession{}).Where("user_id = ?", userID)

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

func (r *breathingRepository) GetUserSessionsToday(userID uint) ([]models.BreathingSession, error) {
	var sessions []models.BreathingSession
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.Preload("Technique").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, today, tomorrow).
		Order("started_at DESC").
		Find(&sessions).Error

	return sessions, err
}

func (r *breathingRepository) GetUserSessionsByDateRange(userID uint, startDate, endDate time.Time) ([]models.BreathingSession, error) {
	var sessions []models.BreathingSession
	err := r.db.Preload("Technique").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, startDate, endDate).
		Order("started_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *breathingRepository) GetUserTotalSessionsCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.BreathingSession{}).
		Where("user_id = ? AND completed = ?", userID, true).
		Count(&count).Error
	return count, err
}

func (r *breathingRepository) GetUserTotalMinutes(userID uint) (int, error) {
	var result struct {
		TotalSeconds int
	}
	err := r.db.Model(&models.BreathingSession{}).
		Select("COALESCE(SUM(duration_seconds), 0) as total_seconds").
		Where("user_id = ?", userID).
		Scan(&result).Error
	return result.TotalSeconds / 60, err
}

func (r *breathingRepository) GetUserTodayXP(userID uint) (int, error) {
	var result struct {
		TotalXP int
	}
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	err := r.db.Model(&models.BreathingSession{}).
		Select("COALESCE(SUM(xp_earned), 0) as total_xp").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, today, tomorrow).
		Scan(&result).Error
	return result.TotalXP, err
}

func (r *breathingRepository) GetMostUsedTechnique(userID uint) (*models.BreathingTechnique, int64, error) {
	var result struct {
		TechniqueID uuid.UUID
		Count       int64
	}

	err := r.db.Model(&models.BreathingSession{}).
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

	technique, err := r.GetTechniqueByID(result.TechniqueID)
	if err != nil {
		return nil, 0, err
	}

	return technique, result.Count, nil
}

func (r *breathingRepository) GetCompletionRate(userID uint) (float64, error) {
	var result struct {
		Total     int64
		Completed int64
	}

	err := r.db.Model(&models.BreathingSession{}).
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

func (r *breathingRepository) CountSessionsSince(userID uint, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&models.BreathingSession{}).
		Where("user_id = ? AND started_at >= ? AND completed = ?", userID, since, true).
		Count(&count).Error
	return count, err
}

// ==========================================
// Preferences Implementation
// ==========================================

func (r *breathingRepository) GetPreferences(userID uint) (*models.BreathingPreference, error) {
	var pref models.BreathingPreference
	err := r.db.Where("user_id = ?", userID).First(&pref).Error
	if err != nil {
		return nil, err
	}
	return &pref, nil
}

func (r *breathingRepository) CreatePreferences(pref *models.BreathingPreference) error {
	return r.db.Create(pref).Error
}

func (r *breathingRepository) UpdatePreferences(pref *models.BreathingPreference) error {
	return r.db.Save(pref).Error
}

func (r *breathingRepository) GetOrCreatePreferences(userID uint) (*models.BreathingPreference, error) {
	pref, err := r.GetPreferences(userID)
	if err == nil {
		return pref, nil
	}

	if err == gorm.ErrRecordNotFound {
		// Create default preferences
		newPref := &models.BreathingPreference{
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

		if err := r.CreatePreferences(newPref); err != nil {
			return nil, err
		}
		return newPref, nil
	}

	return nil, err
}

// ==========================================
// Favorites Implementation
// ==========================================

func (r *breathingRepository) GetFavorites(userID uint) ([]models.BreathingFavorite, error) {
	var favorites []models.BreathingFavorite
	err := r.db.Preload("Technique").
		Where("user_id = ?", userID).
		Order("sort_order ASC").
		Find(&favorites).Error
	return favorites, err
}

func (r *breathingRepository) AddFavorite(favorite *models.BreathingFavorite) error {
	// Get current max sort_order
	var maxOrder int
	r.db.Model(&models.BreathingFavorite{}).
		Where("user_id = ?", favorite.UserID).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder)

	favorite.SortOrder = maxOrder + 1
	return r.db.Create(favorite).Error
}

func (r *breathingRepository) RemoveFavorite(userID uint, techniqueID uuid.UUID) error {
	return r.db.Where("user_id = ? AND technique_id = ?", userID, techniqueID).
		Delete(&models.BreathingFavorite{}).Error
}

func (r *breathingRepository) IsFavorite(userID uint, techniqueID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.BreathingFavorite{}).
		Where("user_id = ? AND technique_id = ?", userID, techniqueID).
		Count(&count).Error
	return count > 0, err
}

func (r *breathingRepository) UpdateFavoriteOrder(userID uint, techniqueIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i, techID := range techniqueIDs {
			if err := tx.Model(&models.BreathingFavorite{}).
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

func (r *breathingRepository) GetDailyStats(userID uint, date time.Time) (int, int, error) {
	dayStart := date.Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)

	var result struct {
		Count        int
		TotalSeconds int
	}

	err := r.db.Model(&models.BreathingSession{}).
		Select("COUNT(*) as count, COALESCE(SUM(duration_seconds), 0) as total_seconds").
		Where("user_id = ? AND started_at >= ? AND started_at < ?", userID, dayStart, dayEnd).
		Scan(&result).Error

	return result.Count, result.TotalSeconds / 60, err
}

func (r *breathingRepository) GetMonthlyCalendar(userID uint, year, month int) ([]DailyCalendarData, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	var rawData []struct {
		Date           time.Time
		SessionsCount  int
		TotalSeconds   int
		TechniqueNames string
	}

	err := r.db.Model(&models.BreathingSession{}).
		Select(`
			DATE(started_at) as date,
			COUNT(*) as sessions_count,
			COALESCE(SUM(duration_seconds), 0) as total_seconds,
			STRING_AGG(DISTINCT bt.name, ',') as technique_names
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
		var techniques []string
		if d.TechniqueNames != "" {
			techniques = splitAndTrim(d.TechniqueNames)
		}
		result[i] = DailyCalendarData{
			Date:           d.Date,
			SessionsCount:  d.SessionsCount,
			TotalMinutes:   d.TotalSeconds / 60,
			TechniquesUsed: techniques,
		}
	}

	return result, nil
}

func (r *breathingRepository) GetTechniqueUsageStats(userID uint) ([]TechniqueUsageData, error) {
	var result []TechniqueUsageData

	err := r.db.Model(&models.BreathingSession{}).
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
