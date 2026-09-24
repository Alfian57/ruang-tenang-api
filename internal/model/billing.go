package model

import "time"

type BillingItemType string

const (
	BillingItemTypeSubscription BillingItemType = "subscription"
	BillingItemTypeTopup        BillingItemType = "topup"
)

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusExpired  PaymentStatus = "expired"
	PaymentStatusCanceled PaymentStatus = "canceled"
	PaymentStatusRefunded PaymentStatus = "refunded"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusExpired  SubscriptionStatus = "expired"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
)

const FeatureKeyChatAIMessages = "chat_ai_messages"

type PremiumPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Code         string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name         string    `gorm:"size:100;not null" json:"name"`
	Description  string    `gorm:"type:text" json:"description"`
	Price        int       `gorm:"not null" json:"price"`
	DurationDays int       `gorm:"not null" json:"duration_days"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (PremiumPlan) TableName() string {
	return "premium_plans"
}

type TopupPackage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Code       string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Coins      int64     `gorm:"not null" json:"coins"`
	BonusCoins int64     `gorm:"default:0" json:"bonus_coins"`
	Price      int       `gorm:"not null" json:"price"`
	IsActive   bool      `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (TopupPackage) TableName() string {
	return "topup_packages"
}

func (p *TopupPackage) TotalCoins() int64 {
	return p.Coins + p.BonusCoins
}

type PaymentTransaction struct {
	ID                           uint            `gorm:"primaryKey" json:"id"`
	OrderID                      string          `gorm:"size:100;not null;uniqueIndex" json:"order_id"`
	UserID                       uint            `gorm:"not null" json:"user_id"`
	ItemType                     BillingItemType `gorm:"size:20;not null" json:"item_type"`
	ItemID                       uint            `gorm:"not null" json:"item_id"`
	ItemName                     string          `gorm:"size:150;not null" json:"item_name"`
	Amount                       int             `gorm:"not null" json:"amount"`
	Currency                     string          `gorm:"size:10;not null;default:'IDR'" json:"currency"`
	Status                       PaymentStatus   `gorm:"size:20;not null;default:'pending'" json:"status"`
	PaymentProvider              string          `gorm:"size:20;not null;default:'midtrans'" json:"payment_provider"`
	ProviderTransactionID        string          `gorm:"size:120" json:"provider_transaction_id"`
	ProviderPaymentType          string          `gorm:"size:50" json:"provider_payment_type"`
	SnapToken                    string          `gorm:"type:text" json:"snap_token"`
	SnapRedirectURL              string          `gorm:"type:text" json:"snap_redirect_url"`
	CallbackPayload              string          `gorm:"type:text" json:"callback_payload"`
	FailureReason                string          `gorm:"type:text" json:"failure_reason"`
	PaidAt                       *time.Time      `json:"paid_at,omitempty"`
	ExpiresAt                    *time.Time      `json:"expires_at,omitempty"`
	RefundedAmount               int64           `gorm:"not null;default:0" json:"refunded_amount"`
	RefundRequestedAmount        int64           `gorm:"not null;default:0" json:"refund_requested_amount"`
	ProviderRefundAmountReported int64           `gorm:"not null;default:0" json:"provider_refund_amount_reported"`
	RefundStatus                 string          `gorm:"size:32;not null;default:'none'" json:"refund_status"`
	RefundReconciliationStatus   string          `gorm:"size:16;not null;default:'not_required'" json:"refund_reconciliation_status"`
	RefundReconciliationReason   string          `gorm:"type:text;not null;default:''" json:"refund_reconciliation_reason,omitempty"`
	CoinsReversed                int64           `gorm:"not null;default:0" json:"coins_reversed"`
	CoinsWrittenOff              int64           `gorm:"not null;default:0" json:"coins_written_off"`
	CreatedAt                    time.Time       `json:"created_at"`
	UpdatedAt                    time.Time       `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

func (t *PaymentTransaction) IsFinalStatus() bool {
	return t.Status == PaymentStatusPaid ||
		t.Status == PaymentStatusFailed ||
		t.Status == PaymentStatusExpired ||
		t.Status == PaymentStatusCanceled ||
		t.Status == PaymentStatusRefunded
}

type UserSubscription struct {
	ID            uint               `gorm:"primaryKey" json:"id"`
	UserID        uint               `gorm:"not null" json:"user_id"`
	PlanID        uint               `gorm:"not null" json:"plan_id"`
	SourceOrderID string             `gorm:"size:100" json:"source_order_id"`
	Status        SubscriptionStatus `gorm:"size:20;not null;default:'active'" json:"status"`
	StartsAt      time.Time          `json:"starts_at"`
	EndsAt        time.Time          `json:"ends_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`

	User User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Plan PremiumPlan `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (UserSubscription) TableName() string {
	return "user_subscriptions"
}

type PaymentRefund struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	PaymentTransactionID uint       `gorm:"not null;uniqueIndex:uq_payment_refunds_refund_key,priority:1;uniqueIndex:uq_payment_refunds_provider_id,priority:1" json:"payment_transaction_id"`
	RefundKey            string     `gorm:"size:120;not null;uniqueIndex:uq_payment_refunds_refund_key,priority:2" json:"refund_key"`
	ProviderRefundID     string     `gorm:"size:120;not null;default:'';uniqueIndex:uq_payment_refunds_provider_id,priority:2" json:"provider_refund_id,omitempty"`
	Amount               int64      `gorm:"not null" json:"amount"`
	Reason               string     `gorm:"size:255;not null;default:''" json:"reason,omitempty"`
	RefundMethod         string     `gorm:"size:32;not null;default:''" json:"refund_method,omitempty"`
	Status               string     `gorm:"size:32;not null;default:'requested'" json:"status"`
	RequestedBy          *uint      `json:"requested_by,omitempty"`
	RequestedAt          time.Time  `gorm:"not null" json:"requested_at"`
	ProviderCreatedAt    *time.Time `json:"provider_created_at,omitempty"`
	BankConfirmedAt      *time.Time `json:"bank_confirmed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (PaymentRefund) TableName() string { return "payment_refunds" }

type PaymentRefundReconciliationEvent struct {
	ID                   uint      `gorm:"primaryKey" json:"id"`
	PaymentTransactionID uint      `gorm:"not null;index" json:"payment_transaction_id"`
	ActorUserID          *uint     `json:"actor_user_id,omitempty"`
	Action               string    `gorm:"size:40;not null" json:"action"`
	Note                 string    `gorm:"type:text;not null" json:"note"`
	RefundedAmount       int64     `gorm:"not null;default:0" json:"refunded_amount"`
	CoinsReversed        int64     `gorm:"not null;default:0" json:"coins_reversed"`
	CoinsWrittenOff      int64     `gorm:"not null;default:0" json:"coins_written_off"`
	PremiumDaysReduced   int       `gorm:"not null;default:0" json:"premium_days_reduced"`
	CreatedAt            time.Time `json:"created_at"`
}

func (PaymentRefundReconciliationEvent) TableName() string {
	return "payment_refund_reconciliation_events"
}

func (s *UserSubscription) IsActive(now time.Time) bool {
	return s.Status == SubscriptionStatusActive && s.EndsAt.After(now)
}

type PaymentWebhookEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Provider    string    `gorm:"size:20;not null" json:"provider"`
	OrderID     string    `gorm:"size:100;not null" json:"order_id"`
	EventKey    string    `gorm:"size:255;not null;uniqueIndex" json:"event_key"`
	Payload     string    `gorm:"type:text;not null" json:"payload"`
	ProcessedAt time.Time `json:"processed_at"`
	CreatedAt   time.Time `json:"created_at"`
}

func (PaymentWebhookEvent) TableName() string {
	return "payment_webhook_events"
}

type UserFeatureUsage struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	UserID           uint      `gorm:"not null" json:"user_id"`
	FeatureKey       string    `gorm:"size:50;not null" json:"feature_key"`
	UsageDate        time.Time `gorm:"type:date;not null" json:"usage_date"`
	UsageWindowStart time.Time `gorm:"not null" json:"usage_window_start"`
	UsedCount        int       `gorm:"not null;default:0" json:"used_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (UserFeatureUsage) TableName() string {
	return "user_feature_usages"
}
