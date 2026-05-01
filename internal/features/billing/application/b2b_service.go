package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/features/billing/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"gorm.io/gorm"
)

var (
	ErrB2BOrganizationNotFound        = errors.New("b2b organization not found")
	ErrB2BPlanNotFound                = errors.New("b2b plan not found")
	ErrB2BSubscriptionNotFound        = errors.New("b2b subscription not found")
	ErrB2BMemberNotFound              = errors.New("b2b member not found")
	ErrB2BMemberNotPendingApproval    = errors.New("member is not awaiting approval")
	ErrB2BForbidden                   = errors.New("forbidden")
	ErrB2BInsufficientSeats           = errors.New("insufficient seats")
	ErrB2BMemberAlreadyActive         = errors.New("member already active")
	ErrB2BInvitationExpired           = errors.New("invitation expired")
	ErrB2BInvalidInviteToken          = errors.New("invalid invitation token")
	ErrB2BEmailMismatch               = errors.New("invitation email mismatch")
	ErrB2BInvalidBillingCycle         = errors.New("invalid billing cycle")
	ErrB2BInvalidSeatCount            = errors.New("invalid seat count for plan")
	ErrB2BInvalidMemberStatus         = errors.New("invalid member status filter")
	ErrB2BInvalidSSOProvider          = errors.New("invalid sso provider")
	ErrB2BSSOEnforcementUnavailable   = errors.New("sso enforcement is not available yet")
	ErrB2BContractedSeatsTooLow       = errors.New("contracted seats cannot be lower than used seats")
	ErrB2BCannotRemoveOwnerMember     = errors.New("cannot remove owner member")
	ErrB2BMitraRoleRequired           = errors.New("mitra role required")
	ErrB2BSubscriptionPaymentRequired = errors.New("b2b subscription payment required")
)

var organizationCodeRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

type B2BNotificationService interface {
	CreateCustomNotification(ctx context.Context, userID uint, notificationType, title, message string, data map[string]string)
}

type B2BService struct {
	repo                *infrastructure.B2BRepository
	notificationService B2BNotificationService
}

func NewB2BService(repo *infrastructure.B2BRepository) *B2BService {
	return &B2BService{repo: repo}
}

func (s *B2BService) SetNotificationService(notificationService B2BNotificationService) {
	s.notificationService = notificationService
}

func toOrganizationDTO(organization *model.Organization) dto.OrganizationDTO {
	if organization == nil {
		return dto.OrganizationDTO{}
	}
	return dto.OrganizationDTO{
		ID:                     organization.ID,
		Code:                   organization.Code,
		Name:                   organization.Name,
		BusinessType:           organization.BusinessType,
		ContactEmail:           organization.ContactEmail,
		Status:                 string(organization.Status),
		RequiresMemberApproval: organization.RequiresMemberApproval,
	}
}

func parseFeaturesJSON(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	features := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &features); err != nil {
		return map[string]any{}
	}
	return features
}

func marshalFeaturesJSON(features map[string]any) string {
	if len(features) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(features)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func toB2BPlanDTO(plan *model.B2BPlan) dto.B2BPlanDTO {
	if plan == nil {
		return dto.B2BPlanDTO{}
	}
	return dto.B2BPlanDTO{
		ID:               plan.ID,
		Code:             plan.Code,
		Name:             plan.Name,
		Description:      plan.Description,
		BillingCycle:     string(plan.BillingCycle),
		BasePricePerSeat: plan.BasePricePerSeat,
		MinSeats:         plan.MinSeats,
		MaxSeats:         plan.MaxSeats,
		Features:         parseFeaturesJSON(plan.FeaturesJSON),
		IsActive:         plan.IsActive,
	}
}

func toB2BSubscriptionDTO(subscription *model.B2BSubscription, plan *model.B2BPlan) *dto.B2BSubscriptionDTO {
	if subscription == nil {
		return nil
	}

	planCode := ""
	planName := ""
	if plan != nil {
		planCode = plan.Code
		planName = plan.Name
	}

	return &dto.B2BSubscriptionDTO{
		ID:              subscription.ID,
		OrganizationID:  subscription.OrganizationID,
		PlanID:          subscription.PlanID,
		PlanCode:        planCode,
		PlanName:        planName,
		Status:          string(subscription.Status),
		ContractedSeats: subscription.ContractedSeats,
		UsedSeats:       subscription.UsedSeats,
		BillingCycle:    string(subscription.BillingCycle),
		UnitPrice:       subscription.UnitPrice,
		Subtotal:        subscription.Subtotal,
		DiscountAmount:  subscription.DiscountAmount,
		TotalAmount:     subscription.TotalAmount,
		StartsAt:        subscription.StartsAt,
		EndsAt:          subscription.EndsAt,
		ActivatedAt:     subscription.ActivatedAt,
	}
}

func toOrganizationMemberDTO(member model.OrganizationMember) dto.OrganizationMemberDTO {
	fullName := strings.TrimSpace(member.FullName)
	if fullName == "" {
		fullName = member.Email
	}

	return dto.OrganizationMemberDTO{
		ID:        member.ID,
		UserID:    member.UserID,
		Email:     member.Email,
		FullName:  fullName,
		Role:      string(member.Role),
		Status:    string(member.Status),
		InvitedAt: member.InvitedAt,
		JoinedAt:  member.JoinedAt,
		RemovedAt: member.RemovedAt,
	}
}

func sanitizeOrganizationCode(nameOrCode string) string {
	trimmed := strings.TrimSpace(nameOrCode)
	if trimmed == "" {
		return ""
	}

	normalized := strings.ToLower(trimmed)
	normalized = organizationCodeRegex.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	if len(normalized) > 60 {
		normalized = normalized[:60]
	}
	return normalized
}

func generateOrganizationCode(name string) string {
	base := sanitizeOrganizationCode(name)
	if base == "" {
		base = "organization"
	}
	return fmt.Sprintf("%s-%d", base, time.Now().Unix()%100000)
}

func normalizeBillingCycle(input string, fallback model.B2BBillingCycle) (model.B2BBillingCycle, error) {
	value := strings.TrimSpace(strings.ToLower(input))
	if value == "" {
		if fallback == model.B2BBillingCycleMonthly || fallback == model.B2BBillingCycleYearly {
			return fallback, nil
		}
		return model.B2BBillingCycleMonthly, nil
	}

	cycle := model.B2BBillingCycle(value)
	if cycle != model.B2BBillingCycleMonthly && cycle != model.B2BBillingCycleYearly {
		return "", ErrB2BInvalidBillingCycle
	}
	return cycle, nil
}

func billingCycleMultiplier(cycle model.B2BBillingCycle) int64 {
	if cycle == model.B2BBillingCycleYearly {
		return 12
	}
	return 1
}

func computeSubscriptionEnd(start time.Time, cycle model.B2BBillingCycle) time.Time {
	if cycle == model.B2BBillingCycleYearly {
		return start.AddDate(1, 0, 0)
	}
	return start.AddDate(0, 1, 0)
}

func calculateVolumeDiscountPercent(seats int) float64 {
	switch {
	case seats >= 500:
		return 0.22
	case seats >= 200:
		return 0.15
	case seats >= 100:
		return 0.10
	case seats >= 50:
		return 0.05
	default:
		return 0
	}
}

func addonPricingRules() map[string]int64 {
	return map[string]int64{
		"dedicated_csm":      1500000,
		"advanced_analytics": 750000,
		"sso":                1000000,
		"custom_branding":    500000,
	}
}

func uniqueStringItems(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	items := make([]string, 0, len(input))
	for _, raw := range input {
		normalized := strings.TrimSpace(strings.ToLower(raw))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	sort.Strings(items)
	return items
}

func (s *B2BService) ensureOrganizationMember(ctx context.Context, requesterID, organizationID uint) (*model.OrganizationMember, error) {
	member, err := s.repo.GetOrganizationMemberByUserID(ctx, organizationID, requesterID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationMemberNotFound) {
			return nil, ErrB2BForbidden
		}
		return nil, err
	}
	if member.Status != model.OrganizationMemberStatusActive {
		return nil, ErrB2BForbidden
	}
	return member, nil
}

func (s *B2BService) ensureOrganizationManager(ctx context.Context, requesterID, organizationID uint) (*model.OrganizationMember, error) {
	member, err := s.ensureOrganizationMember(ctx, requesterID, organizationID)
	if err != nil {
		return nil, err
	}
	if !member.CanManageMembers() {
		return nil, ErrB2BForbidden
	}
	return member, nil
}

func (s *B2BService) ensureMitraRequester(ctx context.Context, requesterID uint) error {
	requester, err := s.repo.FindUserByID(ctx, requesterID)
	if err != nil {
		return err
	}

	if requester.Role != model.RoleMitra {
		return ErrB2BMitraRoleRequired
	}

	return nil
}

func (s *B2BService) CreateOrganization(ctx context.Context, ownerUserID uint, req *dto.CreateOrganizationRequest) (*dto.OrganizationDTO, error) {
	owner, err := s.repo.FindUserByID(ctx, ownerUserID)
	if err != nil {
		return nil, err
	}
	if owner.Role != model.RoleMitra {
		return nil, ErrB2BMitraRoleRequired
	}

	code := sanitizeOrganizationCode(req.Code)
	if code == "" {
		code = generateOrganizationCode(req.Name)
	}

	organization := &model.Organization{
		Code:                   code,
		Name:                   strings.TrimSpace(req.Name),
		BusinessType:           strings.TrimSpace(req.BusinessType),
		ContactEmail:           strings.TrimSpace(strings.ToLower(req.ContactEmail)),
		Status:                 model.OrganizationStatusActive,
		RequiresMemberApproval: true,
		CreatedBy:              &ownerUserID,
	}
	if organization.BusinessType == "" {
		organization.BusinessType = "general"
	}

	now := time.Now()
	err = s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Create(organization).Error; err != nil {
			return err
		}

		ownerMember := &model.OrganizationMember{
			OrganizationID: organization.ID,
			UserID:         &ownerUserID,
			Email:          strings.TrimSpace(strings.ToLower(owner.Email)),
			FullName:       owner.Name,
			Role:           model.OrganizationMemberRoleOwner,
			Status:         model.OrganizationMemberStatusActive,
			InvitedBy:      &ownerUserID,
			InvitedAt:      &now,
			JoinedAt:       &now,
		}
		return tx.Create(ownerMember).Error
	})
	if err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organization.ID, &ownerUserID, "organization.created", "organization", strconv.FormatUint(uint64(organization.ID), 10), map[string]any{
		"name":          organization.Name,
		"contact_email": organization.ContactEmail,
	})

	result := toOrganizationDTO(organization)
	return &result, nil
}

func (s *B2BService) ListPlans(ctx context.Context, activeOnly bool) ([]dto.B2BPlanDTO, error) {
	plans, err := s.repo.ListB2BPlans(ctx, activeOnly)
	if err != nil {
		return nil, err
	}

	result := make([]dto.B2BPlanDTO, 0, len(plans))
	for _, plan := range plans {
		copyPlan := plan
		result = append(result, toB2BPlanDTO(&copyPlan))
	}
	return result, nil
}

func (s *B2BService) AdminListOrganizations(ctx context.Context) ([]dto.OrganizationDTO, error) {
	organizations, err := s.repo.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dto.OrganizationDTO, 0, len(organizations))
	for _, organization := range organizations {
		copyOrganization := organization
		result = append(result, toOrganizationDTO(&copyOrganization))
	}
	return result, nil
}

func (s *B2BService) ListOrganizations(ctx context.Context, requesterID uint) ([]dto.MitraOrganizationListItemDTO, error) {
	if err := s.ensureMitraRequester(ctx, requesterID); err != nil {
		return nil, err
	}

	rows, err := s.repo.ListOrganizationsByUserID(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.MitraOrganizationListItemDTO, 0, len(rows))
	for _, row := range rows {
		organization := dto.OrganizationDTO{
			ID:                     row.OrganizationID,
			Code:                   row.Code,
			Name:                   row.Name,
			BusinessType:           row.BusinessType,
			ContactEmail:           row.ContactEmail,
			Status:                 row.OrganizationStatus,
			RequiresMemberApproval: row.RequiresMemberApproval,
		}

		result = append(result, dto.MitraOrganizationListItemDTO{
			Organization: organization,
			MemberRole:   row.MemberRole,
			MemberStatus: row.MemberStatus,
		})
	}

	return result, nil
}

func (s *B2BService) AdminCreatePlan(ctx context.Context, req *dto.UpsertB2BPlanRequest) (*dto.B2BPlanDTO, error) {
	if req.MinSeats > req.MaxSeats {
		return nil, ErrB2BInvalidSeatCount
	}

	cycle, err := normalizeBillingCycle(req.BillingCycle, model.B2BBillingCycleMonthly)
	if err != nil {
		return nil, err
	}

	plan := &model.B2BPlan{
		Code:             sanitizeOrganizationCode(req.Code),
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		BillingCycle:     cycle,
		BasePricePerSeat: req.BasePricePerSeat,
		MinSeats:         req.MinSeats,
		MaxSeats:         req.MaxSeats,
		FeaturesJSON:     marshalFeaturesJSON(req.Features),
		IsActive:         true,
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	if err := s.repo.CreateB2BPlan(ctx, plan); err != nil {
		return nil, err
	}

	dtoPlan := toB2BPlanDTO(plan)
	return &dtoPlan, nil
}

func (s *B2BService) AdminUpdatePlan(ctx context.Context, planID uint, req *dto.UpsertB2BPlanRequest) (*dto.B2BPlanDTO, error) {
	if req.MinSeats > req.MaxSeats {
		return nil, ErrB2BInvalidSeatCount
	}

	plan, err := s.repo.GetB2BPlanByID(ctx, planID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrB2BPlanNotFound) {
			return nil, ErrB2BPlanNotFound
		}
		return nil, err
	}

	cycle, err := normalizeBillingCycle(req.BillingCycle, plan.BillingCycle)
	if err != nil {
		return nil, err
	}

	code := sanitizeOrganizationCode(req.Code)
	if code == "" {
		code = plan.Code
	}

	plan.Code = code
	plan.Name = strings.TrimSpace(req.Name)
	plan.Description = strings.TrimSpace(req.Description)
	plan.BillingCycle = cycle
	plan.BasePricePerSeat = req.BasePricePerSeat
	plan.MinSeats = req.MinSeats
	plan.MaxSeats = req.MaxSeats
	plan.FeaturesJSON = marshalFeaturesJSON(req.Features)
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateB2BPlan(ctx, plan); err != nil {
		return nil, err
	}

	dtoPlan := toB2BPlanDTO(plan)
	return &dtoPlan, nil
}

func (s *B2BService) AdminCreateSubscription(ctx context.Context, organizationID uint, req *dto.CreateB2BSubscriptionRequest) (*dto.B2BSubscriptionDTO, error) {
	_, err := s.repo.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationNotFound) {
			return nil, ErrB2BOrganizationNotFound
		}
		return nil, err
	}

	plan, err := s.repo.GetB2BPlanByID(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrB2BPlanNotFound) {
			return nil, ErrB2BPlanNotFound
		}
		return nil, err
	}
	if !plan.IsActive {
		return nil, ErrB2BPlanNotFound
	}

	cycle, err := normalizeBillingCycle(req.BillingCycle, plan.BillingCycle)
	if err != nil {
		return nil, err
	}

	if req.ContractedSeats < plan.MinSeats || req.ContractedSeats > plan.MaxSeats {
		return nil, ErrB2BInvalidSeatCount
	}

	startAt := time.Now()
	if req.StartsAt != nil {
		startAt = *req.StartsAt
	}
	endAt := computeSubscriptionEnd(startAt, cycle)

	multiplier := billingCycleMultiplier(cycle)
	subtotal := int64(req.ContractedSeats) * plan.BasePricePerSeat * multiplier

	existing, existingErr := s.repo.GetActiveSubscriptionByOrganizationID(ctx, organizationID)
	if existingErr == nil {
		if existing.UsedSeats > req.ContractedSeats {
			return nil, ErrB2BContractedSeatsTooLow
		}

		existing.PlanID = plan.ID
		existing.Status = model.B2BSubscriptionStatusActive
		existing.ContractedSeats = req.ContractedSeats
		existing.BillingCycle = cycle
		existing.UnitPrice = plan.BasePricePerSeat
		existing.Subtotal = subtotal
		existing.DiscountAmount = 0
		existing.TotalAmount = subtotal
		existing.StartsAt = startAt
		existing.EndsAt = endAt
		now := time.Now()
		existing.ActivatedAt = &now
		existing.MetadataJSON = "{}"

		if err := s.repo.UpdateB2BSubscription(ctx, existing); err != nil {
			return nil, err
		}
		s.appendAuditLog(ctx, organizationID, nil, "subscription.admin_upserted", "b2b_subscription", strconv.FormatUint(uint64(existing.ID), 10), map[string]any{
			"contracted_seats": existing.ContractedSeats,
			"billing_cycle":    existing.BillingCycle,
			"plan_id":          existing.PlanID,
		})
		return toB2BSubscriptionDTO(existing, plan), nil
	}

	if !errors.Is(existingErr, infrastructure.ErrB2BSubscriptionNotFound) {
		return nil, existingErr
	}

	now := time.Now()
	subscription := &model.B2BSubscription{
		OrganizationID:  organizationID,
		PlanID:          plan.ID,
		Status:          model.B2BSubscriptionStatusActive,
		ContractedSeats: req.ContractedSeats,
		UsedSeats:       0,
		BillingCycle:    cycle,
		UnitPrice:       plan.BasePricePerSeat,
		Subtotal:        subtotal,
		DiscountAmount:  0,
		TotalAmount:     subtotal,
		StartsAt:        startAt,
		EndsAt:          endAt,
		ActivatedAt:     &now,
		MetadataJSON:    "{}",
	}

	if err := s.repo.CreateB2BSubscription(ctx, subscription); err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, nil, "subscription.admin_upserted", "b2b_subscription", strconv.FormatUint(uint64(subscription.ID), 10), map[string]any{
		"contracted_seats": subscription.ContractedSeats,
		"billing_cycle":    subscription.BillingCycle,
		"plan_id":          subscription.PlanID,
	})

	return toB2BSubscriptionDTO(subscription, plan), nil
}

func (s *B2BService) GetOrganizationSummary(ctx context.Context, requesterID, organizationID uint) (*dto.OrganizationSummaryResponse, error) {
	if _, err := s.ensureOrganizationMember(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	organization, err := s.repo.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationNotFound) {
			return nil, ErrB2BOrganizationNotFound
		}
		return nil, err
	}

	summary := &dto.OrganizationSummaryResponse{
		Organization: toOrganizationDTO(organization),
		SeatUsage: dto.OrganizationSeatUsageDTO{
			ContractedSeats: 0,
			UsedSeats:       0,
			AvailableSeats:  0,
		},
	}

	subscription, subErr := s.repo.GetActiveSubscriptionByOrganizationID(ctx, organizationID)
	if subErr != nil && !errors.Is(subErr, infrastructure.ErrB2BSubscriptionNotFound) {
		return nil, subErr
	}

	if subscription != nil {
		usedCount, err := s.repo.CountActiveSeatAllocations(ctx, subscription.ID)
		if err != nil {
			return nil, err
		}
		subscription.UsedSeats = int(usedCount)
		summary.Subscription = toB2BSubscriptionDTO(subscription, &subscription.Plan)
		summary.SeatUsage.ContractedSeats = subscription.ContractedSeats
		summary.SeatUsage.UsedSeats = subscription.UsedSeats
		summary.SeatUsage.AvailableSeats = subscription.ContractedSeats - subscription.UsedSeats
		if summary.SeatUsage.AvailableSeats < 0 {
			summary.SeatUsage.AvailableSeats = 0
		}
	}

	return summary, nil
}

func (s *B2BService) ListOrganizationMembers(ctx context.Context, requesterID, organizationID uint, status string) ([]dto.OrganizationMemberDTO, error) {
	if _, err := s.ensureOrganizationMember(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	status = strings.TrimSpace(strings.ToLower(status))
	if status != "" && status != string(model.OrganizationMemberStatusInvited) && status != string(model.OrganizationMemberStatusPendingApproval) && status != string(model.OrganizationMemberStatusActive) && status != string(model.OrganizationMemberStatusRemoved) {
		return nil, ErrB2BInvalidMemberStatus
	}

	members, err := s.repo.ListOrganizationMembers(ctx, organizationID, status)
	if err != nil {
		return nil, err
	}

	result := make([]dto.OrganizationMemberDTO, 0, len(members))
	for _, member := range members {
		result = append(result, toOrganizationMemberDTO(member))
	}
	return result, nil
}

func normalizeInviteRole(raw string) model.OrganizationMemberRole {
	role := strings.TrimSpace(strings.ToLower(raw))
	if role == string(model.OrganizationMemberRoleAdmin) {
		return model.OrganizationMemberRoleAdmin
	}
	return model.OrganizationMemberRoleMember
}

func (s *B2BService) inviteMemberInternal(ctx context.Context, requesterID, organizationID uint, req *dto.InviteOrganizationMemberRequest) (*dto.InviteMemberResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, errors.New("email is required")
	}
	desiredRole := normalizeInviteRole(req.Role)

	invitedUser, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if desiredRole == model.OrganizationMemberRoleAdmin && (invitedUser == nil || invitedUser.Role != model.RoleMitra) {
		return nil, ErrB2BMitraRoleRequired
	}

	invitationToken, err := utils.GenerateRandomString(40)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	expiresAt := now.Add(14 * 24 * time.Hour)

	member, findErr := s.repo.FindOrganizationMemberByEmail(ctx, organizationID, email)
	if findErr != nil && !errors.Is(findErr, infrastructure.ErrOrganizationMemberNotFound) {
		return nil, findErr
	}

	if member != nil {
		if member.Status == model.OrganizationMemberStatusActive {
			return nil, ErrB2BMemberAlreadyActive
		}
		member.Role = desiredRole
		member.FullName = strings.TrimSpace(req.FullName)
		member.Status = model.OrganizationMemberStatusInvited
		member.InvitedBy = &requesterID
		member.InvitedAt = &now
		member.JoinedAt = nil
		member.RemovedAt = nil
		member.UserID = nil
		member.InvitationToken = invitationToken
		member.InvitationExpires = &expiresAt

		if err := s.repo.SaveOrganizationMember(ctx, member); err != nil {
			return nil, err
		}
		s.appendAuditLog(ctx, organizationID, &requesterID, "member.invited", "organization_member", strconv.FormatUint(uint64(member.ID), 10), map[string]any{
			"email": member.Email,
			"role":  member.Role,
		})
		return &dto.InviteMemberResponse{
			MemberID:            member.ID,
			Email:               member.Email,
			InvitationToken:     member.InvitationToken,
			InvitationExpiresAt: member.InvitationExpires,
		}, nil
	}

	newMember := &model.OrganizationMember{
		OrganizationID:    organizationID,
		Email:             email,
		FullName:          strings.TrimSpace(req.FullName),
		Role:              desiredRole,
		Status:            model.OrganizationMemberStatusInvited,
		InvitationToken:   invitationToken,
		InvitationExpires: &expiresAt,
		InvitedBy:         &requesterID,
		InvitedAt:         &now,
	}

	if err := s.repo.CreateOrganizationMember(ctx, newMember); err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "member.invited", "organization_member", strconv.FormatUint(uint64(newMember.ID), 10), map[string]any{
		"email": newMember.Email,
		"role":  newMember.Role,
	})

	return &dto.InviteMemberResponse{
		MemberID:            newMember.ID,
		Email:               newMember.Email,
		InvitationToken:     newMember.InvitationToken,
		InvitationExpiresAt: newMember.InvitationExpires,
	}, nil
}

func (s *B2BService) InviteMember(ctx context.Context, requesterID, organizationID uint, req *dto.InviteOrganizationMemberRequest) (*dto.InviteMemberResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}
	return s.inviteMemberInternal(ctx, requesterID, organizationID, req)
}

func (s *B2BService) BulkInviteMembers(ctx context.Context, requesterID, organizationID uint, req *dto.BulkInviteOrganizationMembersRequest) (*dto.BulkInviteMembersResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	response := &dto.BulkInviteMembersResponse{
		Total:   len(req.Members),
		Invited: 0,
		Skipped: 0,
		Results: make([]dto.BulkInviteMemberResult, 0, len(req.Members)),
	}

	for _, memberReq := range req.Members {
		result := dto.BulkInviteMemberResult{Email: strings.TrimSpace(strings.ToLower(memberReq.Email))}
		_, err := s.inviteMemberInternal(ctx, requesterID, organizationID, &memberReq)
		if err != nil {
			result.Status = "skipped"
			result.Message = err.Error()
			response.Skipped++
			response.Results = append(response.Results, result)
			continue
		}

		result.Status = "invited"
		result.Message = "invitation created"
		response.Invited++
		response.Results = append(response.Results, result)
	}

	return response, nil
}

func (s *B2BService) GetInvitePreview(ctx context.Context, userID uint, invitationToken string) (*dto.OrganizationInvitePreviewResponse, error) {
	trimmedToken := strings.TrimSpace(invitationToken)
	if trimmedToken == "" {
		return nil, ErrB2BInvalidInviteToken
	}

	member, err := s.repo.FindOrganizationMemberByInvitationTokenGlobal(ctx, trimmedToken)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationMemberNotFound) {
			return nil, ErrB2BInvalidInviteToken
		}
		return nil, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	canAccept := true
	message := ""
	switch {
	case member.Status == model.OrganizationMemberStatusActive:
		canAccept = false
		message = "Undangan sudah digunakan."
	case member.InvitationExpires == nil || time.Now().After(*member.InvitationExpires):
		canAccept = false
		message = "Undangan sudah kedaluwarsa."
	case !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(member.Email)):
		canAccept = false
		message = "Undangan ini ditujukan untuk email lain."
	case member.Role == model.OrganizationMemberRoleAdmin && user.Role != model.RoleMitra:
		canAccept = false
		message = "Akun admin organisasi harus dipromosikan menjadi mitra oleh admin internal."
	}

	return &dto.OrganizationInvitePreviewResponse{
		Organization: toOrganizationDTO(&member.Organization),
		MemberID:     member.ID,
		Email:        member.Email,
		FullName:     member.FullName,
		Role:         string(member.Role),
		Status:       string(member.Status),
		ExpiresAt:    member.InvitationExpires,
		CanAccept:    canAccept,
		Message:      message,
	}, nil
}

func (s *B2BService) AcceptInviteByToken(ctx context.Context, userID uint, req *dto.AcceptOrganizationInviteRequest) (*dto.AcceptOrganizationInviteResponse, error) {
	trimmedToken := strings.TrimSpace(req.InvitationToken)
	if trimmedToken == "" {
		return nil, ErrB2BInvalidInviteToken
	}

	member, err := s.repo.FindOrganizationMemberByInvitationTokenGlobal(ctx, trimmedToken)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationMemberNotFound) {
			return nil, ErrB2BInvalidInviteToken
		}
		return nil, err
	}

	return s.acceptInviteMember(ctx, userID, member)
}

func (s *B2BService) AcceptInvite(ctx context.Context, userID, organizationID uint, req *dto.AcceptOrganizationInviteRequest) (*dto.AcceptOrganizationInviteResponse, error) {
	trimmedToken := strings.TrimSpace(req.InvitationToken)
	if trimmedToken == "" {
		return nil, ErrB2BInvalidInviteToken
	}

	member, err := s.repo.FindOrganizationMemberByInvitationToken(ctx, organizationID, trimmedToken)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationMemberNotFound) {
			return nil, ErrB2BInvalidInviteToken
		}
		return nil, err
	}

	return s.acceptInviteMember(ctx, userID, member)
}

func (s *B2BService) acceptInviteMember(ctx context.Context, userID uint, member *model.OrganizationMember) (*dto.AcceptOrganizationInviteResponse, error) {
	organizationID := member.OrganizationID
	if member.Status == model.OrganizationMemberStatusActive {
		return nil, ErrB2BMemberAlreadyActive
	}
	if member.InvitationExpires == nil || time.Now().After(*member.InvitationExpires) {
		return nil, ErrB2BInvitationExpired
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if member.Role == model.OrganizationMemberRoleAdmin && user.Role != model.RoleMitra {
		return nil, ErrB2BMitraRoleRequired
	}
	if !strings.EqualFold(strings.TrimSpace(user.Email), strings.TrimSpace(member.Email)) {
		return nil, ErrB2BEmailMismatch
	}

	var contractedSeats int
	var usedSeats int
	err = s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		organization, lockErr := s.repo.LockOrganizationByID(tx, organizationID)
		if lockErr != nil {
			if errors.Is(lockErr, infrastructure.ErrOrganizationNotFound) {
				return ErrB2BOrganizationNotFound
			}
			return lockErr
		}

		subscription, lockSubErr := s.repo.LockActiveSubscriptionByOrganizationID(tx, organizationID)
		if lockSubErr != nil && !errors.Is(lockSubErr, infrastructure.ErrB2BSubscriptionNotFound) {
			return lockSubErr
		}

		now := time.Now()
		if organization.RequiresMemberApproval {
			member.Status = model.OrganizationMemberStatusPendingApproval
		} else {
			member.Status = model.OrganizationMemberStatusActive
		}
		member.UserID = &userID
		member.JoinedAt = &now
		member.RemovedAt = nil
		member.InvitationToken = ""
		member.InvitationExpires = nil
		if saveErr := s.repo.SaveOrganizationMemberTx(tx, member); saveErr != nil {
			return saveErr
		}

		if organization.RequiresMemberApproval {
			approval := &model.OrganizationMemberApproval{
				OrganizationID:       organizationID,
				OrganizationMemberID: member.ID,
				RequestedBy:          &userID,
				Status:               model.OrganizationMemberApprovalStatusPending,
			}
			if approvalErr := s.repo.CreateOrganizationMemberApprovalTx(tx, approval); approvalErr != nil {
				return approvalErr
			}

			if subscription != nil {
				usedCount, countErr := s.repo.CountActiveSeatAllocationsTx(tx, subscription.ID)
				if countErr != nil {
					return countErr
				}
				contractedSeats = subscription.ContractedSeats
				usedSeats = int(usedCount)
			}
			return nil
		}

		if subscription == nil {
			return ErrB2BSubscriptionNotFound
		}

		hasPaidCoverage, paymentErr := s.repo.HasPaidBillingCoverageForSubscriptionTx(tx, subscription.ID, now)
		if paymentErr != nil {
			return paymentErr
		}
		if !hasPaidCoverage {
			return ErrB2BSubscriptionPaymentRequired
		}

		usedCount, countErr := s.repo.CountActiveSeatAllocationsTx(tx, subscription.ID)
		if countErr != nil {
			return countErr
		}
		if int(usedCount) >= subscription.ContractedSeats {
			return ErrB2BInsufficientSeats
		}

		allocation := &model.B2BSeatAllocation{
			SubscriptionID:       subscription.ID,
			OrganizationMemberID: member.ID,
			AllocatedAt:          now,
		}
		if allocErr := s.repo.CreateSeatAllocationTx(tx, allocation); allocErr != nil {
			return allocErr
		}

		subscription.UsedSeats = int(usedCount) + 1
		if subErr := s.repo.SaveB2BSubscriptionTx(tx, subscription); subErr != nil {
			return subErr
		}

		contractedSeats = subscription.ContractedSeats
		usedSeats = subscription.UsedSeats
		return nil
	})
	if err != nil {
		return nil, err
	}

	action := "member.invite_accepted"
	if member.Status == model.OrganizationMemberStatusPendingApproval {
		action = "member.pending_approval"
	}
	s.appendAuditLog(ctx, organizationID, &userID, action, "organization_member", strconv.FormatUint(uint64(member.ID), 10), map[string]any{
		"email":  member.Email,
		"status": member.Status,
	})

	response := &dto.AcceptOrganizationInviteResponse{
		OrganizationID: organizationID,
		MemberID:       member.ID,
		SeatUsage: dto.OrganizationSeatUsageDTO{
			ContractedSeats: contractedSeats,
			UsedSeats:       usedSeats,
			AvailableSeats:  contractedSeats - usedSeats,
		},
	}
	if response.SeatUsage.AvailableSeats < 0 {
		response.SeatUsage.AvailableSeats = 0
	}

	return response, nil
}

func (s *B2BService) RemoveMember(ctx context.Context, requesterID, organizationID, memberID uint) (*dto.OrganizationSeatUsageDTO, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	member, err := s.repo.GetOrganizationMemberByID(ctx, organizationID, memberID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrOrganizationMemberNotFound) {
			return nil, ErrB2BMemberNotFound
		}
		return nil, err
	}

	if member.Role == model.OrganizationMemberRoleOwner {
		return nil, ErrB2BCannotRemoveOwnerMember
	}

	var contractedSeats int
	var usedSeats int
	err = s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		now := time.Now()
		member.Status = model.OrganizationMemberStatusRemoved
		member.RemovedAt = &now
		member.InvitationToken = ""
		member.InvitationExpires = nil

		subscription, lockErr := s.repo.LockActiveSubscriptionByOrganizationID(tx, organizationID)
		if lockErr != nil && !errors.Is(lockErr, infrastructure.ErrB2BSubscriptionNotFound) {
			return lockErr
		}

		if subscription != nil {
			allocation, allocationErr := s.repo.FindActiveSeatAllocationByMemberTx(tx, subscription.ID, member.ID)
			if allocationErr != nil {
				return allocationErr
			}
			if allocation != nil {
				if releaseErr := s.repo.ReleaseSeatAllocationTx(tx, allocation.ID, "member_removed"); releaseErr != nil {
					return releaseErr
				}

				usedCount, countErr := s.repo.CountActiveSeatAllocationsTx(tx, subscription.ID)
				if countErr != nil {
					return countErr
				}
				subscription.UsedSeats = int(usedCount)
				if subErr := s.repo.SaveB2BSubscriptionTx(tx, subscription); subErr != nil {
					return subErr
				}
			}

			contractedSeats = subscription.ContractedSeats
			usedSeats = subscription.UsedSeats
		}

		return s.repo.SaveOrganizationMemberTx(tx, member)
	})
	if err != nil {
		return nil, err
	}

	usage := &dto.OrganizationSeatUsageDTO{
		ContractedSeats: contractedSeats,
		UsedSeats:       usedSeats,
		AvailableSeats:  contractedSeats - usedSeats,
	}
	if usage.AvailableSeats < 0 {
		usage.AvailableSeats = 0
	}
	s.appendAuditLog(ctx, organizationID, &requesterID, "member.removed", "organization_member", strconv.FormatUint(uint64(memberID), 10), map[string]any{
		"email": member.Email,
		"role":  member.Role,
	})
	return usage, nil
}

func (s *B2BService) CreateQuote(ctx context.Context, creatorUserID uint, req *dto.CreateB2BQuoteRequest) (*dto.CreateB2BQuoteResponse, error) {
	if err := s.ensureMitraRequester(ctx, creatorUserID); err != nil {
		return nil, err
	}

	if req.OrganizationID != nil {
		if _, err := s.repo.GetOrganizationByID(ctx, *req.OrganizationID); err != nil {
			if errors.Is(err, infrastructure.ErrOrganizationNotFound) {
				return nil, ErrB2BOrganizationNotFound
			}
			return nil, err
		}
	}

	plan, err := s.repo.GetB2BPlanByID(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrB2BPlanNotFound) {
			return nil, ErrB2BPlanNotFound
		}
		return nil, err
	}
	if !plan.IsActive {
		return nil, ErrB2BPlanNotFound
	}

	billingCycle, err := normalizeBillingCycle(req.BillingCycle, plan.BillingCycle)
	if err != nil {
		return nil, err
	}

	if req.RequestedSeats < plan.MinSeats || req.RequestedSeats > plan.MaxSeats {
		return nil, ErrB2BInvalidSeatCount
	}

	multiplier := billingCycleMultiplier(billingCycle)
	grossAmount := int64(req.RequestedSeats) * plan.BasePricePerSeat * multiplier

	volumeDiscountPercent := calculateVolumeDiscountPercent(req.RequestedSeats)
	volumeDiscountAmount := int64(float64(grossAmount) * volumeDiscountPercent)

	annualDiscountAmount := int64(0)
	if billingCycle == model.B2BBillingCycleYearly {
		annualDiscountAmount = int64(float64(grossAmount-volumeDiscountAmount) * 0.10)
	}

	selectedAddOns := uniqueStringItems(req.SelectedAddOns)
	addOnRule := addonPricingRules()
	addOnAmount := int64(0)
	for _, addOn := range selectedAddOns {
		if monthlyFee, exists := addOnRule[addOn]; exists {
			addOnAmount += monthlyFee * multiplier
		}
	}

	finalAmount := grossAmount - volumeDiscountAmount - annualDiscountAmount + addOnAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	quoteCode := "BQ-" + strings.ToUpper(strconv.FormatInt(time.Now().UnixNano(), 36))
	addOnsJSON, marshalErr := json.Marshal(selectedAddOns)
	if marshalErr != nil {
		return nil, marshalErr
	}

	validUntil := time.Now().Add(7 * 24 * time.Hour)
	quote := &model.B2BPricingQuote{
		QuoteCode:            quoteCode,
		OrganizationID:       req.OrganizationID,
		PlanID:               &plan.ID,
		RequestedSeats:       req.RequestedSeats,
		BillingCycle:         billingCycle,
		SelectedAddOnsJSON:   string(addOnsJSON),
		BasePricePerSeat:     plan.BasePricePerSeat,
		GrossAmount:          grossAmount,
		VolumeDiscountAmount: volumeDiscountAmount,
		AnnualDiscountAmount: annualDiscountAmount,
		AddOnAmount:          addOnAmount,
		FinalAmount:          finalAmount,
		Currency:             "IDR",
		ValidUntil:           validUntil,
		Status:               model.B2BQuoteStatusDraft,
		CreatedBy:            &creatorUserID,
	}
	if err := s.repo.CreateB2BPricingQuote(ctx, quote); err != nil {
		return nil, err
	}

	appliedRules := make([]string, 0, 4)
	if volumeDiscountPercent > 0 {
		appliedRules = append(appliedRules, fmt.Sprintf("volume_discount_%.0f%%", volumeDiscountPercent*100))
	}
	if annualDiscountAmount > 0 {
		appliedRules = append(appliedRules, "annual_discount_10%")
	}
	if len(selectedAddOns) > 0 {
		appliedRules = append(appliedRules, fmt.Sprintf("addons_%d", len(selectedAddOns)))
	}

	return &dto.CreateB2BQuoteResponse{
		QuoteCode:            quote.QuoteCode,
		OrganizationID:       quote.OrganizationID,
		PlanID:               plan.ID,
		RequestedSeats:       quote.RequestedSeats,
		BillingCycle:         string(quote.BillingCycle),
		BasePricePerSeat:     quote.BasePricePerSeat,
		GrossAmount:          quote.GrossAmount,
		VolumeDiscountAmount: quote.VolumeDiscountAmount,
		AnnualDiscountAmount: quote.AnnualDiscountAmount,
		AddOnAmount:          quote.AddOnAmount,
		FinalAmount:          quote.FinalAmount,
		Currency:             quote.Currency,
		ValidUntil:           quote.ValidUntil,
		SelectedAddOns:       selectedAddOns,
		AppliedRules:         appliedRules,
	}, nil
}

func (s *B2BService) IsUserEntitledB2BPremium(ctx context.Context, userID uint) (bool, *uint, error) {
	orgID, err := s.repo.GetActiveB2BOrganizationForUser(ctx, userID, time.Now())
	if err != nil {
		return false, nil, err
	}
	if orgID == nil {
		return false, nil, nil
	}
	return true, orgID, nil
}
