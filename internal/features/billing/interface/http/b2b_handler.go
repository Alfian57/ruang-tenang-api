package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	billingapp "github.com/Alfian57/ruang-tenang-api/internal/features/billing/application"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

type B2BHandler struct {
	service *billingapp.B2BService
}

func NewB2BHandler(service *billingapp.B2BService) *B2BHandler {
	return &B2BHandler{service: service}
}

func (h *B2BHandler) requireUserID(c *gin.Context) (uint, bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{Success: false, Message: "Unauthorized"})
		return 0, false
	}
	return userID, true
}

func (h *B2BHandler) requireMitraRole(c *gin.Context) bool {
	role, ok := middleware.GetUserRole(c)
	if !ok || strings.ToLower(strings.TrimSpace(role)) != "mitra" {
		c.JSON(http.StatusForbidden, dto.ErrorResponse("Mitra access required"))
		return false
	}

	return true
}

func (h *B2BHandler) requireMitraUserID(c *gin.Context) (uint, bool) {
	if !h.requireMitraRole(c) {
		return 0, false
	}

	return h.requireUserID(c)
}

func (h *B2BHandler) parseUintPathParam(c *gin.Context, key string) (uint, bool) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(c.Param(key)), 10, 32)
	if err != nil || parsed == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid path parameter: "+key))
		return 0, false
	}
	return uint(parsed), true
}

func (h *B2BHandler) writeServiceError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, billingapp.ErrB2BForbidden):
		c.JSON(http.StatusForbidden, dto.ErrorResponse("forbidden"))
	case errors.Is(err, billingapp.ErrB2BMitraRoleRequired):
		c.JSON(http.StatusForbidden, dto.ErrorResponse(err.Error()))
	case errors.Is(err, billingapp.ErrB2BOrganizationNotFound),
		errors.Is(err, billingapp.ErrB2BPlanNotFound),
		errors.Is(err, billingapp.ErrB2BMemberNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse(err.Error()))
	case errors.Is(err, billingapp.ErrB2BInsufficientSeats),
		errors.Is(err, billingapp.ErrB2BMemberAlreadyActive),
		errors.Is(err, billingapp.ErrB2BMemberNotPendingApproval),
		errors.Is(err, billingapp.ErrB2BContractedSeatsTooLow),
		errors.Is(err, billingapp.ErrB2BCannotRemoveOwnerMember),
		errors.Is(err, billingapp.ErrB2BSubscriptionNotFound):
		c.JSON(http.StatusConflict, dto.ErrorResponse(err.Error()))
	case errors.Is(err, billingapp.ErrB2BBlockedByPersonalPremium):
		c.JSON(http.StatusConflict, dto.ErrorResponse("Kamu masih memiliki Premium pribadi yang aktif. Batalkan/biarkan masa berlakunya selesai sebelum bergabung sebagai anggota Premium B2B."))
	case errors.Is(err, billingapp.ErrB2BSubscriptionPaymentRequired):
		c.JSON(http.StatusPaymentRequired, dto.ErrorResponse("Langganan B2B belum memiliki pembayaran aktif untuk periode ini"))
	case errors.Is(err, billingapp.ErrB2BInvalidInviteToken),
		errors.Is(err, billingapp.ErrB2BInvitationExpired),
		errors.Is(err, billingapp.ErrB2BEmailMismatch),
		errors.Is(err, billingapp.ErrB2BInvalidBillingCycle),
		errors.Is(err, billingapp.ErrB2BInvalidSeatCount),
		errors.Is(err, billingapp.ErrB2BInvalidMemberStatus),
		errors.Is(err, billingapp.ErrB2BInvalidSSOProvider),
		errors.Is(err, billingapp.ErrB2BSSOEnforcementUnavailable):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse(fallback))
	}
}

func (h *B2BHandler) CreateOrganization(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	var req dto.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.CreateOrganization(ctx, userID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to create organization")
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(result, "Organization created"))
}

func (h *B2BHandler) ListPlans(c *gin.Context) {
	ctx := c.Request.Context()
	if !h.requireMitraRole(c) {
		return
	}

	activeOnly := true
	if raw := strings.TrimSpace(c.Query("active_only")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid active_only query"))
			return
		}
		activeOnly = parsed
	}

	result, err := h.service.ListPlans(ctx, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("failed to load b2b plans"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) ListOrganizations(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	result, err := h.service.ListOrganizations(ctx, userID)
	if err != nil {
		h.writeServiceError(c, err, "failed to list organizations")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) GetOrganizationSummary(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	result, err := h.service.GetOrganizationSummary(ctx, userID, organizationID)
	if err != nil {
		h.writeServiceError(c, err, "failed to load organization summary")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) ListOrganizationMembers(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	status := strings.TrimSpace(strings.ToLower(c.Query("status")))
	result, err := h.service.ListOrganizationMembers(ctx, userID, organizationID, status)
	if err != nil {
		h.writeServiceError(c, err, "failed to list organization members")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) InviteMember(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.InviteOrganizationMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.InviteMember(ctx, userID, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to invite member")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Invitation created"))
}

func (h *B2BHandler) BulkInviteMembers(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.BulkInviteOrganizationMembersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.BulkInviteMembers(ctx, userID, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to process bulk invitations")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Bulk invitation processed"))
}

func (h *B2BHandler) AcceptInvite(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.AcceptOrganizationInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.AcceptInvite(ctx, userID, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to accept invitation")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Invitation accepted"))
}

func (h *B2BHandler) GetInvitePreview(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	invitationToken := strings.TrimSpace(c.Param("token"))
	if invitationToken == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid invitation token"))
		return
	}

	result, err := h.service.GetInvitePreview(ctx, userID, invitationToken)
	if err != nil {
		h.writeServiceError(c, err, "failed to load invitation")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) AcceptInviteByToken(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	var req dto.AcceptOrganizationInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.AcceptInviteByToken(ctx, userID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to accept invitation")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Invitation accepted"))
}

func (h *B2BHandler) RemoveMember(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}
	memberID, ok := h.parseUintPathParam(c, "member_id")
	if !ok {
		return
	}

	result, err := h.service.RemoveMember(ctx, userID, organizationID, memberID)
	if err != nil {
		h.writeServiceError(c, err, "failed to remove member")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Member removed"))
}

func (h *B2BHandler) CreateQuote(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	var req dto.CreateB2BQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.CreateQuote(ctx, userID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to create quote")
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(result, "Quote created"))
}

func (h *B2BHandler) ApproveMember(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}
	memberID, ok := h.parseUintPathParam(c, "member_id")
	if !ok {
		return
	}

	var req dto.MemberApprovalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.ApproveMember(ctx, userID, organizationID, memberID, req.Note)
	if err != nil {
		h.writeServiceError(c, err, "failed to approve member")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Member approved"))
}

func (h *B2BHandler) RejectMember(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}
	memberID, ok := h.parseUintPathParam(c, "member_id")
	if !ok {
		return
	}

	var req dto.MemberApprovalDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.RejectMember(ctx, userID, organizationID, memberID, req.Note)
	if err != nil {
		h.writeServiceError(c, err, "failed to reject member")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Member rejected"))
}

func (h *B2BHandler) GetOrganizationAnalytics(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	days := 30
	if raw := strings.TrimSpace(c.Query("days")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid days query"))
			return
		}
		days = parsed
	}

	result, err := h.service.GetOrganizationAnalytics(ctx, userID, organizationID, days)
	if err != nil {
		h.writeServiceError(c, err, "failed to load analytics")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) GetImpactReport(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	days := 30
	if raw := strings.TrimSpace(c.Query("days")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid days query"))
			return
		}
		days = parsed
	}

	result, err := h.service.GetImpactReport(ctx, userID, organizationID, days)
	if err != nil {
		h.writeServiceError(c, err, "failed to load impact report")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) ListOrganizationAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	page := 1
	if raw := strings.TrimSpace(c.Query("page")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid page query"))
			return
		}
		page = parsed
	}

	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid limit query"))
			return
		}
		limit = parsed
	}

	action := strings.TrimSpace(strings.ToLower(c.Query("action")))
	result, err := h.service.ListAuditLogs(ctx, userID, organizationID, action, page, limit)
	if err != nil {
		h.writeServiceError(c, err, "failed to load audit logs")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) GetOnboardingTemplate(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	role := strings.TrimSpace(strings.ToLower(c.Query("role")))
	result, err := h.service.GetOnboardingTemplate(ctx, userID, organizationID, role)
	if err != nil {
		h.writeServiceError(c, err, "failed to load onboarding template")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) UpsertOnboardingTemplate(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.UpsertOrganizationOnboardingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.UpsertOnboardingTemplate(ctx, userID, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to save onboarding template")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Onboarding template saved"))
}

func (h *B2BHandler) SeatUpgradeSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.SeatUpgradeSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.SelfServiceSeatUpgrade(ctx, userID, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to upgrade seats")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Seat upgrade applied"))
}

func (h *B2BHandler) RunOrganizationReminders(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	result, err := h.service.RunOrganizationReminders(ctx, userID, organizationID)
	if err != nil {
		h.writeServiceError(c, err, "failed to run reminders")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Reminders processed"))
}

func (h *B2BHandler) GetSSOConfig(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	result, err := h.service.GetSSOConfig(ctx, userID, organizationID)
	if err != nil {
		h.writeServiceError(c, err, "failed to load sso config")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) UpsertSSOConfig(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.UpsertOrganizationSSOConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.UpsertSSOConfig(ctx, userID, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to update sso config")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "SSO config updated"))
}

func (h *B2BHandler) GetPricingRecommendation(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireMitraUserID(c)
	if !ok {
		return
	}

	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	result, err := h.service.GetPricingRecommendation(ctx, userID, organizationID)
	if err != nil {
		h.writeServiceError(c, err, "failed to generate pricing recommendation")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) AdminCreatePlan(c *gin.Context) {
	ctx := c.Request.Context()

	var req dto.UpsertB2BPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.AdminCreatePlan(ctx, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to create b2b plan")
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(result, "B2B plan created"))
}

func (h *B2BHandler) AdminListPlans(c *gin.Context) {
	ctx := c.Request.Context()

	activeOnly := false
	if raw := strings.TrimSpace(c.Query("active_only")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid active_only query"))
			return
		}
		activeOnly = parsed
	}

	result, err := h.service.ListPlans(ctx, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("failed to load b2b plans"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) AdminListOrganizations(c *gin.Context) {
	ctx := c.Request.Context()

	result, err := h.service.AdminListOrganizations(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("failed to load b2b organizations"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *B2BHandler) AdminUpdatePlan(c *gin.Context) {
	ctx := c.Request.Context()
	planID, ok := h.parseUintPathParam(c, "plan_id")
	if !ok {
		return
	}

	var req dto.UpsertB2BPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.AdminUpdatePlan(ctx, planID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to update b2b plan")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "B2B plan updated"))
}

func (h *B2BHandler) AdminCreateSubscription(c *gin.Context) {
	ctx := c.Request.Context()
	organizationID, ok := h.parseUintPathParam(c, "organization_id")
	if !ok {
		return
	}

	var req dto.CreateB2BSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.AdminCreateSubscription(ctx, organizationID, &req)
	if err != nil {
		h.writeServiceError(c, err, "failed to upsert b2b subscription")
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "B2B subscription saved"))
}
