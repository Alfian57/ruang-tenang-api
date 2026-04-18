package dto

import "time"

type CreateOrganizationRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=150"`
	Code         string `json:"code" binding:"omitempty,min=3,max=60"`
	BusinessType string `json:"business_type" binding:"omitempty,max=50"`
	ContactEmail string `json:"contact_email" binding:"required,email"`
}

type OrganizationDTO struct {
	ID                     uint   `json:"id"`
	Code                   string `json:"code"`
	Name                   string `json:"name"`
	BusinessType           string `json:"business_type"`
	ContactEmail           string `json:"contact_email"`
	Status                 string `json:"status"`
	RequiresMemberApproval bool   `json:"requires_member_approval"`
}

type OrganizationSeatUsageDTO struct {
	ContractedSeats int `json:"contracted_seats"`
	UsedSeats       int `json:"used_seats"`
	AvailableSeats  int `json:"available_seats"`
}

type B2BSubscriptionDTO struct {
	ID              uint       `json:"id"`
	OrganizationID  uint       `json:"organization_id"`
	PlanID          uint       `json:"plan_id"`
	PlanCode        string     `json:"plan_code"`
	PlanName        string     `json:"plan_name"`
	Status          string     `json:"status"`
	ContractedSeats int        `json:"contracted_seats"`
	UsedSeats       int        `json:"used_seats"`
	BillingCycle    string     `json:"billing_cycle"`
	UnitPrice       int64      `json:"unit_price"`
	Subtotal        int64      `json:"subtotal"`
	DiscountAmount  int64      `json:"discount_amount"`
	TotalAmount     int64      `json:"total_amount"`
	StartsAt        time.Time  `json:"starts_at"`
	EndsAt          time.Time  `json:"ends_at"`
	ActivatedAt     *time.Time `json:"activated_at,omitempty"`
}

type OrganizationSummaryResponse struct {
	Organization OrganizationDTO          `json:"organization"`
	Subscription *B2BSubscriptionDTO      `json:"subscription,omitempty"`
	SeatUsage    OrganizationSeatUsageDTO `json:"seat_usage"`
}

type UpsertB2BPlanRequest struct {
	Code             string         `json:"code" binding:"required,min=3,max=60"`
	Name             string         `json:"name" binding:"required,min=2,max=120"`
	Description      string         `json:"description"`
	BillingCycle     string         `json:"billing_cycle" binding:"required,oneof=monthly yearly"`
	BasePricePerSeat int64          `json:"base_price_per_seat" binding:"required,gte=0"`
	MinSeats         int            `json:"min_seats" binding:"required,gt=0"`
	MaxSeats         int            `json:"max_seats" binding:"required,gt=0"`
	Features         map[string]any `json:"features"`
	IsActive         *bool          `json:"is_active"`
}

type B2BPlanDTO struct {
	ID               uint           `json:"id"`
	Code             string         `json:"code"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	BillingCycle     string         `json:"billing_cycle"`
	BasePricePerSeat int64          `json:"base_price_per_seat"`
	MinSeats         int            `json:"min_seats"`
	MaxSeats         int            `json:"max_seats"`
	Features         map[string]any `json:"features"`
	IsActive         bool           `json:"is_active"`
}

type CreateB2BSubscriptionRequest struct {
	PlanID          uint       `json:"plan_id" binding:"required"`
	ContractedSeats int        `json:"contracted_seats" binding:"required,gt=0"`
	BillingCycle    string     `json:"billing_cycle" binding:"omitempty,oneof=monthly yearly"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
}

type InviteOrganizationMemberRequest struct {
	Email    string `json:"email" binding:"required,email"`
	FullName string `json:"full_name" binding:"omitempty,max=150"`
	Role     string `json:"role" binding:"omitempty,oneof=admin member"`
}

type BulkInviteOrganizationMembersRequest struct {
	Members []InviteOrganizationMemberRequest `json:"members" binding:"required,min=1,dive"`
}

type OrganizationMemberDTO struct {
	ID        uint       `json:"id"`
	UserID    *uint      `json:"user_id,omitempty"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	Role      string     `json:"role"`
	Status    string     `json:"status"`
	InvitedAt *time.Time `json:"invited_at,omitempty"`
	JoinedAt  *time.Time `json:"joined_at,omitempty"`
	RemovedAt *time.Time `json:"removed_at,omitempty"`
}

type InviteMemberResponse struct {
	MemberID            uint       `json:"member_id"`
	Email               string     `json:"email"`
	InvitationToken     string     `json:"invitation_token"`
	InvitationExpiresAt *time.Time `json:"invitation_expires_at,omitempty"`
}

type BulkInviteMemberResult struct {
	Email   string `json:"email"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BulkInviteMembersResponse struct {
	Total   int                      `json:"total"`
	Invited int                      `json:"invited"`
	Skipped int                      `json:"skipped"`
	Results []BulkInviteMemberResult `json:"results"`
}

type AcceptOrganizationInviteRequest struct {
	InvitationToken string `json:"invitation_token" binding:"required"`
}

type AcceptOrganizationInviteResponse struct {
	OrganizationID uint                     `json:"organization_id"`
	MemberID       uint                     `json:"member_id"`
	SeatUsage      OrganizationSeatUsageDTO `json:"seat_usage"`
}

type CreateB2BQuoteRequest struct {
	OrganizationID *uint    `json:"organization_id,omitempty"`
	PlanID         uint     `json:"plan_id" binding:"required"`
	RequestedSeats int      `json:"requested_seats" binding:"required,gt=0"`
	BillingCycle   string   `json:"billing_cycle" binding:"omitempty,oneof=monthly yearly"`
	SelectedAddOns []string `json:"selected_add_ons"`
}

type CreateB2BQuoteResponse struct {
	QuoteCode            string    `json:"quote_code"`
	OrganizationID       *uint     `json:"organization_id,omitempty"`
	PlanID               uint      `json:"plan_id"`
	RequestedSeats       int       `json:"requested_seats"`
	BillingCycle         string    `json:"billing_cycle"`
	BasePricePerSeat     int64     `json:"base_price_per_seat"`
	GrossAmount          int64     `json:"gross_amount"`
	VolumeDiscountAmount int64     `json:"volume_discount_amount"`
	AnnualDiscountAmount int64     `json:"annual_discount_amount"`
	AddOnAmount          int64     `json:"add_on_amount"`
	FinalAmount          int64     `json:"final_amount"`
	Currency             string    `json:"currency"`
	ValidUntil           time.Time `json:"valid_until"`
	SelectedAddOns       []string  `json:"selected_add_ons"`
	AppliedRules         []string  `json:"applied_rules"`
}

type MemberApprovalDecisionRequest struct {
	Note string `json:"note" binding:"omitempty,max=1000"`
}

type MemberApprovalDecisionResponse struct {
	OrganizationID uint                     `json:"organization_id"`
	MemberID       uint                     `json:"member_id"`
	Status         string                   `json:"status"`
	SeatUsage      OrganizationSeatUsageDTO `json:"seat_usage"`
	DecidedAt      time.Time                `json:"decided_at"`
	Note           string                   `json:"note,omitempty"`
}

type DailyUsageMetricDTO struct {
	MetricDate       string `json:"metric_date"`
	ActiveMembers    int    `json:"active_members"`
	InvitedMembers   int    `json:"invited_members"`
	PendingApprovals int    `json:"pending_approvals"`
	ContractedSeats  int    `json:"contracted_seats"`
	UsedSeats        int    `json:"used_seats"`
	MessagesSent     int    `json:"messages_sent"`
}

type OrganizationAnalyticsResponse struct {
	OrganizationID     uint                     `json:"organization_id"`
	WindowDays         int                      `json:"window_days"`
	MemberStatusCounts map[string]int           `json:"member_status_counts"`
	SeatUsage          OrganizationSeatUsageDTO `json:"seat_usage"`
	SeatUtilizationPct float64                  `json:"seat_utilization_pct"`
	Trend              []DailyUsageMetricDTO    `json:"trend"`
	GeneratedAt        time.Time                `json:"generated_at"`
}

type OrganizationAuditLogDTO struct {
	ID          uint           `json:"id"`
	ActorUserID *uint          `json:"actor_user_id,omitempty"`
	Action      string         `json:"action"`
	EntityType  string         `json:"entity_type"`
	EntityID    string         `json:"entity_id,omitempty"`
	Metadata    map[string]any `json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
}

type OrganizationAuditLogListResponse struct {
	Items []OrganizationAuditLogDTO `json:"items"`
	Page  int                       `json:"page"`
	Limit int                       `json:"limit"`
	Total int64                     `json:"total"`
}

type UpsertOrganizationOnboardingTemplateRequest struct {
	Role           string   `json:"role" binding:"required,oneof=owner admin member"`
	Title          string   `json:"title" binding:"required,min=2,max=150"`
	WelcomeMessage string   `json:"welcome_message" binding:"omitempty,max=2000"`
	Checklist      []string `json:"checklist"`
	IsDefault      bool     `json:"is_default"`
	IsActive       *bool    `json:"is_active"`
}

type OrganizationOnboardingTemplateDTO struct {
	ID             uint      `json:"id"`
	OrganizationID uint      `json:"organization_id"`
	Role           string    `json:"role"`
	Title          string    `json:"title"`
	WelcomeMessage string    `json:"welcome_message"`
	Checklist      []string  `json:"checklist"`
	IsDefault      bool      `json:"is_default"`
	IsActive       bool      `json:"is_active"`
	CreatedBy      *uint     `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SeatUpgradeSubscriptionRequest struct {
	ContractedSeats int    `json:"contracted_seats" binding:"required,gt=0"`
	BillingCycle    string `json:"billing_cycle" binding:"omitempty,oneof=monthly yearly"`
}

type SeatUpgradeSubscriptionResponse struct {
	Subscription *B2BSubscriptionDTO      `json:"subscription"`
	SeatUsage    OrganizationSeatUsageDTO `json:"seat_usage"`
}

type RunOrganizationRemindersResponse struct {
	OrganizationID uint      `json:"organization_id"`
	Generated      int       `json:"generated"`
	Sent           int       `json:"sent"`
	Failed         int       `json:"failed"`
	ProcessedAt    time.Time `json:"processed_at"`
}

type UpsertOrganizationSSOConfigRequest struct {
	Provider       string         `json:"provider" binding:"omitempty,oneof=saml oidc google_workspace azure_ad okta"`
	IssuerURL      string         `json:"issuer_url" binding:"omitempty,max=500"`
	EntrypointURL  string         `json:"entrypoint_url" binding:"omitempty,max=500"`
	Audience       string         `json:"audience" binding:"omitempty,max=255"`
	CertificatePEM string         `json:"certificate_pem"`
	IsEnabled      *bool          `json:"is_enabled"`
	EnforceSSO     *bool          `json:"enforce_sso"`
	Metadata       map[string]any `json:"metadata"`
}

type OrganizationSSOConfigDTO struct {
	OrganizationID uint           `json:"organization_id"`
	Provider       string         `json:"provider"`
	IssuerURL      string         `json:"issuer_url"`
	EntrypointURL  string         `json:"entrypoint_url"`
	Audience       string         `json:"audience"`
	CertificatePEM string         `json:"certificate_pem"`
	IsEnabled      bool           `json:"is_enabled"`
	EnforceSSO     bool           `json:"enforce_sso"`
	Metadata       map[string]any `json:"metadata"`
	CreatedAt      *time.Time     `json:"created_at,omitempty"`
	UpdatedAt      *time.Time     `json:"updated_at,omitempty"`
}

type PricingRecommendationResponse struct {
	OrganizationID          uint      `json:"organization_id"`
	GeneratedForDate        string    `json:"generated_for_date"`
	RecommendedPlanID       *uint     `json:"recommended_plan_id,omitempty"`
	RecommendedBillingCycle string    `json:"recommended_billing_cycle"`
	RecommendedSeats        int       `json:"recommended_seats"`
	EstimatedMonthlyCost    int64     `json:"estimated_monthly_cost"`
	EstimatedYearlySaving   int64     `json:"estimated_yearly_saving"`
	ConfidenceScore         float64   `json:"confidence_score"`
	Reasons                 []string  `json:"reasons"`
	CreatedAt               time.Time `json:"created_at"`
}
