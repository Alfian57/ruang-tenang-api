package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ForumCategoryHandler struct {
	service services.ForumCategoryService
}

func NewForumCategoryHandler(service services.ForumCategoryService) *ForumCategoryHandler {
	return &ForumCategoryHandler{service}
}

// @Summary Create a new forum category
// @Description Create a new forum category (Admin only)
// @Tags forum-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{name=string} true "Category request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/forum-categories [post]
func (h *ForumCategoryHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateCategory(req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Category created successfully"})
}

// @Summary Get all forum categories
// @Description Get all forum categories
// @Tags forum-category
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /forum-categories [get]
func (h *ForumCategoryHandler) GetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// @Summary Get all forum categories for admin
// @Description Get all forum categories for admin
// @Tags forum-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /admin/forum-categories [get]
func (h *ForumCategoryHandler) AdminGetAllCategories(c *gin.Context) {
	categories, err := h.service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

// @Summary Update a forum category
// @Description Update a forum category (Admin only)
// @Tags forum-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param request body object{name=string} true "Category request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/forum-categories/{id} [put]
func (h *ForumCategoryHandler) UpdateCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.UpdateCategory(uint(id), req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category updated successfully"})
}

// @Summary Delete a forum category
// @Description Delete a forum category (Admin only)
// @Tags forum-category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /admin/forum-categories/{id} [delete]
func (h *ForumCategoryHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.DeleteCategory(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
