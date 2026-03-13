package handler

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// CreateAppeal godoc
// @Summary Submit an appeal
// @Description Submit an appeal for suspension or ban
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateAppealRequest true "Appeal data"
// @Success 201 {object} dto.Response
// @Router /appeals [post]
func (h *ModerationHandler) CreateAppeal(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	appeal, err := h.moderationService.CreateAppeal(ctx, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(appeal, "Appeal submitted successfully"))
}

// GetAppeals godoc
// @Summary Get appeals
// @Description Get list of appeals for moderation
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status: pending, approved, rejected"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Router /moderation/appeals [get]
func (h *ModerationHandler) GetAppeals(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	appeals, total, err := h.moderationService.GetAppeals(ctx, status, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get appeals"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(appeals, page, limit, total))
}

// ReviewAppeal godoc
// @Summary Review an appeal
// @Description Approve or reject an appeal
// @Tags Moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Appeal ID"
// @Param request body dto.ReviewAppealRequest true "Review data"
// @Success 200 {object} dto.Response
// @Router /moderation/appeals/{id} [put]
func (h *ModerationHandler) ReviewAppeal(c *gin.Context) {
	ctx := c.Request.Context()
	appealID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid appeal ID"))
		return
	}

	moderatorID, _ := middleware.GetUserID(c)

	var req dto.ReviewAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.moderationService.ReviewAppeal(ctx, uint(appealID), moderatorID, &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Appeal reviewed successfully"))
}
