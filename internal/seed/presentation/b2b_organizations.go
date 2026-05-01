package presentation

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedB2BOrganizations seeds B2B entities including orgs, members, subscriptions, and enterprise fixtures.
func SeedB2BOrganizations(db *gorm.DB) error {
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	owner, err := getUserByEmail(db, "mitra@ruang-tenang.com")
	if err != nil {
		return err
	}
	if owner == nil {
		return nil
	}

	member, err := getUserByEmail(db, "dery@gmail.com")
	if err != nil {
		return err
	}
	if member == nil {
		return nil
	}

	var plan model.B2BPlan
	planErr := db.Where("code = ?", "b2b-campus-yearly").First(&plan).Error
	if planErr != nil {
		planErr = db.Where("code = ?", "b2b-campus-monthly").First(&plan).Error
	}
	if planErr != nil {
		if !errors.Is(planErr, gorm.ErrRecordNotFound) {
			return planErr
		}
		plan = model.B2BPlan{
			Code:             "b2b-campus-yearly",
			Name:             "B2B Campus Yearly",
			Description:      "Fallback plan created by presentation seeder.",
			BillingCycle:     model.B2BBillingCycleYearly,
			BasePricePerSeat: 52000,
			MinSeats:         25,
			MaxSeats:         5000,
			FeaturesJSON:     `{"sso":true,"seat_management":true,"analytics":true}`,
			IsActive:         true,
		}
		if err := db.Create(&plan).Error; err != nil {
			return err
		}
	}

	org := model.Organization{
		Code:                   "demo-kampus-nusantara",
		Name:                   "Demo Kampus Nusantara",
		BusinessType:           "education",
		ContactEmail:           "it@kampus-demo.ac.id",
		Status:                 model.OrganizationStatusActive,
		RequiresMemberApproval: true,
		CreatedBy:              uintPtr(owner.ID),
	}

	var existingOrg model.Organization
	err = db.Where("code = ?", org.Code).First(&existingOrg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&org).Error; err != nil {
				return err
			}
			existingOrg = org
		} else {
			return err
		}
	} else {
		if err := db.Model(&existingOrg).Updates(map[string]interface{}{
			"name":                     org.Name,
			"business_type":            org.BusinessType,
			"contact_email":            org.ContactEmail,
			"status":                   org.Status,
			"requires_member_approval": org.RequiresMemberApproval,
			"created_by":               org.CreatedBy,
		}).Error; err != nil {
			return err
		}
	}

	if err := removeObsoleteOrganizationMembers(db, existingOrg.ID, []string{
		"demo.utama@ruang-tenang.com",
		"demo.cadangan@ruang-tenang.com",
		"usertest@gmail.com",
		"mitra.invited@ruang-tenang.com",
		"andhika@gmail.com",
	}, now); err != nil {
		return err
	}

	_, err = upsertOrganizationMember(db, model.OrganizationMember{
		OrganizationID: existingOrg.ID,
		UserID:         uintPtr(owner.ID),
		Email:          owner.Email,
		FullName:       owner.Name,
		Role:           model.OrganizationMemberRoleOwner,
		Status:         model.OrganizationMemberStatusActive,
		InvitedBy:      uintPtr(owner.ID),
		InvitedAt:      timePtr(now.AddDate(0, 0, -40)),
		JoinedAt:       timePtr(now.AddDate(0, 0, -39)),
	})
	if err != nil {
		return err
	}

	var activeMember model.OrganizationMember
	activeMember, err = upsertOrganizationMember(db, model.OrganizationMember{
		OrganizationID: existingOrg.ID,
		UserID:         uintPtr(member.ID),
		Email:          member.Email,
		FullName:       member.Name,
		Role:           model.OrganizationMemberRoleMember,
		Status:         model.OrganizationMemberStatusActive,
		InvitedBy:      uintPtr(owner.ID),
		InvitedAt:      timePtr(now.AddDate(0, 0, -19)),
		JoinedAt:       timePtr(now.AddDate(0, 0, -18)),
	})
	if err != nil {
		return err
	}

	var pendingMember model.OrganizationMember
	pendingMember, err = upsertOrganizationMember(db, model.OrganizationMember{
		OrganizationID:  existingOrg.ID,
		Email:           "calon.member@kampus-demo.ac.id",
		FullName:        "Calon Member Demo",
		Role:            model.OrganizationMemberRoleMember,
		Status:          model.OrganizationMemberStatusPendingApproval,
		InvitationToken: "approve-calon-member-demo",
		InvitedBy:       uintPtr(owner.ID),
		InvitedAt:       timePtr(now.AddDate(0, 0, -1)),
	})
	if err != nil {
		return err
	}

	contractedSeats := 120
	unitPrice := plan.BasePricePerSeat
	subtotal := int64(contractedSeats) * unitPrice
	total := subtotal - 250000
	activatedAt := now.AddDate(0, 0, -35)
	subscription := model.B2BSubscription{
		OrganizationID:  existingOrg.ID,
		PlanID:          plan.ID,
		Status:          model.B2BSubscriptionStatusActive,
		ContractedSeats: contractedSeats,
		BillingCycle:    plan.BillingCycle,
		UnitPrice:       unitPrice,
		Subtotal:        subtotal,
		DiscountAmount:  250000,
		TotalAmount:     total,
		StartsAt:        startOfToday.AddDate(0, 0, -35),
		EndsAt:          startOfToday.AddDate(0, 11, 0),
		ActivatedAt:     &activatedAt,
		MetadataJSON:    `{"source":"presentation-seeder","sales_owner":"Enterprise Demo Team"}`,
	}

	var existingSub model.B2BSubscription
	err = db.Where("organization_id = ? AND status = ?", existingOrg.ID, model.B2BSubscriptionStatusActive).First(&existingSub).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&subscription).Error; err != nil {
				return err
			}
			existingSub = subscription
		} else {
			return err
		}
	} else {
		if err := db.Model(&existingSub).Updates(map[string]interface{}{
			"plan_id":          subscription.PlanID,
			"contracted_seats": subscription.ContractedSeats,
			"billing_cycle":    subscription.BillingCycle,
			"unit_price":       subscription.UnitPrice,
			"subtotal":         subscription.Subtotal,
			"discount_amount":  subscription.DiscountAmount,
			"total_amount":     subscription.TotalAmount,
			"starts_at":        subscription.StartsAt,
			"ends_at":          subscription.EndsAt,
			"activated_at":     subscription.ActivatedAt,
			"metadata_json":    subscription.MetadataJSON,
			"status":           subscription.Status,
		}).Error; err != nil {
			return err
		}
	}

	seatMembers := []model.OrganizationMember{activeMember}

	for _, memberWithSeat := range seatMembers {
		if err := upsertSeatAllocation(db, existingSub.ID, memberWithSeat.ID, now.AddDate(0, 0, -20)); err != nil {
			return err
		}
	}

	if err := db.Model(&model.B2BSubscription{}).Where("id = ?", existingSub.ID).Update("used_seats", len(seatMembers)).Error; err != nil {
		return err
	}

	if pendingMember.ID != 0 {
		if err := upsertMemberApproval(db, model.OrganizationMemberApproval{
			OrganizationID:       existingOrg.ID,
			OrganizationMemberID: pendingMember.ID,
			Status:               model.OrganizationMemberApprovalStatusPending,
			Note:                 "Menunggu approval dari admin organisasi untuk seat allocation.",
		}); err != nil {
			return err
		}
	}

	if activeMember.ID != 0 {
		approvedAt := now.AddDate(0, 0, -18)
		if err := upsertMemberApproval(db, model.OrganizationMemberApproval{
			OrganizationID:       existingOrg.ID,
			OrganizationMemberID: activeMember.ID,
			RequestedBy:          uintPtr(owner.ID),
			ApproverUserID:       uintPtr(owner.ID),
			Status:               model.OrganizationMemberApprovalStatusApproved,
			Note:                 "Approval otomatis untuk anggota demo aktif.",
			DecidedAt:            &approvedAt,
		}); err != nil {
			return err
		}
	}

	if err := ensureAuditLog(db, existingOrg.ID, uintPtr(owner.ID), "member_invited", "organization_member", owner.Email, `{"count":2}`); err != nil {
		return err
	}
	if err := ensureAuditLog(db, existingOrg.ID, uintPtr(owner.ID), "subscription_activated", "b2b_subscription", fmt.Sprintf("%d", existingSub.ID), `{"contracted_seats":120}`); err != nil {
		return err
	}
	if err := ensureAuditLog(db, existingOrg.ID, uintPtr(owner.ID), "sso_configured", "b2b_sso_config", fmt.Sprintf("org-%d", existingOrg.ID), `{"provider":"oidc","is_enabled":true}`); err != nil {
		return err
	}

	templates := []model.OrganizationOnboardingTemplate{
		{
			OrganizationID: existingOrg.ID,
			Role:           model.OrganizationMemberRoleOwner,
			Title:          "Onboarding Owner",
			WelcomeMessage: "Selamat datang Owner. Atur kebijakan approval anggota, seat, dan SSO.",
			ChecklistJSON:  `["Lengkapi profil organisasi","Atur konfigurasi SSO","Undang admin pertama"]`,
			IsDefault:      true,
			IsActive:       true,
			CreatedBy:      uintPtr(owner.ID),
		},
		{
			OrganizationID: existingOrg.ID,
			Role:           model.OrganizationMemberRoleAdmin,
			Title:          "Onboarding Admin",
			WelcomeMessage: "Selamat datang Admin. Kelola approvals dan health score tim.",
			ChecklistJSON:  `["Review pending approvals","Pantau penggunaan seat","Aktifkan reminder job"]`,
			IsDefault:      true,
			IsActive:       true,
			CreatedBy:      uintPtr(owner.ID),
		},
		{
			OrganizationID: existingOrg.ID,
			Role:           model.OrganizationMemberRoleMember,
			Title:          "Onboarding Member",
			WelcomeMessage: "Mulai perjalanan tenangmu bersama organisasi.",
			ChecklistJSON:  `["Lengkapi profil personal","Coba sesi chat refleksi","Selesaikan daily challenge"]`,
			IsDefault:      true,
			IsActive:       true,
			CreatedBy:      uintPtr(owner.ID),
		},
	}
	for _, template := range templates {
		if err := upsertOnboardingTemplate(db, template); err != nil {
			return err
		}
	}

	for i := 6; i >= 0; i-- {
		metricDate := startOfToday.AddDate(0, 0, -i)
		if err := upsertUsageMetric(db, model.B2BUsageDailyMetric{
			OrganizationID:   existingOrg.ID,
			MetricDate:       metricDate,
			ActiveMembers:    2,
			InvitedMembers:   1,
			PendingApprovals: 1,
			ContractedSeats:  contractedSeats,
			UsedSeats:        len(seatMembers),
			MessagesSent:     18 + (6-i)*4,
		}); err != nil {
			return err
		}
	}

	seatPayload := `{"threshold":0.8,"used_seats":1,"contracted_seats":120}`
	if err := upsertReminderJob(db, model.B2BReminderJob{
		OrganizationID: existingOrg.ID,
		SubscriptionID: uintPtr(existingSub.ID),
		JobType:        model.B2BReminderJobTypeSeatThreshold,
		Status:         model.B2BReminderJobStatusPending,
		DueAt:          startOfToday.Add(9 * time.Hour),
		PayloadJSON:    seatPayload,
		AttemptCount:   0,
	}); err != nil {
		return err
	}

	if err := upsertReminderJob(db, model.B2BReminderJob{
		OrganizationID: existingOrg.ID,
		SubscriptionID: uintPtr(existingSub.ID),
		JobType:        model.B2BReminderJobTypeSubscriptionExpiry,
		Status:         model.B2BReminderJobStatusPending,
		DueAt:          existingSub.EndsAt.AddDate(0, 0, -14),
		PayloadJSON:    `{"days_before_expiry":14}`,
		AttemptCount:   0,
	}); err != nil {
		return err
	}

	sso := model.B2BSSOConfig{
		OrganizationID: existingOrg.ID,
		Provider:       model.B2BSSOProviderOIDC,
		IssuerURL:      "https://id.demo-kampus.ac.id",
		EntrypointURL:  "https://id.demo-kampus.ac.id/oauth2/auth",
		Audience:       "ruang-tenang-b2b",
		IsEnabled:      true,
		EnforceSSO:     false,
		MetadataJSON:   `{"client_id":"demo-campus-oidc","domain":"kampus-demo.ac.id"}`,
		CreatedBy:      uintPtr(owner.ID),
		UpdatedBy:      uintPtr(owner.ID),
	}
	if err := upsertSSOConfig(db, sso); err != nil {
		return err
	}

	recommendation := model.B2BPricingRecommendation{
		OrganizationID:          existingOrg.ID,
		GeneratedForDate:        startOfToday,
		RecommendedPlanID:       uintPtr(plan.ID),
		RecommendedBillingCycle: model.B2BBillingCycleYearly,
		RecommendedSeats:        150,
		EstimatedMonthlyCost:    7800000,
		EstimatedYearlySaving:   9200000,
		ConfidenceScore:         87.40,
		ReasonsJSON:             `["Tren penggunaan kursi meningkat 18% dalam 30 hari","Penggunaan fitur analytics tinggi","Prediksi onboarding anggota baru aktif"]`,
	}
	if err := upsertPricingRecommendation(db, recommendation); err != nil {
		return err
	}

	invoicePaidAt := now.AddDate(0, 0, -5)
	if err := upsertBillingHistory(db, model.B2BBillingHistory{
		SubscriptionID:     existingSub.ID,
		InvoiceNumber:      "INV-B2B-DEMO-2025-0001",
		BillingPeriodStart: subscription.StartsAt,
		BillingPeriodEnd:   subscription.EndsAt,
		SeatsBilled:        contractedSeats,
		Amount:             total,
		Status:             model.B2BBillingHistoryStatusPaid,
		PaidAt:             &invoicePaidAt,
	}); err != nil {
		return err
	}

	quoteValidUntil := now.AddDate(0, 0, 14)
	if err := upsertPricingQuote(db, model.B2BPricingQuote{
		QuoteCode:            "QUOTE-DEMO-KAMPUS-001",
		OrganizationID:       uintPtr(existingOrg.ID),
		PlanID:               uintPtr(plan.ID),
		RequestedSeats:       150,
		BillingCycle:         model.B2BBillingCycleYearly,
		SelectedAddOnsJSON:   `["priority_support","advanced_analytics"]`,
		BasePricePerSeat:     plan.BasePricePerSeat,
		GrossAmount:          int64(150) * plan.BasePricePerSeat,
		VolumeDiscountAmount: 450000,
		AnnualDiscountAmount: 700000,
		AddOnAmount:          350000,
		FinalAmount:          int64(150)*plan.BasePricePerSeat - 800000,
		Currency:             "IDR",
		ValidUntil:           quoteValidUntil,
		Status:               model.B2BQuoteStatusAccepted,
		CreatedBy:            uintPtr(owner.ID),
	}); err != nil {
		return err
	}

	return nil
}

func getUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	var user model.User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func removeObsoleteOrganizationMembers(db *gorm.DB, organizationID uint, emails []string, removedAt time.Time) error {
	var members []model.OrganizationMember
	if err := db.Where("organization_id = ? AND email IN ?", organizationID, emails).Find(&members).Error; err != nil {
		return err
	}
	if len(members) == 0 {
		return nil
	}

	memberIDs := make([]uint, 0, len(members))
	for _, member := range members {
		memberIDs = append(memberIDs, member.ID)
	}

	if err := db.Model(&model.B2BSeatAllocation{}).
		Where("organization_member_id IN ? AND released_at IS NULL", memberIDs).
		Updates(map[string]interface{}{
			"released_at":    removedAt,
			"release_reason": "Seeder account cleanup",
		}).Error; err != nil {
		return err
	}

	return db.Model(&model.OrganizationMember{}).
		Where("id IN ?", memberIDs).
		Updates(map[string]interface{}{
			"status":                model.OrganizationMemberStatusRemoved,
			"user_id":               nil,
			"invitation_token":      "",
			"invitation_expires_at": nil,
			"removed_at":            removedAt,
		}).Error
}

func upsertOrganizationMember(db *gorm.DB, member model.OrganizationMember) (model.OrganizationMember, error) {
	if member.Status == model.OrganizationMemberStatusInvited && member.InvitationExpires == nil {
		expires := time.Now().UTC().AddDate(0, 0, 7)
		member.InvitationExpires = &expires
	}

	var invitationToken interface{}
	if member.InvitationToken != "" {
		invitationToken = member.InvitationToken
	}

	var existing model.OrganizationMember
	err := db.Where("organization_id = ? AND email = ?", member.OrganizationID, member.Email).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Model(&model.OrganizationMember{}).Create(map[string]interface{}{
				"organization_id":       member.OrganizationID,
				"user_id":               member.UserID,
				"email":                 member.Email,
				"full_name":             member.FullName,
				"role":                  member.Role,
				"status":                member.Status,
				"invitation_token":      invitationToken,
				"invitation_expires_at": member.InvitationExpires,
				"invited_by":            member.InvitedBy,
				"invited_at":            member.InvitedAt,
				"joined_at":             member.JoinedAt,
				"removed_at":            member.RemovedAt,
				"created_at":            time.Now().UTC(),
				"updated_at":            time.Now().UTC(),
			}).Error; err != nil {
				return model.OrganizationMember{}, err
			}
			if err := db.Where("organization_id = ? AND email = ?", member.OrganizationID, member.Email).First(&member).Error; err != nil {
				return model.OrganizationMember{}, err
			}
			return member, nil
		}
		return model.OrganizationMember{}, err
	}

	updates := map[string]interface{}{
		"user_id":               member.UserID,
		"full_name":             member.FullName,
		"role":                  member.Role,
		"status":                member.Status,
		"invitation_token":      invitationToken,
		"invitation_expires_at": member.InvitationExpires,
		"invited_by":            member.InvitedBy,
		"invited_at":            member.InvitedAt,
		"joined_at":             member.JoinedAt,
		"removed_at":            member.RemovedAt,
	}
	if err := db.Model(&existing).Updates(updates).Error; err != nil {
		return model.OrganizationMember{}, err
	}

	if err := db.Where("id = ?", existing.ID).First(&existing).Error; err != nil {
		return model.OrganizationMember{}, err
	}
	return existing, nil
}

func upsertSeatAllocation(db *gorm.DB, subscriptionID, memberID uint, allocatedAt time.Time) error {
	var existing model.B2BSeatAllocation
	err := db.Where("subscription_id = ? AND organization_member_id = ?", subscriptionID, memberID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			allocation := model.B2BSeatAllocation{
				SubscriptionID:       subscriptionID,
				OrganizationMemberID: memberID,
				AllocatedAt:          allocatedAt,
			}
			return db.Create(&allocation).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"allocated_at":   allocatedAt,
		"released_at":    nil,
		"release_reason": "",
	}).Error
}

func upsertMemberApproval(db *gorm.DB, approval model.OrganizationMemberApproval) error {
	var existing model.OrganizationMemberApproval
	err := db.Where(
		"organization_id = ? AND organization_member_id = ? AND status = ?",
		approval.OrganizationID,
		approval.OrganizationMemberID,
		approval.Status,
	).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&approval).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"requested_by":     approval.RequestedBy,
		"approver_user_id": approval.ApproverUserID,
		"note":             approval.Note,
		"decided_at":       approval.DecidedAt,
	}).Error
}

func ensureAuditLog(db *gorm.DB, organizationID uint, actorUserID *uint, action, entityType, entityID, metadataJSON string) error {
	var existing model.OrganizationAuditLog
	err := db.Where(
		"organization_id = ? AND action = ? AND entity_type = ? AND entity_id = ?",
		organizationID,
		action,
		entityType,
		entityID,
	).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			entry := model.OrganizationAuditLog{
				OrganizationID: organizationID,
				ActorUserID:    actorUserID,
				Action:         action,
				EntityType:     entityType,
				EntityID:       entityID,
				MetadataJSON:   metadataJSON,
			}
			return db.Create(&entry).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"actor_user_id": actorUserID,
		"metadata_json": metadataJSON,
	}).Error
}

func upsertOnboardingTemplate(db *gorm.DB, template model.OrganizationOnboardingTemplate) error {
	var existing model.OrganizationOnboardingTemplate
	err := db.Where("organization_id = ? AND role = ?", template.OrganizationID, template.Role).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&template).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"title":           template.Title,
		"welcome_message": template.WelcomeMessage,
		"checklist_json":  template.ChecklistJSON,
		"is_default":      template.IsDefault,
		"is_active":       template.IsActive,
		"created_by":      template.CreatedBy,
	}).Error
}

func upsertUsageMetric(db *gorm.DB, metric model.B2BUsageDailyMetric) error {
	var existing model.B2BUsageDailyMetric
	err := db.Where("organization_id = ? AND metric_date = ?", metric.OrganizationID, metric.MetricDate).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&metric).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"active_members":    metric.ActiveMembers,
		"invited_members":   metric.InvitedMembers,
		"pending_approvals": metric.PendingApprovals,
		"contracted_seats":  metric.ContractedSeats,
		"used_seats":        metric.UsedSeats,
		"messages_sent":     metric.MessagesSent,
	}).Error
}

func upsertReminderJob(db *gorm.DB, job model.B2BReminderJob) error {
	var existing model.B2BReminderJob
	err := db.Where("organization_id = ? AND job_type = ? AND due_at = ?", job.OrganizationID, job.JobType, job.DueAt).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&job).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"subscription_id": job.SubscriptionID,
		"status":          job.Status,
		"payload_json":    job.PayloadJSON,
		"attempt_count":   job.AttemptCount,
		"last_error":      job.LastError,
		"sent_at":         job.SentAt,
	}).Error
}

func upsertSSOConfig(db *gorm.DB, cfg model.B2BSSOConfig) error {
	var existing model.B2BSSOConfig
	err := db.Where("organization_id = ?", cfg.OrganizationID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&cfg).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"provider":        cfg.Provider,
		"issuer_url":      cfg.IssuerURL,
		"entrypoint_url":  cfg.EntrypointURL,
		"audience":        cfg.Audience,
		"certificate_pem": cfg.CertificatePEM,
		"is_enabled":      cfg.IsEnabled,
		"enforce_sso":     cfg.EnforceSSO,
		"metadata_json":   cfg.MetadataJSON,
		"created_by":      cfg.CreatedBy,
		"updated_by":      cfg.UpdatedBy,
	}).Error
}

func upsertPricingRecommendation(db *gorm.DB, rec model.B2BPricingRecommendation) error {
	var existing model.B2BPricingRecommendation
	err := db.Where("organization_id = ? AND generated_for_date = ?", rec.OrganizationID, rec.GeneratedForDate).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&rec).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"recommended_plan_id":       rec.RecommendedPlanID,
		"recommended_billing_cycle": rec.RecommendedBillingCycle,
		"recommended_seats":         rec.RecommendedSeats,
		"estimated_monthly_cost":    rec.EstimatedMonthlyCost,
		"estimated_yearly_saving":   rec.EstimatedYearlySaving,
		"confidence_score":          rec.ConfidenceScore,
		"reasons_json":              rec.ReasonsJSON,
	}).Error
}

func upsertBillingHistory(db *gorm.DB, history model.B2BBillingHistory) error {
	var existing model.B2BBillingHistory
	err := db.Where("invoice_number = ?", history.InvoiceNumber).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&history).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"subscription_id":      history.SubscriptionID,
		"billing_period_start": history.BillingPeriodStart,
		"billing_period_end":   history.BillingPeriodEnd,
		"seats_billed":         history.SeatsBilled,
		"amount":               history.Amount,
		"status":               history.Status,
		"paid_at":              history.PaidAt,
	}).Error
}

func upsertPricingQuote(db *gorm.DB, quote model.B2BPricingQuote) error {
	var existing model.B2BPricingQuote
	err := db.Where("quote_code = ?", quote.QuoteCode).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&quote).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"organization_id":        quote.OrganizationID,
		"plan_id":                quote.PlanID,
		"requested_seats":        quote.RequestedSeats,
		"billing_cycle":          quote.BillingCycle,
		"selected_addons_json":   quote.SelectedAddOnsJSON,
		"base_price_per_seat":    quote.BasePricePerSeat,
		"gross_amount":           quote.GrossAmount,
		"volume_discount_amount": quote.VolumeDiscountAmount,
		"annual_discount_amount": quote.AnnualDiscountAmount,
		"add_on_amount":          quote.AddOnAmount,
		"final_amount":           quote.FinalAmount,
		"currency":               quote.Currency,
		"valid_until":            quote.ValidUntil,
		"status":                 quote.Status,
		"created_by":             quote.CreatedBy,
	}).Error
}

func uintPtr(v uint) *uint {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}
