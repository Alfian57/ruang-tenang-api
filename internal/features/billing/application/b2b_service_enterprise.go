package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/features/billing/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func parseStringSliceJSON(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []string{}
	}
	return items
}

func marshalStringSliceJSON(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	encoded, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func parseJSONMap(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]any{}
	}
	return result
}

func marshalJSONMap(data map[string]any) string {
	if len(data) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func parseRole(raw string) (model.OrganizationMemberRole, error) {
	role := model.OrganizationMemberRole(strings.TrimSpace(strings.ToLower(raw)))
	switch role {
	case model.OrganizationMemberRoleOwner, model.OrganizationMemberRoleAdmin, model.OrganizationMemberRoleMember:
		return role, nil
	default:
		return "", ErrB2BInvalidMemberStatus
	}
}

func normalizeChecklist(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func normalizeSSOProvider(raw string) (model.B2BSSOProvider, error) {
	provider := model.B2BSSOProvider(strings.TrimSpace(strings.ToLower(raw)))
	switch provider {
	case "":
		return "", nil
	case model.B2BSSOProviderSAML, model.B2BSSOProviderOIDC, model.B2BSSOProviderGoogleWorkspace, model.B2BSSOProviderAzureAD, model.B2BSSOProviderOkta:
		return provider, nil
	default:
		return "", ErrB2BInvalidSSOProvider
	}
}

func computeSeatUtilization(contractedSeats, usedSeats int) float64 {
	if contractedSeats <= 0 {
		return 0
	}
	pct := (float64(usedSeats) / float64(contractedSeats)) * 100
	return math.Round(pct*100) / 100
}

func toOnboardingTemplateDTO(template *model.OrganizationOnboardingTemplate) *dto.OrganizationOnboardingTemplateDTO {
	if template == nil {
		return nil
	}
	return &dto.OrganizationOnboardingTemplateDTO{
		ID:             template.ID,
		OrganizationID: template.OrganizationID,
		Role:           string(template.Role),
		Title:          template.Title,
		WelcomeMessage: template.WelcomeMessage,
		Checklist:      parseStringSliceJSON(template.ChecklistJSON),
		IsDefault:      template.IsDefault,
		IsActive:       template.IsActive,
		CreatedBy:      template.CreatedBy,
		CreatedAt:      template.CreatedAt,
		UpdatedAt:      template.UpdatedAt,
	}
}

func (s *B2BService) appendAuditLog(ctx context.Context, organizationID uint, actorUserID *uint, action, entityType, entityID string, metadata map[string]any) {
	auditLog := &model.OrganizationAuditLog{
		OrganizationID: organizationID,
		ActorUserID:    actorUserID,
		Action:         strings.TrimSpace(action),
		EntityType:     strings.TrimSpace(entityType),
		EntityID:       strings.TrimSpace(entityID),
		MetadataJSON:   marshalJSONMap(metadata),
	}
	_ = s.repo.CreateOrganizationAuditLog(ctx, auditLog)
}

func (s *B2BService) snapshotDailyMetric(ctx context.Context, organizationID uint) (*model.B2BUsageDailyMetric, error) {
	statusCounts, err := s.repo.CountOrganizationMembersByStatus(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	contractedSeats := 0
	usedSeats := 0
	subscription, subErr := s.repo.GetActiveSubscriptionByOrganizationID(ctx, organizationID)
	if subErr != nil && !errors.Is(subErr, infrastructure.ErrB2BSubscriptionNotFound) {
		return nil, subErr
	}
	if subscription != nil {
		contractedSeats = subscription.ContractedSeats
		usedCount, countErr := s.repo.CountActiveSeatAllocations(ctx, subscription.ID)
		if countErr != nil {
			return nil, countErr
		}
		usedSeats = int(usedCount)
	}
	today := startOfDay(time.Now())
	messageCounts, err := s.repo.CountOrganizationUserChatMessagesByDate(ctx, organizationID, today, today)
	if err != nil {
		return nil, err
	}

	metric := &model.B2BUsageDailyMetric{
		OrganizationID:   organizationID,
		MetricDate:       today,
		ActiveMembers:    statusCounts[string(model.OrganizationMemberStatusActive)],
		InvitedMembers:   statusCounts[string(model.OrganizationMemberStatusInvited)],
		PendingApprovals: statusCounts[string(model.OrganizationMemberStatusPendingApproval)],
		ContractedSeats:  contractedSeats,
		UsedSeats:        usedSeats,
		MessagesSent:     messageCounts[today.Format("2006-01-02")],
	}

	if err := s.repo.UpsertUsageDailyMetric(ctx, metric); err != nil {
		return nil, err
	}
	return metric, nil
}

func (s *B2BService) ApproveMember(ctx context.Context, requesterID, organizationID, memberID uint, note string) (*dto.MemberApprovalDecisionResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	trimmedNote := strings.TrimSpace(note)
	decidedAt := time.Now()
	contractedSeats := 0
	usedSeats := 0

	err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		member, memberErr := s.repo.LockOrganizationMemberByID(tx, organizationID, memberID)
		if memberErr != nil {
			if errors.Is(memberErr, infrastructure.ErrOrganizationMemberNotFound) {
				return ErrB2BMemberNotFound
			}
			return memberErr
		}
		if member.Status != model.OrganizationMemberStatusPendingApproval {
			return ErrB2BMemberNotPendingApproval
		}

		subscription, subErr := s.repo.LockActiveSubscriptionByOrganizationID(tx, organizationID)
		if subErr != nil {
			if errors.Is(subErr, infrastructure.ErrB2BSubscriptionNotFound) {
				return ErrB2BSubscriptionNotFound
			}
			return subErr
		}

		hasPaidCoverage, paymentErr := s.repo.HasPaidBillingCoverageForSubscriptionTx(tx, subscription.ID, decidedAt)
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
			AllocatedAt:          decidedAt,
		}
		if allocErr := s.repo.CreateSeatAllocationTx(tx, allocation); allocErr != nil {
			return allocErr
		}

		member.Status = model.OrganizationMemberStatusActive
		member.RemovedAt = nil
		if saveErr := s.repo.SaveOrganizationMemberTx(tx, member); saveErr != nil {
			return saveErr
		}

		approval := &model.OrganizationMemberApproval{
			OrganizationID:       organizationID,
			OrganizationMemberID: member.ID,
			RequestedBy:          member.UserID,
			ApproverUserID:       &requesterID,
			Status:               model.OrganizationMemberApprovalStatusApproved,
			Note:                 trimmedNote,
			DecidedAt:            &decidedAt,
		}
		if approvalErr := s.repo.CreateOrganizationMemberApprovalTx(tx, approval); approvalErr != nil {
			return approvalErr
		}

		subscription.UsedSeats = int(usedCount) + 1
		if saveSubErr := s.repo.SaveB2BSubscriptionTx(tx, subscription); saveSubErr != nil {
			return saveSubErr
		}

		contractedSeats = subscription.ContractedSeats
		usedSeats = subscription.UsedSeats
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "member.approved", "organization_member", strconv.FormatUint(uint64(memberID), 10), map[string]any{
		"note": trimmedNote,
	})

	member, memberErr := s.repo.GetOrganizationMemberByID(ctx, organizationID, memberID)
	if memberErr == nil && member.UserID != nil && s.notificationService != nil {
		s.notificationService.CreateCustomNotification(ctx, *member.UserID, "b2b_member_approved", "Akses B2B Disetujui", "Permintaan akses organisasimu sudah disetujui admin.", map[string]string{
			"organization_id": strconv.FormatUint(uint64(organizationID), 10),
			"member_id":       strconv.FormatUint(uint64(memberID), 10),
		})
	}

	return &dto.MemberApprovalDecisionResponse{
		OrganizationID: organizationID,
		MemberID:       memberID,
		Status:         string(model.OrganizationMemberStatusActive),
		SeatUsage: dto.OrganizationSeatUsageDTO{
			ContractedSeats: contractedSeats,
			UsedSeats:       usedSeats,
			AvailableSeats:  max(0, contractedSeats-usedSeats),
		},
		DecidedAt: decidedAt,
		Note:      trimmedNote,
	}, nil
}

func (s *B2BService) RejectMember(ctx context.Context, requesterID, organizationID, memberID uint, note string) (*dto.MemberApprovalDecisionResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	trimmedNote := strings.TrimSpace(note)
	decidedAt := time.Now()

	err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		member, memberErr := s.repo.LockOrganizationMemberByID(tx, organizationID, memberID)
		if memberErr != nil {
			if errors.Is(memberErr, infrastructure.ErrOrganizationMemberNotFound) {
				return ErrB2BMemberNotFound
			}
			return memberErr
		}
		if member.Status != model.OrganizationMemberStatusPendingApproval {
			return ErrB2BMemberNotPendingApproval
		}

		member.Status = model.OrganizationMemberStatusRemoved
		member.RemovedAt = &decidedAt
		if saveErr := s.repo.SaveOrganizationMemberTx(tx, member); saveErr != nil {
			return saveErr
		}

		approval := &model.OrganizationMemberApproval{
			OrganizationID:       organizationID,
			OrganizationMemberID: member.ID,
			RequestedBy:          member.UserID,
			ApproverUserID:       &requesterID,
			Status:               model.OrganizationMemberApprovalStatusRejected,
			Note:                 trimmedNote,
			DecidedAt:            &decidedAt,
		}
		return s.repo.CreateOrganizationMemberApprovalTx(tx, approval)
	})
	if err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "member.rejected", "organization_member", strconv.FormatUint(uint64(memberID), 10), map[string]any{
		"note": trimmedNote,
	})

	member, memberErr := s.repo.GetOrganizationMemberByID(ctx, organizationID, memberID)
	if memberErr == nil && member.UserID != nil && s.notificationService != nil {
		s.notificationService.CreateCustomNotification(ctx, *member.UserID, "b2b_member_rejected", "Permintaan Akses Ditolak", "Permintaan akses organisasi ditolak admin. Silakan hubungi admin organisasimu.", map[string]string{
			"organization_id": strconv.FormatUint(uint64(organizationID), 10),
			"member_id":       strconv.FormatUint(uint64(memberID), 10),
		})
	}

	return &dto.MemberApprovalDecisionResponse{
		OrganizationID: organizationID,
		MemberID:       memberID,
		Status:         string(model.OrganizationMemberStatusRemoved),
		SeatUsage:      dto.OrganizationSeatUsageDTO{},
		DecidedAt:      decidedAt,
		Note:           trimmedNote,
	}, nil
}

func (s *B2BService) GetOrganizationAnalytics(ctx context.Context, requesterID, organizationID uint, days int) (*dto.OrganizationAnalyticsResponse, error) {
	if _, err := s.ensureOrganizationMember(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	snapshot, err := s.snapshotDailyMetric(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	endDate := startOfDay(time.Now())
	startDate := endDate.AddDate(0, 0, -(days - 1))
	metrics, err := s.repo.ListUsageDailyMetrics(ctx, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(metrics) == 0 && snapshot != nil {
		metrics = append(metrics, *snapshot)
	}
	messageCounts, err := s.repo.CountOrganizationUserChatMessagesByDate(ctx, organizationID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	statusCounts, err := s.repo.CountOrganizationMembersByStatus(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	contractedSeats := snapshot.ContractedSeats
	usedSeats := snapshot.UsedSeats
	seatUsage := dto.OrganizationSeatUsageDTO{
		ContractedSeats: contractedSeats,
		UsedSeats:       usedSeats,
		AvailableSeats:  max(0, contractedSeats-usedSeats),
	}

	trend := make([]dto.DailyUsageMetricDTO, 0, len(metrics))
	for _, metric := range metrics {
		metricDate := metric.MetricDate.Format("2006-01-02")
		messagesSent := metric.MessagesSent
		if countedMessages, exists := messageCounts[metricDate]; exists {
			messagesSent = countedMessages
		}
		trend = append(trend, dto.DailyUsageMetricDTO{
			MetricDate:       metricDate,
			ActiveMembers:    metric.ActiveMembers,
			InvitedMembers:   metric.InvitedMembers,
			PendingApprovals: metric.PendingApprovals,
			ContractedSeats:  metric.ContractedSeats,
			UsedSeats:        metric.UsedSeats,
			MessagesSent:     messagesSent,
		})
	}

	return &dto.OrganizationAnalyticsResponse{
		OrganizationID:     organizationID,
		WindowDays:         days,
		MemberStatusCounts: statusCounts,
		SeatUsage:          seatUsage,
		SeatUtilizationPct: computeSeatUtilization(contractedSeats, usedSeats),
		Trend:              trend,
		GeneratedAt:        time.Now(),
	}, nil
}

func (s *B2BService) ListAuditLogs(ctx context.Context, requesterID, organizationID uint, action string, page, limit int) (*dto.OrganizationAuditLogListResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}

	logs, total, err := s.repo.ListOrganizationAuditLogs(ctx, organizationID, action, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.OrganizationAuditLogDTO, 0, len(logs))
	for _, item := range logs {
		items = append(items, dto.OrganizationAuditLogDTO{
			ID:          item.ID,
			ActorUserID: item.ActorUserID,
			Action:      item.Action,
			EntityType:  item.EntityType,
			EntityID:    item.EntityID,
			Metadata:    parseJSONMap(item.MetadataJSON),
			CreatedAt:   item.CreatedAt,
		})
	}

	return &dto.OrganizationAuditLogListResponse{
		Items: items,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *B2BService) UpsertOnboardingTemplate(ctx context.Context, requesterID, organizationID uint, req *dto.UpsertOrganizationOnboardingTemplateRequest) (*dto.OrganizationOnboardingTemplateDTO, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	role, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	template := &model.OrganizationOnboardingTemplate{
		OrganizationID: organizationID,
		Role:           role,
		Title:          strings.TrimSpace(req.Title),
		WelcomeMessage: strings.TrimSpace(req.WelcomeMessage),
		ChecklistJSON:  marshalStringSliceJSON(normalizeChecklist(req.Checklist)),
		IsDefault:      req.IsDefault,
		IsActive:       isActive,
		CreatedBy:      &requesterID,
	}
	if err := s.repo.UpsertOrganizationOnboardingTemplate(ctx, template); err != nil {
		return nil, err
	}

	stored, err := s.repo.GetOrganizationOnboardingTemplate(ctx, organizationID, role)
	if err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "onboarding_template.upserted", "onboarding_template", string(role), map[string]any{
		"is_default": req.IsDefault,
		"is_active":  isActive,
	})

	return toOnboardingTemplateDTO(stored), nil
}

func (s *B2BService) GetOnboardingTemplate(ctx context.Context, requesterID, organizationID uint, roleRaw string) (*dto.OrganizationOnboardingTemplateDTO, error) {
	if _, err := s.ensureOrganizationMember(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	role := model.OrganizationMemberRoleMember
	if strings.TrimSpace(roleRaw) != "" {
		parsedRole, err := parseRole(roleRaw)
		if err != nil {
			return nil, err
		}
		role = parsedRole
	}

	template, err := s.repo.GetOrganizationOnboardingTemplate(ctx, organizationID, role)
	if err != nil {
		return nil, err
	}
	if template == nil {
		return &dto.OrganizationOnboardingTemplateDTO{
			OrganizationID: organizationID,
			Role:           string(role),
			Title:          "Selamat datang di organisasi",
			WelcomeMessage: "Pelajari panduan awal, aktivasi akses premium, dan mulai gunakan Ruang Tenang bersama tim.",
			Checklist: []string{
				"Perbarui profil dan data kontak",
				"Ikuti panduan etika diskusi tim",
				"Coba fitur premium pertama",
			},
			IsDefault: true,
			IsActive:  true,
		}, nil
	}

	return toOnboardingTemplateDTO(template), nil
}

func (s *B2BService) SelfServiceSeatUpgrade(ctx context.Context, requesterID, organizationID uint, req *dto.SeatUpgradeSubscriptionRequest) (*dto.SeatUpgradeSubscriptionResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	var updatedSubscription *model.B2BSubscription
	var updatedPlan *model.B2BPlan
	var seatUsage dto.OrganizationSeatUsageDTO

	err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		subscription, subErr := s.repo.LockActiveSubscriptionByOrganizationID(tx, organizationID)
		if subErr != nil {
			if errors.Is(subErr, infrastructure.ErrB2BSubscriptionNotFound) {
				return ErrB2BSubscriptionNotFound
			}
			return subErr
		}

		plan, planErr := s.repo.GetB2BPlanByID(ctx, subscription.PlanID)
		if planErr != nil {
			if errors.Is(planErr, infrastructure.ErrB2BPlanNotFound) {
				return ErrB2BPlanNotFound
			}
			return planErr
		}

		if req.ContractedSeats < subscription.UsedSeats {
			return ErrB2BContractedSeatsTooLow
		}
		if req.ContractedSeats < plan.MinSeats || req.ContractedSeats > plan.MaxSeats {
			return ErrB2BInvalidSeatCount
		}

		cycle, cycleErr := normalizeBillingCycle(req.BillingCycle, subscription.BillingCycle)
		if cycleErr != nil {
			return cycleErr
		}

		multiplier := billingCycleMultiplier(cycle)
		subtotal := int64(req.ContractedSeats) * subscription.UnitPrice * multiplier
		subscription.ContractedSeats = req.ContractedSeats
		subscription.BillingCycle = cycle
		subscription.Subtotal = subtotal
		subscription.DiscountAmount = 0
		subscription.TotalAmount = subtotal

		if saveErr := s.repo.SaveB2BSubscriptionTx(tx, subscription); saveErr != nil {
			return saveErr
		}

		updatedSubscription = subscription
		updatedPlan = plan
		seatUsage = dto.OrganizationSeatUsageDTO{
			ContractedSeats: subscription.ContractedSeats,
			UsedSeats:       subscription.UsedSeats,
			AvailableSeats:  max(0, subscription.ContractedSeats-subscription.UsedSeats),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "subscription.seat_upgraded", "b2b_subscription", strconv.FormatUint(uint64(updatedSubscription.ID), 10), map[string]any{
		"contracted_seats": updatedSubscription.ContractedSeats,
		"billing_cycle":    updatedSubscription.BillingCycle,
	})

	return &dto.SeatUpgradeSubscriptionResponse{
		Subscription: toB2BSubscriptionDTO(updatedSubscription, updatedPlan),
		SeatUsage:    seatUsage,
	}, nil
}

func (s *B2BService) RunOrganizationReminders(ctx context.Context, requesterID, organizationID uint) (*dto.RunOrganizationRemindersResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	now := time.Now()
	generated := 0
	sent := 0
	failed := 0

	subscription, subErr := s.repo.GetActiveSubscriptionByOrganizationID(ctx, organizationID)
	if subErr != nil && !errors.Is(subErr, infrastructure.ErrB2BSubscriptionNotFound) {
		return nil, subErr
	}

	if subscription != nil {
		usedCount, countErr := s.repo.CountActiveSeatAllocations(ctx, subscription.ID)
		if countErr != nil {
			return nil, countErr
		}
		utilizationPct := computeSeatUtilization(subscription.ContractedSeats, int(usedCount))

		if utilizationPct >= 80 {
			payload := map[string]any{
				"subscription_id":  subscription.ID,
				"used_seats":       int(usedCount),
				"contracted_seats": subscription.ContractedSeats,
				"utilization_pct":  utilizationPct,
			}
			job := &model.B2BReminderJob{
				OrganizationID: organizationID,
				SubscriptionID: &subscription.ID,
				JobType:        model.B2BReminderJobTypeSeatThreshold,
				Status:         model.B2BReminderJobStatusPending,
				DueAt:          startOfDay(now),
				PayloadJSON:    marshalJSONMap(payload),
			}
			if err := s.repo.UpsertReminderJob(ctx, job); err == nil {
				generated++
			}
		}

		daysToExpiry := int(math.Ceil(subscription.EndsAt.Sub(now).Hours() / 24))
		if daysToExpiry <= 7 {
			payload := map[string]any{
				"subscription_id": subscription.ID,
				"expires_at":      subscription.EndsAt.Format(time.RFC3339),
				"days_remaining":  daysToExpiry,
			}
			job := &model.B2BReminderJob{
				OrganizationID: organizationID,
				SubscriptionID: &subscription.ID,
				JobType:        model.B2BReminderJobTypeSubscriptionExpiry,
				Status:         model.B2BReminderJobStatusPending,
				DueAt:          subscription.EndsAt,
				PayloadJSON:    marshalJSONMap(payload),
			}
			if err := s.repo.UpsertReminderJob(ctx, job); err == nil {
				generated++
			}
		}
	}

	jobs, err := s.repo.ListDueReminderJobs(ctx, organizationID, now)
	if err != nil {
		return nil, err
	}

	managers, err := s.repo.ListOrganizationManagers(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		if s.notificationService == nil {
			_ = s.repo.MarkReminderJobFailed(ctx, job.ID, "notification service unavailable")
			failed++
			continue
		}

		title := "Reminder organisasi B2B"
		message := "Ada kondisi langganan yang perlu ditindaklanjuti admin organisasi."
		switch job.JobType {
		case model.B2BReminderJobTypeSeatThreshold:
			title = "Kapasitas seat hampir penuh"
			message = "Penggunaan seat timmu melewati 80%. Pertimbangkan upgrade seat agar onboarding tidak tertahan."
		case model.B2BReminderJobTypeSubscriptionExpiry:
			title = "Langganan B2B akan berakhir"
			message = "Langganan B2B akan segera berakhir. Segera perpanjang agar akses premium tim tetap aktif."
		}

		for _, manager := range managers {
			if manager.UserID == nil {
				continue
			}
			s.notificationService.CreateCustomNotification(ctx, *manager.UserID, "b2b_reminder", title, message, map[string]string{
				"organization_id": strconv.FormatUint(uint64(organizationID), 10),
				"job_id":          strconv.FormatUint(uint64(job.ID), 10),
				"job_type":        string(job.JobType),
			})
		}

		if markErr := s.repo.MarkReminderJobSent(ctx, job.ID); markErr != nil {
			_ = s.repo.MarkReminderJobFailed(ctx, job.ID, markErr.Error())
			failed++
			continue
		}
		sent++
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "reminder.jobs_run", "b2b_reminder_job", "", map[string]any{
		"generated": generated,
		"sent":      sent,
		"failed":    failed,
	})

	return &dto.RunOrganizationRemindersResponse{
		OrganizationID: organizationID,
		Generated:      generated,
		Sent:           sent,
		Failed:         failed,
		ProcessedAt:    now,
	}, nil
}

func (s *B2BService) GetSSOConfig(ctx context.Context, requesterID, organizationID uint) (*dto.OrganizationSSOConfigDTO, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	cfg, err := s.repo.GetSSOConfigByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return &dto.OrganizationSSOConfigDTO{
			OrganizationID: organizationID,
			Metadata:       map[string]any{},
		}, nil
	}

	createdAt := cfg.CreatedAt
	updatedAt := cfg.UpdatedAt
	return &dto.OrganizationSSOConfigDTO{
		OrganizationID: cfg.OrganizationID,
		Provider:       string(cfg.Provider),
		IssuerURL:      cfg.IssuerURL,
		EntrypointURL:  cfg.EntrypointURL,
		Audience:       cfg.Audience,
		CertificatePEM: cfg.CertificatePEM,
		IsEnabled:      cfg.IsEnabled,
		EnforceSSO:     cfg.EnforceSSO,
		Metadata:       parseJSONMap(cfg.MetadataJSON),
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}, nil
}

func (s *B2BService) UpsertSSOConfig(ctx context.Context, requesterID, organizationID uint, req *dto.UpsertOrganizationSSOConfigRequest) (*dto.OrganizationSSOConfigDTO, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}
	if req.EnforceSSO != nil && *req.EnforceSSO {
		return nil, ErrB2BSSOEnforcementUnavailable
	}

	existing, err := s.repo.GetSSOConfigByOrganizationID(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	provider, err := normalizeSSOProvider(req.Provider)
	if err != nil {
		return nil, err
	}
	if provider == "" && existing != nil {
		provider = existing.Provider
	}

	isEnabled := existing != nil && existing.IsEnabled
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}
	enforceSSO := false

	metadataJSON := "{}"
	if existing != nil {
		metadataJSON = existing.MetadataJSON
	}
	if req.Metadata != nil {
		metadataJSON = marshalJSONMap(req.Metadata)
	}

	cfg := &model.B2BSSOConfig{
		OrganizationID: organizationID,
		Provider:       provider,
		IssuerURL:      strings.TrimSpace(req.IssuerURL),
		EntrypointURL:  strings.TrimSpace(req.EntrypointURL),
		Audience:       strings.TrimSpace(req.Audience),
		CertificatePEM: strings.TrimSpace(req.CertificatePEM),
		IsEnabled:      isEnabled,
		EnforceSSO:     enforceSSO,
		MetadataJSON:   metadataJSON,
		CreatedBy:      &requesterID,
		UpdatedBy:      &requesterID,
	}

	if existing != nil {
		if cfg.IssuerURL == "" {
			cfg.IssuerURL = existing.IssuerURL
		}
		if cfg.EntrypointURL == "" {
			cfg.EntrypointURL = existing.EntrypointURL
		}
		if cfg.Audience == "" {
			cfg.Audience = existing.Audience
		}
		if cfg.CertificatePEM == "" {
			cfg.CertificatePEM = existing.CertificatePEM
		}
		cfg.CreatedBy = existing.CreatedBy
	}

	if err := s.repo.UpsertSSOConfig(ctx, cfg); err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "sso_config.updated", "b2b_sso_config", strconv.FormatUint(uint64(organizationID), 10), map[string]any{
		"provider":    provider,
		"is_enabled":  isEnabled,
		"enforce_sso": enforceSSO,
	})

	return s.GetSSOConfig(ctx, requesterID, organizationID)
}

func (s *B2BService) GetPricingRecommendation(ctx context.Context, requesterID, organizationID uint) (*dto.PricingRecommendationResponse, error) {
	if _, err := s.ensureOrganizationManager(ctx, requesterID, organizationID); err != nil {
		return nil, err
	}

	subscription, err := s.repo.GetActiveSubscriptionByOrganizationID(ctx, organizationID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrB2BSubscriptionNotFound) {
			return nil, ErrB2BSubscriptionNotFound
		}
		return nil, err
	}

	plan, err := s.repo.GetB2BPlanByID(ctx, subscription.PlanID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrB2BPlanNotFound) {
			return nil, ErrB2BPlanNotFound
		}
		return nil, err
	}

	metric, err := s.repo.GetLatestUsageDailyMetric(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if metric == nil {
		metric, err = s.snapshotDailyMetric(ctx, organizationID)
		if err != nil {
			return nil, err
		}
	}

	utilizationPct := computeSeatUtilization(metric.ContractedSeats, metric.UsedSeats)
	recommendedSeats := subscription.ContractedSeats
	confidenceScore := 72.5
	reasons := []string{"usage_stable"}

	switch {
	case utilizationPct >= 90:
		recommendedSeats = int(math.Ceil(float64(subscription.ContractedSeats) * 1.20))
		confidenceScore = 89.0
		reasons = []string{"high_seat_utilization", "prevent_onboarding_blockers"}
	case utilizationPct >= 75:
		recommendedSeats = int(math.Ceil(float64(subscription.ContractedSeats) * 1.10))
		confidenceScore = 83.0
		reasons = []string{"seat_utilization_trending_up", "capacity_buffer_recommended"}
	case utilizationPct < 45:
		recommendedSeats = int(math.Ceil(float64(max(metric.UsedSeats, 1)) * 1.15))
		confidenceScore = 80.0
		reasons = []string{"low_seat_utilization", "cost_efficiency_optimization"}
	}

	if recommendedSeats < plan.MinSeats {
		recommendedSeats = plan.MinSeats
	}
	if recommendedSeats > plan.MaxSeats {
		recommendedSeats = plan.MaxSeats
	}

	recommendedCycle := subscription.BillingCycle
	if recommendedCycle == model.B2BBillingCycleMonthly && recommendedSeats >= 50 {
		recommendedCycle = model.B2BBillingCycleYearly
		reasons = append(reasons, "annual_cycle_discount_opportunity")
	}

	estimatedMonthlyCost := int64(recommendedSeats) * subscription.UnitPrice
	estimatedYearlySaving := int64(0)
	if recommendedCycle == model.B2BBillingCycleYearly {
		estimatedYearlySaving = int64(math.Round(float64(estimatedMonthlyCost*12) * 0.10))
	}

	reasonJSON := marshalStringSliceJSON(reasons)
	recommendation := &model.B2BPricingRecommendation{
		OrganizationID:          organizationID,
		GeneratedForDate:        startOfDay(time.Now()),
		RecommendedPlanID:       &subscription.PlanID,
		RecommendedBillingCycle: recommendedCycle,
		RecommendedSeats:        recommendedSeats,
		EstimatedMonthlyCost:    estimatedMonthlyCost,
		EstimatedYearlySaving:   estimatedYearlySaving,
		ConfidenceScore:         confidenceScore,
		ReasonsJSON:             reasonJSON,
	}
	if err := s.repo.UpsertPricingRecommendation(ctx, recommendation); err != nil {
		return nil, err
	}

	s.appendAuditLog(ctx, organizationID, &requesterID, "pricing_recommendation.generated", "b2b_pricing_recommendation", strconv.FormatUint(uint64(organizationID), 10), map[string]any{
		"recommended_seats": recommendedSeats,
		"recommended_cycle": recommendedCycle,
	})

	return &dto.PricingRecommendationResponse{
		OrganizationID:          organizationID,
		GeneratedForDate:        recommendation.GeneratedForDate.Format("2006-01-02"),
		RecommendedPlanID:       recommendation.RecommendedPlanID,
		RecommendedBillingCycle: string(recommendation.RecommendedBillingCycle),
		RecommendedSeats:        recommendation.RecommendedSeats,
		EstimatedMonthlyCost:    recommendation.EstimatedMonthlyCost,
		EstimatedYearlySaving:   recommendation.EstimatedYearlySaving,
		ConfidenceScore:         recommendation.ConfidenceScore,
		Reasons:                 parseStringSliceJSON(recommendation.ReasonsJSON),
		CreatedAt:               recommendation.CreatedAt,
	}, nil
}

func (s *B2BService) maybeNotifyManagers(ctx context.Context, organizationID uint, title, message string, data map[string]string) {
	if s.notificationService == nil {
		return
	}
	managers, err := s.repo.ListOrganizationManagers(ctx, organizationID)
	if err != nil {
		return
	}
	for _, manager := range managers {
		if manager.UserID == nil {
			continue
		}
		s.notificationService.CreateCustomNotification(ctx, *manager.UserID, "b2b_notice", title, message, data)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clampSeatCount(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func buildReminderMessage(jobType model.B2BReminderJobType, payload map[string]any) string {
	switch jobType {
	case model.B2BReminderJobTypeSeatThreshold:
		utilization, _ := payload["utilization_pct"].(float64)
		return fmt.Sprintf("Penggunaan seat sudah mencapai %.2f%%. Pertimbangkan upgrade seat.", utilization)
	case model.B2BReminderJobTypeSubscriptionExpiry:
		daysRemaining, _ := payload["days_remaining"].(float64)
		return fmt.Sprintf("Langganan akan berakhir dalam %.0f hari. Segera lakukan perpanjangan.", daysRemaining)
	default:
		return "Ada pengingat langganan B2B yang memerlukan perhatian admin."
	}
}
