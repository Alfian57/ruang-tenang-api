package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a single cached item with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// CacheService provides a generic in-memory caching solution
type CacheService struct {
	data map[string]*CacheEntry
	mu   sync.RWMutex

	// Configurable TTLs for different cache types
	DefaultTTL     time.Duration
	LevelConfigTTL time.Duration
	CategoryTTL    time.Duration
}

// NewCacheService creates a new CacheService with default TTLs
func NewCacheService() *CacheService {
	return &CacheService{
		data:           make(map[string]*CacheEntry),
		DefaultTTL:     30 * time.Minute,
		LevelConfigTTL: 1 * time.Hour, // Level configs rarely change
		CategoryTTL:    30 * time.Minute,
	}
}

// Get retrieves an item from cache, returns nil if not found or expired
func (c *CacheService) Get(key string) interface{} {
	c.mu.RLock()
	entry, exists := c.data[key]
	c.mu.RUnlock()

	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry.Data
}

// Set stores an item in cache with the default TTL
func (c *CacheService) Set(key string, data interface{}) {
	c.SetWithTTL(key, data, c.DefaultTTL)
}

// SetWithTTL stores an item in cache with a custom TTL
func (c *CacheService) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes an item from cache
func (c *CacheService) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// DeletePrefix removes all items with keys starting with the given prefix
func (c *CacheService) DeletePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.data {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.data, key)
		}
	}
}

// Clear removes all items from cache
func (c *CacheService) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*CacheEntry)
}

// Cache keys constants
const (
	CacheKeyLevelConfigs      = "level_configs:all"
	CacheKeyLevelByExp        = "level_configs:exp:"
	CacheKeySongCategories    = "song_categories:all"
	CacheKeyForumCategories   = "forum_categories:all"
	CacheKeyArticleCategories = "article_categories:all"
)
