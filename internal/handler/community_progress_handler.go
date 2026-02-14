package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type CommunityProgressHandler struct {
	communityService *service.CommunityProgressService
}

func NewCommunityProgressHandler(communityService *service.CommunityProgressService) *CommunityProgressHandler {
	return &CommunityProgressHandler{
		communityService: communityService,
	}
}

// GetCommunityStats godoc
// @Summary Get community statistics
// @Description Get aggregated community progress statistics
// @Tags community
// @Accept json
// @Produce json
// @Success 200 {object} dto.CommunityStatsResponse
// @Router /api/v1/community/stats [get]
func (h *CommunityProgressHandler) GetCommunityStats(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := h.communityService.GetCommunityStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil statistik komunitas"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stats, "Statistik komunitas berhasil diambil"))
}

// GetLevelHallOfFame godoc
// @Summary Get level hall of fame
// @Description Get top users within a specific level tier
// @Tags community
// @Accept json
// @Produce json
// @Param level path int true "Level number"
// @Param limit query int false "Limit results (default 10)"
// @Success 200 {object} dto.LevelHallOfFameResponse
// @Router /api/v1/community/hall-of-fame/level/{level} [get]
func (h *CommunityProgressHandler) GetLevelHallOfFame(c *gin.Context) {
	ctx := c.Request.Context()
	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 10 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level tidak valid (harus 1-10)"))
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit > 50 {
		limit = 50
	}

	hallOfFame, err := h.communityService.GetLevelHallOfFame(ctx, level, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data hall of fame"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(hallOfFame, "Hall of fame berhasil diambil"))
}

// GetMonthlyHallOfFame godoc
// @Summary Get monthly hall of fame
// @Description Get monthly hall of fame by category
// @Tags community
// @Accept json
// @Produce json
// @Param month query int true "Month (1-12)"
// @Param year query int true "Year"
// @Param category query string false "Category filter"
// @Success 200 {object} []dto.HallOfFameEntry
// @Router /api/v1/community/hall-of-fame/monthly [get]
func (h *CommunityProgressHandler) GetMonthlyHallOfFame(c *gin.Context) {
	ctx := c.Request.Context()
	month, _ := strconv.Atoi(c.Query("month"))
	year, _ := strconv.Atoi(c.Query("year"))
	category := c.Query("category")

	if month < 1 || month > 12 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Bulan tidak valid (harus 1-12)"))
		return
	}

	if year < 2020 || year > 2100 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Tahun tidak valid"))
		return
	}

	entries, err := h.communityService.GetMonthlyHallOfFame(ctx, month, year, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data hall of fame bulanan"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(entries, "Hall of fame bulanan berhasil diambil"))
}

// GetHallOfFameCategories godoc
// @Summary Get hall of fame categories
// @Description Get available hall of fame categories
// @Tags community
// @Accept json
// @Produce json
// @Success 200 {object} []string
// @Router /api/v1/community/hall-of-fame/categories [get]
func (h *CommunityProgressHandler) GetHallOfFameCategories(c *gin.Context) {
	ctx := c.Request.Context()
	categories := h.communityService.GetAvailableHallOfFameCategories(ctx)

	c.JSON(http.StatusOK, dto.SuccessResponse(categories, "Kategori hall of fame berhasil diambil"))
}

// GetPersonalJourney godoc
// @Summary Get personal journey
// @Description Get user's personal progress journey
// @Tags community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.PersonalJourneyResponse
// @Router /api/v1/community/my-journey [get]
func (h *CommunityProgressHandler) GetPersonalJourney(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	journey, err := h.communityService.GetPersonalJourney(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data perjalanan personal"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(journey, "Perjalanan personal berhasil diambil"))
}

// GetWeeklyProgress godoc
// @Summary Get weekly progress
// @Description Get user's weekly progress breakdown
// @Tags community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.WeeklyProgressResponse
// @Router /api/v1/community/my-progress/weekly [get]
func (h *CommunityProgressHandler) GetWeeklyProgress(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	progress, err := h.communityService.GetWeeklyProgress(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data progress mingguan"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(progress, "Progress mingguan berhasil diambil"))
}

// GetMonthlyProgress godoc
// @Summary Get monthly progress
// @Description Get user's monthly progress breakdown
// @Tags community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MonthlyProgressResponse
// @Router /api/v1/community/my-progress/monthly [get]
func (h *CommunityProgressHandler) GetMonthlyProgress(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	progress, err := h.communityService.GetMonthlyProgress(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data progress bulanan"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(progress, "Progress bulanan berhasil diambil"))
}

// GetAllTimeStats godoc
// @Summary Get all-time stats
// @Description Get user's all-time statistics
// @Tags community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.AllTimeStatsResponse
// @Router /api/v1/community/my-stats [get]
func (h *CommunityProgressHandler) GetAllTimeStats(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	stats, err := h.communityService.GetAllTimeStats(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data statistik"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stats, "Statistik berhasil diambil"))
}

// GetLevelUpCelebration godoc
// @Summary Get level up celebration
// @Description Get celebration data for a level up event
// @Tags community
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param level path int true "New level"
// @Success 200 {object} dto.LevelUpCelebrationResponse
// @Router /api/v1/community/celebrate/{level} [get]
func (h *CommunityProgressHandler) GetLevelUpCelebration(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	levelStr := c.Param("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 1 || level > 10 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Level tidak valid"))
		return
	}

	celebration, err := h.communityService.GetLevelUpCelebration(ctx, userID.(uint), level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data perayaan"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(celebration, "Data perayaan berhasil diambil"))
}
