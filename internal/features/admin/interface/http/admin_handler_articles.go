package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

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

	// Validate the referenced category exists for a clear 400 instead of a
	// generic 500 from a foreign-key violation.
	var categoryCount int64
	h.db.WithContext(ctx).Model(&model.ArticleCategory{}).Where("id = ?", req.CategoryID).Count(&categoryCount)
	if categoryCount == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Kategori artikel tidak ditemukan"))
		return
	}

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

// GetArticle godoc
// @Summary Get an article by ID
// @Description Get an article by ID (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Success 200 {object} dto.Response
// @Router /admin/articles/{id} [get]
func (h *AdminHandler) GetArticle(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var article model.Article

	query := h.db.WithContext(ctx).Preload("Category").Preload("Author")
	if _, err := strconv.Atoi(id); err == nil {
		query = query.Where("id = ?", id)
	} else {
		query = query.Where("slug = ?", id)
	}
	if err := query.First(&article).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	result := gin.H{
		"id":                article.ID,
		"slug":              article.Slug,
		"title":             article.Title,
		"content":           article.Content,
		"thumbnail":         article.Thumbnail,
		"category_id":       article.ArticleCategoryID,
		"category":          gin.H{"id": article.Category.ID, "name": article.Category.Name},
		"status":            article.Status,
		"moderation_status": article.ModerationStatus,
		"user_id":           article.UserID,
		"created_at":        article.CreatedAt,
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
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

	// Validate the referenced category exists.
	var categoryCount int64
	h.db.WithContext(ctx).Model(&model.ArticleCategory{}).Where("id = ?", req.CategoryID).Count(&categoryCount)
	if categoryCount == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Kategori artikel tidak ditemukan"))
		return
	}

	article.Title = req.Title
	article.Content = req.Content
	article.ArticleCategoryID = req.CategoryID
	// Only replace the thumbnail when a new one is provided so an omitted
	// thumbnail doesn't silently wipe the existing image.
	if req.Thumbnail != "" {
		article.Thumbnail = req.Thumbnail
	}

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
			"slug":              a.Slug,
			"title":             a.Title,
			"thumbnail":         a.Thumbnail,
			"category_id":       a.ArticleCategoryID,
			"category":          gin.H{"id": a.Category.ID, "name": a.Category.Name},
			"status":            a.Status,
			"moderation_status": a.ModerationStatus,
			"user_id":           a.UserID,
			"is_user_generated": a.IsUserGenerated,
			"created_at":        a.CreatedAt,
		}
		if a.Author != nil {
			item["author"] = gin.H{"id": a.Author.ID, "name": a.Author.Name}
		}
		result[i] = item
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(result, params.Page, params.Limit, total))
}
