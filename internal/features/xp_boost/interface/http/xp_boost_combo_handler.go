package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/xp_boost/application")

type XPBoostComboHandler struct {
	boostService *application.XPBoostComboService
}

func NewXPBoostComboHandler(boostService *application.XPBoostComboService) *XPBoostComboHandler {
	return &XPBoostComboHandler{boostService: boostService}
}

// @Summary Get active XP boost
// @Description Get current active XP boost status
// @Tags xp-boost-combo
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.XPBoostResponse
// @Router /api/v1/xp-boost/active [get]
func (h *XPBoostComboHandler) GetActiveBoost(c *gin.Context) {
	userID, _ := c.Get("user_id")

	boost, err := h.boostService.GetActiveBoost(c.Request.Context(), userID.(uint))
	if err != nil {
		switch err {
		case application.ErrNoActiveBoost:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(boost, "XP Boost aktif"))
}

// @Summary Get combo status
// @Description Get current combo chain status
// @Tags xp-boost-combo
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.ComboStatusResponse
// @Router /api/v1/combo/status [get]
func (h *XPBoostComboHandler) GetComboStatus(c *gin.Context) {
	userID, _ := c.Get("user_id")

	status, err := h.boostService.GetComboStatus(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(status, "Status combo"))
}

// @Summary Get effective multiplier
// @Description Get combined XP multiplier from boost + combo
// @Tags xp-boost-combo
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/xp-boost/multiplier [get]
func (h *XPBoostComboHandler) GetEffectiveMultiplier(c *gin.Context) {
	userID, _ := c.Get("user_id")

	multiplier := h.boostService.GetEffectiveMultiplier(c.Request.Context(), userID.(uint))

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"effective_multiplier": multiplier,
	}, "Multiplier efektif"))
}
