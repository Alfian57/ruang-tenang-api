package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type DailySpinHandler struct {
	spinService *service.DailySpinService
}

func NewDailySpinHandler(spinService *service.DailySpinService) *DailySpinHandler {
	return &DailySpinHandler{spinService: spinService}
}

// @Summary Get spin wheel
// @Description Get the daily roulette wheel with all reward slots and spin status
// @Tags daily-spin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SpinWheelResponse
// @Router /api/v1/daily-spin/wheel [get]
func (h *DailySpinHandler) GetWheel(c *gin.Context) {
	userID, _ := c.Get("user_id")

	wheel, err := h.spinService.GetWheel(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(wheel, "Data roulette"))
}

// @Summary Spin the wheel
// @Description Perform the daily spin to win a reward
// @Tags daily-spin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SpinResultResponse
// @Router /api/v1/daily-spin/spin [post]
func (h *DailySpinHandler) Spin(c *gin.Context) {
	userID, _ := c.Get("user_id")

	result, err := h.spinService.Spin(c.Request.Context(), userID.(uint))
	if err != nil {
		switch err {
		case service.ErrAlreadySpunToday:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case service.ErrNoSpinRewards:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Selamat! Kamu mendapatkan hadiah!"))
}
