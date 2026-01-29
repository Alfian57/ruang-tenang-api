package handlers

import (
	"net/http"
	"strconv"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/services"
	"github.com/gin-gonic/gin"
)

type ModerationHandler struct {
	moderationService *services.ModerationService
}

func NewModerationHandler(moderationService *services.ModerationService) *ModerationHandler {
	return &ModerationHandler{moderationService: moderationService}
}

// ========================
// Moderation Dashboard
// ========================

// GetModerationStats godoc
// @Summary Get moderation statistics
// @Description Get moderation dashboard statistics
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /moderation/stats [get]
func (h *ModerationHandler) GetModerationStats(c *gin.Context) {
	stats, err := h.moderationService.GetModerationStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get moderation stats"))
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(stats, ""))
}

// GetModerationQueue godoc
// @Summary Get article moderation queue
// @Description Get articles pending moderation
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status: pending, flagged"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Router /moderation/queue [get]
func (h *ModerationHandler) GetModerationQueue(c *gin.Context) {
	var params dto.ModerationQueueParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 20
	}

	items, total, err := h.moderationService.GetModerationQueue(params.Status, params.Page, params.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get moderation queue"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(items, params.Page, params.Limit, total))
}

// ModerateArticle godoc
// @Summary Moderate an article
// @Description Approve, reject, or request edit for an article
// @Tags Moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Article ID"
// @Param request body dto.ModerateArticleRequest true "Moderation action"
// @Success 200 {object} dto.Response
// @Router /moderation/articles/{id} [put]
func (h *ModerationHandler) ModerateArticle(c *gin.Context) {
	articleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid article ID"))
		return
	}

	moderatorID, _ := middleware.GetUserID(c)

	var req dto.ModerateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.moderationService.ModerateArticle(uint(articleID), moderatorID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Article moderated successfully"))
}

// ========================
// Reports Management
// ========================

// GetReports godoc
// @Summary Get reports
// @Description Get user reports for moderation
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status: pending, reviewing, resolved, dismissed"
// @Param report_type query string false "Filter by type: article, forum, forum_post, user"
// @Param reason query string false "Filter by reason"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Router /moderation/reports [get]
func (h *ModerationHandler) GetReports(c *gin.Context) {
	var params dto.ReportQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 20
	}

	reports, total, err := h.moderationService.GetReports(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get reports"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(reports, params.Page, params.Limit, total))
}

// HandleReport godoc
// @Summary Handle a report
// @Description Process a report with action
// @Tags Moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Report ID"
// @Param request body dto.HandleReportRequest true "Handle action"
// @Success 200 {object} dto.Response
// @Router /moderation/reports/{id} [put]
func (h *ModerationHandler) HandleReport(c *gin.Context) {
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid report ID"))
		return
	}

	moderatorID, _ := middleware.GetUserID(c)

	var req dto.HandleReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.moderationService.HandleReport(uint(reportID), moderatorID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Report handled successfully"))
}

// ========================
// User Report Submission
// ========================

// CreateReport godoc
// @Summary Create a report
// @Description Submit a report for content or user
// @Tags Reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateReportRequest true "Report data"
// @Success 201 {object} dto.Response
// @Router /reports [post]
func (h *ModerationHandler) CreateReport(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req dto.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	report, err := h.moderationService.CreateReport(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(report, "Report submitted successfully"))
}

// ========================
// User Blocking
// ========================

// BlockUser godoc
// @Summary Block a user
// @Description Block another user
// @Tags Blocking
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BlockUserRequest true "Block data"
// @Success 201 {object} dto.Response
// @Router /blocks [post]
func (h *ModerationHandler) BlockUser(c *gin.Context) {
	blockerID, _ := middleware.GetUserID(c)

	var req dto.BlockUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.moderationService.BlockUser(blockerID, req.UserID, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(nil, "User blocked successfully"))
}

// UnblockUser godoc
// @Summary Unblock a user
// @Description Remove a block on another user
// @Tags Blocking
// @Produce json
// @Security BearerAuth
// @Param id path int true "Blocked User ID"
// @Success 200 {object} dto.Response
// @Router /blocks/{id} [delete]
func (h *ModerationHandler) UnblockUser(c *gin.Context) {
	blockerID, _ := middleware.GetUserID(c)

	blockedID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid user ID"))
		return
	}

	if err := h.moderationService.UnblockUser(blockerID, uint(blockedID)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "User unblocked successfully"))
}

// GetBlockedUsers godoc
// @Summary Get blocked users
// @Description Get list of users blocked by current user
// @Tags Blocking
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BlockListResponse
// @Router /blocks [get]
func (h *ModerationHandler) GetBlockedUsers(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	blocks, count, err := h.moderationService.GetBlockedUsers(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get blocked users"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(dto.BlockListResponse{
		Blocks:     blocks,
		TotalCount: count,
	}, ""))
}

// ========================
// User Strikes (Moderator)
// ========================

// GetUserStrikes godoc
// @Summary Get user strikes
// @Description Get strikes for a specific user
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param active_only query bool false "Only active strikes"
// @Success 200 {object} dto.Response
// @Router /moderation/users/{id}/strikes [get]
func (h *ModerationHandler) GetUserStrikes(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid user ID"))
		return
	}

	activeOnly := c.Query("active_only") == "true"

	strikes, err := h.moderationService.GetUserStrikes(uint(userID), activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get user strikes"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(strikes, ""))
}

// ========================
// Trigger Warnings
// ========================

// AddTriggerWarnings godoc
// @Summary Add trigger warnings
// @Description Add trigger warning tags to content
// @Tags Moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.TriggerWarningRequest true "Trigger warning data"
// @Success 200 {object} dto.Response
// @Router /moderation/trigger-warnings [post]
func (h *ModerationHandler) AddTriggerWarnings(c *gin.Context) {
	moderatorID, _ := middleware.GetUserID(c)

	var req dto.TriggerWarningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.moderationService.AddTriggerWarnings(req.ContentType, req.ContentID, moderatorID, req.TriggerWarnings); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Trigger warnings added successfully"))
}

// ========================
// AI Disclaimer & Settings
// ========================

// AcceptAIDisclaimer godoc
// @Summary Accept AI disclaimer
// @Description Mark that user has accepted the AI chat disclaimer
// @Tags User Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /user/accept-ai-disclaimer [post]
func (h *ModerationHandler) AcceptAIDisclaimer(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	if err := h.moderationService.AcceptAIDisclaimer(userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to accept disclaimer"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "AI disclaimer accepted"))
}

// UpdateContentWarningPreference godoc
// @Summary Update content warning preference
// @Description Update user's preference for content warnings
// @Tags User Settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.UpdateContentWarningPreferenceRequest true "Preference"
// @Success 200 {object} dto.Response
// @Router /user/content-warning-preference [put]
func (h *ModerationHandler) UpdateContentWarningPreference(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)

	var req dto.UpdateContentWarningPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if err := h.moderationService.UpdateContentWarningPreference(userID, req.Preference); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to update preference"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Content warning preference updated"))
}

// ========================
// Moderator Actions Log
// ========================

// GetModeratorActions godoc
// @Summary Get moderator actions log
// @Description Get audit log of moderator actions
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Param moderator_id query int false "Filter by moderator"
// @Param action_type query string false "Filter by action type"
// @Param target_type query string false "Filter by target type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Router /moderation/actions [get]
func (h *ModerationHandler) GetModeratorActions(c *gin.Context) {
	var params dto.ModeratorActionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 50 {
		params.Limit = 20
	}

	actions, total, err := h.moderationService.GetModeratorActions(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get moderator actions"))
		return
	}

	c.JSON(http.StatusOK, dto.NewPaginatedResponse(actions, params.Page, params.Limit, total))
}

// ========================
// Crisis Keywords Management
// ========================

// GetCrisisKeywords godoc
// @Summary Get crisis keywords
// @Description Get all crisis detection keywords
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /moderation/crisis-keywords [get]
func (h *ModerationHandler) GetCrisisKeywords(c *gin.Context) {
	keywords, err := h.moderationService.GetCrisisKeywords()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to get crisis keywords"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(keywords, ""))
}

// CreateCrisisKeyword godoc
// @Summary Create crisis keyword
// @Description Add a new crisis detection keyword
// @Tags Moderation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCrisisKeywordRequest true "Keyword data"
// @Success 201 {object} dto.Response
// @Router /moderation/crisis-keywords [post]
func (h *ModerationHandler) CreateCrisisKeyword(c *gin.Context) {
	var req dto.CreateCrisisKeywordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	keyword, err := h.moderationService.CreateCrisisKeyword(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(keyword, "Crisis keyword created successfully"))
}

// DeleteCrisisKeyword godoc
// @Summary Delete crisis keyword
// @Description Remove a crisis detection keyword
// @Tags Moderation
// @Produce json
// @Security BearerAuth
// @Param id path int true "Keyword ID"
// @Success 200 {object} dto.Response
// @Router /moderation/crisis-keywords/{id} [delete]
func (h *ModerationHandler) DeleteCrisisKeyword(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid keyword ID"))
		return
	}

	if err := h.moderationService.DeleteCrisisKeyword(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Crisis keyword deleted successfully"))
}
