package dto

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type MidtransProviderID string

func (id *MidtransProviderID) UnmarshalJSON(data []byte) error {
	value := strings.TrimSpace(string(data))
	if value == "null" || value == "" {
		*id = ""
		return nil
	}
	if strings.HasPrefix(value, "\"") {
		var decoded string
		if err := json.Unmarshal(data, &decoded); err != nil {
			return err
		}
		*id = MidtransProviderID(strings.TrimSpace(decoded))
		return nil
	}
	if _, err := strconv.ParseUint(value, 10, 64); err != nil {
		return fmt.Errorf("invalid Midtrans provider ID")
	}
	*id = MidtransProviderID(value)
	return nil
}

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
	BusinessPlans []B2BPlanDTO      `json:"business_plans"`
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
	IsPremium         bool                        `json:"is_premium"`
	EntitlementSource string                      `json:"entitlement_source"`
	B2BOrganizationID *uint                       `json:"b2b_organization_id,omitempty"`
	PremiumSince      *time.Time                  `json:"premium_since,omitempty"`
	PremiumExpiresAt  *time.Time                  `json:"premium_expires_at,omitempty"`
	GoldCoins         int64                       `json:"gold_coins"`
	ChatQuota         ChatQuotaDTO                `json:"chat_quota"`
	Subscription      *BillingSubscriptionInfoDTO `json:"subscription,omitempty"`
}

type PaymentTransactionDTO struct {
	ID                           uint       `json:"id"`
	OrderID                      string     `json:"order_id"`
	UserID                       uint       `json:"user_id"`
	ItemType                     string     `json:"item_type"`
	ItemID                       uint       `json:"item_id"`
	ItemName                     string     `json:"item_name"`
	Amount                       int        `json:"amount"`
	Currency                     string     `json:"currency"`
	Status                       string     `json:"status"`
	PaymentProvider              string     `json:"payment_provider"`
	ProviderTransactionID        string     `json:"provider_transaction_id,omitempty"`
	ProviderPaymentType          string     `json:"provider_payment_type,omitempty"`
	FailureReason                string     `json:"failure_reason,omitempty"`
	SnapToken                    string     `json:"snap_token,omitempty"`
	SnapURL                      string     `json:"snap_url,omitempty"`
	PaidAt                       *time.Time `json:"paid_at,omitempty"`
	RefundedAmount               int64      `json:"refunded_amount"`
	RefundRequestedAmount        int64      `json:"refund_requested_amount"`
	ProviderRefundAmountReported int64      `json:"provider_refund_amount_reported"`
	RefundStatus                 string     `json:"refund_status"`
	RefundReconciliationStatus   string     `json:"refund_reconciliation_status"`
	RefundReconciliationReason   string     `json:"refund_reconciliation_reason,omitempty"`
	CreatedAt                    time.Time  `json:"created_at"`
	UpdatedAt                    time.Time  `json:"updated_at"`
}

type MidtransWebhookRequest struct {
	OrderID            string                 `json:"order_id"`
	StatusCode         string                 `json:"status_code"`
	GrossAmount        string                 `json:"gross_amount"`
	SignatureKey       string                 `json:"signature_key"`
	TransactionStatus  string                 `json:"transaction_status"`
	FraudStatus        string                 `json:"fraud_status"`
	PaymentType        string                 `json:"payment_type"`
	TransactionID      string                 `json:"transaction_id"`
	TransactionTime    string                 `json:"transaction_time"`
	SettlementTime     string                 `json:"settlement_time"`
	RefundAmount       string                 `json:"refund_amount"`
	RefundChargebackID MidtransProviderID     `json:"refund_chargeback_id" swaggertype:"string"`
	RefundKey          string                 `json:"refund_key"`
	RefundReason       string                 `json:"reason"`
	RefundMethod       string                 `json:"refund_method"`
	BankConfirmedAt    string                 `json:"bank_confirmed_at"`
	Refunds            []MidtransRefundDetail `json:"refunds"`
}

type MidtransRefundDetail struct {
	RefundChargebackID MidtransProviderID `json:"refund_chargeback_id" swaggertype:"string"`
	RefundAmount       string             `json:"refund_amount"`
	CreatedAt          string             `json:"created_at"`
	Reason             string             `json:"reason"`
	RefundKey          string             `json:"refund_key"`
	RefundMethod       string             `json:"refund_method"`
	BankConfirmedAt    string             `json:"bank_confirmed_at"`
}

type AdminRefundRequest struct {
	Amount int64  `json:"amount" binding:"required,gt=0"`
	Reason string `json:"reason" binding:"required,max=255"`
}

type AdminRefundResponse struct {
	OrderID   string `json:"order_id"`
	RefundKey string `json:"refund_key"`
	Amount    int64  `json:"amount"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

type AdminRefundReconciliationRequest struct {
	Action                     string `json:"action" binding:"required,oneof=deduct_remaining_coins accept_consumed_coins revoke_premium_days revoke_refunded_subscription retain_entitlement mark_refund_rejected complete_manual_review"`
	PremiumDaysToRevoke        int    `json:"premium_days_to_revoke" binding:"omitempty,gte=0,lte=3650"`
	ProviderRejectionConfirmed bool   `json:"provider_rejection_confirmed"`
	ManualReviewConfirmed      bool   `json:"manual_review_confirmed"`
	Note                       string `json:"note" binding:"required,min=3,max=1000"`
}

type AdminRefundReconciliationResponse struct {
	OrderID                    string `json:"order_id"`
	RefundReconciliationStatus string `json:"refund_reconciliation_status"`
	RefundReconciliationReason string `json:"refund_reconciliation_reason"`
	CoinsReversed              int64  `json:"coins_reversed"`
	CoinsWrittenOff            int64  `json:"coins_written_off"`
	PremiumDaysReduced         int    `json:"premium_days_reduced"`
}
