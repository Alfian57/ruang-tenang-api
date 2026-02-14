package handler

import (
	"net/http"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/repository"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db           *gorm.DB
	userRepo     *repository.UserRepository
	articleRepo  *repository.ArticleRepository
	cacheService *service.CacheService
}

func NewAdminHandler(db *gorm.DB, userRepo *repository.UserRepository, articleRepo *repository.ArticleRepository, cacheService *service.CacheService) *AdminHandler {
	return &AdminHandler{
		db:           db,
		userRepo:     userRepo,
		articleRepo:  articleRepo,
		cacheService: cacheService,
	}
}

// GetDashboardStats godoc
// @Summary Get dashboard statistics
// @Description Get platform statistics for admin dashboard
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /admin/stats [get]
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastMonthStart := monthStart.AddDate(0, -1, 0)
	lastMonthEnd := monthStart.Add(-time.Second)

	role, exists := c.Get("user_role")
	isModerator := exists && role == "moderator"

	// User stats
	var totalUsers int64
	var activeUsers int64
	var blockedUsers int64
	var usersThisMonth int64
	var userChartData []int64 = make([]int64, 7)
	var userGrowth float64
	var recentUsersDTO []gin.H = make([]gin.H, 0)

	// Only fetch user stats if not moderator
	if !isModerator {
		h.db.WithContext(ctx).Model(&model.User{}).Count(&totalUsers)
		h.db.WithContext(ctx).Model(&model.User{}).Where("is_blocked = ?", false).Count(&activeUsers)
		h.db.WithContext(ctx).Model(&model.User{}).Where("is_blocked = ?", true).Count(&blockedUsers)
		h.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= ?", monthStart).Count(&usersThisMonth)

		var usersLastMonth int64
		h.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= ? AND created_at <= ?", lastMonthStart, lastMonthEnd).Count(&usersLastMonth)
		if usersLastMonth > 0 {
			userGrowth = float64(usersThisMonth-usersLastMonth) / float64(usersLastMonth) * 100
		}

		// Weekly chart data for users
		for i := 6; i >= 0; i-- {
			dayStart := todayStart.AddDate(0, 0, -i)
			dayEnd := dayStart.Add(24 * time.Hour)
			var count int64
			h.db.WithContext(ctx).Model(&model.User{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&count)
			userChartData[6-i] = count
		}

		// Recent users
		var recentUsers []model.User
		h.db.WithContext(ctx).Order("created_at DESC").Limit(5).Find(&recentUsers)
		for _, u := range recentUsers {
			recentUsersDTO = append(recentUsersDTO, gin.H{
				"id":         u.ID,
				"name":       u.Name,
				"email":      u.Email,
				"role":       u.Role,
				"is_blocked": u.IsBlocked,
				"created_at": u.CreatedAt,
			})
		}
	}

	// Article stats
	var totalArticles int64
	h.db.WithContext(ctx).Model(&model.Article{}).Count(&totalArticles)

	var articlesThisMonth int64
	h.db.WithContext(ctx).Model(&model.Article{}).Where("created_at >= ?", monthStart).Count(&articlesThisMonth)

	// Chat stats
	var totalChatSessions int64
	h.db.WithContext(ctx).Model(&model.ChatSession{}).Count(&totalChatSessions)

	var chatSessionsToday int64
	h.db.WithContext(ctx).Model(&model.ChatSession{}).Where("created_at >= ?", todayStart).Count(&chatSessionsToday)

	var totalMessages int64
	h.db.WithContext(ctx).Model(&model.ChatMessage{}).Count(&totalMessages)

	var messagesToday int64
	h.db.WithContext(ctx).Model(&model.ChatMessage{}).Where("created_at >= ?", todayStart).Count(&messagesToday)

	// Song stats
	var totalSongs int64
	h.db.WithContext(ctx).Model(&model.Song{}).Count(&totalSongs)

	var totalSongCategories int64
	h.db.WithContext(ctx).Model(&model.SongCategory{}).Count(&totalSongCategories)

	// Mood stats
	var totalMoods int64
	h.db.WithContext(ctx).Model(&model.UserMood{}).Count(&totalMoods)

	var moodsToday int64
	h.db.WithContext(ctx).Model(&model.UserMood{}).Where("created_at >= ?", todayStart).Count(&moodsToday)

	// Weekly chart data for chat sessions
	chatChartData := make([]int64, 7)
	for i := 6; i >= 0; i-- {
		dayStart := todayStart.AddDate(0, 0, -i)
		dayEnd := dayStart.Add(24 * time.Hour)
		var count int64
		h.db.WithContext(ctx).Model(&model.ChatSession{}).Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).Count(&count)
		chatChartData[6-i] = count
	}

	// Article category stats
	var totalArticleCategories int64
	h.db.WithContext(ctx).Model(&model.ArticleCategory{}).Count(&totalArticleCategories)

	// Pending/blocked articles
	var blockedArticles int64
	h.db.WithContext(ctx).Model(&model.Article{}).Where("status = ?", "blocked").Count(&blockedArticles)

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"users": gin.H{
			"total":      totalUsers,
			"active":     activeUsers,
			"blocked":    blockedUsers,
			"this_month": usersThisMonth,
			"growth":     userGrowth,
			"chart_data": userChartData,
		},
		"articles": gin.H{
			"total":      totalArticles,
			"this_month": articlesThisMonth,
			"blocked":    blockedArticles,
			"categories": totalArticleCategories,
		},
		"chat_sessions": gin.H{
			"total":      totalChatSessions,
			"today":      chatSessionsToday,
			"chart_data": chatChartData,
		},
		"messages": gin.H{
			"total": totalMessages,
			"today": messagesToday,
		},
		"songs": gin.H{
			"total":      totalSongs,
			"categories": totalSongCategories,
		},
		"moods": gin.H{
			"total": totalMoods,
			"today": moodsToday,
		},
		"recent_users": recentUsersDTO,
		"is_moderator": isModerator,
	}, ""))
}

// GetUsers godoc
// @Summary Get all users
// @Description Get paginated list of users for admin
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by name or email"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.PaginatedResponse
// @Router /admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		Search string `form:"search"`
		Page   int    `form:"page"`
		Limit  int    `form:"limit"`
	}
	c.ShouldBindQuery(&params)

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 10
	}

	var users []model.User
	var total int64

	query := h.db.WithContext(ctx).Model(&model.User{})
	if params.Search != "" {
		searchTerm := "%" + params.Search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	query.Count(&total)
	query.Offset((params.Page - 1) * params.Limit).Limit(params.Limit).Order("created_at DESC").Find(&users)

	result := make([]gin.H, len(users))
	for i, u := range users {
		result[i] = gin.H{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"role":       u.Role,
			"is_blocked": u.IsBlocked,
			"created_at": u.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(result, params.Page, params.Limit, total))
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	result := h.db.WithContext(ctx).Delete(&model.User{}, id)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User deleted"))
}

// BlockUser godoc
// @Summary Block a user
// @Description Block a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/block [put]
func (h *AdminHandler) BlockUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	// Prevent blocking admin users
	if user.Role == model.RoleAdmin {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Cannot block admin users"))
		return
	}

	user.IsBlocked = true
	if err := h.db.WithContext(ctx).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to block user"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User blocked"))
}

// UnblockUser godoc
// @Summary Unblock a user
// @Description Unblock a user by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Response
// @Router /admin/users/{id}/unblock [put]
func (h *AdminHandler) UnblockUser(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var user model.User
	if err := h.db.WithContext(ctx).First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("User not found"))
		return
	}

	user.IsBlocked = false
	if err := h.db.WithContext(ctx).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to unblock user"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User unblocked"))
}

// CreateArticle godoc
// @Summary Create an article
// @Description Create a new article (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateArticleRequest true "Article data"
// @Success 201 {object} dto.Response
// @Router /admin/articles [post]
func (h *AdminHandler) CreateArticle(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}
	uid := userID.(uint)

	article := model.Article{
		Title:             req.Title,
		Thumbnail:         req.Thumbnail,
		Content:           req.Content,
		ArticleCategoryID: req.CategoryID,
		UserID:            uid,
		Status:            model.ArticleStatusPublished,
	}

	if err := h.db.WithContext(ctx).Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create article"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{"id": article.ID}, "Article created"))
}

// UpdateArticle godoc
// @Summary Update an article
// @Description Update an existing article
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Param request body dto.UpdateArticleRequest true "Article data"
// @Success 200 {object} dto.Response
// @Router /admin/articles/{id} [put]
func (h *AdminHandler) UpdateArticle(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var article model.Article
	if err := h.db.WithContext(ctx).First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	var req dto.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	article.Title = req.Title
	article.Thumbnail = req.Thumbnail
	article.Content = req.Content
	article.ArticleCategoryID = req.CategoryID

	if err := h.db.WithContext(ctx).Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update article"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article updated"))
}

// DeleteArticle godoc
// @Summary Delete an article
// @Description Delete an article by ID
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Success 200 {object} dto.Response
// @Router /admin/articles/{id} [delete]
func (h *AdminHandler) DeleteArticle(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	result := h.db.WithContext(ctx).Delete(&model.Article{}, id)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article deleted"))
}

// BlockArticle godoc
// @Summary Block an article
// @Description Block an article by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Success 200 {object} dto.Response
// @Router /admin/articles/{id}/block [put]
func (h *AdminHandler) BlockArticle(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var article model.Article
	if err := h.db.WithContext(ctx).First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	article.Status = model.ArticleStatusBlocked
	if err := h.db.WithContext(ctx).Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to block article"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article blocked"))
}

// UnblockArticle godoc
// @Summary Unblock an article
// @Description Unblock an article by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Success 200 {object} dto.Response
// @Router /admin/articles/{id}/unblock [put]
func (h *AdminHandler) UnblockArticle(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var article model.Article
	if err := h.db.WithContext(ctx).First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	article.Status = model.ArticleStatusPublished
	if err := h.db.WithContext(ctx).Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to unblock article"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article unblocked"))
}

// GetAllArticles godoc
// @Summary Get all articles for admin
// @Description Get paginated list of all articles with optional filtering (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param category_id query int false "Filter by category ID"
// @Param search query string false "Search by title"
// @Param status query string false "Filter by status (published, draft, blocked)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.PaginatedResponse
// @Router /admin/articles [get]
func (h *AdminHandler) GetAllArticles(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		CategoryID uint   `form:"category_id"`
		Search     string `form:"search"`
		Status     string `form:"status"`
		Page       int    `form:"page"`
		Limit      int    `form:"limit"`
	}
	c.ShouldBindQuery(&params)

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 10
	}

	articles, total, err := h.articleRepo.FindAll(ctx, params.CategoryID, params.Search, params.Page, params.Limit, params.Status, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get articles"))
		return
	}

	result := make([]gin.H, len(articles))
	for i, a := range articles {
		item := gin.H{
			"id":                a.ID,
			"title":             a.Title,
			"thumbnail":         a.Thumbnail,
			"category_id":       a.ArticleCategoryID,
			"category":          gin.H{"id": a.Category.ID, "name": a.Category.Name},
			"status":            a.Status,
			"moderation_status": a.ModerationStatus,
			"user_id":           a.UserID,
			"created_at":        a.CreatedAt,
		}
		if a.Author != nil {
			item["author"] = gin.H{"id": a.Author.ID, "name": a.Author.Name}
		}
		result[i] = item
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(result, params.Page, params.Limit, total))
}

// CreateArticleCategory godoc
// @Summary Create article category
// @Description Create a new article category
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateArticleCategoryRequest true "Category data"
// @Success 201 {object} dto.Response
// @Router /admin/article-categories [post]
func (h *AdminHandler) CreateArticleCategory(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateArticleCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	category := model.ArticleCategory{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := h.db.WithContext(ctx).Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create category"))
		return
	}

	// Invalidate article categories cache
	if h.cacheService != nil {
		h.cacheService.Delete(service.CacheKeyArticleCategories)
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{"id": category.ID}, "Category created"))
}

// DeleteArticleCategory godoc
// @Summary Delete article category
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} dto.Response
// @Router /admin/article-categories/{id} [delete]
func (h *AdminHandler) DeleteArticleCategory(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	// Check if category has related articles
	var articleCount int64
	h.db.WithContext(ctx).Model(&model.Article{}).Where("article_category_id = ?", id).Count(&articleCount)
	if articleCount > 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Kategori tidak dapat dihapus karena masih memiliki artikel terkait"))
		return
	}

	result := h.db.WithContext(ctx).Delete(&model.ArticleCategory{}, id)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Category not found"))
		return
	}

	// Invalidate article categories cache
	if h.cacheService != nil {
		h.cacheService.Delete(service.CacheKeyArticleCategories)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Category deleted"))
}

// UpdateArticleCategory godoc
// @Summary Update article category
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body dto.CreateArticleCategoryRequest true "Category data"
// @Success 200 {object} dto.Response
// @Router /admin/article-categories/{id} [put]
func (h *AdminHandler) UpdateArticleCategory(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var category model.ArticleCategory
	if err := h.db.WithContext(ctx).First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Category not found"))
		return
	}

	var req dto.CreateArticleCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	category.Name = req.Name
	category.Description = req.Description
	if err := h.db.WithContext(ctx).Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update category"))
		return
	}

	// Invalidate article categories cache
	if h.cacheService != nil {
		h.cacheService.Delete(service.CacheKeyArticleCategories)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Category updated"))
}

// GetArticleCategories godoc
// @Summary Get all article categories for admin
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /admin/article-categories [get]
func (h *AdminHandler) GetArticleCategories(c *gin.Context) {
	ctx := c.Request.Context()
	var categories []model.ArticleCategory
	if err := h.db.WithContext(ctx).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get categories"))
		return
	}

	result := make([]gin.H, len(categories))
	for i, cat := range categories {
		var articleCount int64
		h.db.WithContext(ctx).Model(&model.Article{}).Where("article_category_id = ?", cat.ID).Count(&articleCount)
		result[i] = gin.H{
			"id":            cat.ID,
			"name":          cat.Name,
			"description":   cat.Description,
			"article_count": articleCount,
			"created_at":    cat.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

// CreateSongCategory godoc
// @Summary Create song category
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSongCategoryRequest true "Category data"
// @Success 201 {object} dto.Response
// @Router /admin/song-categories [post]
func (h *AdminHandler) CreateSongCategory(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateSongCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	category := model.SongCategory{Name: req.Name, Thumbnail: req.Thumbnail}
	if err := h.db.WithContext(ctx).Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create category"))
		return
	}

	// Invalidate song categories cache
	if h.cacheService != nil {
		h.cacheService.Delete(service.CacheKeySongCategories)
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{"id": category.ID}, "Category created"))
}

// DeleteSongCategory godoc
// @Summary Delete song category
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} dto.Response
// @Router /admin/song-categories/{id} [delete]
func (h *AdminHandler) DeleteSongCategory(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	result := h.db.WithContext(ctx).Delete(&model.SongCategory{}, id)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Category not found"))
		return
	}

	// Invalidate song categories cache
	if h.cacheService != nil {
		h.cacheService.Delete(service.CacheKeySongCategories)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Category deleted"))
}

// UpdateSongCategory godoc
// @Summary Update song category
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body dto.CreateSongCategoryRequest true "Category data"
// @Success 200 {object} dto.Response
// @Router /admin/song-categories/{id} [put]
func (h *AdminHandler) UpdateSongCategory(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var category model.SongCategory
	if err := h.db.WithContext(ctx).First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Category not found"))
		return
	}

	var req dto.CreateSongCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	category.Name = req.Name
	category.Thumbnail = req.Thumbnail
	if err := h.db.WithContext(ctx).Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update category"))
		return
	}

	// Invalidate song categories cache
	if h.cacheService != nil {
		h.cacheService.Delete(service.CacheKeySongCategories)
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Category updated"))
}

// CreateSong godoc
// @Summary Create a song
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSongRequest true "Song data"
// @Success 201 {object} dto.Response
// @Router /admin/songs [post]
func (h *AdminHandler) CreateSong(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	song := model.Song{
		Title:          req.Title,
		FilePath:       req.FilePath,
		Thumbnail:      req.Thumbnail,
		SongCategoryID: req.CategoryID,
	}

	if err := h.db.WithContext(ctx).Create(&song).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create song"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{"id": song.ID}, "Song created"))
}

// UpdateSong godoc
// @Summary Update a song
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Song ID"
// @Param request body dto.CreateSongRequest true "Song data"
// @Success 200 {object} dto.Response
// @Router /admin/songs/{id} [put]
func (h *AdminHandler) UpdateSong(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var song model.Song
	if err := h.db.WithContext(ctx).First(&song, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Song not found"))
		return
	}

	var req dto.CreateSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	song.Title = req.Title
	song.FilePath = req.FilePath
	song.Thumbnail = req.Thumbnail
	song.SongCategoryID = req.CategoryID

	if err := h.db.WithContext(ctx).Save(&song).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update song"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Song updated"))
}

// GetAllSongs godoc
// @Summary Get all songs for admin
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param category_id query int false "Filter by category ID"
// @Success 200 {object} dto.Response
// @Router /admin/songs [get]
func (h *AdminHandler) GetAllSongs(c *gin.Context) {
	ctx := c.Request.Context()
	var params struct {
		CategoryID uint `form:"category_id"`
	}
	c.ShouldBindQuery(&params)

	var songs []model.Song
	query := h.db.WithContext(ctx).Preload("Category")

	if params.CategoryID > 0 {
		query = query.Where("song_category_id = ?", params.CategoryID)
	}

	if err := query.Find(&songs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get songs"))
		return
	}

	result := make([]gin.H, len(songs))
	for i, s := range songs {
		result[i] = gin.H{
			"id":          s.ID,
			"title":       s.Title,
			"file_path":   s.FilePath,
			"thumbnail":   s.Thumbnail,
			"category_id": s.SongCategoryID,
			"category":    gin.H{"id": s.Category.ID, "name": s.Category.Name},
			"created_at":  s.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

// DeleteSong godoc
// @Summary Delete a song
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "Song ID"
// @Success 200 {object} dto.Response
// @Router /admin/songs/{id} [delete]
func (h *AdminHandler) DeleteSong(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	result := h.db.WithContext(ctx).Delete(&model.Song{}, id)
	if result.Error != nil || result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Song not found"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Song deleted"))
}

// ClearCache godoc
// @Summary Clear all cache
// @Description Clear all in-memory cache (useful after seeding)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /admin/cache/clear [post]
func (h *AdminHandler) ClearCache(c *gin.Context) {

	if h.cacheService != nil {
		h.cacheService.Clear()
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Cache cleared successfully"))
}

// ToggleForumFlag godoc
// @Summary Toggle forum flag status
// @Description Block or unblock a forum by toggling is_flagged
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Forum ID"
// @Success 200 {object} dto.Response
// @Router /admin/forums/{id}/toggle-flag [post]
func (h *AdminHandler) ToggleForumFlag(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var forum model.Forum
	if err := h.db.WithContext(ctx).First(&forum, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Forum not found"))
		return
	}

	// Toggle the flag
	newFlagState := !forum.IsFlagged
	updates := map[string]interface{}{
		"is_flagged":     newFlagState,
		"flagged_reason": "",
	}
	if newFlagState {
		updates["flagged_reason"] = "Diblokir oleh admin"
	}

	if err := h.db.WithContext(ctx).Model(&forum).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update forum"))
		return
	}

	message := "Forum berhasil dibuka"
	if newFlagState {
		message = "Forum berhasil diblokir"
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"is_flagged": newFlagState}, message))
}
