package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	foruminfra "github.com/Alfian57/ruang-tenang-api/internal/features/forum/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
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

func (h *AdminHandler) findForumByIdentifier(c *gin.Context) (*model.Forum, bool) {
	ctx := c.Request.Context()
	identifier := c.Param("identifier")
	if identifier == "" {
		identifier = c.Param("id")
	}

	if id, err := strconv.ParseUint(identifier, 10, 32); err == nil {
		forum, err := h.forumRepo.GetForumByID(ctx, uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Forum not found"))
			return nil, false
		}
		return forum, true
	}

	forum, err := h.forumRepo.GetForumBySlug(ctx, identifier)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Forum not found"))
		return nil, false
	}
	return forum, true
}

func parseAdminLimitOffset(c *gin.Context, defaultLimit int) (int, int) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultLimit)))
	if err != nil || limit <= 0 {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	return limit, offset
}

func parseAdminPostSort(value string) foruminfra.PostSortOption {
	switch value {
	case string(foruminfra.SortByNewest):
		return foruminfra.SortByNewest
	case string(foruminfra.SortByOldest):
		return foruminfra.SortByOldest
	default:
		return foruminfra.SortByTop
	}
}

// ToggleForumFlag godoc
// @Summary Toggle forum flag status
// @Description Block or unblock a forum by toggling is_flagged
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param identifier path string true "Forum ID or slug"
// @Success 200 {object} dto.Response
// @Router /admin/forums/{identifier}/toggle-flag [post]
func (h *AdminHandler) ToggleForumFlag(c *gin.Context) {
	ctx := c.Request.Context()
	forum, ok := h.findForumByIdentifier(c)
	if !ok {
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

// GetForum godoc
// @Summary Get forum details
// @Description Get details of a specific forum by ID or slug (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param identifier path string true "Forum ID or slug"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Failure 500 {object} dto.Response
// @Router /admin/forums/{identifier} [get]
func (h *AdminHandler) GetForum(c *gin.Context) {
	ctx := c.Request.Context()
	forum, ok := h.findForumByIdentifier(c)
	if !ok {
		return
	}

	var likesCount int64
	h.db.WithContext(ctx).Model(&model.ForumLike{}).Where("forum_id = ?", forum.ID).Count(&likesCount)
	forum.LikesCount = likesCount

	var repliesCount int64
	h.db.WithContext(ctx).Model(&model.ForumPost{}).Where("forum_id = ?", forum.ID).Count(&repliesCount)
	forum.RepliesCount = repliesCount

	c.JSON(http.StatusOK, dto.SuccessResponse(forum, "Forum retrieved successfully"))
}

// GetForumPosts godoc
// @Summary Get forum posts for admin
// @Description Get replies for a forum topic by ID or slug (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param identifier path string true "Forum ID or slug"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param sort query string false "Sort"
// @Success 200 {object} dto.PaginatedResponse
// @Router /admin/forums/{identifier}/posts [get]
func (h *AdminHandler) GetForumPosts(c *gin.Context) {
	ctx := c.Request.Context()
	forum, ok := h.findForumByIdentifier(c)
	if !ok {
		return
	}

	limit, offset := parseAdminLimitOffset(c, 20)
	sort := parseAdminPostSort(c.DefaultQuery("sort", "top"))
	posts, total, err := h.forumRepo.GetForumPostsSorted(ctx, forum.ID, limit, offset, sort, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get forum posts"))
		return
	}

	page := offset/limit + 1
	c.JSON(http.StatusOK, dto.NewPaginatedResponse(posts, page, limit, total))
}

// CreateForumPost godoc
// @Summary Create forum post as admin
// @Description Create a reply on a forum topic as admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param identifier path string true "Forum ID or slug"
// @Success 201 {object} dto.Response
// @Router /admin/forums/{identifier}/posts [post]
func (h *AdminHandler) CreateForumPost(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	forum, ok := h.findForumByIdentifier(c)
	if !ok {
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid request: "+err.Error()))
		return
	}

	post := &model.ForumPost{
		UserID:  userID,
		ForumID: forum.ID,
		Content: req.Content,
	}
	if err := h.forumRepo.CreateForumPost(ctx, post); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create forum post"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(post, "Post created successfully"))
}

// DeleteForum godoc
// @Summary Delete forum as admin
// @Description Delete a forum by ID or slug (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param identifier path string true "Forum ID or slug"
// @Success 200 {object} dto.Response
// @Router /admin/forums/{identifier} [delete]
func (h *AdminHandler) DeleteForum(c *gin.Context) {
	ctx := c.Request.Context()
	forum, ok := h.findForumByIdentifier(c)
	if !ok {
		return
	}

	if err := h.forumRepo.DeleteForum(ctx, forum.ID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to delete forum"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Forum deleted successfully"))
}

// DeleteForumPost godoc
// @Summary Delete forum post as admin
// @Description Delete a forum reply (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} dto.Response
// @Router /admin/forum-posts/{id} [delete]
func (h *AdminHandler) DeleteForumPost(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid post ID"))
		return
	}

	if err := h.forumRepo.DeleteForumPost(ctx, uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to delete forum post"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Post deleted successfully"))
}

// ToggleAcceptedAnswer godoc
// @Summary Toggle accepted answer as admin
// @Description Mark or unmark a forum reply as accepted answer (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} dto.Response
// @Router /admin/forum-posts/{id}/accepted-answer [put]
func (h *AdminHandler) ToggleAcceptedAnswer(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid post ID"))
		return
	}

	post, err := h.forumRepo.GetForumPostByID(ctx, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("Post not found"))
		return
	}

	if post.IsAcceptedAnswer {
		if err := h.forumRepo.UnmarkAcceptedAnswer(ctx, post.ForumID); err != nil {
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to unmark accepted answer"))
			return
		}
		c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"is_accepted_answer": false}, "Accepted answer removed"))
		return
	}

	if err := h.forumRepo.MarkAsAcceptedAnswer(ctx, post.ID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to mark accepted answer"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"is_accepted_answer": true}, "Accepted answer updated"))
}
