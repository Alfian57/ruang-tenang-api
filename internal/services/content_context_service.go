package services

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/repositories"
)

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
	articleRepo      *repositories.ArticleRepository
	songRepo         *repositories.SongRepository
	songCategoryRepo *repositories.SongCategoryRepository
	forumRepo        repositories.ForumRepository

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
	articleRepo *repositories.ArticleRepository,
	songRepo *repositories.SongRepository,
	songCategoryRepo *repositories.SongCategoryRepository,
	forumRepo repositories.ForumRepository,
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
	go s.backgroundSync()

	return s
}

// GetContentContext returns the content context for AI
// Returns partial context if cache is still loading
func (s *ContentContextService) GetContentContext() string {
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

	s.cachedContext = s.buildContextFromMaps()
	s.contextDirty = false

	return s.cachedContext
}

// NotifyArticleChange notifies the cache of an article change (for event-driven updates)
func (s *ContentContextService) NotifyArticleChange(article *models.Article) {
	if article == nil || article.Status != models.ArticleStatusPublished {
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
func (s *ContentContextService) NotifyArticleDelete(articleID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.articles, articleID)
	s.contextDirty = true
}

// NotifyForumChange notifies the cache of a forum change
func (s *ContentContextService) NotifyForumChange(forum *models.Forum) {
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
func (s *ContentContextService) NotifyForumDelete(forumID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.forums, forumID)
	s.contextDirty = true
}

// Stop stops the background sync goroutine
func (s *ContentContextService) Stop() {
	close(s.stopChan)
}

// backgroundSync runs the background pre-warm and periodic sync
func (s *ContentContextService) backgroundSync() {
	// Initial pre-warm
	fmt.Println("ContentContextService: Starting initial pre-warm...")
	s.fullSync()
	fmt.Printf("ContentContextService: Pre-warm complete. Articles: %d, Songs: %d categories, Forums: %d\n",
		len(s.articles), len(s.songCategories), len(s.forums))

	// Periodic sync
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.incrementalSync()
		case <-s.stopChan:
			fmt.Println("ContentContextService: Background sync stopped")
			return
		}
	}
}

// fullSync fetches all data from database
func (s *ContentContextService) fullSync() {
	s.syncArticles(time.Time{}) // Zero time = fetch all
	s.syncForums(time.Time{})
	s.syncMusic()

	s.mu.Lock()
	s.isReady = true
	s.contextDirty = true
	s.mu.Unlock()
}

// incrementalSync fetches only updated records since last sync
func (s *ContentContextService) incrementalSync() {
	s.syncArticles(s.lastArticleSync)
	s.syncForums(s.lastForumSync)
	s.syncMusic() // Music is typically small, just refresh all
}

// syncArticles syncs articles since the given time
func (s *ContentContextService) syncArticles(since time.Time) {
	var articles []models.Article
	var err error

	if since.IsZero() {
		// Full sync - get all published articles
		articles, _, err = s.articleRepo.FindPublished(0, "", 1, 10000)
	} else {
		// Incremental sync - get updated since last sync
		articles, err = s.articleRepo.FindUpdatedSince(since)
	}

	if err != nil {
		fmt.Printf("ContentContextService: Failed to sync articles: %v\n", err)
		return
	}

	s.mu.Lock()
	for _, article := range articles {
		if article.Status == models.ArticleStatusPublished {
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
func (s *ContentContextService) syncForums(since time.Time) {
	var forums []models.Forum
	var err error

	if since.IsZero() {
		// Full sync
		forums, _, err = s.forumRepo.GetForums(10000, 0, "", nil)
	} else {
		// Incremental sync
		forums, err = s.forumRepo.FindUpdatedSince(since)
	}

	if err != nil {
		fmt.Printf("ContentContextService: Failed to sync forums: %v\n", err)
		return
	}

	s.mu.Lock()
	for _, forum := range forums {
		repliesCount, _ := s.forumRepo.GetRepliesCount(forum.ID)
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
func (s *ContentContextService) syncMusic() {
	categories, err := s.songCategoryRepo.FindAll()
	if err != nil {
		fmt.Printf("ContentContextService: Failed to sync music: %v\n", err)
		return
	}

	s.mu.Lock()
	// Clear and rebuild
	s.songCategories = make(map[uint]*SongCategorySummary)
	for _, category := range categories {
		songCount := s.songRepo.CountByCategoryID(category.ID)
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
func (s *ContentContextService) buildContextFromMaps() string {
	var context strings.Builder

	if !s.isReady {
		context.WriteString("\n## KONTEN APLIKASI\n")
		context.WriteString("(Data sedang dimuat, rekomendasi konten mungkin belum lengkap)\n\n")
	} else {
		context.WriteString("\n## KONTEN APLIKASI (Untuk Rekomendasi ke Pengguna)\n")
		context.WriteString("Berikut adalah SEMUA konten yang tersedia di aplikasi Ruang Tenang.\n\n")
	}

	// Articles
	s.buildArticlesContextFromMap(&context)

	// Music
	s.buildMusicContextFromMap(&context)

	// Forums
	s.buildForumsContextFromMap(&context)

	// App features (static)
	s.buildAppFeaturesContext(&context)

	return context.String()
}

func (s *ContentContextService) buildArticlesContextFromMap(context *strings.Builder) {
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

func (s *ContentContextService) buildMusicContextFromMap(context *strings.Builder) {
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

func (s *ContentContextService) buildForumsContextFromMap(context *strings.Builder) {
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

func (s *ContentContextService) buildAppFeaturesContext(context *strings.Builder) {
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
