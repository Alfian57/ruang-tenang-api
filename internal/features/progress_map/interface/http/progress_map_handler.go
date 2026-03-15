package handler

import (
	"net/http"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Alfian57/ruang-tenang-api/internal/features/progress_map/application"
)

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

// AdminGetAllLandmarks godoc
// @Summary Get all map landmarks (Admin)
// @Description Get all map landmarks for admin management
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /api/v1/admin/map-landmarks [get]
func (h *ProgressMapHandler) AdminGetAllLandmarks(c *gin.Context) {
	ctx := c.Request.Context()

	landmarks, err := h.mapService.AdminGetAllLandmarks(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal mengambil data landmark"))
		return
	}

	responses := make([]dto.AdminMapLandmarkResponse, len(landmarks))
	for i, item := range landmarks {
		responses[i] = dto.AdminMapLandmarkResponse{
			ID:             item.ID,
			RegionID:       item.RegionID,
			RegionName:     item.Region.Name,
			RegionKey:      item.Region.RegionKey,
			LandmarkKey:    item.LandmarkKey,
			Name:           item.Name,
			Description:    item.Description,
			Icon:           item.Icon,
			UnlockType:     string(item.UnlockType),
			UnlockActivity: item.UnlockActivity,
			UnlockValue:    item.UnlockValue,
			PositionX:      item.PositionX,
			PositionY:      item.PositionY,
			XPReward:       item.XPReward,
			CoinReward:     item.CoinReward,
			DisplayOrder:   item.DisplayOrder,
			IsActive:       item.IsActive,
			CreatedAt:      item.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(responses, "Data landmark berhasil diambil"))
}

// AdminCreateLandmark godoc
// @Summary Create map landmark (Admin)
// @Description Create a map landmark unlock criteria and reward config
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AdminCreateMapLandmarkRequest true "Create landmark payload"
// @Success 201 {object} dto.Response
// @Router /api/v1/admin/map-landmarks [post]
func (h *ProgressMapHandler) AdminCreateLandmark(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.AdminCreateMapLandmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Payload tidak valid"))
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	landmark := &model.MapLandmark{
		RegionID:       req.RegionID,
		LandmarkKey:    req.LandmarkKey,
		Name:           req.Name,
		Description:    req.Description,
		Icon:           req.Icon,
		UnlockType:     model.MapUnlockType(req.UnlockType),
		UnlockActivity: req.UnlockActivity,
		UnlockValue:    req.UnlockValue,
		PositionX:      req.PositionX,
		PositionY:      req.PositionY,
		XPReward:       req.XPReward,
		CoinReward:     req.CoinReward,
		DisplayOrder:   req.DisplayOrder,
		IsActive:       isActive,
	}

	if err := h.mapService.AdminCreateLandmark(ctx, landmark); err != nil {
		switch err {
		case application.ErrInvalidUnlockType:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		case application.ErrRegionNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal membuat landmark"))
		}
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(nil, "Landmark berhasil dibuat"))
}

// AdminUpdateLandmark godoc
// @Summary Update map landmark (Admin)
// @Description Update map landmark unlock criteria and reward config
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Landmark UUID"
// @Param request body dto.AdminUpdateMapLandmarkRequest true "Update landmark payload"
// @Success 200 {object} dto.Response
// @Router /api/v1/admin/map-landmarks/{id} [put]
func (h *ProgressMapHandler) AdminUpdateLandmark(c *gin.Context) {
	ctx := c.Request.Context()

	landmarkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID landmark tidak valid"))
		return
	}

	var req dto.AdminUpdateMapLandmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Payload tidak valid"))
		return
	}

	payload := &model.MapLandmark{
		RegionID:       req.RegionID,
		LandmarkKey:    req.LandmarkKey,
		Name:           req.Name,
		Description:    req.Description,
		Icon:           req.Icon,
		UnlockType:     model.MapUnlockType(req.UnlockType),
		UnlockActivity: req.UnlockActivity,
		UnlockValue:    req.UnlockValue,
		PositionX:      req.PositionX,
		PositionY:      req.PositionY,
		XPReward:       req.XPReward,
		CoinReward:     req.CoinReward,
		DisplayOrder:   req.DisplayOrder,
		IsActive:       req.IsActive,
	}

	if err := h.mapService.AdminUpdateLandmark(ctx, landmarkID, payload); err != nil {
		switch err {
		case application.ErrInvalidUnlockType:
			c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		case application.ErrRegionNotFound, application.ErrLandmarkNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal memperbarui landmark"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Landmark berhasil diperbarui"))
}

// AdminDeleteLandmark godoc
// @Summary Delete map landmark (Admin)
// @Description Soft delete map landmark by setting it inactive
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Landmark UUID"
// @Success 200 {object} dto.Response
// @Router /api/v1/admin/map-landmarks/{id} [delete]
func (h *ProgressMapHandler) AdminDeleteLandmark(c *gin.Context) {
	ctx := c.Request.Context()

	landmarkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("ID landmark tidak valid"))
		return
	}

	if err := h.mapService.AdminDeleteLandmark(ctx, landmarkID); err != nil {
		switch err {
		case application.ErrLandmarkNotFound:
			c.JSON(http.StatusNotFound, dto.ErrorResponseWithCode(dto.ErrCodeNotFound, err.Error()))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Gagal menghapus landmark"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Landmark berhasil dihapus"))
}
