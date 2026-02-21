package handler

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newAdminHandlerForGuardTests(t *testing.T) *AdminHandler {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Article{},
		&model.ArticleCategory{},
		&model.Forum{},
		&model.ForumPost{},
		&model.ForumLike{},
		&model.ForumCategory{},
		&model.Song{},
		&model.SongCategory{},
		&model.UserMood{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS chat_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create chat_sessions table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create chat_messages table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS journals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		share_with_ai BOOLEAN
	)`).Error; err != nil {
		t.Fatalf("create journals table: %v", err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS journal_settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER UNIQUE NOT NULL,
		allow_ai_access BOOLEAN,
		ai_context_days INTEGER,
		ai_context_max_entries INTEGER,
		default_share_with_ai BOOLEAN,
		is_blocked BOOLEAN,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create journal_settings table: %v", err)
	}

	journalSvc := service.NewJournalService(
		repository.NewJournalRepository(db),
		repository.NewJournalSettingsRepository(db),
		repository.NewJournalAIAccessLogRepository(db),
		repository.NewUserMoodRepository(db),
		nil,
	)

	return &AdminHandler{db: db, forumRepo: repository.NewForumRepository(db), journalService: journalSvc}
}

func newAdminCtx(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestAdminHandler_GuardAndValidationBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAdminHandlerForGuardTests(t)

	// invalid JSON branches
	for name, run := range map[string]func(*gin.Context){
		"create-article":          h.CreateArticle,
		"create-article-category": h.CreateArticleCategory,
		"create-song-category":    h.CreateSongCategory,
		"create-song":             h.CreateSong,
	} {
		t.Run(name+"-invalid-json", func(t *testing.T) {
			c, w := newAdminCtx(http.MethodPost, "/", "{")
			run(c)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}

	// unauthorized branch for CreateArticle after successful bind
	{
		c, w := newAdminCtx(http.MethodPost, "/admin/articles", `{"title":"T","content":"C","category_id":1}`)
		h.CreateArticle(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 CreateArticle unauthorized, got %d", w.Code)
		}
	}

	// not found branches
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/articles/999", `{"title":"T","content":"C","category_id":1}`)
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UpdateArticle(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 UpdateArticle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/articles/999", "")
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.DeleteArticle(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 DeleteArticle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/articles/999/block", "")
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.BlockArticle(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 BlockArticle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/articles/999/unblock", "")
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UnblockArticle(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 UnblockArticle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/article-categories/999", `{"name":"x"}`)
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UpdateArticleCategory(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 UpdateArticleCategory, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/song-categories/999", `{"name":"x"}`)
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UpdateSongCategory(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 UpdateSongCategory, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/songs/999", `{"title":"t","file_path":"f","category_id":1}`)
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.UpdateSong(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 UpdateSong, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/songs/999", "")
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.DeleteSong(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 DeleteSong, got %d", w.Code)
		}
	}
}

func TestAdminHandler_DashboardUsersAndForumsBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAdminHandlerForGuardTests(t)

	admin := model.User{Name: "Admin", Email: "admin@test.local", Username: "admin1", Password: "x", Role: model.RoleAdmin}
	member := model.User{Name: "Member", Email: "member@test.local", Username: "member1", Password: "x", Role: model.RoleMember}
	if err := h.db.Create(&admin).Error; err != nil {
		t.Fatalf("seed admin user: %v", err)
	}
	if err := h.db.Create(&member).Error; err != nil {
		t.Fatalf("seed member user: %v", err)
	}

	category := model.ForumCategory{Name: "General"}
	if err := h.db.Create(&category).Error; err != nil {
		t.Fatalf("seed forum category: %v", err)
	}
	categoryID := category.ID
	forum := model.Forum{Title: "Forum 1", Content: "Content", Slug: "forum-1", UserID: member.ID, CategoryID: &categoryID}
	if err := h.db.Create(&forum).Error; err != nil {
		t.Fatalf("seed forum: %v", err)
	}

	articleCategory := model.ArticleCategory{Name: "Article Cat", Description: "desc"}
	if err := h.db.Create(&articleCategory).Error; err != nil {
		t.Fatalf("seed article category: %v", err)
	}

	// dashboard and users list happy paths
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/stats", "")
		h.GetDashboardStats(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetDashboardStats, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/stats", "")
		c.Set("user_role", "moderator")
		h.GetDashboardStats(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetDashboardStats moderator, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/users?page=0&limit=999", "")
		h.GetUsers(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetUsers, got %d", w.Code)
		}
	}

	// user management branches
	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/users/99999", "")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		h.DeleteUser(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 DeleteUser, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/99999/block", "")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		h.BlockUser(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 BlockUser, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(admin.ID))+"/block", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(admin.ID))}}
		h.BlockUser(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 BlockUser admin, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/99999/unblock", "")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		h.UnblockUser(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 UnblockUser, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(member.ID))+"/block", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(member.ID))}}
		h.BlockUser(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 BlockUser member, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(member.ID))+"/unblock", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(member.ID))}}
		h.UnblockUser(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 UnblockUser member, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/99999/block-journal", "")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		h.ToggleJournalBlock(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 ToggleJournalBlock, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(member.ID))+"/block-journal", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(member.ID))}}
		h.ToggleJournalBlock(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ToggleJournalBlock member first toggle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(member.ID))+"/block-journal", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(member.ID))}}
		h.ToggleJournalBlock(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ToggleJournalBlock member second toggle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(admin.ID))+"/block-forum", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(admin.ID))}}
		h.ToggleForumBlock(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 ToggleForumBlock admin, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(member.ID))+"/block-forum", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(member.ID))}}
		h.ToggleForumBlock(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ToggleForumBlock member first toggle, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/users/"+strconv.Itoa(int(member.ID))+"/block-forum", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(member.ID))}}
		h.ToggleForumBlock(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ToggleForumBlock member second toggle, got %d", w.Code)
		}
	}

	// forum/cache branches
	{
		hWithCache := *h
		hWithCache.cacheService = service.NewCacheService()
		c, w := newAdminCtx(http.MethodPost, "/admin/cache/clear", "")
		hWithCache.ClearCache(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ClearCache with cache service, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPost, "/admin/cache/clear", "")
		h.ClearCache(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ClearCache, got %d", w.Code)
		}
	}
	{
		hNoCache := *h
		hNoCache.cacheService = nil
		c, w := newAdminCtx(http.MethodPost, "/admin/cache/clear", "")
		hNoCache.ClearCache(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ClearCache without cache service, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPost, "/admin/forums/99999/toggle-flag", "")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		h.ToggleForumFlag(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 ToggleForumFlag, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPost, "/admin/forums/"+strconv.Itoa(int(forum.ID))+"/toggle-flag", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(forum.ID))}}
		h.ToggleForumFlag(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 ToggleForumFlag existing, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/forums/99999", "")
		c.Params = gin.Params{{Key: "id", Value: "99999"}}
		h.GetForum(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 GetForum not found, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/forums/"+strconv.Itoa(int(forum.ID)), "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(forum.ID))}}
		h.GetForum(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetForum existing, got %d", w.Code)
		}
	}

	// article branches
	var articleID uint
	{
		body := `{"title":"Admin Article","content":"Body","category_id":` + strconv.Itoa(int(articleCategory.ID)) + `}`
		c, w := newAdminCtx(http.MethodPost, "/admin/articles", body)
		c.Set("user_id", admin.ID)
		h.CreateArticle(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 CreateArticle success, got %d", w.Code)
		}

		var created model.Article
		if err := h.db.Order("id desc").First(&created).Error; err != nil {
			t.Fatalf("query created article: %v", err)
		}
		articleID = created.ID
	}
	{
		body := `{"title":"Admin Article Updated","content":"Body Updated","category_id":` + strconv.Itoa(int(articleCategory.ID)) + `}`
		c, w := newAdminCtx(http.MethodPut, "/admin/articles/"+strconv.Itoa(int(articleID)), body)
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleID))}}
		h.UpdateArticle(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 UpdateArticle success, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/articles/"+strconv.Itoa(int(articleID))+"/block", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleID))}}
		h.BlockArticle(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 BlockArticle success, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodPut, "/admin/articles/"+strconv.Itoa(int(articleID))+"/unblock", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleID))}}
		h.UnblockArticle(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 UnblockArticle success, got %d", w.Code)
		}
	}
}

func TestAdminHandler_ListEndpointsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAdminHandlerForGuardTests(t)
	h.articleRepo = repository.NewArticleRepository(h.db)

	user := model.User{Name: "List User", Email: "list@test.local", Username: "listuser", Password: "x", Role: model.RoleMember}
	if err := h.db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	articleCategory := model.ArticleCategory{Name: "Cat", Description: "d"}
	if err := h.db.Create(&articleCategory).Error; err != nil {
		t.Fatalf("create article category: %v", err)
	}
	songCategory := model.SongCategory{Name: "Ambient"}
	if err := h.db.Create(&songCategory).Error; err != nil {
		t.Fatalf("create song category: %v", err)
	}
	article := model.Article{Title: "Article 1", Content: "content", ArticleCategoryID: articleCategory.ID, UserID: user.ID, Status: model.ArticleStatusPublished}
	if err := h.db.Create(&article).Error; err != nil {
		t.Fatalf("create article: %v", err)
	}
	forumCategory := model.ForumCategory{Name: "Forum Cat"}
	if err := h.db.Create(&forumCategory).Error; err != nil {
		t.Fatalf("create forum category: %v", err)
	}
	forumCategoryID := forumCategory.ID
	forum := model.Forum{Title: "Forum 1", Content: "content", Slug: "forum-1", UserID: user.ID, CategoryID: &forumCategoryID}
	if err := h.db.Create(&forum).Error; err != nil {
		t.Fatalf("create forum: %v", err)
	}
	song := model.Song{Title: "Song 1", FilePath: "/tmp/song.mp3", SongCategoryID: songCategory.ID}
	if err := h.db.Create(&song).Error; err != nil {
		t.Fatalf("create song: %v", err)
	}

	{
		c, w := newAdminCtx(http.MethodGet, "/admin/articles?page=0&limit=100", "")
		h.GetAllArticles(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetAllArticles, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/article-categories", "")
		h.GetArticleCategories(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetArticleCategories, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/songs", "")
		h.GetAllSongs(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetAllSongs, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodGet, "/admin/forums?page=0&limit=999", "")
		h.GetForums(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 GetForums, got %d", w.Code)
		}
	}
}

func TestAdminHandler_DeleteCategoryBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAdminHandlerForGuardTests(t)

	user := model.User{Name: "User", Email: "user@test.local", Username: "user1", Password: "x", Role: model.RoleMember}
	if err := h.db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	articleCategoryInUse := model.ArticleCategory{Name: "InUse", Description: "d"}
	if err := h.db.Create(&articleCategoryInUse).Error; err != nil {
		t.Fatalf("seed article category in use: %v", err)
	}
	articleCategoryFree := model.ArticleCategory{Name: "Free", Description: "d"}
	if err := h.db.Create(&articleCategoryFree).Error; err != nil {
		t.Fatalf("seed article category free: %v", err)
	}
	article := model.Article{Title: "A1", Content: "C", ArticleCategoryID: articleCategoryInUse.ID, UserID: user.ID, Status: model.ArticleStatusPublished}
	if err := h.db.Create(&article).Error; err != nil {
		t.Fatalf("seed article: %v", err)
	}

	songCategory := model.SongCategory{Name: "Song Cat"}
	if err := h.db.Create(&songCategory).Error; err != nil {
		t.Fatalf("seed song category: %v", err)
	}

	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/article-categories/999", "")
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.DeleteArticleCategory(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 DeleteArticleCategory not found, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/article-categories/in-use", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleCategoryInUse.ID))}}
		h.DeleteArticleCategory(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 DeleteArticleCategory in use, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/article-categories/free", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleCategoryFree.ID))}}
		h.DeleteArticleCategory(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 DeleteArticleCategory success, got %d", w.Code)
		}
	}

	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/song-categories/999", "")
		c.Params = gin.Params{{Key: "id", Value: "999"}}
		h.DeleteSongCategory(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404 DeleteSongCategory not found, got %d", w.Code)
		}
	}
	{
		c, w := newAdminCtx(http.MethodDelete, "/admin/song-categories/existing", "")
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(songCategory.ID))}}
		h.DeleteSongCategory(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 DeleteSongCategory success, got %d", w.Code)
		}
	}
}

func TestAdminHandler_CategorySongCreateUpdateSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newAdminHandlerForGuardTests(t)

	articleCategory := model.ArticleCategory{Name: "Old Article Cat", Description: "old"}
	songCategory := model.SongCategory{Name: "Old Song Cat", Thumbnail: "old.png"}
	if err := h.db.Create(&articleCategory).Error; err != nil {
		t.Fatalf("seed article category: %v", err)
	}
	if err := h.db.Create(&songCategory).Error; err != nil {
		t.Fatalf("seed song category: %v", err)
	}

	t.Run("create-article-category-success", func(t *testing.T) {
		c, w := newAdminCtx(http.MethodPost, "/admin/article-categories", `{"name":"Mindfulness","description":"desc"}`)
		h.CreateArticleCategory(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 create article category, got %d", w.Code)
		}
	})

	t.Run("update-article-category-success", func(t *testing.T) {
		c, w := newAdminCtx(http.MethodPut, "/admin/article-categories/1", `{"name":"Updated Article Cat","description":"new"}`)
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleCategory.ID))}}
		h.UpdateArticleCategory(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 update article category, got %d", w.Code)
		}
	})

	t.Run("create-song-category-success", func(t *testing.T) {
		c, w := newAdminCtx(http.MethodPost, "/admin/song-categories", `{"name":"Calm","thumbnail":"calm.png"}`)
		h.CreateSongCategory(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 create song category, got %d", w.Code)
		}
	})

	t.Run("update-song-category-success", func(t *testing.T) {
		c, w := newAdminCtx(http.MethodPut, "/admin/song-categories/1", `{"name":"Updated Song Cat","thumbnail":"new.png"}`)
		c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(songCategory.ID))}}
		h.UpdateSongCategory(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 update song category, got %d", w.Code)
		}
	})

	t.Run("create-and-update-song-success", func(t *testing.T) {
		reqCreate := `{"title":"Track 1","file_path":"/audio/track1.mp3","thumbnail":"t.png","category_id":` + strconv.Itoa(int(songCategory.ID)) + `}`
		c1, w1 := newAdminCtx(http.MethodPost, "/admin/songs", reqCreate)
		h.CreateSong(c1)
		if w1.Code != http.StatusCreated {
			t.Fatalf("expected 201 create song, got %d", w1.Code)
		}

		var created model.Song
		if err := h.db.Order("id desc").First(&created).Error; err != nil {
			t.Fatalf("query created song: %v", err)
		}

		reqUpdate := `{"title":"Track 1 Updated","file_path":"/audio/track1-updated.mp3","thumbnail":"u.png","category_id":` + strconv.Itoa(int(songCategory.ID)) + `}`
		c2, w2 := newAdminCtx(http.MethodPut, "/admin/songs/"+strconv.Itoa(int(created.ID)), reqUpdate)
		c2.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(created.ID))}}
		h.UpdateSong(c2)
		if w2.Code != http.StatusOK {
			t.Fatalf("expected 200 update song, got %d", w2.Code)
		}
	})
}

func TestAdminHandler_CategorySong_InternalAndBadJSONBranches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("internal-server-branches-on-missing-tables", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		h := &AdminHandler{db: db}

		c1, w1 := newAdminCtx(http.MethodPost, "/admin/article-categories", `{"name":"A","description":"D"}`)
		h.CreateArticleCategory(c1)
		if w1.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 CreateArticleCategory, got %d", w1.Code)
		}

		c2, w2 := newAdminCtx(http.MethodGet, "/admin/article-categories", "")
		h.GetArticleCategories(c2)
		if w2.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 GetArticleCategories, got %d", w2.Code)
		}

		c3, w3 := newAdminCtx(http.MethodPost, "/admin/song-categories", `{"name":"S","thumbnail":"t"}`)
		h.CreateSongCategory(c3)
		if w3.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 CreateSongCategory, got %d", w3.Code)
		}

		c4, w4 := newAdminCtx(http.MethodPost, "/admin/songs", `{"title":"Track","file_path":"/a.mp3","thumbnail":"t","category_id":1}`)
		h.CreateSong(c4)
		if w4.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 CreateSong, got %d", w4.Code)
		}

		c5, w5 := newAdminCtx(http.MethodGet, "/admin/songs", "")
		h.GetAllSongs(c5)
		if w5.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 GetAllSongs, got %d", w5.Code)
		}
	})

	t.Run("update-bad-json-branches", func(t *testing.T) {
		h := newAdminHandlerForGuardTests(t)
		articleCategory := model.ArticleCategory{Name: "AC", Description: "d"}
		songCategory := model.SongCategory{Name: "SC", Thumbnail: "t"}
		song := model.Song{Title: "Song", FilePath: "/s.mp3", SongCategoryID: 1}

		if err := h.db.Create(&articleCategory).Error; err != nil {
			t.Fatalf("seed article category: %v", err)
		}
		if err := h.db.Create(&songCategory).Error; err != nil {
			t.Fatalf("seed song category: %v", err)
		}
		song.SongCategoryID = songCategory.ID
		if err := h.db.Create(&song).Error; err != nil {
			t.Fatalf("seed song: %v", err)
		}

		c1, w1 := newAdminCtx(http.MethodPut, "/admin/article-categories/1", "{")
		c1.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(articleCategory.ID))}}
		h.UpdateArticleCategory(c1)
		if w1.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 UpdateArticleCategory bad json, got %d", w1.Code)
		}

		c2, w2 := newAdminCtx(http.MethodPut, "/admin/song-categories/1", "{")
		c2.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(songCategory.ID))}}
		h.UpdateSongCategory(c2)
		if w2.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 UpdateSongCategory bad json, got %d", w2.Code)
		}

		c3, w3 := newAdminCtx(http.MethodPut, "/admin/songs/1", "{")
		c3.Params = gin.Params{{Key: "id", Value: strconv.Itoa(int(song.ID))}}
		h.UpdateSong(c3)
		if w3.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 UpdateSong bad json, got %d", w3.Code)
		}
	})
}
