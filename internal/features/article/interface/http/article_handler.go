package handler

import (
	dailytaskapp "github.com/Alfian57/ruang-tenang-api/internal/features/daily_task/application"
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Alfian57/ruang-tenang-api/internal/features/article/application")

type ArticleHandler struct {
	articleService   *application.ArticleService
	dailyTaskService dailytaskapp.DailyTaskService
}

func NewArticleHandler(articleService *application.ArticleService) *ArticleHandler {
	return &ArticleHandler{articleService: articleService}
}

// SetDailyTaskService sets the daily task service for progress tracking
func (h *ArticleHandler) SetDailyTaskService(dailyTaskService dailytaskapp.DailyTaskService) {
	h.dailyTaskService = dailyTaskService
}

// GetArticles godoc
// @Summary Get articles list
// @Description Get paginated list of published articles with optional filtering
// @Tags Articles
// @Produce json
// @Param category_id query int false "Filter by category ID"
// @Param search query string false "Search by title"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.PaginatedResponse
// @Router /articles [get]
func (h *ArticleHandler) GetArticles(c *gin.Context) {
	ctx := c.Request.Context()
	var params dto.ArticleQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 10
	}

	// Public endpoint: only published articles
	articles, total, err := h.articleService.GetPublishedArticles(ctx, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get articles"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(articles, params.Page, params.Limit, total))
}

// GetArticle godoc
// @Summary Get article by slug
// @Description Get full article details by slug (only published)
// @Tags Articles
// @Produce json
// @Param slug path string true "Article Slug"
// @Success 200 {object} dto.ArticleDTO
// @Failure 404 {object} dto.Response
// @Router /articles/{slug} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("slug")

	// Public endpoint: only published articles
	article, err := h.articleService.GetPublishedArticleBySlug(ctx, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	// Update daily task progress for reading article (if user is authenticated)
	if h.dailyTaskService != nil {
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok && uid > 0 {
				if err := h.dailyTaskService.UpdateTaskProgress(ctx, uid, model.TaskTypeReadArticle); err != nil {
					logger.Warn("failed to update daily task progress for reading article",
						zap.Uint("user_id", uid), zap.Error(err))
				}
			}
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(article, ""))
}

// GetCategories godoc
// @Summary Get article categories
// @Description Get all article categories
// @Tags Articles
// @Produce json
// @Success 200 {object} dto.Response
// @Router /article-categories [get]
func (h *ArticleHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.articleService.GetCategories(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get categories"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(categories, ""))
}

// GetMyArticles godoc
// @Summary Get user's own articles
// @Description Get paginated list of articles owned by the authenticated user
// @Tags Articles
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.PaginatedResponse
// @Router /my-articles [get]
func (h *ArticleHandler) GetMyArticles(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	articles, total, err := h.articleService.GetUserArticles(ctx, userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get articles"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(articles, page, limit, total))
}

// CreateMyArticle godoc
// @Summary Create a new article
// @Description Create a new article as the authenticated user
// @Tags Articles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateUserArticleRequest true "Article data"
// @Success 201 {object} dto.Response
// @Router /my-articles [post]
func (h *ArticleHandler) CreateMyArticle(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	var req dto.CreateUserArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	article, err := h.articleService.CreateUserArticle(ctx, userID.(uint), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(gin.H{"id": article.ID}, "Article created successfully"))
}

// UpdateMyArticle godoc
// @Summary Update user's own article
// @Description Update an article owned by the authenticated user
// @Tags Articles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param slug path string true "Article Slug"
// @Param request body dto.UpdateUserArticleRequest true "Article data"
// @Success 200 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Router /my-articles/{slug} [put]
func (h *ArticleHandler) UpdateMyArticle(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	slug := c.Param("slug")

	var req dto.UpdateUserArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	_, err := h.articleService.UpdateUserArticleBySlug(ctx, userID.(uint), slug, &req)
	if err != nil {
		if err.Error() == "not authorized to update this article" {
			c.JSON(http.StatusForbidden, dto.ErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article updated successfully"))
}

// DeleteMyArticle godoc
// @Summary Delete user's own article
// @Description Delete an article owned by the authenticated user
// @Tags Articles
// @Security BearerAuth
// @Produce json
// @Param slug path string true "Article Slug"
// @Success 200 {object} dto.Response
// @Failure 403 {object} dto.Response
// @Router /my-articles/{slug} [delete]
func (h *ArticleHandler) DeleteMyArticle(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	slug := c.Param("slug")

	err := h.articleService.DeleteUserArticleBySlug(ctx, userID.(uint), slug)
	if err != nil {
		if err.Error() == "not authorized to delete this article" {
			c.JSON(http.StatusForbidden, dto.ErrorResponse(err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article deleted successfully"))
}

// GetArticleByIDForUser godoc
// @Summary Get article by slug for authenticated user
// @Description Get full article details (including own unpublished articles)
// @Tags Articles
// @Security BearerAuth
// @Produce json
// @Param slug path string true "Article Slug"
// @Success 200 {object} dto.ArticleDTO
// @Failure 404 {object} dto.Response
// @Router /my-articles/{slug} [get]
func (h *ArticleHandler) GetArticleByIDForUser(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	slug := c.Param("slug")

	article, err := h.articleService.GetArticleBySlug(ctx, slug)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
		return
	}

	// Check if user owns this article or if it's published
	if article.UserID != userID.(uint) {
		if article.Status != string(model.ArticleStatusPublished) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Article not found"))
			return
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(article, ""))
}
