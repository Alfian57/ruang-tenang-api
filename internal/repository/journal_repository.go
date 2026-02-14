package repository

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// JournalRepository handles database operations for journals
type JournalRepository struct {
	db *gorm.DB
}

// NewJournalRepository creates a new JournalRepository instance
func NewJournalRepository(db *gorm.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

// ===== Journal CRUD =====

// Create creates a new journal entry
func (r *JournalRepository) Create(journal *model.Journal) error {
	return r.db.Create(journal).Error
}

// FindByID finds a journal by ID
func (r *JournalRepository) FindByID(id uint) (*model.Journal, error) {
	var journal model.Journal
	err := r.db.Preload("Mood").First(&journal, id).Error
	if err != nil {
		return nil, err
	}
	return &journal, nil
}

// FindByIDAndUserID finds a journal by ID and user ID (for authorization)
func (r *JournalRepository) FindByIDAndUserID(id, userID uint) (*model.Journal, error) {
	var journal model.Journal
	err := r.db.Preload("Mood").Where("id = ? AND user_id = ?", id, userID).First(&journal).Error
	if err != nil {
		return nil, err
	}
	return &journal, nil
}

// Update updates a journal entry
func (r *JournalRepository) Update(journal *model.Journal) error {
	return r.db.Save(journal).Error
}

// Delete deletes a journal entry
func (r *JournalRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Journal{}).Error
}

// FindByUserID finds all journals for a user with pagination
func (r *JournalRepository) FindByUserID(userID uint, page, limit int, tags []string, startDate, endDate *time.Time) ([]model.Journal, int64, error) {
	var journals []model.Journal
	var total int64

	query := r.db.Model(&model.Journal{}).Where("user_id = ?", userID)

	// Filter by tags if provided
	if len(tags) > 0 {
		query = query.Where("tags && ?", pq.StringArray(tags))
	}

	// Filter by date range
	if startDate != nil {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != nil {
		query = query.Where("created_at <= ?", endDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginate and fetch
	offset := (page - 1) * limit
	err := query.Preload("Mood").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&journals).Error

	return journals, total, err
}

// SearchByContent searches journals by content
func (r *JournalRepository) SearchByContent(userID uint, query string, limit int) ([]model.Journal, error) {
	var journals []model.Journal
	err := r.db.Preload("Mood").
		Where("user_id = ? AND (content ILIKE ? OR title ILIKE ?)", userID, "%"+query+"%", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&journals).Error
	return journals, err
}

// ===== AI Context Functions =====

// FindForAIContext finds journal entries that are shared with AI
func (r *JournalRepository) FindForAIContext(userID uint, daysBack, maxEntries int) ([]model.Journal, error) {
	var journals []model.Journal

	cutoffDate := time.Now().AddDate(0, 0, -daysBack)

	err := r.db.Preload("Mood").
		Where("user_id = ? AND share_with_ai = ? AND created_at >= ?", userID, true, cutoffDate).
		Order("created_at DESC").
		Limit(maxEntries).
		Find(&journals).Error

	return journals, err
}

// FindRelevantForAIContext finds journal entries relevant to a query
func (r *JournalRepository) FindRelevantForAIContext(userID uint, query string, maxEntries int) ([]model.Journal, error) {
	var journals []model.Journal

	err := r.db.Preload("Mood").
		Where("user_id = ? AND share_with_ai = ? AND (content ILIKE ? OR title ILIKE ? OR ? = ANY(tags))",
			userID, true, "%"+query+"%", "%"+query+"%", query).
		Order("created_at DESC").
		Limit(maxEntries).
		Find(&journals).Error

	return journals, err
}

// UpdateAIAccessedAt updates the AI accessed timestamp for a journal
func (r *JournalRepository) UpdateAIAccessedAt(id uint) error {
	now := time.Now()
	return r.db.Model(&model.Journal{}).Where("id = ?", id).Update("ai_accessed_at", now).Error
}

// CountByUserID counts total journals for a user
func (r *JournalRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Journal{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountSharedWithAI counts journals shared with AI
func (r *JournalRepository) CountSharedWithAI(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Journal{}).Where("user_id = ? AND share_with_ai = ?", userID, true).Count(&count).Error
	return count, err
}

// ===== Analytics Functions =====

// GetMoodDistribution gets mood distribution for a user's journals
func (r *JournalRepository) GetMoodDistribution(userID uint) (map[string]int, error) {
	type MoodCount struct {
		Mood  string
		Count int
	}

	var results []MoodCount
	err := r.db.Model(&model.Journal{}).
		Select("user_moods.mood, COUNT(*) as count").
		Joins("LEFT JOIN user_moods ON journals.mood_id = user_moods.id").
		Where("journals.user_id = ? AND journals.mood_id IS NOT NULL", userID).
		Group("user_moods.mood").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	distribution := make(map[string]int)
	for _, r := range results {
		distribution[r.Mood] = r.Count
	}

	return distribution, nil
}

// GetTagFrequency gets tag frequency for a user's journals
func (r *JournalRepository) GetTagFrequency(userID uint) (map[string]int, error) {
	type TagCount struct {
		Tag   string
		Count int
	}

	var results []TagCount
	err := r.db.Raw(`
		SELECT unnest(tags) as tag, COUNT(*) as count 
		FROM journals 
		WHERE user_id = ? AND tags IS NOT NULL AND array_length(tags, 1) > 0
		GROUP BY unnest(tags)
		ORDER BY count DESC
		LIMIT 20
	`, userID).Scan(&results).Error

	if err != nil {
		return nil, err
	}

	frequency := make(map[string]int)
	for _, r := range results {
		frequency[r.Tag] = r.Count
	}

	return frequency, nil
}

// GetTotalWordCount gets total word count for a user's journals
func (r *JournalRepository) GetTotalWordCount(userID uint) (int, error) {
	var total int
	err := r.db.Model(&model.Journal{}).
		Select("COALESCE(SUM(word_count), 0)").
		Where("user_id = ?", userID).
		Scan(&total).Error
	return total, err
}

// GetEntriesByMonth gets entry count by month
func (r *JournalRepository) GetEntriesByMonth(userID uint, months int) ([]struct {
	Month string
	Count int
}, error) {
	var results []struct {
		Month string
		Count int
	}

	err := r.db.Raw(`
		SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COUNT(*) as count 
		FROM journals 
		WHERE user_id = ? AND created_at >= NOW() - INTERVAL '1 month' * ?
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month DESC
	`, userID, months).Scan(&results).Error

	return results, err
}

// GetWritingStreak gets current writing streak
func (r *JournalRepository) GetWritingStreak(userID uint) (int, error) {
	var dates []time.Time
	err := r.db.Model(&model.Journal{}).
		Select("DATE(created_at)").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Distinct().
		Limit(365).
		Pluck("DATE(created_at)", &dates).Error

	if err != nil || len(dates) == 0 {
		return 0, err
	}

	streak := 0
	today := time.Now().Truncate(24 * time.Hour)

	for i, date := range dates {
		expectedDate := today.AddDate(0, 0, -i)
		if date.Truncate(24 * time.Hour).Equal(expectedDate) {
			streak++
		} else {
			break
		}
	}

	return streak, nil
}

// ===== Settings Functions =====

// JournalSettingsRepository handles journal settings database operations
type JournalSettingsRepository struct {
	db *gorm.DB
}

// NewJournalSettingsRepository creates a new JournalSettingsRepository instance
func NewJournalSettingsRepository(db *gorm.DB) *JournalSettingsRepository {
	return &JournalSettingsRepository{db: db}
}

// FindByUserID finds settings by user ID
func (r *JournalSettingsRepository) FindByUserID(userID uint) (*model.JournalSettings, error) {
	var settings model.JournalSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// FindOrCreate finds or creates settings for a user
func (r *JournalSettingsRepository) FindOrCreate(userID uint) (*model.JournalSettings, error) {
	var settings model.JournalSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		settings = model.JournalSettings{
			UserID:              userID,
			AllowAIAccess:       false,
			AIContextDays:       7,
			AIContextMaxEntries: 5,
			DefaultShareWithAI:  false,
		}
		if err := r.db.Create(&settings).Error; err != nil {
			return nil, err
		}
		return &settings, nil
	}
	return &settings, err
}

// Update updates journal settings
func (r *JournalSettingsRepository) Update(settings *model.JournalSettings) error {
	return r.db.Save(settings).Error
}

// ===== AI Access Log Functions =====

// JournalAIAccessLogRepository handles AI access log database operations
type JournalAIAccessLogRepository struct {
	db *gorm.DB
}

// NewJournalAIAccessLogRepository creates a new JournalAIAccessLogRepository instance
func NewJournalAIAccessLogRepository(db *gorm.DB) *JournalAIAccessLogRepository {
	return &JournalAIAccessLogRepository{db: db}
}

// Create creates a new AI access log entry
func (r *JournalAIAccessLogRepository) Create(log *model.JournalAIAccessLog) error {
	return r.db.Create(log).Error
}

// FindByUserID finds AI access logs for a user
func (r *JournalAIAccessLogRepository) FindByUserID(userID uint, limit int) ([]model.JournalAIAccessLog, error) {
	var logs []model.JournalAIAccessLog
	err := r.db.Preload("Journal").
		Where("user_id = ?", userID).
		Order("accessed_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// FindByJournalID finds AI access logs for a specific journal
func (r *JournalAIAccessLogRepository) FindByJournalID(journalID uint) ([]model.JournalAIAccessLog, error) {
	var logs []model.JournalAIAccessLog
	err := r.db.Where("journal_id = ?", journalID).
		Order("accessed_at DESC").
		Find(&logs).Error
	return logs, err
}
