package handlers

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type MoodHandler struct {
	moodService *services.MoodService
}

func NewMoodHandler(moodService *services.MoodService) *MoodHandler {
	return &MoodHandler{moodService: moodService}
}

// RecordMood godoc
// @Summary Record user mood
// @Description Record a new mood entry for the user
// @Tags Mood
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateMoodRequest true "Mood data"
// @Success 201 {object} dto.UserMoodDTO
// @Router /user-moods [post]
func (h *MoodHandler) RecordMood(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateMoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	mood, err := h.moodService.RecordMood(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to record mood"))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(mood, "Mood recorded"))
}

// GetMoodHistory godoc
// @Summary Get mood history
// @Description Get user's mood history with optional date filtering
// @Tags Mood
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(30)
// @Success 200 {object} dto.MoodHistoryDTO
// @Router /user-moods [get]
func (h *MoodHandler) GetMoodHistory(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var params dto.MoodQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 30
	}

	history, err := h.moodService.GetMoodHistory(userID, &params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get mood history"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(history, ""))
}

// GetLatestMood godoc
// @Summary Get latest mood
// @Description Get user's most recent mood entry
// @Tags Mood
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.UserMoodDTO
// @Failure 404 {object} dto.Response
// @Router /user-moods/latest [get]
func (h *MoodHandler) GetLatestMood(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	mood, err := h.moodService.GetLatestMood(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse("No mood recorded yet"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(mood, ""))
}

// GetMoodStats godoc
// @Summary Get mood statistics
// @Description Get mood statistics for the last N days
// @Tags Mood
// @Produce json
// @Security BearerAuth
// @Param days query int false "Number of days" default(30)
// @Success 200 {object} dto.Response
// @Router /user-moods/stats [get]
func (h *MoodHandler) GetMoodStats(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	days := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := c.GetQuery("days"); err {
			_ = parsed
		}
	}

	stats, err := h.moodService.GetMoodStats(userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get statistics"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(stats, ""))
}
