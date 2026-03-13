package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/application")

type ProgressMapHandler struct {
	mapService *application.ProgressMapService
}

func NewProgressMapHandler(mapService *application.ProgressMapService) *ProgressMapHandler {
	return &ProgressMapHandler{
		mapService: mapService,
	}
}

// GetFullMap godoc
// @Summary Get full progress map
// @Description Get the complete map with all regions and user progress
// @Tags progress-map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.FullMapResponse
// @Router /api/v1/map [get]
func (h *ProgressMapHandler) GetFullMap(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	fullMap, err := h.mapService.GetFullMap(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil peta progress"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(fullMap, "Peta progress berhasil diambil"))
}

// GetRegionDetail godoc
// @Summary Get region detail
// @Description Get detailed information about a specific map region
// @Tags progress-map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key path string true "Region key"
// @Success 200 {object} dto.MapRegionResponse
// @Router /api/v1/map/regions/{key} [get]
func (h *ProgressMapHandler) GetRegionDetail(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	regionKey := c.Param("key")

	region, err := h.mapService.GetRegionDetail(ctx, regionKey, userID.(uint))
	if err != nil {
		if err == application.ErrRegionNotFound {
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil detail region"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(region, "Detail region berhasil diambil"))
}

// GetProgressSummary godoc
// @Summary Get map progress summary
// @Description Get a brief summary of the user's map progress
// @Tags progress-map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MapProgressSummary
// @Router /api/v1/map/summary [get]
func (h *ProgressMapHandler) GetProgressSummary(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	summary, err := h.mapService.GetProgressSummary(ctx, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil ringkasan progress"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(summary, "Ringkasan progress berhasil diambil"))
}

// ClaimLandmarkReward godoc
// @Summary Claim landmark reward
// @Description Claim XP and coin rewards for an unlocked landmark
// @Tags progress-map
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Landmark UUID"
// @Success 200 {object} dto.Response
// @Router /api/v1/map/landmarks/{id}/claim [post]
func (h *ProgressMapHandler) ClaimLandmarkReward(c *gin.Context) {
	ctx := c.Request.Context()
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Unauthorized"))
		return
	}

	landmarkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID landmark tidak valid"))
		return
	}

	if err := h.mapService.ClaimLandmarkReward(ctx, userID.(uint), landmarkID); err != nil {
		switch err {
		case application.ErrLandmarkNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		case application.ErrLandmarkNotUnlocked:
			c.JSON(http.StatusBadRequest, dto.ErrorResponseWithCode(dto.ErrCodeBadRequest, err.Error()))
		case application.ErrRewardAlreadyClaimed:
			c.JSON(http.StatusConflict, dto.ErrorResponseWithCode(dto.ErrCodeConflict, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengklaim reward"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Reward landmark berhasil diklaim"))
}
