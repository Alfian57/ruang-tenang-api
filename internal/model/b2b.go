package model

import "time"

type OrganizationStatus string

const (
	OrganizationStatusActive    OrganizationStatus = "active"
	OrganizationStatusInactive  OrganizationStatus = "inactive"
	OrganizationStatusSuspended OrganizationStatus = "suspended"
)

type OrganizationMemberRole string

const (
	OrganizationMemberRoleOwner  OrganizationMemberRole = "owner"
	OrganizationMemberRoleAdmin  OrganizationMemberRole = "admin"
	OrganizationMemberRoleMember OrganizationMemberRole = "member"
)

type OrganizationMemberStatus string

const (
	OrganizationMemberStatusInvited         OrganizationMemberStatus = "invited"
	OrganizationMemberStatusPendingApproval OrganizationMemberStatus = "pending_approval"
	OrganizationMemberStatusActive          OrganizationMemberStatus = "active"
	OrganizationMemberStatusRemoved         OrganizationMemberStatus = "removed"
)

type B2BBillingCycle string

const (
	B2BBillingCycleMonthly B2BBillingCycle = "monthly"
	B2BBillingCycleYearly  B2BBillingCycle = "yearly"
)

type B2BSubscriptionStatus string

const (
	B2BSubscriptionStatusDraft     B2BSubscriptionStatus = "draft"
	B2BSubscriptionStatusActive    B2BSubscriptionStatus = "active"
	B2BSubscriptionStatusSuspended B2BSubscriptionStatus = "suspended"
	B2BSubscriptionStatusExpired   B2BSubscriptionStatus = "expired"
	B2BSubscriptionStatusCanceled  B2BSubscriptionStatus = "canceled"
)

type B2BQuoteStatus string

const (
	B2BQuoteStatusDraft    B2BQuoteStatus = "draft"
	B2BQuoteStatusAccepted B2BQuoteStatus = "accepted"
	B2BQuoteStatusExpired  B2BQuoteStatus = "expired"
)

type Organization struct {
	ID                     uint               `gorm:"primaryKey" json:"id"`
	Code                   string             `gorm:"size:60;not null;uniqueIndex" json:"code"`
	Name                   string             `gorm:"size:150;not null" json:"name"`
	BusinessType           string             `gorm:"size:50;not null;default:'general'" json:"business_type"`
	ContactEmail           string             `gorm:"size:255;not null" json:"contact_email"`
	Status                 OrganizationStatus `gorm:"size:20;not null;default:'active'" json:"status"`
	RequiresMemberApproval bool               `gorm:"not null;default:true" json:"requires_member_approval"`
	CreatedBy              *uint              `json:"created_by,omitempty"`
	CreatedAt              time.Time          `json:"created_at"`
	UpdatedAt              time.Time          `json:"updated_at"`
}

func (Organization) TableName() string {
	return "organizations"
}

type B2BPlan struct {
	ID               uint            `gorm:"primaryKey" json:"id"`
	Code             string          `gorm:"size:60;not null;uniqueIndex" json:"code"`
	Name             string          `gorm:"size:120;not null" json:"name"`
	Description      string          `gorm:"type:text" json:"description"`
	BillingCycle     B2BBillingCycle `gorm:"size:20;not null" json:"billing_cycle"`
	BasePricePerSeat int64           `gorm:"not null" json:"base_price_per_seat"`
	MinSeats         int             `gorm:"not null;default:1" json:"min_seats"`
	MaxSeats         int             `gorm:"not null;default:100000" json:"max_seats"`
	FeaturesJSON     string          `gorm:"type:jsonb;not null;default:'{}'" json:"features_json"`
	IsActive         bool            `gorm:"not null;default:true" json:"is_active"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (B2BPlan) TableName() string {
	return "b2b_plans"
}

type B2BSubscription struct {
	ID              uint                  `gorm:"primaryKey" json:"id"`
	OrganizationID  uint                  `gorm:"not null" json:"organization_id"`
	PlanID          uint                  `gorm:"not null" json:"plan_id"`
	Status          B2BSubscriptionStatus `gorm:"size:20;not null;default:'draft'" json:"status"`
	ContractedSeats int                   `gorm:"not null" json:"contracted_seats"`
	UsedSeats       int                   `gorm:"not null;default:0" json:"used_seats"`
	BillingCycle    B2BBillingCycle       `gorm:"size:20;not null" json:"billing_cycle"`
	UnitPrice       int64                 `gorm:"not null" json:"unit_price"`
	Subtotal        int64                 `gorm:"not null" json:"subtotal"`
	DiscountAmount  int64                 `gorm:"not null;default:0" json:"discount_amount"`
	TotalAmount     int64                 `gorm:"not null" json:"total_amount"`
	StartsAt        time.Time             `json:"starts_at"`
	EndsAt          time.Time             `json:"ends_at"`
	ActivatedAt     *time.Time            `json:"activated_at,omitempty"`
	MetadataJSON    string                `gorm:"type:jsonb;not null;default:'{}'" json:"metadata_json"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`

	Organization Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Plan         B2BPlan      `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (B2BSubscription) TableName() string {
	return "b2b_subscriptions"
}

func (s *B2BSubscription) IsActiveAt(now time.Time) bool {
	if s == nil {
		return false
	}
	return s.Status == B2BSubscriptionStatusActive && !now.Before(s.StartsAt) && now.Before(s.EndsAt)
}

type OrganizationMember struct {
	ID                uint                     `gorm:"primaryKey" json:"id"`
	OrganizationID    uint                     `gorm:"not null" json:"organization_id"`
	UserID            *uint                    `json:"user_id,omitempty"`
	Email             string                   `gorm:"size:255;not null" json:"email"`
	FullName          string                   `gorm:"size:150" json:"full_name"`
	Role              OrganizationMemberRole   `gorm:"size:20;not null;default:'member'" json:"role"`
	Status            OrganizationMemberStatus `gorm:"size:20;not null;default:'invited'" json:"status"`
	InvitationToken   string                   `gorm:"size:120" json:"-"`
	InvitationExpires *time.Time               `gorm:"column:invitation_expires_at" json:"invitation_expires_at,omitempty"`
	InvitedBy         *uint                    `json:"invited_by,omitempty"`
	InvitedAt         *time.Time               `json:"invited_at,omitempty"`
	JoinedAt          *time.Time               `json:"joined_at,omitempty"`
	RemovedAt         *time.Time               `json:"removed_at,omitempty"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`

	Organization Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}

func (m *OrganizationMember) CanManageMembers() bool {
	if m == nil || m.Status != OrganizationMemberStatusActive {
		return false
	}
	return m.Role == OrganizationMemberRoleOwner || m.Role == OrganizationMemberRoleAdmin
}

type B2BSeatAllocation struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	SubscriptionID       uint       `gorm:"not null" json:"subscription_id"`
	OrganizationMemberID uint       `gorm:"not null" json:"organization_member_id"`
	AllocatedAt          time.Time  `json:"allocated_at"`
	ReleasedAt           *time.Time `json:"released_at,omitempty"`
	ReleaseReason        string     `gorm:"type:text" json:"release_reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (B2BSeatAllocation) TableName() string {
	return "b2b_seat_allocations"
}

type B2BPricingQuote struct {
	ID                   uint            `gorm:"primaryKey" json:"id"`
	QuoteCode            string          `gorm:"size:60;not null;uniqueIndex" json:"quote_code"`
	OrganizationID       *uint           `json:"organization_id,omitempty"`
	PlanID               *uint           `json:"plan_id,omitempty"`
	RequestedSeats       int             `gorm:"not null" json:"requested_seats"`
	BillingCycle         B2BBillingCycle `gorm:"size:20;not null" json:"billing_cycle"`
	SelectedAddOnsJSON   string          `gorm:"column:selected_addons_json;type:jsonb;not null;default:'[]'" json:"selected_addons_json"`
	BasePricePerSeat     int64           `gorm:"not null" json:"base_price_per_seat"`
	GrossAmount          int64           `gorm:"not null" json:"gross_amount"`
	VolumeDiscountAmount int64           `gorm:"not null;default:0" json:"volume_discount_amount"`
	AnnualDiscountAmount int64           `gorm:"not null;default:0" json:"annual_discount_amount"`
	AddOnAmount          int64           `gorm:"not null;default:0" json:"add_on_amount"`
	FinalAmount          int64           `gorm:"not null" json:"final_amount"`
	Currency             string          `gorm:"size:10;not null;default:'IDR'" json:"currency"`
	ValidUntil           time.Time       `json:"valid_until"`
	Status               B2BQuoteStatus  `gorm:"size:20;not null;default:'draft'" json:"status"`
	CreatedBy            *uint           `json:"created_by,omitempty"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (B2BPricingQuote) TableName() string {
	return "b2b_pricing_quotes"
}

type B2BBillingHistory struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	SubscriptionID     uint       `gorm:"not null" json:"subscription_id"`
	InvoiceNumber      string     `gorm:"size:100;not null;uniqueIndex" json:"invoice_number"`
	BillingPeriodStart time.Time  `json:"billing_period_start"`
	BillingPeriodEnd   time.Time  `json:"billing_period_end"`
	SeatsBilled        int        `gorm:"not null" json:"seats_billed"`
	Amount             int64      `gorm:"not null" json:"amount"`
	Status             string     `gorm:"size:20;not null;default:'pending'" json:"status"`
	PaidAt             *time.Time `json:"paid_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (B2BBillingHistory) TableName() string {
	return "b2b_billing_histories"
}

type OrganizationMemberApprovalStatus string

const (
	OrganizationMemberApprovalStatusPending  OrganizationMemberApprovalStatus = "pending"
	OrganizationMemberApprovalStatusApproved OrganizationMemberApprovalStatus = "approved"
	OrganizationMemberApprovalStatusRejected OrganizationMemberApprovalStatus = "rejected"
)

type OrganizationMemberApproval struct {
	ID                   uint                             `gorm:"primaryKey" json:"id"`
	OrganizationID       uint                             `gorm:"not null" json:"organization_id"`
	OrganizationMemberID uint                             `gorm:"not null" json:"organization_member_id"`
	RequestedBy          *uint                            `json:"requested_by,omitempty"`
	ApproverUserID       *uint                            `json:"approver_user_id,omitempty"`
	Status               OrganizationMemberApprovalStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	Note                 string                           `gorm:"type:text" json:"note,omitempty"`
	DecidedAt            *time.Time                       `json:"decided_at,omitempty"`
	CreatedAt            time.Time                        `json:"created_at"`
	UpdatedAt            time.Time                        `json:"updated_at"`
}

func (OrganizationMemberApproval) TableName() string {
	return "organization_member_approvals"
}

type OrganizationAuditLog struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	OrganizationID uint      `gorm:"not null" json:"organization_id"`
	ActorUserID    *uint     `json:"actor_user_id,omitempty"`
	Action         string    `gorm:"size:80;not null" json:"action"`
	EntityType     string    `gorm:"size:80;not null" json:"entity_type"`
	EntityID       string    `gorm:"size:120" json:"entity_id,omitempty"`
	MetadataJSON   string    `gorm:"type:jsonb;not null;default:'{}'" json:"metadata_json"`
	CreatedAt      time.Time `json:"created_at"`
}

func (OrganizationAuditLog) TableName() string {
	return "organization_audit_logs"
}

type OrganizationOnboardingTemplate struct {
	ID             uint                   `gorm:"primaryKey" json:"id"`
	OrganizationID uint                   `gorm:"not null" json:"organization_id"`
	Role           OrganizationMemberRole `gorm:"size:20;not null" json:"role"`
	Title          string                 `gorm:"size:150;not null" json:"title"`
	WelcomeMessage string                 `gorm:"type:text" json:"welcome_message,omitempty"`
	ChecklistJSON  string                 `gorm:"type:jsonb;not null;default:'[]'" json:"checklist_json"`
	IsDefault      bool                   `gorm:"not null;default:false" json:"is_default"`
	IsActive       bool                   `gorm:"not null;default:true" json:"is_active"`
	CreatedBy      *uint                  `json:"created_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func (OrganizationOnboardingTemplate) TableName() string {
	return "organization_onboarding_templates"
}

type B2BUsageDailyMetric struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	OrganizationID   uint      `gorm:"not null" json:"organization_id"`
	MetricDate       time.Time `gorm:"type:date;not null" json:"metric_date"`
	ActiveMembers    int       `gorm:"not null;default:0" json:"active_members"`
	InvitedMembers   int       `gorm:"not null;default:0" json:"invited_members"`
	PendingApprovals int       `gorm:"not null;default:0" json:"pending_approvals"`
	ContractedSeats  int       `gorm:"not null;default:0" json:"contracted_seats"`
	UsedSeats        int       `gorm:"not null;default:0" json:"used_seats"`
	MessagesSent     int       `gorm:"not null;default:0" json:"messages_sent"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (B2BUsageDailyMetric) TableName() string {
	return "b2b_usage_daily_metrics"
}

type B2BReminderJobStatus string

const (
	B2BReminderJobStatusPending B2BReminderJobStatus = "pending"
	B2BReminderJobStatusSent    B2BReminderJobStatus = "sent"
	B2BReminderJobStatusFailed  B2BReminderJobStatus = "failed"
)

type B2BReminderJobType string

const (
	B2BReminderJobTypeSeatThreshold      B2BReminderJobType = "seat_threshold"
	B2BReminderJobTypeSubscriptionExpiry B2BReminderJobType = "subscription_expiry"
	B2BReminderJobTypeInvoiceDue         B2BReminderJobType = "invoice_due"
)

type B2BReminderJob struct {
	ID             uint                 `gorm:"primaryKey" json:"id"`
	OrganizationID uint                 `gorm:"not null" json:"organization_id"`
	SubscriptionID *uint                `json:"subscription_id,omitempty"`
	JobType        B2BReminderJobType   `gorm:"size:50;not null" json:"job_type"`
	Status         B2BReminderJobStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	DueAt          time.Time            `json:"due_at"`
	PayloadJSON    string               `gorm:"type:jsonb;not null;default:'{}'" json:"payload_json"`
	AttemptCount   int                  `gorm:"not null;default:0" json:"attempt_count"`
	LastError      string               `gorm:"type:text" json:"last_error,omitempty"`
	SentAt         *time.Time           `json:"sent_at,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func (B2BReminderJob) TableName() string {
	return "b2b_reminder_jobs"
}

type B2BSSOProvider string

const (
	B2BSSOProviderSAML            B2BSSOProvider = "saml"
	B2BSSOProviderOIDC            B2BSSOProvider = "oidc"
	B2BSSOProviderGoogleWorkspace B2BSSOProvider = "google_workspace"
	B2BSSOProviderAzureAD         B2BSSOProvider = "azure_ad"
	B2BSSOProviderOkta            B2BSSOProvider = "okta"
)

type B2BSSOConfig struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	OrganizationID uint           `gorm:"not null;uniqueIndex" json:"organization_id"`
	Provider       B2BSSOProvider `gorm:"size:30" json:"provider"`
	IssuerURL      string         `gorm:"size:500" json:"issuer_url,omitempty"`
	EntrypointURL  string         `gorm:"size:500" json:"entrypoint_url,omitempty"`
	Audience       string         `gorm:"size:255" json:"audience,omitempty"`
	CertificatePEM string         `gorm:"type:text" json:"certificate_pem,omitempty"`
	IsEnabled      bool           `gorm:"not null;default:false" json:"is_enabled"`
	EnforceSSO     bool           `gorm:"not null;default:false" json:"enforce_sso"`
	MetadataJSON   string         `gorm:"type:jsonb;not null;default:'{}'" json:"metadata_json"`
	CreatedBy      *uint          `json:"created_by,omitempty"`
	UpdatedBy      *uint          `json:"updated_by,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (B2BSSOConfig) TableName() string {
	return "b2b_sso_configs"
}

type B2BPricingRecommendation struct {
	ID                      uint            `gorm:"primaryKey" json:"id"`
	OrganizationID          uint            `gorm:"not null" json:"organization_id"`
	GeneratedForDate        time.Time       `gorm:"type:date;not null" json:"generated_for_date"`
	RecommendedPlanID       *uint           `json:"recommended_plan_id,omitempty"`
	RecommendedBillingCycle B2BBillingCycle `gorm:"size:20;not null" json:"recommended_billing_cycle"`
	RecommendedSeats        int             `gorm:"not null" json:"recommended_seats"`
	EstimatedMonthlyCost    int64           `gorm:"not null;default:0" json:"estimated_monthly_cost"`
	EstimatedYearlySaving   int64           `gorm:"not null;default:0" json:"estimated_yearly_saving"`
	ConfidenceScore         float64         `gorm:"type:numeric(5,2);not null;default:0" json:"confidence_score"`
	ReasonsJSON             string          `gorm:"type:jsonb;not null;default:'[]'" json:"reasons_json"`
	CreatedAt               time.Time       `json:"created_at"`
	UpdatedAt               time.Time       `json:"updated_at"`
}

func (B2BPricingRecommendation) TableName() string {
	return "b2b_pricing_recommendations"
}
