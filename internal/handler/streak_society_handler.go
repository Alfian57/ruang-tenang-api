package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
)

type StreakSocietyHandler struct {
	societyService *service.StreakSocietyService
}

func NewStreakSocietyHandler(societyService *service.StreakSocietyService) *StreakSocietyHandler {
	return &StreakSocietyHandler{societyService: societyService}
}

// @Summary Get streak society overview
// @Description Get all societies and user's current membership status
// @Tags streak-society
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.StreakSocietyOverviewResponse
// @Router /api/v1/streak-society/overview [get]
func (h *StreakSocietyHandler) GetOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")

	overview, err := h.societyService.GetOverview(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(overview, "Data streak society"))
}

// @Summary Join streak society
// @Description Automatically join the highest eligible streak society
// @Tags streak-society
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.StreakSocietyResponse
// @Router /api/v1/streak-society/join [post]
func (h *StreakSocietyHandler) JoinSociety(c *gin.Context) {
	userID, _ := c.Get("user_id")

	society, err := h.societyService.JoinSociety(c.Request.Context(), userID.(uint))
	if err != nil {
		switch err {
		case service.ErrStreakTooLow:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		case service.ErrAlreadySocietyMember:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(society, "Berhasil bergabung ke streak society!"))
}

// @Summary Get society members
// @Description Get members of a specific streak society
// @Tags streak-society
// @Produce json
// @Security BearerAuth
// @Param id path int true "Society ID"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} dto.PaginatedResponse
// @Router /api/v1/streak-society/{id}/members [get]
func (h *StreakSocietyHandler) GetMembers(c *gin.Context) {
	societyID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID society tidak valid"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	members, total, err := h.societyService.GetSocietyMembers(c.Request.Context(), societyID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(members, page, limit, total))
}
