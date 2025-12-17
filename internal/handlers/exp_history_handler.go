package handlers

import (
	"math"
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ExpHistoryHandler struct {
	expHistoryService  *services.ExpHistoryService
	levelConfigService *services.LevelConfigService
}

func NewExpHistoryHandler(expHistoryService *services.ExpHistoryService, levelConfigService *services.LevelConfigService) *ExpHistoryHandler {
	return &ExpHistoryHandler{
		expHistoryService:  expHistoryService,
		levelConfigService: levelConfigService,
	}
}

// GetHistory godoc
// @Summary Get EXP history
// @Description Get authenticated user's EXP earning history with optional filters
// @Tags EXP
// @Produce json
// @Security BearerAuth
// @Param activity_type query string false "Activity type filter"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.ExpHistoryResponse
// @Failure 401 {object} dto.Response
// @Router /exp-history [get]
func (h *ExpHistoryHandler) GetHistory(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var filter dto.ExpHistoryFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	histories, total, err := h.expHistoryService.GetHistory(userID, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get history"))
		return
	}

	// Convert to DTOs
	historyDTOs := make([]dto.ExpHistoryDTO, len(histories))
	for i, h := range histories {
		historyDTOs[i] = dto.ExpHistoryDTO{
			ID:           h.ID,
			ActivityType: h.ActivityType,
			Points:       h.Points,
			Description:  h.Description,
			CreatedAt:    h.CreatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.ExpHistoryResponse{
		Data:       historyDTOs,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, ""))
}

// GetActivityTypes godoc
// @Summary Get activity types
// @Description Get list of activity types for filtering
// @Tags EXP
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Router /exp-history/activity-types [get]
func (h *ExpHistoryHandler) GetActivityTypes(c *gin.Context) {
	types, err := h.expHistoryService.GetActivityTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get activity types"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(types, ""))
}
