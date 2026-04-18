package dto

import "time"

type PremiumPlanDTO struct {
	ID           uint   `json:"id"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int    `json:"price"`
	DurationDays int    `json:"duration_days"`
	IsActive     bool   `json:"is_active"`
}

type TopupPackageDTO struct {
	ID         uint   `json:"id"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	Coins      int64  `json:"coins"`
	BonusCoins int64  `json:"bonus_coins"`
	TotalCoins int64  `json:"total_coins"`
	Price      int    `json:"price"`
	IsActive   bool   `json:"is_active"`
}

type ChatQuotaDTO struct {
	FeatureKey  string `json:"feature_key"`
	Limit       int    `json:"limit"`
	Used        int    `json:"used"`
	Remaining   int    `json:"remaining"`
	IsUnlimited bool   `json:"is_unlimited"`
	ResetAt     string `json:"reset_at"`
}

type BillingCatalogResponse struct {
	Plans         []PremiumPlanDTO  `json:"plans"`
	TopupPackages []TopupPackageDTO `json:"topup_packages"`
	ChatQuota     ChatQuotaDTO      `json:"chat_quota"`
}

type CreateCheckoutRequest struct {
	ItemType string `json:"item_type" binding:"required,oneof=subscription topup"`
	ItemID   uint   `json:"item_id" binding:"required"`
}

type CreateCheckoutResponse struct {
	TransactionID uint       `json:"transaction_id"`
	OrderID       string     `json:"order_id"`
	ItemType      string     `json:"item_type"`
	ItemID        uint       `json:"item_id"`
	ItemName      string     `json:"item_name"`
	Amount        int        `json:"amount"`
	Currency      string     `json:"currency"`
	Status        string     `json:"status"`
	SnapToken     string     `json:"snap_token"`
	SnapURL       string     `json:"snap_url"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

type BillingSubscriptionInfoDTO struct {
	PlanID        uint      `json:"plan_id"`
	PlanCode      string    `json:"plan_code"`
	PlanName      string    `json:"plan_name"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	SourceOrderID string    `json:"source_order_id"`
}

type BillingStatusResponse struct {
	IsPremium        bool                        `json:"is_premium"`
	PremiumSince     *time.Time                  `json:"premium_since,omitempty"`
	PremiumExpiresAt *time.Time                  `json:"premium_expires_at,omitempty"`
	GoldCoins        int64                       `json:"gold_coins"`
	ChatQuota        ChatQuotaDTO                `json:"chat_quota"`
	Subscription     *BillingSubscriptionInfoDTO `json:"subscription,omitempty"`
}

type PaymentTransactionDTO struct {
	ID                    uint       `json:"id"`
	OrderID               string     `json:"order_id"`
	UserID                uint       `json:"user_id"`
	ItemType              string     `json:"item_type"`
	ItemID                uint       `json:"item_id"`
	ItemName              string     `json:"item_name"`
	Amount                int        `json:"amount"`
	Currency              string     `json:"currency"`
	Status                string     `json:"status"`
	PaymentProvider       string     `json:"payment_provider"`
	ProviderTransactionID string     `json:"provider_transaction_id,omitempty"`
	ProviderPaymentType   string     `json:"provider_payment_type,omitempty"`
	FailureReason         string     `json:"failure_reason,omitempty"`
	PaidAt                *time.Time `json:"paid_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type MidtransWebhookRequest struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	PaymentType       string `json:"payment_type"`
	TransactionID     string `json:"transaction_id"`
	SettlementTime    string `json:"settlement_time"`
}
