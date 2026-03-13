package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"

	"github.com/Alfian57/ruang-tenang-api/internal/features/weekly_league/application")

type WeeklyLeagueHandler struct {
	leagueService *application.WeeklyLeagueService
}

func NewWeeklyLeagueHandler(leagueService *application.WeeklyLeagueService) *WeeklyLeagueHandler {
	return &WeeklyLeagueHandler{leagueService: leagueService}
}

// @Summary Get league overview
// @Description Get current weekly league overview including leaderboard and user position
// @Tags weekly-leagues
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.LeagueOverviewResponse
// @Router /api/v1/leagues/overview [get]
func (h *WeeklyLeagueHandler) GetOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")

	overview, err := h.leagueService.GetOverview(c.Request.Context(), userID.(uint))
	if err != nil {
		switch err {
		case application.ErrLeagueSeasonNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(overview, "Berhasil mengambil data liga"))
}

// @Summary Get all league divisions
// @Description Get list of all league divisions/tiers
// @Tags weekly-leagues
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.LeagueDivisionResponse
// @Router /api/v1/leagues/divisions [get]
func (h *WeeklyLeagueHandler) GetDivisions(c *gin.Context) {
	divisions, err := h.leagueService.GetDivisions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(divisions, "Berhasil mengambil data divisi"))
}
