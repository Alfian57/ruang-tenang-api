package userctx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/cache"
)

// MoodRepo is the port interface for mood repository access
type MoodRepo interface {
	FindTodayByUserID(ctx context.Context, userID uint) (*model.UserMood, error)
}

// UserContextCache provides fast user-context lookups for AI chat
// using the existing in-memory CacheService (no Redis needed).
type UserContextCache struct {
	cacheService *cache.CacheService
	moodRepo     MoodRepo
}

// NewUserContextCache creates a new UserContextCache
func NewUserContextCache(cacheService *cache.CacheService, moodRepo MoodRepo) *UserContextCache {
	return &UserContextCache{
		cacheService: cacheService,
		moodRepo:     moodRepo,
	}
}

// Cache key prefixes and TTLs
const (
	moodCachePrefix    = "user_mood:"
	journalCachePrefix = "user_journal_summary:"
	moodCacheTTL       = 24 * time.Hour
	journalCacheTTL    = 6 * time.Hour
)

// MoodContext holds cached mood info for a user
type MoodContext struct {
	Mood  string
	Emoji string
}

// --- Mood Context ---

// GetMoodContext returns the cached mood for a user, or fetches from DB if not cached.
func (c *UserContextCache) GetMoodContext(ctx context.Context, userID uint) *MoodContext {
	key := fmt.Sprintf("%s%d", moodCachePrefix, userID)

	if cached := c.cacheService.Get(key); cached != nil {
		if mc, ok := cached.(*MoodContext); ok {
			return mc
		}
	}

	// Cache miss: fetch from DB and cache
	mood, err := c.moodRepo.FindTodayByUserID(ctx, userID)
	if err != nil || mood == nil {
		return nil
	}

	mc := &MoodContext{
		Mood:  string(mood.Mood),
		Emoji: mood.GetMoodEmoji(),
	}
	c.cacheService.SetWithTTL(key, mc, moodCacheTTL)
	return mc
}

// SetMoodContext updates the cached mood after user records a new mood.
func (c *UserContextCache) SetMoodContext(userID uint, mood string, emoji string) {
	key := fmt.Sprintf("%s%d", moodCachePrefix, userID)
	c.cacheService.SetWithTTL(key, &MoodContext{
		Mood:  mood,
		Emoji: emoji,
	}, moodCacheTTL)
}

// --- Journal Summary Context ---

// GetJournalSummary returns the cached journal summary for a user.
func (c *UserContextCache) GetJournalSummary(userID uint) string {
	key := fmt.Sprintf("%s%d", journalCachePrefix, userID)
	if cached := c.cacheService.Get(key); cached != nil {
		if s, ok := cached.(string); ok {
			return s
		}
	}
	return ""
}

// SetJournalSummary caches a journal summary for a user.
func (c *UserContextCache) SetJournalSummary(userID uint, summary string) {
	key := fmt.Sprintf("%s%d", journalCachePrefix, userID)
	c.cacheService.SetWithTTL(key, summary, journalCacheTTL)
}

// InvalidateUserContext clears all cached context for a user.
func (c *UserContextCache) InvalidateUserContext(userID uint) {
	c.cacheService.Delete(fmt.Sprintf("%s%d", moodCachePrefix, userID))
	c.cacheService.Delete(fmt.Sprintf("%s%d", journalCachePrefix, userID))
}

// BuildUserContextPrompt generates a concise context string for the AI system prompt.
func (c *UserContextCache) BuildUserContextPrompt(ctx context.Context, userID uint) string {
	var parts []string

	if mc := c.GetMoodContext(ctx, userID); mc != nil {
		parts = append(parts, fmt.Sprintf("Mood user hari ini: %s %s", mc.Emoji, mc.Mood))
	}

	if summary := c.GetJournalSummary(userID); summary != "" {
		parts = append(parts, fmt.Sprintf("Ringkasan jurnal terbaru: %s", summary))
	}

	if len(parts) == 0 {
		return ""
	}

	return "\n\n=== KONTEKS USER SAAT INI ===\n" + strings.Join(parts, "\n") + "\n=== AKHIR KONTEKS USER ===\n"
}
