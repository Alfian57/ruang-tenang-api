package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ForumHandler struct {
	service services.ForumService
}

func NewForumHandler(service services.ForumService) *ForumHandler {
	return &ForumHandler{service}
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
	userRole := c.GetString("role")
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

	c.JSON(http.StatusCreated, gin.H{"message": "Post created successfully"})
}

// @Summary Get forum posts
// @Description Get replies for a forum topic
// @Tags forum
// @Accept json
// @Produce json
// @Param id path int true "Forum ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /forums/{id}/posts [get]
func (h *ForumHandler) GetForumPosts(c *gin.Context) {
	forumID, _ := strconv.Atoi(c.Param("id"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, total, err := h.service.GetForumPosts(uint(forumID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  posts,
		"total": total,
		"limit": limit,
		"page":  offset/limit + 1,
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
	userRole := c.GetString("role")
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
