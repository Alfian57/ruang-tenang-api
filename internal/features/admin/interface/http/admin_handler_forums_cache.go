package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

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

// GetForums godoc
// @Summary Get all forums for admin
// @Description Get paginated list of all forums with optional filtering (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param search query string false "Search by title"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.PaginatedResponse
// @Router /admin/forums [get]
func (h *AdminHandler) GetForums(c *gin.Context) {
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

	forums, total, err := h.forumRepo.GetForums(ctx, params.Limit, (params.Page-1)*params.Limit, params.Search, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get forums"))
		return
	}

	result := make([]gin.H, len(forums))
	for i, f := range forums {
		item := gin.H{
			"id":         f.ID,
			"title":      f.Title,
			"content":    f.Content,
			"slug":       f.Slug,
			"is_flagged": f.IsFlagged,
			"created_at": f.CreatedAt,
			"updated_at": f.UpdatedAt,
		}

		if f.Category.ID != 0 {
			item["category"] = gin.H{
				"id":   f.Category.ID,
				"name": f.Category.Name,
			}
		}

		if f.User.ID != 0 {
			item["user"] = gin.H{
				"id":   f.User.ID,
				"name": f.User.Name,
			}
		}

		result[i] = item
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(result, params.Page, params.Limit, total))
}

// Copy of GetForumByID from ForumHandler but tailored for Admin
// GetForum godoc
// @Summary Get forum details
// @Description Get details of a specific forum by ID (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Forum ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /admin/forums/{id} [get]
func (h *AdminHandler) GetForum(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	var forum model.Forum
	if err := h.db.WithContext(ctx).Preload("User").Preload("Category").First(&forum, id).Error; err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Forum not found"))
		return
	}

	var likesCount int64
	h.db.WithContext(ctx).Model(&model.ForumLike{}).Where("forum_id = ?", forum.ID).Count(&likesCount)
	forum.LikesCount = likesCount

	var repliesCount int64
	h.db.WithContext(ctx).Model(&model.ForumPost{}).Where("forum_id = ?", forum.ID).Count(&repliesCount)
	forum.RepliesCount = repliesCount

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"data": forum}, ""))
}
