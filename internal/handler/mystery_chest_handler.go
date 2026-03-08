package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MysteryChestHandler struct {
	chestService *service.MysteryChestService
}

func NewMysteryChestHandler(chestService *service.MysteryChestService) *MysteryChestHandler {
	return &MysteryChestHandler{chestService: chestService}
}

// @Summary Get my chests
// @Description Get paginated list of user's chests
// @Tags mystery-chests
// @Produce json
// @Security BearerAuth
// @Param is_opened query bool false "Filter by opened status"
// @Param rarity query string false "Filter by rarity (common, rare, epic, legendary)"
// @Param page query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 10)"
// @Success 200 {object} dto.PaginatedResponse
// @Router /api/v1/chests [get]
func (h *MysteryChestHandler) GetMyChests(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var filter dto.ChestFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Parameter tidak valid"))
		return
	}

	chests, total, err := h.chestService.GetMyChests(c.Request.Context(), userID.(uint), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(chests, filter.Page, filter.Limit, total))
}

// @Summary Open a chest
// @Description Open a mystery chest and reveal the reward
// @Tags mystery-chests
// @Produce json
// @Security BearerAuth
// @Param id path string true "Chest UUID"
// @Success 200 {object} dto.OpenChestResponse
// @Router /api/v1/chests/{id}/open [post]
func (h *MysteryChestHandler) OpenChest(c *gin.Context) {
	userID, _ := c.Get("user_id")

	chestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID chest tidak valid"))
		return
	}

	result, err := h.chestService.OpenChest(c.Request.Context(), userID.(uint), chestID)
	if err != nil {
		switch err {
		case service.ErrChestNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case service.ErrChestAlreadyOpen:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		case service.ErrNotChestOwner:
			c.JSON(http.StatusForbidden, dto.ErrorResponseWithCode(dto.ErrCodeForbidden, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Chest berhasil dibuka!"))
}
