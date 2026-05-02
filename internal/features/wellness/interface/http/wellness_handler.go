package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	wellnessapp "github.com/Alfian57/ruang-tenang-api/internal/features/wellness/application"
	wellnessinfra "github.com/Alfian57/ruang-tenang-api/internal/features/wellness/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WellnessHandler struct {
	service *wellnessapp.WellnessService
}

func NewWellnessHandler(service *wellnessapp.WellnessService) *WellnessHandler {
	return &WellnessHandler{service: service}
}

func (h *WellnessHandler) requireUserID(c *gin.Context) (uint, bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return 0, false
	}
	return userID, true
}

func (h *WellnessHandler) writeError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, wellnessinfra.ErrWellnessProfileNotFound),
		errors.Is(err, wellnessinfra.ErrWellnessPlanNotFound),
		errors.Is(err, wellnessinfra.ErrWellnessItemNotFound),
		errors.Is(err, wellnessinfra.ErrWeeklyInsightNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse(err.Error()))
	case errors.Is(err, wellnessapp.ErrWellnessOnboardingRequired),
		errors.Is(err, wellnessapp.ErrUnsupportedNeedCondition):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(fallback))
	}
}

func (h *WellnessHandler) GetOnboarding(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetOnboarding(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err, "failed to load wellness onboarding")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *WellnessHandler) CompleteOnboarding(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	var req dto.WellnessOnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}
	result, err := h.service.CompleteOnboarding(c.Request.Context(), userID, &req)
	if err != nil {
		h.writeError(c, err, "failed to save wellness onboarding")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Wellness onboarding saved"))
}

func (h *WellnessHandler) GetCurrentPlan(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetCurrentPlan(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err, "failed to load wellness plan")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *WellnessHandler) CompletePlanItem(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	itemID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid plan item id"))
		return
	}
	result, err := h.service.CompletePlanItem(c.Request.Context(), userID, itemID)
	if err != nil {
		h.writeError(c, err, "failed to complete plan item")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Plan item completed"))
}

func (h *WellnessHandler) NeedNow(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	var req dto.WellnessNeedNowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}
	result, err := h.service.NeedNow(c.Request.Context(), userID, req.Condition)
	if err != nil {
		h.writeError(c, err, "failed to create wellness recommendation")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *WellnessHandler) GetWeeklyInsight(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetWeeklyInsight(c.Request.Context(), userID, c.Query("week_start"))
	if err != nil {
		h.writeError(c, err, "failed to load weekly insight")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *WellnessHandler) CompleteTour(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	result, err := h.service.CompleteTour(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err, "failed to complete tour")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Tour completed"))
}

func (h *WellnessHandler) GetJourneyMap(c *gin.Context) {
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}
	result, err := h.service.GetJourneyMap(c.Request.Context(), userID)
	if err != nil {
		h.writeError(c, err, "failed to load journey map")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}
