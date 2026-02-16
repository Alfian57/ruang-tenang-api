package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InspiringStoryHandler struct {
	storyService *service.InspiringStoryService
}

func NewInspiringStoryHandler(storyService *service.InspiringStoryService) *InspiringStoryHandler {
	return &InspiringStoryHandler{
		storyService: storyService,
	}
}

// ==========================================
// Categories
// ==========================================

// GetCategories godoc
// @Summary Get story categories
// @Description Get all story categories with story counts
// @Tags stories
// @Accept json
// @Produce json
// @Success 200 {object} []dto.StoryCategoryResponse
// @Router /api/v1/stories/categories [get]
func (h *InspiringStoryHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories, err := h.storyService.GetCategories(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil kategori"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(categories, "Success"))
}

// ==========================================
// Story CRUD
// ==========================================

// CreateStory godoc
// @Summary Create a new story
// @Description Submit a new inspiring story for moderation
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param story body dto.CreateStoryRequest true "Story data"
// @Success 201 {object} dto.StoryResponse
// @Router /api/v1/stories [post]
func (h *InspiringStoryHandler) CreateStory(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	var req dto.CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid: "+err.Error()))
		return
	}

	story, err := h.storyService.CreateStory(ctx, userID.(uint), &req)
	if err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal membuat cerita"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(story, "Cerita berhasil dikirim untuk ditinjau"))
}

// UpdateStory godoc
// @Summary Update a story
// @Description Update an existing story (only pending or revision_requested)
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param story body dto.UpdateStoryRequest true "Story data"
// @Success 200 {object} dto.StoryResponse
// @Router /api/v1/stories/{id} [put]
func (h *InspiringStoryHandler) UpdateStory(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	var req dto.UpdateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid: "+err.Error()))
		return
	}

	story, err := h.storyService.UpdateStory(ctx, storyID, userID.(uint), &req)
	if err != nil {
		if se, ok := service.IsServiceError(err); ok {
			status := http.StatusBadRequest
			if se.Code == "UNAUTHORIZED" {
				status = http.StatusForbidden
			} else if se.Code == "STORY_NOT_FOUND" {
				status = http.StatusNotFound
			}
			c.JSON(status, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal memperbarui cerita"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(story, "Cerita berhasil diperbarui"))
}

// DeleteStory godoc
// @Summary Delete a story
// @Description Delete a story owned by the user
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/stories/{id} [delete]
func (h *InspiringStoryHandler) DeleteStory(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	if err := h.storyService.DeleteStory(ctx, storyID, userID.(uint)); err != nil {
		if se, ok := service.IsServiceError(err); ok {
			status := http.StatusBadRequest
			if se.Code == "UNAUTHORIZED" {
				status = http.StatusForbidden
			} else if se.Code == "STORY_NOT_FOUND" {
				status = http.StatusNotFound
			}
			c.JSON(status, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal menghapus cerita"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Cerita berhasil dihapus"))
}

// GetStory godoc
// @Summary Get a story
// @Description Get a single story by ID
// @Tags stories
// @Accept json
// @Produce json
// @Param id path string true "Story ID"
// @Success 200 {object} dto.StoryResponse
// @Router /api/v1/stories/{id} [get]
func (h *InspiringStoryHandler) GetStory(c *gin.Context) {
	ctx := c.Request.Context()
	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	// Get viewer ID if authenticated
	var viewerID uint
	if userID, exists := c.Get("user_id"); exists {
		viewerID = userID.(uint)
	}

	story, err := h.storyService.GetStory(ctx, storyID, viewerID)
	if err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusNotFound, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil cerita"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(story, "Success"))
}

// GetStories godoc
// @Summary Get stories
// @Description Get paginated list of approved stories
// @Tags stories
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Param category_id query string false "Filter by category ID"
// @Param search query string false "Search in title and content"
// @Param sort_by query string false "Sort by: recent, hearts, featured"
// @Success 200 {object} dto.StoriesListResponse
// @Router /api/v1/stories [get]
func (h *InspiringStoryHandler) GetStories(c *gin.Context) {
	ctx := c.Request.Context()
	var filter dto.StoryFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Parameter tidak valid"))
		return
	}

	// Get viewer ID if authenticated
	var viewerID uint
	if userID, exists := c.Get("user_id"); exists {
		viewerID = userID.(uint)
	}

	stories, err := h.storyService.GetStories(ctx, &filter, viewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil daftar cerita"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(stories.Stories, stories.Page, stories.Limit, stories.Total))
}

// GetMyStories godoc
// @Summary Get my stories
// @Description Get current user's stories
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Param status query string false "Filter by status"
// @Success 200 {object} dto.StoriesListResponse
// @Router /api/v1/stories/my-stories [get]
func (h *InspiringStoryHandler) GetMyStories(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	stories, err := h.storyService.GetUserStories(ctx, userID.(uint), status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil cerita"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(stories.Stories, stories.Page, stories.Limit, stories.Total))
}

// GetFeaturedStories godoc
// @Summary Get featured stories
// @Description Get featured stories
// @Tags stories
// @Accept json
// @Produce json
// @Param limit query int false "Number of stories (default 5)"
// @Success 200 {object} []dto.StoryCardResponse
// @Router /api/v1/stories/featured [get]
func (h *InspiringStoryHandler) GetFeaturedStories(c *gin.Context) {
	ctx := c.Request.Context()
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	stories, err := h.storyService.GetFeaturedStories(ctx, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil cerita unggulan"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stories, "Success"))
}

// ==========================================
// Hearts
// ==========================================

// ToggleHeart godoc
// @Summary Toggle heart on story
// @Description Toggle heart (like) on a story
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/stories/{id}/heart [post]
func (h *InspiringStoryHandler) ToggleHeart(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	hasHearted, heartCount, err := h.storyService.ToggleHeart(ctx, storyID, userID.(uint))
	if err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal memproses heart"))
		return
	}

	message := "Heart berhasil ditambahkan"
	if !hasHearted {
		message = "Heart berhasil dihapus"
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"has_hearted": hasHearted,
		"heart_count": heartCount,
	}, message))
}

// ==========================================
// Comments
// ==========================================

// CreateComment godoc
// @Summary Create a comment
// @Description Add a supportive comment to a story
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param comment body dto.CreateStoryCommentRequest true "Comment data"
// @Success 201 {object} dto.StoryCommentResponse
// @Router /api/v1/stories/{id}/comments [post]
func (h *InspiringStoryHandler) CreateComment(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	var req dto.CreateStoryCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid: "+err.Error()))
		return
	}

	comment, err := h.storyService.CreateComment(ctx, storyID, userID.(uint), &req)
	if err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal membuat komentar"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(comment, "Komentar berhasil ditambahkan"))
}

// GetComments godoc
// @Summary Get story comments
// @Description Get comments for a story
// @Tags stories
// @Accept json
// @Produce json
// @Param id path string true "Story ID"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20)"
// @Success 200 {object} dto.StoryCommentsListResponse
// @Router /api/v1/stories/{id}/comments [get]
func (h *InspiringStoryHandler) GetComments(c *gin.Context) {
	ctx := c.Request.Context()
	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get viewer ID if authenticated
	var viewerID uint
	if userID, exists := c.Get("user_id"); exists {
		viewerID = userID.(uint)
	}

	comments, err := h.storyService.GetComments(ctx, storyID, viewerID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil komentar"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(comments, "Success"))
}

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment owned by the user
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param commentId path string true "Comment ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/stories/{id}/comments/{commentId} [delete]
func (h *InspiringStoryHandler) DeleteComment(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID komentar tidak valid"))
		return
	}

	if err := h.storyService.DeleteComment(ctx, commentID, userID.(uint)); err != nil {
		if se, ok := service.IsServiceError(err); ok {
			status := http.StatusBadRequest
			if se.Code == "UNAUTHORIZED" {
				status = http.StatusForbidden
			} else if se.Code == "COMMENT_NOT_FOUND" {
				status = http.StatusNotFound
			}
			c.JSON(status, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal menghapus komentar"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Komentar berhasil dihapus"))
}

// ToggleCommentHeart godoc
// @Summary Toggle heart on comment
// @Description Toggle heart on a comment
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param commentId path string true "Comment ID"
// @Success 200 {object} dto.Response
// @Router /api/v1/stories/{id}/comments/{commentId}/heart [post]
func (h *InspiringStoryHandler) ToggleCommentHeart(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID komentar tidak valid"))
		return
	}

	hasHearted, heartCount, err := h.storyService.ToggleCommentHeart(ctx, commentID, userID.(uint))
	if err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal memproses heart"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"has_hearted": hasHearted,
		"heart_count": heartCount,
	}, "Comment heart status retrieved"))
}

// ==========================================
// Stats
// ==========================================

// GetMyStats godoc
// @Summary Get my story stats
// @Description Get current user's story statistics
// @Tags stories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.StoryStatsResponse
// @Router /api/v1/stories/my-stats [get]
func (h *InspiringStoryHandler) GetMyStats(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	stats, err := h.storyService.GetAuthorStats(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil statistik"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stats, "Success"))
}

// GetMostAppreciated godoc
// @Summary Get most appreciated stories
// @Description Get most appreciated stories for a month
// @Tags stories
// @Accept json
// @Produce json
// @Param month query int false "Month (default current month)"
// @Param year query int false "Year (default current year)"
// @Param limit query int false "Limit (default 10)"
// @Success 200 {object} dto.MostAppreciatedStoriesResponse
// @Router /api/v1/stories/most-appreciated [get]
func (h *InspiringStoryHandler) GetMostAppreciated(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now()
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	stories, err := h.storyService.GetMostAppreciatedStories(ctx, month, year, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil cerita terpopuler"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stories, "Success"))
}

// ==========================================
// Moderation (Admin/Moderator)
// ==========================================

// GetPendingStories godoc
// @Summary Get pending stories
// @Description Get stories pending moderation (admin/moderator only)
// @Tags stories-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20)"
// @Success 200 {object} dto.StoriesListResponse
// @Router /api/v1/admin/stories/pending [get]
func (h *InspiringStoryHandler) GetPendingStories(c *gin.Context) {
	ctx := c.Request.Context()
	// Role check should be done by middleware
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	role := userRole.(model.UserRole)
	if role != model.RoleAdmin && role != model.RoleModerator {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("Akses ditolak"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	stories, err := h.storyService.GetPendingStories(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil cerita pending"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(stories.Stories, stories.Page, stories.Limit, stories.Total))
}

// ModerateStory godoc
// @Summary Moderate a story
// @Description Approve, reject, or request revision for a story
// @Tags stories-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param moderation body dto.ModerateStoryRequest true "Moderation data"
// @Success 200 {object} dto.Response
// @Router /api/v1/admin/stories/{id}/moderate [post]
func (h *InspiringStoryHandler) ModerateStory(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	role := userRole.(model.UserRole)
	if role != model.RoleAdmin && role != model.RoleModerator {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("Akses ditolak"))
		return
	}

	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	var req dto.ModerateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid: "+err.Error()))
		return
	}

	if err := h.storyService.ModerateStory(ctx, storyID, userID.(uint), &req); err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal moderasi cerita"))
		return
	}

	message := "Cerita berhasil dimoderasi"
	if req.Status == "approved" {
		message = "Cerita berhasil disetujui dan dipublikasikan"
	} else if req.Status == "rejected" {
		message = "Cerita ditolak"
	} else if req.Status == "revision_requested" {
		message = "Permintaan revisi telah dikirim ke penulis"
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, message))
}

// SetFeatured godoc
// @Summary Set featured status
// @Description Set or remove featured status for a story
// @Tags stories-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param featured query bool true "Featured status"
// @Success 200 {object} dto.Response
// @Router /api/v1/admin/stories/{id}/featured [post]
func (h *InspiringStoryHandler) SetFeatured(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	role := userRole.(model.UserRole)
	if role != model.RoleAdmin && role != model.RoleModerator {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("Akses ditolak"))
		return
	}

	storyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID cerita tidak valid"))
		return
	}

	featured := c.Query("featured") == "true"

	if err := h.storyService.SetFeatured(ctx, storyID, featured, userID.(uint)); err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengubah status unggulan"))
		return
	}

	message := "Cerita berhasil dijadikan unggulan"
	if !featured {
		message = "Status unggulan cerita berhasil dihapus"
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, message))
}

// HideComment godoc
// @Summary Hide a comment
// @Description Hide a comment (moderator action)
// @Tags stories-admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Story ID"
// @Param commentId path string true "Comment ID"
// @Param request body dto.HideStoryCommentRequest true "Hide reason"
// @Success 200 {object} dto.Response
// @Router /api/v1/admin/stories/{id}/comments/{commentId}/hide [post]
func (h *InspiringStoryHandler) HideComment(c *gin.Context) {
	ctx := c.Request.Context()
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	role := userRole.(model.UserRole)
	if role != model.RoleAdmin && role != model.RoleModerator {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("Akses ditolak"))
		return
	}

	commentID, err := uuid.Parse(c.Param("commentId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID komentar tidak valid"))
		return
	}

	var req dto.HideStoryCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Data tidak valid: "+err.Error()))
		return
	}

	if err := h.storyService.HideComment(ctx, commentID, &req); err != nil {
		if se, ok := service.IsServiceError(err); ok {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(se.Message))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal menyembunyikan komentar"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Komentar berhasil disembunyikan"))
}
