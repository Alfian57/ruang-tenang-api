package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ForumHandler struct {
	service          services.ForumService
	dailyTaskService services.DailyTaskService
}

func NewForumHandler(service services.ForumService) *ForumHandler {
	return &ForumHandler{service: service}
}

// SetDailyTaskService sets the daily task service for progress tracking
func (h *ForumHandler) SetDailyTaskService(dailyTaskService services.DailyTaskService) {
	h.dailyTaskService = dailyTaskService
}

// @Summary Create a new forum topic
// @Description Create a new forum topic
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{title=string,content=string,category_id=int} true "Forum request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forums [post]
func (h *ForumHandler) CreateForum(c *gin.Context) {
	userID := c.GetUint("user_id")
	var req struct {
		Title      string `json:"title" binding:"required"`
		Content    string `json:"content"`
		CategoryID *uint  `json:"category_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateForum(userID, req.Title, req.Content, req.CategoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Forum created successfully"})
}

// @Summary Get list of forums
// @Description Get list of forums with pagination and filters
// @Tags forum
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search term"
// @Param category_id query int false "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /forums [get]
func (h *ForumHandler) GetForums(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	search := c.Query("search")

	var categoryID *uint
	if catStr := c.Query("category_id"); catStr != "" {
		id, _ := strconv.Atoi(catStr)
		uid := uint(id)
		categoryID = &uid
	}

	forums, total, err := h.service.GetForums(limit, offset, search, categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  forums,
		"total": total,
		"limit": limit,
		"page":  offset/limit + 1,
	})
}

// @Summary Get forum details
// @Description Get details of a specific forum
// @Tags forum
// @Accept json
// @Produce json
// @Param id path int true "Forum ID"
// @Success 200 {object} models.Forum
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forums/{id} [get]
func (h *ForumHandler) GetForumByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	id, _ := strconv.Atoi(c.Param("id"))
	forum, err := h.service.GetForumByID(userID, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Forum not found"})
		return
	}

	c.JSON(http.StatusOK, forum)
}

// @Summary Delete a forum
// @Description Delete a forum (Owner or Admin only)
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Forum ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forums/{id} [delete]
func (h *ForumHandler) DeleteForum(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.DeleteForum(userID, userRole, uint(id)); err != nil {
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this forum"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Forum deleted successfully"})
}

// @Summary Create a forum post (reply)
// @Description Create a reply to a forum topic
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Forum ID"
// @Param request body object{content=string} true "Post request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forums/{id} [post]
func (h *ForumHandler) CreateForumPost(c *gin.Context) {
	userID := c.GetUint("user_id")
	forumID, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateForumPost(userID, uint(forumID), req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update daily task progress for commenting in forum
	if h.dailyTaskService != nil {
		_ = h.dailyTaskService.UpdateTaskProgress(userID, models.TaskTypeCommentForum)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Post created successfully"})
}

// @Summary Get forum posts
// @Description Get replies for a forum topic with sorting options
// @Tags forum
// @Accept json
// @Produce json
// @Param id path int true "Forum ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param sort query string false "Sort by: top, newest, oldest" default(top)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /forums/{id}/posts [get]
func (h *ForumHandler) GetForumPosts(c *gin.Context) {
	userID := c.GetUint("user_id")
	forumID, _ := strconv.Atoi(c.Param("id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sort := c.DefaultQuery("sort", "top")

	posts, total, err := h.service.GetForumPostsSorted(uint(forumID), limit, offset, sort, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  posts,
		"total": total,
		"limit": limit,
		"page":  offset/limit + 1,
		"sort":  sort,
	})
}

// @Summary Delete a forum post
// @Description Delete a forum post (Owner or Admin only)
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id} [delete]
func (h *ForumHandler) DeleteForumPost(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.DeleteForumPost(userID, userRole, uint(id)); err != nil {
		if err.Error() == "unauthorized" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this post"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// @Summary Toggle forum like
// @Description Like or unlike a forum
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Forum ID"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /forums/{id}/like [put]
func (h *ForumHandler) ToggleLike(c *gin.Context) {
	userID := c.GetUint("user_id")
	forumID, _ := strconv.Atoi(c.Param("id"))

	liked, err := h.service.ToggleLike(userID, uint(forumID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	message := "Forum liked"
	if !liked {
		message = "Forum unliked"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
		"liked":   liked,
	})
}

// ==================== Post Voting Handlers ====================

// @Summary Upvote a forum post
// @Description Upvote a forum post/answer
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id}/upvote [put]
func (h *ForumHandler) UpvotePost(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.VotePost(userID, uint(postID), "upvote"); err != nil {
		if err.Error() == "cannot vote on your own post" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post upvoted successfully",
		"voted":   true,
	})
}

// @Summary Downvote a forum post
// @Description Downvote a forum post/answer (if enabled)
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id}/downvote [put]
func (h *ForumHandler) DownvotePost(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.VotePost(userID, uint(postID), "downvote"); err != nil {
		if err.Error() == "cannot vote on your own post" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Post downvoted successfully",
		"voted":   true,
	})
}

// @Summary Remove vote from a forum post
// @Description Remove user's vote from a forum post
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id}/vote [delete]
func (h *ForumHandler) RemovePostVote(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.RemovePostVote(userID, uint(postID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote removed successfully",
		"voted":   false,
	})
}

// ==================== Best Answer Handlers ====================

// @Summary Mark post as accepted answer
// @Description Mark a forum post as the accepted answer (OP only)
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id}/accept [put]
func (h *ForumHandler) MarkAcceptedAnswer(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	postID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.MarkAsAcceptedAnswer(userID, userRole, uint(postID)); err != nil {
		if err.Error() == "only the thread creator can mark an accepted answer" ||
			err.Error() == "cannot mark your own answer as accepted" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Answer marked as accepted",
	})
}

// @Summary Unmark accepted answer for a forum
// @Description Remove the accepted answer mark from a forum (OP only)
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Forum ID"
// @Success 200 {object} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forums/{id}/accepted-answer [delete]
func (h *ForumHandler) UnmarkAcceptedAnswer(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	forumID, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.UnmarkAcceptedAnswer(userID, userRole, uint(forumID)); err != nil {
		if err.Error() == "only the thread creator can unmark an accepted answer" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Accepted answer unmarked",
	})
}

// @Summary Get accepted answer for a forum
// @Description Get the accepted answer for a forum topic
// @Tags forum
// @Accept json
// @Produce json
// @Param id path int true "Forum ID"
// @Success 200 {object} models.ForumPost
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /forums/{id}/accepted-answer [get]
func (h *ForumHandler) GetAcceptedAnswer(c *gin.Context) {
	forumID, _ := strconv.Atoi(c.Param("id"))

	post, err := h.service.GetAcceptedAnswer(uint(forumID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No accepted answer found"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// ==================== Report Handlers ====================

// @Summary Report a forum post
// @Description Report a forum post for moderation
// @Tags forum
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param request body object{reason=string,description=string} true "Report request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /posts/{id}/report [post]
func (h *ForumHandler) ReportPost(c *gin.Context) {
	userID := c.GetUint("user_id")
	postID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Reason      string `json:"reason" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReportPost(userID, uint(postID), req.Reason, req.Description); err != nil {
		if err.Error() == "invalid report reason" ||
			err.Error() == "cannot report your own post" ||
			err.Error() == "you have already reported this post" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Post reported successfully",
	})
}

// @Summary Get pending post reports
// @Description Get all pending post reports for moderation (Moderator only)
// @Tags moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /moderation/post-reports [get]
func (h *ForumHandler) GetPendingPostReports(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	reports, total, err := h.service.GetPendingPostReports(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reports,
		"total": total,
		"limit": limit,
		"page":  offset/limit + 1,
	})
}

// @Summary Review a post report
// @Description Review and action a post report (Moderator only)
// @Tags moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Report ID"
// @Param request body object{status=string,notes=string} true "Review request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /moderation/post-reports/{id} [put]
func (h *ForumHandler) ReviewPostReport(c *gin.Context) {
	reviewerID := c.GetUint("user_id")
	reportID, _ := strconv.Atoi(c.Param("id"))

	var req struct {
		Status string `json:"status" binding:"required"` // reviewed, dismissed, actioned
		Notes  string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReviewPostReport(reviewerID, uint(reportID), req.Status, req.Notes); err != nil {
		if err.Error() == "invalid report status" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report reviewed successfully",
	})
}

// @Summary Get report reasons
// @Description Get available report reasons for forum posts
// @Tags forum
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /forums/report-reasons [get]
func (h *ForumHandler) GetReportReasons(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"reasons": dto.PostReportReasons(),
	})
}

// @Summary Get sort options
// @Description Get available sort options for forum posts
// @Tags forum
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /forums/sort-options [get]
func (h *ForumHandler) GetSortOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"options": dto.GetForumPostSortOptions(),
	})
}
