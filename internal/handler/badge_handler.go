package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type BadgeHandler struct {
	badgeService *service.BadgeService
}

func NewBadgeHandler(badgeService *service.BadgeService) *BadgeHandler {
	return &BadgeHandler{
		badgeService: badgeService,
	}
}

// GetAllBadges godoc
// @Summary Get all badges
// @Description Get all badge definitions
// @Tags badges
// @Accept json
// @Produce json
// @Success 200 {object} []dto.BadgeDefinitionResponse
// @Router /api/v1/badges [get]
func (h *BadgeHandler) GetAllBadges(c *gin.Context) {
	ctx := c.Request.Context()
	badges, err := h.badgeService.GetAllBadges(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil daftar badge"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(badges, "Daftar badge berhasil diambil"))
}

// GetBadgesByCategory godoc
// @Summary Get badges by category
// @Description Get badges filtered by category
// @Tags badges
// @Accept json
// @Produce json
// @Param category path string true "Category key"
// @Success 200 {object} []dto.BadgeDefinitionResponse
// @Router /api/v1/badges/category/{category} [get]
func (h *BadgeHandler) GetBadgesByCategory(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Param("category")

	badges, err := h.badgeService.GetBadgesByCategory(ctx, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil badge berdasarkan kategori"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(badges, "Badge berdasarkan kategori berhasil diambil"))
}

// GetBadgeCategories godoc
// @Summary Get badge categories
// @Description Get all badge categories
// @Tags badges
// @Accept json
// @Produce json
// @Success 200 {object} []dto.BadgeCategoryInfo
// @Router /api/v1/badges/categories [get]
func (h *BadgeHandler) GetBadgeCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories := h.badgeService.GetBadgeCategories(ctx)

	c.JSON(http.StatusOK, dto.SuccessResponse(categories, "Kategori badge berhasil diambil"))
}

// GetUserBadges godoc
// @Summary Get user badges
// @Description Get current user's earned badges
// @Tags badges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserBadgesResponse
// @Router /api/v1/badges/my-badges [get]
func (h *BadgeHandler) GetUserBadges(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	badges, err := h.badgeService.GetUserBadges(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil badge user"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(badges, "Badge user berhasil diambil"))
}

// GetBadgeProgress godoc
// @Summary Get badge progress
// @Description Get progress towards all badges for current user
// @Tags badges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []dto.BadgeProgressResponse
// @Router /api/v1/badges/progress [get]
func (h *BadgeHandler) GetBadgeProgress(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	progress, err := h.badgeService.GetBadgeProgress(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil progress badge"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(progress, "Progress badge berhasil diambil"))
}

// GetRecentlyEarnedBadges godoc
// @Summary Get recently earned badges
// @Description Get badges earned within a specified number of days
// @Tags badges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days (default 7)"
// @Success 200 {object} []dto.BadgeResponse
// @Router /api/v1/badges/recent [get]
func (h *BadgeHandler) GetRecentlyEarnedBadges(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))
	if days > 365 {
		days = 365
	}

	badges, err := h.badgeService.GetRecentlyEarnedBadges(ctx, userID.(uint), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil badge terbaru"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(badges, "Badge terbaru berhasil diambil"))
}

// CheckNewBadges godoc
// @Summary Check for new badges
// @Description Check if user qualifies for any new badges
// @Tags badges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []dto.BadgeResponse
// @Router /api/v1/badges/check [post]
func (h *BadgeHandler) CheckNewBadges(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	newBadges, err := h.badgeService.CheckAndAwardBadges(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal memeriksa badge baru"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(newBadges, "Badge berhasil diperiksa"))
}

// GetDisplayBadges godoc
// @Summary Get display badges
// @Description Get badges for profile display (limited)
// @Tags badges
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of badges (default 5)"
// @Success 200 {object} []dto.BadgeResponse
// @Router /api/v1/badges/display [get]
func (h *BadgeHandler) GetDisplayBadges(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if limit > 10 {
		limit = 10
	}

	badges, err := h.badgeService.GetDisplayBadges(ctx, userID.(uint), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil badge display"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(badges, "Badge display berhasil diambil"))
}
