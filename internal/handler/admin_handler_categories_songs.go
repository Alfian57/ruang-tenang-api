package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

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
