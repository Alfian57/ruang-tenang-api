package contentctx

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"

	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"go.uber.org/zap"
)

// Port interfaces for cross-feature dependencies
type ArticleRepo interface {
	FindPublished(ctx context.Context, categoryID uint, search string, page, limit int) ([]model.Article, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Article, error)
}

type SongRepo interface {
	CountByCategoryID(ctx context.Context, categoryID uint) int64
}

type SongCategoryRepo interface {
	FindAll(ctx context.Context) ([]model.SongCategory, error)
}

type ForumRepo interface {
	GetForums(ctx context.Context, limit, offset int, search string, categoryID *uint) ([]model.Forum, int64, error)
	FindUpdatedSince(ctx context.Context, since time.Time) ([]model.Forum, error)
	GetRepliesCount(ctx context.Context, forumID uint) (int64, error)
}

// ContentSummary types for lightweight caching
type ArticleSummary struct {
	ID       uint
	Title    string
	Category string
}

type SongCategorySummary struct {
	ID        uint
	Name      string
	SongCount int64
}

type ForumSummary struct {
	ID           uint
	Title        string
	RepliesCount int64
}

// ContentContextService provides cached content context for AI chat
// Uses background goroutine for pre-warming and incremental sync
type ContentContextService struct {
	articleRepo      ArticleRepo
	songRepo         SongRepo
	songCategoryRepo SongCategoryRepo
	forumRepo        ForumRepo

	// Map-based cache for efficient updates
	articles       map[uint]*ArticleSummary
	songCategories map[uint]*SongCategorySummary
	forums         map[uint]*ForumSummary

	// Sync tracking
	lastArticleSync time.Time
	lastForumSync   time.Time
	lastMusicSync   time.Time
	syncInterval    time.Duration

	// Status
	isReady       bool
	cachedContext string
	contextDirty  bool
	mu            sync.RWMutex

	// Shutdown
	stopChan chan struct{}
}

// NewContentContextService creates a new ContentContextService and starts background sync
func NewContentContextService(
	articleRepo ArticleRepo,
	songRepo SongRepo,
	songCategoryRepo SongCategoryRepo,
	forumRepo ForumRepo,
) *ContentContextService {
	s := &ContentContextService{
		articleRepo:      articleRepo,
		songRepo:         songRepo,
		songCategoryRepo: songCategoryRepo,
		forumRepo:        forumRepo,

		articles:       make(map[uint]*ArticleSummary),
		songCategories: make(map[uint]*SongCategorySummary),
		forums:         make(map[uint]*ForumSummary),

		syncInterval: 1 * time.Minute,
		stopChan:     make(chan struct{}),
	}

	// Start background pre-warm and sync
	go s.backgroundSync(context.Background())

	return s
}

// GetContentContext returns the content context for AI
// Returns partial context if cache is still loading
func (s *ContentContextService) GetContentContext(ctx context.Context) string {
	s.mu.RLock()

	// Return cached context if available and not dirty
	if s.cachedContext != "" && !s.contextDirty {
		defer s.mu.RUnlock()
		return s.cachedContext
	}

	// Need to upgrade to write lock - release read lock first
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.cachedContext != "" && !s.contextDirty {
		return s.cachedContext
	}

	s.cachedContext = s.buildContextFromMaps(ctx)
	s.contextDirty = false

	return s.cachedContext
}

// NotifyArticleChange notifies the cache of an article change (for event-driven updates)
func (s *ContentContextService) NotifyArticleChange(ctx context.Context, article *model.Article) {
	if article == nil || article.Status != model.ArticleStatusPublished {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	categoryName := "Umum"
	if article.Category.Name != "" {
		categoryName = article.Category.Name
	}

	s.articles[article.ID] = &ArticleSummary{
		ID:       article.ID,
		Title:    article.Title,
		Category: categoryName,
	}
	s.contextDirty = true
}

// NotifyArticleDelete notifies the cache of an article deletion
func (s *ContentContextService) NotifyArticleDelete(ctx context.Context, articleID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.articles, articleID)
	s.contextDirty = true
}

// NotifyForumChange notifies the cache of a forum change
func (s *ContentContextService) NotifyForumChange(ctx context.Context, forum *model.Forum) {
	if forum == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.forums[forum.ID] = &ForumSummary{
		ID:           forum.ID,
		Title:        forum.Title,
		RepliesCount: forum.RepliesCount,
	}
	s.contextDirty = true
}

// NotifyForumDelete notifies the cache of a forum deletion
func (s *ContentContextService) NotifyForumDelete(ctx context.Context, forumID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.forums, forumID)
	s.contextDirty = true
}

// Stop stops the background sync goroutine
func (s *ContentContextService) Stop(ctx context.Context) {
	close(s.stopChan)
}

// backgroundSync runs the background pre-warm and periodic sync
func (s *ContentContextService) backgroundSync(ctx context.Context) {
	// Initial pre-warm
	logger.Info("content context: starting initial pre-warm")
	s.fullSync(ctx)
	logger.Info("content context: pre-warm complete",
		zap.Int("articles", len(s.articles)),
		zap.Int("song_categories", len(s.songCategories)),
		zap.Int("forums", len(s.forums)))

	// Periodic sync
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.incrementalSync(ctx)
		case <-s.stopChan:
			logger.Info("content context: background sync stopped")
			return
		}
	}
}

// fullSync fetches all data from database
func (s *ContentContextService) fullSync(ctx context.Context) {
	s.syncArticles(ctx, time.Time{}) // Zero time = fetch all
	s.syncForums(ctx, time.Time{})
	s.syncMusic(ctx)

	s.mu.Lock()
	s.isReady = true
	s.contextDirty = true
	s.mu.Unlock()
}

// incrementalSync fetches only updated records since last sync
func (s *ContentContextService) incrementalSync(ctx context.Context) {
	s.syncArticles(ctx, s.lastArticleSync)
	s.syncForums(ctx, s.lastForumSync)
	s.syncMusic(ctx) // Music is typically small, just refresh all
}

// syncArticles syncs articles since the given time
func (s *ContentContextService) syncArticles(ctx context.Context, since time.Time) {
	var articles []model.Article
	var err error

	if since.IsZero() {
		// Full sync - get all published articles
		articles, _, err = s.articleRepo.FindPublished(ctx, 0, "", 1, 10000)
	} else {
		// Incremental sync - get updated since last sync
		articles, err = s.articleRepo.FindUpdatedSince(ctx, since)
	}

	if err != nil {
		logger.Error("content context: failed to sync articles", zap.Error(err))
		return
	}

	s.mu.Lock()
	for _, article := range articles {
		if article.Status == model.ArticleStatusPublished {
			categoryName := "Umum"
			if article.Category.Name != "" {
				categoryName = article.Category.Name
			}
			s.articles[article.ID] = &ArticleSummary{
				ID:       article.ID,
				Title:    article.Title,
				Category: categoryName,
			}
		} else {
			// Article unpublished or blocked, remove from cache
			delete(s.articles, article.ID)
		}
	}
	if len(articles) > 0 {
		s.contextDirty = true
	}
	s.lastArticleSync = time.Now()
	s.mu.Unlock()
}

// syncForums syncs forums since the given time
func (s *ContentContextService) syncForums(ctx context.Context, since time.Time) {
	var forums []model.Forum
	var err error

	if since.IsZero() {
		// Full sync
		forums, _, err = s.forumRepo.GetForums(ctx, 10000, 0, "", nil)
	} else {
		// Incremental sync
		forums, err = s.forumRepo.FindUpdatedSince(ctx, since)
	}

	if err != nil {
		logger.Error("content context: failed to sync forums", zap.Error(err))
		return
	}

	s.mu.Lock()
	for _, forum := range forums {
		repliesCount, _ := s.forumRepo.GetRepliesCount(ctx, forum.ID)
		s.forums[forum.ID] = &ForumSummary{
			ID:           forum.ID,
			Title:        forum.Title,
			RepliesCount: repliesCount,
		}
	}
	if len(forums) > 0 {
		s.contextDirty = true
	}
	s.lastForumSync = time.Now()
	s.mu.Unlock()
}

// syncMusic syncs music categories (typically small, always full sync)
func (s *ContentContextService) syncMusic(ctx context.Context) {
	categories, err := s.songCategoryRepo.FindAll(ctx)
	if err != nil {
		logger.Error("content context: failed to sync music", zap.Error(err))
		return
	}

	s.mu.Lock()
	// Clear and rebuild
	s.songCategories = make(map[uint]*SongCategorySummary)
	for _, category := range categories {
		songCount := s.songRepo.CountByCategoryID(ctx, category.ID)
		s.songCategories[category.ID] = &SongCategorySummary{
			ID:        category.ID,
			Name:      category.Name,
			SongCount: songCount,
		}
	}
	s.lastMusicSync = time.Now()
	s.contextDirty = true
	s.mu.Unlock()
}

// buildContextFromMaps builds the context string from cached maps
func (s *ContentContextService) buildContextFromMaps(ctx context.Context) string {
	var context strings.Builder

	if !s.isReady {
		context.WriteString("\n## KONTEN APLIKASI\n")
		context.WriteString("(Data sedang dimuat, rekomendasi konten mungkin belum lengkap)\n\n")
	} else {
		context.WriteString("\n## KONTEN APLIKASI (Untuk Rekomendasi ke Pengguna)\n")
		context.WriteString("Berikut adalah SEMUA konten yang tersedia di aplikasi Ruang Tenang.\n\n")
	}

	// Articles
	s.buildArticlesContextFromMap(ctx, &context)

	// Music
	s.buildMusicContextFromMap(ctx, &context)

	// Forums
	s.buildForumsContextFromMap(ctx, &context)

	// App features (static)
	s.buildAppFeaturesContext(ctx, &context)

	return context.String()
}

func (s *ContentContextService) buildArticlesContextFromMap(ctx context.Context, context *strings.Builder) {
	if len(s.articles) == 0 {
		return
	}

	context.WriteString("### ARTIKEL KESEHATAN MENTAL\n")
	context.WriteString(fmt.Sprintf("Total: %d artikel tersedia\n", len(s.articles)))
	context.WriteString("INGAT: Gunakan format markdown clickable: [Judul](https://ruang-tenang.site/articles/{id})\n\n")

	for _, article := range s.articles {
		title := article.Title
		if len(title) > 60 {
			title = title[:60] + "..."
		}
		context.WriteString(fmt.Sprintf("- ID:%d | \"%s\" | Kategori: %s | URL: https://ruang-tenang.site/articles/%d\n", article.ID, title, article.Category, article.ID))
	}
	context.WriteString("\n")
}

func (s *ContentContextService) buildMusicContextFromMap(ctx context.Context, context *strings.Builder) {
	if len(s.songCategories) == 0 {
		return
	}

	context.WriteString("### MUSIK RELAKSASI\n")
	context.WriteString("URL: https://ruang-tenang.site/dashboard/music\n")
	context.WriteString("Gunakan format: [Musik Relaksasi](https://ruang-tenang.site/dashboard/music)\n\n")

	for _, category := range s.songCategories {
		context.WriteString(fmt.Sprintf("- \"%s\" - %d lagu tersedia\n", category.Name, category.SongCount))
	}
	context.WriteString("\n")
}

func (s *ContentContextService) buildForumsContextFromMap(ctx context.Context, context *strings.Builder) {
	if len(s.forums) == 0 {
		return
	}

	context.WriteString("### FORUM KOMUNITAS\n")
	context.WriteString(fmt.Sprintf("Total: %d topik forum\n", len(s.forums)))
	context.WriteString("INGAT: Gunakan format markdown clickable: [Judul](https://ruang-tenang.site/dashboard/forum/{id})\n\n")

	for _, forum := range s.forums {
		title := forum.Title
		if len(title) > 50 {
			title = title[:50] + "..."
		}
		context.WriteString(fmt.Sprintf("- ID:%d | \"%s\" | %d balasan | URL: https://ruang-tenang.site/dashboard/forum/%d\n", forum.ID, title, forum.RepliesCount, forum.ID))
	}
	context.WriteString("\n")
}

func (s *ContentContextService) buildAppFeaturesContext(ctx context.Context, context *strings.Builder) {
	context.WriteString("### FITUR APLIKASI RUANG TENANG\n")
	context.WriteString("Rekomendasikan fitur yang sesuai dengan kebutuhan pengguna.\n")
	context.WriteString("INGAT: Gunakan format markdown clickable [Nama](URL) untuk semua link!\n\n")

	features := []struct {
		name        string
		url         string
		description string
	}{
		{"Chat AI (Runa)", "https://ruang-tenang.site/dashboard/chat", "Curhat dengan AI kapan saja, bisa teks atau voice."},
		{"Musik Relaksasi", "https://ruang-tenang.site/dashboard/music", "Musik menenangkan untuk relaksasi dan tidur."},
		{"Artikel", "https://ruang-tenang.site/articles", "Artikel kesehatan mental dan pengembangan diri."},
		{"Forum Komunitas", "https://ruang-tenang.site/dashboard/forum", "Diskusi anonim dengan pengguna lain."},
		{"Mood Tracker", "https://ruang-tenang.site/dashboard/mood", "Catat suasana hati harian."},
		{"Profil", "https://ruang-tenang.site/dashboard/profile", "Lihat level dan EXP."},
	}

	for _, f := range features {
		context.WriteString(fmt.Sprintf("- %s (%s): %s\n", f.name, f.url, f.description))
	}
	context.WriteString("\n")
}
