package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *B2BRepository) LockOrganizationByID(tx *gorm.DB, organizationID uint) (*model.Organization, error) {
	var organization model.Organization
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", organizationID).
		First(&organization).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}
	return &organization, nil
}

func (r *B2BRepository) LockOrganizationMemberByID(tx *gorm.DB, organizationID, memberID uint) (*model.OrganizationMember, error) {
	var member model.OrganizationMember
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("organization_id = ? AND id = ?", organizationID, memberID).
		First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrganizationMemberNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *B2BRepository) CountOrganizationMembersByStatus(ctx context.Context, organizationID uint) (map[string]int, error) {
	type statusCountRow struct {
		Status string
		Total  int64
	}

	var rows []statusCountRow
	err := r.db.WithContext(ctx).
		Table("organization_members").
		Select("status, COUNT(*) AS total").
		Where("organization_id = ?", organizationID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := map[string]int{
		string(model.OrganizationMemberStatusInvited):         0,
		string(model.OrganizationMemberStatusPendingApproval): 0,
		string(model.OrganizationMemberStatusActive):          0,
		string(model.OrganizationMemberStatusRemoved):         0,
	}
	for _, row := range rows {
		counts[row.Status] = int(row.Total)
	}
	return counts, nil
}

func (r *B2BRepository) ListOrganizationManagers(ctx context.Context, organizationID uint) ([]model.OrganizationMember, error) {
	var members []model.OrganizationMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND role IN ? AND user_id IS NOT NULL",
			organizationID,
			model.OrganizationMemberStatusActive,
			[]model.OrganizationMemberRole{model.OrganizationMemberRoleOwner, model.OrganizationMemberRoleAdmin},
		).
		Order("id ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *B2BRepository) CreateOrganizationMemberApprovalTx(tx *gorm.DB, approval *model.OrganizationMemberApproval) error {
	return tx.Create(approval).Error
}

func (r *B2BRepository) CreateOrganizationAuditLog(ctx context.Context, auditLog *model.OrganizationAuditLog) error {
	return r.db.WithContext(ctx).Create(auditLog).Error
}

func (r *B2BRepository) CreateOrganizationAuditLogTx(tx *gorm.DB, auditLog *model.OrganizationAuditLog) error {
	return tx.Create(auditLog).Error
}

func (r *B2BRepository) ListOrganizationAuditLogs(ctx context.Context, organizationID uint, action string, page, limit int) ([]model.OrganizationAuditLog, int64, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	query := r.db.WithContext(ctx).Model(&model.OrganizationAuditLog{}).Where("organization_id = ?", organizationID)
	if action != "" {
		query = query.Where("action = ?", strings.TrimSpace(action))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.OrganizationAuditLog
	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *B2BRepository) UpsertOrganizationOnboardingTemplate(ctx context.Context, template *model.OrganizationOnboardingTemplate) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}, {Name: "role"}},
		DoUpdates: clause.Assignments(map[string]any{
			"title":           template.Title,
			"welcome_message": template.WelcomeMessage,
			"checklist_json":  template.ChecklistJSON,
			"is_default":      template.IsDefault,
			"is_active":       template.IsActive,
			"updated_at":      time.Now(),
		}),
	}).Create(template).Error
}

func (r *B2BRepository) GetOrganizationOnboardingTemplate(ctx context.Context, organizationID uint, role model.OrganizationMemberRole) (*model.OrganizationOnboardingTemplate, error) {
	var template model.OrganizationOnboardingTemplate
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND role = ?", organizationID, role).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &template, nil
}

func (r *B2BRepository) ListOrganizationOnboardingTemplates(ctx context.Context, organizationID uint, activeOnly bool) ([]model.OrganizationOnboardingTemplate, error) {
	query := r.db.WithContext(ctx).Where("organization_id = ?", organizationID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	var templates []model.OrganizationOnboardingTemplate
	err := query.Order("role ASC").Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func (r *B2BRepository) UpsertUsageDailyMetric(ctx context.Context, metric *model.B2BUsageDailyMetric) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}, {Name: "metric_date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"active_members":    metric.ActiveMembers,
			"invited_members":   metric.InvitedMembers,
			"pending_approvals": metric.PendingApprovals,
			"contracted_seats":  metric.ContractedSeats,
			"used_seats":        metric.UsedSeats,
			"messages_sent":     metric.MessagesSent,
			"updated_at":        time.Now(),
		}),
	}).Create(metric).Error
}

func (r *B2BRepository) ListUsageDailyMetrics(ctx context.Context, organizationID uint, startDate, endDate time.Time) ([]model.B2BUsageDailyMetric, error) {
	var metrics []model.B2BUsageDailyMetric
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND metric_date >= ? AND metric_date <= ?", organizationID, startDate, endDate).
		Order("metric_date ASC").
		Find(&metrics).Error
	if err != nil {
		return nil, err
	}
	return metrics, nil
}

func (r *B2BRepository) GetLatestUsageDailyMetric(ctx context.Context, organizationID uint) (*model.B2BUsageDailyMetric, error) {
	var metric model.B2BUsageDailyMetric
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Order("metric_date DESC").
		First(&metric).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &metric, nil
}

func (r *B2BRepository) CountOrganizationUserChatMessagesByDate(ctx context.Context, organizationID uint, startDate, endDate time.Time) (map[string]int, error) {
	type row struct {
		MetricDate string
		Total      int64
	}

	var rows []row
	err := r.db.WithContext(ctx).
		Table("chat_messages AS cm").
		Select("DATE(cm.created_at)::text AS metric_date, COUNT(*) AS total").
		Joins("JOIN chat_sessions AS cs ON cs.id = cm.chat_session_id").
		Joins("JOIN organization_members AS om ON om.user_id = cs.user_id").
		Where("om.organization_id = ?", organizationID).
		Where("cm.role = ?", model.ChatRoleUser).
		Where("cm.created_at >= ? AND cm.created_at < ?", startDate, endDate.AddDate(0, 0, 1)).
		Group("DATE(cm.created_at)").
		Order("DATE(cm.created_at) ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[row.MetricDate] = int(row.Total)
	}
	return result, nil
}

func (r *B2BRepository) UpsertReminderJob(ctx context.Context, job *model.B2BReminderJob) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}, {Name: "job_type"}, {Name: "due_at"}},
		DoUpdates: clause.Assignments(map[string]any{
			"subscription_id": job.SubscriptionID,
			"status":          model.B2BReminderJobStatusPending,
			"payload_json":    job.PayloadJSON,
			"last_error":      "",
			"updated_at":      time.Now(),
		}),
	}).Create(job).Error
}

func (r *B2BRepository) ListDueReminderJobs(ctx context.Context, organizationID uint, now time.Time) ([]model.B2BReminderJob, error) {
	var jobs []model.B2BReminderJob
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND status = ? AND due_at <= ?", organizationID, model.B2BReminderJobStatusPending, now).
		Order("due_at ASC").
		Find(&jobs).Error
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *B2BRepository) MarkReminderJobSent(ctx context.Context, jobID uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.B2BReminderJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":        model.B2BReminderJobStatusSent,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"last_error":    "",
			"sent_at":       now,
			"updated_at":    now,
		}).Error
}

func (r *B2BRepository) MarkReminderJobFailed(ctx context.Context, jobID uint, message string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.B2BReminderJob{}).
		Where("id = ?", jobID).
		Updates(map[string]any{
			"status":        model.B2BReminderJobStatusFailed,
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"last_error":    strings.TrimSpace(message),
			"updated_at":    now,
		}).Error
}

func (r *B2BRepository) UpsertSSOConfig(ctx context.Context, cfg *model.B2BSSOConfig) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}},
		DoUpdates: clause.Assignments(map[string]any{
			"provider":        cfg.Provider,
			"issuer_url":      cfg.IssuerURL,
			"entrypoint_url":  cfg.EntrypointURL,
			"audience":        cfg.Audience,
			"certificate_pem": cfg.CertificatePEM,
			"is_enabled":      cfg.IsEnabled,
			"enforce_sso":     cfg.EnforceSSO,
			"metadata_json":   cfg.MetadataJSON,
			"updated_by":      cfg.UpdatedBy,
			"updated_at":      time.Now(),
		}),
	}).Create(cfg).Error
}

func (r *B2BRepository) GetSSOConfigByOrganizationID(ctx context.Context, organizationID uint) (*model.B2BSSOConfig, error) {
	var cfg model.B2BSSOConfig
	err := r.db.WithContext(ctx).Where("organization_id = ?", organizationID).First(&cfg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *B2BRepository) UpsertPricingRecommendation(ctx context.Context, recommendation *model.B2BPricingRecommendation) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "organization_id"}, {Name: "generated_for_date"}},
		DoUpdates: clause.Assignments(map[string]any{
			"recommended_plan_id":       recommendation.RecommendedPlanID,
			"recommended_billing_cycle": recommendation.RecommendedBillingCycle,
			"recommended_seats":         recommendation.RecommendedSeats,
			"estimated_monthly_cost":    recommendation.EstimatedMonthlyCost,
			"estimated_yearly_saving":   recommendation.EstimatedYearlySaving,
			"confidence_score":          recommendation.ConfidenceScore,
			"reasons_json":              recommendation.ReasonsJSON,
			"updated_at":                time.Now(),
		}),
	}).Create(recommendation).Error
}

func (r *B2BRepository) GetLatestPricingRecommendation(ctx context.Context, organizationID uint) (*model.B2BPricingRecommendation, error) {
	var recommendation model.B2BPricingRecommendation
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", organizationID).
		Order("generated_for_date DESC").
		First(&recommendation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &recommendation, nil
}
