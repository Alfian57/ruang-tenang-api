package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/features/billing/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/Alfian57/ruang-tenang-api/internal/shared/entitlement"
	"github.com/Alfian57/ruang-tenang-api/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultDailyChatMessageLimit     = 100
	defaultChatQuotaResetInterval    = 24 * time.Hour
	minChatQuotaResetInterval        = time.Minute
	featureKeyChatAI                 = model.FeatureKeyChatAIMessages
	premiumEntitlementSourceFree     = "free"
	premiumEntitlementSourcePersonal = "personal"
	premiumEntitlementSourceB2B      = "b2b"
)

var (
	ErrItemNotFound                      = errors.New("billing item not found")
	ErrItemNotActive                     = errors.New("billing item is not active")
	ErrMidtransNotConfigured             = errors.New("midtrans is not configured")
	ErrWebhookSignatureInvalid           = errors.New("invalid webhook signature")
	ErrWebhookPayloadInvalid             = errors.New("invalid webhook payment details")
	ErrWebhookDuplicate                  = errors.New("webhook event already processed")
	ErrRefundNotAllowed                  = errors.New("refund is not allowed for this transaction")
	ErrRefundAmountInvalid               = errors.New("refund amount exceeds the refundable amount")
	ErrRefundCoinsInsufficient           = errors.New("coin balance is insufficient for this refund")
	ErrRefundSubmissionUnknown           = errors.New("refund submission status is unknown; reconcile it in Midtrans before retrying")
	ErrRefundReconciliationNotRequired   = errors.New("transaction does not require refund reconciliation")
	ErrRefundReconciliationWaiting       = errors.New("refund is waiting for Midtrans confirmation")
	ErrRefundReconciliationActionInvalid = errors.New("refund reconciliation action is not valid for this transaction")
	ErrChatQuotaExceeded                 = errors.New("chat quota exceeded")
	ErrPremiumPlanNotFound               = errors.New("premium plan not found")
	ErrTopupPackageNotFound              = errors.New("topup package not found")
	// ErrPersonalPremiumBlockedByB2B is returned when a user covered by an active
	// B2B premium seat tries to purchase a personal premium subscription.
	ErrPersonalPremiumBlockedByB2B = errors.New("personal premium is unavailable while you have active B2B premium access")
)

type ServiceConfig struct {
	MidtransServerKey string
	DefaultDailyLimit int
	ResetInterval     string
}

type Service struct {
	repo               *infrastructure.BillingRepository
	midtransClient     MidtransClient
	serverKey          string
	dailyChatLimit     int
	quotaResetInterval time.Duration
	b2bService         *B2BService
}

type premiumEntitlement struct {
	Active            bool
	Source            string
	B2BOrganizationID *uint
}

type chatQuotaWindow struct {
	Start time.Time
	End   time.Time
}

func NewService(
	repo *infrastructure.BillingRepository,
	midtransClient MidtransClient,
	cfg ServiceConfig,
) *Service {
	limit := cfg.DefaultDailyLimit
	if limit <= 0 {
		limit = defaultDailyChatMessageLimit
	}
	resetInterval := parseChatQuotaResetInterval(cfg.ResetInterval)

	return &Service{
		repo:               repo,
		midtransClient:     midtransClient,
		serverKey:          strings.TrimSpace(cfg.MidtransServerKey),
		dailyChatLimit:     limit,
		quotaResetInterval: resetInterval,
	}
}

type TransactionListParams struct {
	UserID                     *uint
	Status                     string
	ItemType                   string
	RefundReconciliationStatus string
	StartDate                  *time.Time
	EndDate                    *time.Time
	Page                       int
	Limit                      int
}

type TransactionListResult struct {
	Transactions []dto.PaymentTransactionDTO `json:"transactions"`
	Total        int64                       `json:"total"`
	Page         int                         `json:"page"`
	Limit        int                         `json:"limit"`
	TotalPages   int                         `json:"total_pages"`
}

type ExportCSVResult struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

func (s *Service) SetB2BService(b2bService *B2BService) {
	s.b2bService = b2bService
}

func (s *Service) resolvePremiumEntitlement(ctx context.Context, user *model.User) (*premiumEntitlement, error) {
	if user != nil && user.IsPremium && (user.PremiumExpiresAt == nil || user.PremiumExpiresAt.After(time.Now())) {
		return &premiumEntitlement{
			Active: true,
			Source: premiumEntitlementSourcePersonal,
		}, nil
	}

	if s.b2bService == nil || user == nil {
		return &premiumEntitlement{Source: premiumEntitlementSourceFree}, nil
	}

	entitled, organizationID, err := s.b2bService.IsUserEntitledB2BPremium(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if entitled {
		return &premiumEntitlement{
			Active:            true,
			Source:            premiumEntitlementSourceB2B,
			B2BOrganizationID: organizationID,
		}, nil
	}

	return &premiumEntitlement{Source: premiumEntitlementSourceFree}, nil
}

func (s *Service) hasPremiumEntitlement(ctx context.Context, user *model.User) (bool, error) {
	entitlementStatus, err := s.resolvePremiumEntitlement(ctx, user)
	if err != nil {
		return false, err
	}

	return entitlementStatus.Active, nil
}

func (s *Service) buildQuota(hasUnlimitedAccess bool, used int, resetAt time.Time) dto.ChatQuotaDTO {
	if hasUnlimitedAccess {
		return dto.ChatQuotaDTO{
			FeatureKey:  featureKeyChatAI,
			Limit:       0,
			Used:        0,
			Remaining:   0,
			IsUnlimited: true,
			ResetAt:     resetAt.Format(time.RFC3339),
		}
	}

	remaining := s.dailyChatLimit - used
	if remaining < 0 {
		remaining = 0
	}

	return dto.ChatQuotaDTO{
		FeatureKey:  featureKeyChatAI,
		Limit:       s.dailyChatLimit,
		Used:        used,
		Remaining:   remaining,
		IsUnlimited: false,
		ResetAt:     resetAt.Format(time.RFC3339),
	}
}

func parseChatQuotaResetInterval(value string) time.Duration {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" || trimmed == "daily" || trimmed == "day" {
		return defaultChatQuotaResetInterval
	}

	if strings.HasSuffix(trimmed, "d") {
		daysRaw := strings.TrimSpace(strings.TrimSuffix(trimmed, "d"))
		days, err := strconv.ParseFloat(daysRaw, 64)
		if err == nil && days > 0 {
			duration := time.Duration(days * float64(24*time.Hour))
			if duration < minChatQuotaResetInterval {
				return minChatQuotaResetInterval
			}
			return duration
		}
	}

	duration, err := time.ParseDuration(trimmed)
	if err != nil || duration <= 0 {
		return defaultChatQuotaResetInterval
	}
	if duration < minChatQuotaResetInterval {
		return minChatQuotaResetInterval
	}

	return duration
}

func (s *Service) currentChatQuotaWindow(now time.Time) chatQuotaWindow {
	interval := s.quotaResetInterval
	if interval <= 0 {
		interval = defaultChatQuotaResetInterval
	}

	return chatQuotaWindowFor(now, interval)
}

func chatQuotaWindowFor(now time.Time, interval time.Duration) chatQuotaWindow {
	if interval <= 0 {
		interval = defaultChatQuotaResetInterval
	}

	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if interval < 24*time.Hour {
		elapsed := now.Sub(startOfDay)
		windowIndex := int64(elapsed / interval)
		start := startOfDay.Add(time.Duration(windowIndex) * interval).Truncate(time.Second)
		return chatQuotaWindow{
			Start: start,
			End:   start.Add(interval).Truncate(time.Second),
		}
	}

	if interval%(24*time.Hour) == 0 {
		daysPerWindow := int(interval / (24 * time.Hour))
		if daysPerWindow <= 0 {
			daysPerWindow = 1
		}

		anchor := time.Date(1970, 1, 1, 0, 0, 0, 0, now.Location())
		daysSinceAnchor := int(startOfDay.Sub(anchor) / (24 * time.Hour))
		windowIndex := daysSinceAnchor / daysPerWindow
		start := anchor.AddDate(0, 0, windowIndex*daysPerWindow).Truncate(time.Second)
		return chatQuotaWindow{
			Start: start,
			End:   start.AddDate(0, 0, daysPerWindow).Truncate(time.Second),
		}
	}

	anchor := time.Date(1970, 1, 1, 0, 0, 0, 0, now.Location())
	elapsed := now.Sub(anchor)
	windowIndex := int64(elapsed / interval)
	start := anchor.Add(time.Duration(windowIndex) * interval).Truncate(time.Second)
	return chatQuotaWindow{
		Start: start,
		End:   start.Add(interval).Truncate(time.Second),
	}
}

// ChatQuotaWindowStart returns the storage key for the quota window used by
// the chat entitlement checks. Seeders use the same calculation so seeded
// usage is visible to the runtime quota enforcement.
func ChatQuotaWindowStart(now time.Time, resetInterval string) time.Time {
	return chatQuotaWindowFor(now, parseChatQuotaResetInterval(resetInterval)).Start
}

func toPlanDTO(plan model.PremiumPlan) dto.PremiumPlanDTO {
	return dto.PremiumPlanDTO{
		ID:           plan.ID,
		Code:         plan.Code,
		Name:         plan.Name,
		Description:  plan.Description,
		Price:        plan.Price,
		DurationDays: plan.DurationDays,
		IsActive:     plan.IsActive,
	}
}

func toTopupDTO(pkg model.TopupPackage) dto.TopupPackageDTO {
	return dto.TopupPackageDTO{
		ID:         pkg.ID,
		Code:       pkg.Code,
		Name:       pkg.Name,
		Coins:      pkg.Coins,
		BonusCoins: pkg.BonusCoins,
		TotalCoins: pkg.TotalCoins(),
		Price:      pkg.Price,
		IsActive:   pkg.IsActive,
	}
}

func toTransactionDTO(tx model.PaymentTransaction) dto.PaymentTransactionDTO {
	return dto.PaymentTransactionDTO{
		ID:                           tx.ID,
		OrderID:                      tx.OrderID,
		UserID:                       tx.UserID,
		ItemType:                     string(tx.ItemType),
		ItemID:                       tx.ItemID,
		ItemName:                     tx.ItemName,
		Amount:                       tx.Amount,
		Currency:                     tx.Currency,
		Status:                       string(tx.Status),
		PaymentProvider:              tx.PaymentProvider,
		ProviderTransactionID:        tx.ProviderTransactionID,
		ProviderPaymentType:          tx.ProviderPaymentType,
		FailureReason:                tx.FailureReason,
		SnapToken:                    tx.SnapToken,
		SnapURL:                      tx.SnapRedirectURL,
		PaidAt:                       tx.PaidAt,
		RefundedAmount:               tx.RefundedAmount,
		RefundRequestedAmount:        tx.RefundRequestedAmount,
		ProviderRefundAmountReported: tx.ProviderRefundAmountReported,
		RefundStatus:                 tx.RefundStatus,
		RefundReconciliationStatus:   tx.RefundReconciliationStatus,
		RefundReconciliationReason:   tx.RefundReconciliationReason,
		CreatedAt:                    tx.CreatedAt,
		UpdatedAt:                    tx.UpdatedAt,
	}
}

func (s *Service) GetCatalog(ctx context.Context, userID uint) (*dto.BillingCatalogResponse, error) {
	plans, err := s.repo.GetActivePlans(ctx)
	if err != nil {
		return nil, err
	}
	packages, err := s.repo.GetActiveTopupPackages(ctx)
	if err != nil {
		return nil, err
	}

	businessPlans := make([]dto.B2BPlanDTO, 0)
	if s.b2bService != nil {
		plansFromB2B, listErr := s.b2bService.ListPlans(ctx, true)
		if listErr != nil {
			return nil, listErr
		}
		businessPlans = plansFromB2B
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	quotaWindow := s.currentChatQuotaWindow(now)
	used, err := s.repo.GetFeatureUsage(ctx, userID, featureKeyChatAI, quotaWindow.Start)
	if err != nil {
		return nil, err
	}

	entitlementStatus, err := s.resolvePremiumEntitlement(ctx, user)
	if err != nil {
		return nil, err
	}

	planDTOs := make([]dto.PremiumPlanDTO, 0, len(plans))
	for _, plan := range plans {
		planDTOs = append(planDTOs, toPlanDTO(plan))
	}

	topupDTOs := make([]dto.TopupPackageDTO, 0, len(packages))
	for _, pkg := range packages {
		topupDTOs = append(topupDTOs, toTopupDTO(pkg))
	}

	return &dto.BillingCatalogResponse{
		Plans:         planDTOs,
		TopupPackages: topupDTOs,
		BusinessPlans: businessPlans,
		ChatQuota:     s.buildQuota(entitlementStatus.Active, used, quotaWindow.End),
	}, nil
}

func (s *Service) GetStatus(ctx context.Context, userID uint) (*dto.BillingStatusResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	quotaWindow := s.currentChatQuotaWindow(now)
	used, err := s.repo.GetFeatureUsage(ctx, userID, featureKeyChatAI, quotaWindow.Start)
	if err != nil {
		return nil, err
	}

	entitlementStatus, err := s.resolvePremiumEntitlement(ctx, user)
	if err != nil {
		return nil, err
	}

	latest, err := s.repo.GetLatestSubscription(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := &dto.BillingStatusResponse{
		IsPremium:         entitlementStatus.Active,
		EntitlementSource: entitlementStatus.Source,
		B2BOrganizationID: entitlementStatus.B2BOrganizationID,
		PremiumSince:      user.PremiumSince,
		PremiumExpiresAt:  user.PremiumExpiresAt,
		GoldCoins:         user.GoldCoins,
		ChatQuota:         s.buildQuota(entitlementStatus.Active, used, quotaWindow.End),
	}

	if latest != nil {
		resp.Subscription = &dto.BillingSubscriptionInfoDTO{
			PlanID:        latest.PlanID,
			PlanCode:      latest.Plan.Code,
			PlanName:      latest.Plan.Name,
			Status:        string(latest.Status),
			StartsAt:      latest.StartsAt,
			EndsAt:        latest.EndsAt,
			SourceOrderID: latest.SourceOrderID,
		}
	}

	return resp, nil
}

func (s *Service) CreateCheckout(ctx context.Context, userID uint, req *dto.CreateCheckoutRequest) (*dto.CreateCheckoutResponse, error) {
	if s.midtransClient == nil || !s.midtransClient.IsConfigured() {
		return nil, ErrMidtransNotConfigured
	}

	itemType := model.BillingItemType(req.ItemType)
	if itemType != model.BillingItemTypeSubscription && itemType != model.BillingItemTypeTopup {
		return nil, ErrItemNotFound
	}

	orderID := fmt.Sprintf("RT-%d-%d", userID, time.Now().UnixNano())
	transaction := &model.PaymentTransaction{
		OrderID:         orderID,
		UserID:          userID,
		ItemType:        itemType,
		ItemID:          req.ItemID,
		Currency:        "IDR",
		Status:          model.PaymentStatusPending,
		PaymentProvider: "midtrans",
	}

	var amount int
	var itemName string
	var itemCode string

	switch itemType {
	case model.BillingItemTypeSubscription:
		// Mutual exclusion: a user already covered by an active B2B premium seat
		// cannot also buy a personal premium subscription.
		if s.b2bService != nil {
			entitledB2B, _, entErr := s.b2bService.IsUserEntitledB2BPremium(ctx, userID)
			if entErr != nil {
				return nil, entErr
			}
			if entitledB2B {
				return nil, ErrPersonalPremiumBlockedByB2B
			}
		}

		plan, err := s.repo.GetPlanByID(ctx, req.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrItemNotFound
			}
			return nil, err
		}
		if !plan.IsActive {
			return nil, ErrItemNotActive
		}
		amount = plan.Price
		itemName = plan.Name
		itemCode = plan.Code
	case model.BillingItemTypeTopup:
		pkg, err := s.repo.GetTopupPackageByID(ctx, req.ItemID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrItemNotFound
			}
			return nil, err
		}
		if !pkg.IsActive {
			return nil, ErrItemNotActive
		}
		amount = pkg.Price
		itemName = pkg.Name
		itemCode = pkg.Code
	}

	transaction.Amount = amount
	transaction.ItemName = itemName

	expiresAt := time.Now().Add(24 * time.Hour)
	transaction.ExpiresAt = &expiresAt

	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	snapReq := MidtransSnapRequest{
		TransactionDetails: MidtransTransactionDetails{
			OrderID:     orderID,
			GrossAmount: amount,
		},
		CustomerDetails: MidtransCustomerDetails{
			FirstName: user.Name,
			Email:     user.Email,
		},
		ItemDetails: []MidtransItemDetails{
			{
				ID:       itemCode,
				Price:    amount,
				Quantity: 1,
				Name:     itemName,
			},
		},
		Callbacks: &MidtransCallbacks{
			Finish: func() string {
				if url := os.Getenv("FRONTEND_URL"); url != "" {
					return url + "/payment/success"
				}
				return "http://localhost:3000/payment/success"
			}(),
		},
	}

	snapResp, err := s.midtransClient.CreateSnapTransaction(ctx, snapReq)
	if err != nil {
		transaction.Status = model.PaymentStatusFailed
		transaction.FailureReason = err.Error()
		_ = s.repo.UpdateTransaction(ctx, transaction)
		return nil, err
	}

	transaction.SnapToken = snapResp.Token
	transaction.SnapRedirectURL = snapResp.RedirectURL
	if err := s.repo.UpdateTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	return &dto.CreateCheckoutResponse{
		TransactionID: transaction.ID,
		OrderID:       transaction.OrderID,
		ItemType:      string(transaction.ItemType),
		ItemID:        transaction.ItemID,
		ItemName:      transaction.ItemName,
		Amount:        transaction.Amount,
		Currency:      transaction.Currency,
		Status:        string(transaction.Status),
		SnapToken:     transaction.SnapToken,
		SnapURL:       transaction.SnapRedirectURL,
		ExpiresAt:     transaction.ExpiresAt,
	}, nil
}

func (s *Service) mapTransactionStatus(midtransStatus string, fraudStatus string) model.PaymentStatus {
	switch strings.ToLower(strings.TrimSpace(midtransStatus)) {
	case "capture", "settlement":
		fraud := strings.ToLower(strings.TrimSpace(fraudStatus))
		if fraud == "deny" {
			return model.PaymentStatusFailed
		}
		if fraud != "" && fraud != "accept" {
			return model.PaymentStatusPending
		}
		return model.PaymentStatusPaid
	case "pending":
		return model.PaymentStatusPending
	case "deny", "failure":
		return model.PaymentStatusFailed
	case "cancel":
		return model.PaymentStatusCanceled
	case "expire":
		return model.PaymentStatusExpired
	case "refund", "chargeback":
		return model.PaymentStatusRefunded
	case "partial_refund", "partial_chargeback":
		return model.PaymentStatusPaid
	default:
		return model.PaymentStatusPending
	}
}

func (s *Service) parseSettlementTime(value string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return &parsed
		}
	}
	return nil
}

func (s *Service) verifyWebhookSignature(payload *dto.MidtransWebhookRequest) bool {
	if strings.TrimSpace(s.serverKey) == "" {
		return false
	}

	raw := payload.OrderID + payload.StatusCode + payload.GrossAmount + s.serverKey
	hash := sha512.Sum512([]byte(raw))
	expected := hex.EncodeToString(hash[:])
	actual := strings.ToLower(strings.TrimSpace(payload.SignatureKey))
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func webhookAmountMatches(grossAmount string, amount int) bool {
	value, ok := new(big.Rat).SetString(strings.TrimSpace(grossAmount))
	return ok && value.Cmp(new(big.Rat).SetInt64(int64(amount))) == 0
}

func (s *Service) applySuccessfulPayment(tx *gorm.DB, transaction *model.PaymentTransaction, paidAt *time.Time) error {
	if transaction.ItemType == model.BillingItemTypeSubscription {
		plan, err := s.repo.GetPlanByIDTx(tx, transaction.ItemID)
		if err != nil {
			return err
		}

		// The plan may have been deactivated between checkout and payment. We
		// still honor the completed payment (the user already paid) but flag it
		// so admins can review.
		if !plan.IsActive {
			logger.Warn("billing: honoring paid subscription for an inactive plan",
				zap.String("order_id", transaction.OrderID), zap.Uint("plan_id", plan.ID))
		}

		user, err := s.repo.LockUserByID(tx, transaction.UserID)
		if err != nil {
			return err
		}

		now := time.Now()
		startAt := now
		if user.PremiumExpiresAt != nil && user.PremiumExpiresAt.After(now) {
			startAt = *user.PremiumExpiresAt
		}
		endAt := startAt.AddDate(0, 0, plan.DurationDays)

		user.IsPremium = true
		if user.PremiumSince == nil {
			user.PremiumSince = &now
		}
		user.PremiumExpiresAt = &endAt

		if err := s.repo.SaveUser(tx, user); err != nil {
			return err
		}

		subscription := &model.UserSubscription{
			UserID:        user.ID,
			PlanID:        plan.ID,
			SourceOrderID: transaction.OrderID,
			Status:        model.SubscriptionStatusActive,
			StartsAt:      startAt,
			EndsAt:        endAt,
		}
		if err := s.repo.CreateSubscription(tx, subscription); err != nil {
			return err
		}
	}

	if transaction.ItemType == model.BillingItemTypeTopup {
		pkg, err := s.repo.GetTopupPackageByIDTx(tx, transaction.ItemID)
		if err != nil {
			return err
		}
		if !pkg.IsActive {
			logger.Warn("billing: honoring paid topup for an inactive package",
				zap.String("order_id", transaction.OrderID), zap.Uint("package_id", pkg.ID))
		}
		if err := s.repo.AddUserGoldCoins(tx, transaction.UserID, pkg.TotalCoins()); err != nil {
			return err
		}
	}

	if paidAt == nil {
		now := time.Now()
		paidAt = &now
	}
	transaction.PaidAt = paidAt

	return nil
}

func (s *Service) reverseSubscriptionPayment(tx *gorm.DB, transaction *model.PaymentTransaction) error {
	if transaction.ItemType != model.BillingItemTypeSubscription {
		return nil
	}
	if err := tx.Model(&model.UserSubscription{}).
		Where("user_id = ? AND source_order_id = ?", transaction.UserID, transaction.OrderID).
		Update("status", model.SubscriptionStatusCanceled).Error; err != nil {
		return err
	}
	user, err := s.repo.LockUserByID(tx, transaction.UserID)
	if err != nil {
		return err
	}
	now := time.Now()
	var remaining []model.UserSubscription
	if err := tx.Where("user_id = ? AND status = ? AND ends_at > ?", user.ID, model.SubscriptionStatusActive, now).
		Order("starts_at ASC").Find(&remaining).Error; err != nil {
		return err
	}
	if len(remaining) == 0 {
		user.IsPremium = false
		user.PremiumExpiresAt = nil
		return s.repo.SaveUser(tx, user)
	}
	cursor := now
	for i := range remaining {
		subscription := &remaining[i]
		if subscription.StartsAt.After(cursor) {
			duration := subscription.EndsAt.Sub(subscription.StartsAt)
			subscription.StartsAt = cursor
			subscription.EndsAt = cursor.Add(duration)
			if err := tx.Save(subscription).Error; err != nil {
				return err
			}
		}
		if subscription.EndsAt.After(cursor) {
			cursor = subscription.EndsAt
		}
	}
	user.IsPremium = true
	user.PremiumExpiresAt = &cursor
	return s.repo.SaveUser(tx, user)
}

func (s *Service) revokeRefundedSubscription(tx *gorm.DB, transaction *model.PaymentTransaction) (int, error) {
	var subscription model.UserSubscription
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND source_order_id = ?", transaction.UserID, transaction.OrderID).
		First(&subscription).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	daysReduced := 0
	if err == nil {
		duration := subscription.EndsAt.Sub(subscription.StartsAt)
		if duration > 0 {
			daysReduced = int((duration + 24*time.Hour - 1) / (24 * time.Hour))
		}
	}
	if err := s.reverseSubscriptionPayment(tx, transaction); err != nil {
		return 0, err
	}
	return daysReduced, nil
}

func generateRefundKey() (string, error) {
	var token [12]byte
	if _, err := rand.Read(token[:]); err != nil {
		return "", err
	}
	return "rt-" + hex.EncodeToString(token[:]), nil
}

func proportionalCoinReversal(totalCoins, refundAmount int64, transactionAmount int) int64 {
	if totalCoins <= 0 || refundAmount <= 0 || transactionAmount <= 0 {
		return 0
	}
	if refundAmount >= int64(transactionAmount) {
		return totalCoins
	}
	numerator := new(big.Int).Mul(big.NewInt(totalCoins), big.NewInt(refundAmount))
	return numerator.Div(numerator, big.NewInt(int64(transactionAmount))).Int64()
}

func parseRefundID(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return ""
	}
	if strings.HasPrefix(trimmed, "\"") {
		var value string
		if json.Unmarshal(raw, &value) != nil {
			return ""
		}
		return strings.TrimSpace(value)
	}
	return trimmed
}

func parseProviderReference(value dto.MidtransProviderID) string {
	return strings.TrimSpace(string(value))
}

func parseIDRAmount(raw string) (int64, bool) {
	amount, ok := new(big.Rat).SetString(strings.TrimSpace(raw))
	if !ok || !amount.IsInt() || !amount.Num().IsInt64() || amount.Sign() <= 0 {
		return 0, false
	}
	return amount.Num().Int64(), true
}

func parseProviderTimestamp(value string) *time.Time {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		var parsed time.Time
		var err error
		if layout == "2006-01-02 15:04:05" {
			parsed, err = time.ParseInLocation(layout, trimmed, time.FixedZone("WIB", 7*60*60))
		} else {
			parsed, err = time.Parse(layout, trimmed)
		}
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func refundIdentifier(detail dto.MidtransRefundDetail, orderID string) (refundKey, providerID string) {
	providerID = parseProviderReference(detail.RefundChargebackID)
	refundKey = strings.TrimSpace(detail.RefundKey)
	if refundKey == "" && providerID != "" {
		refundKey = "midtrans-" + providerID
	}
	if providerID == "" && refundKey != "" {
		providerID = "key:" + refundKey
	}
	if refundKey == "" && providerID == "" && strings.TrimSpace(detail.CreatedAt) != "" {
		identity := fmt.Sprintf("%s:%s:%s:%s", orderID, detail.RefundAmount, detail.CreatedAt, detail.Reason)
		hash := sha256.Sum256([]byte(identity))
		refundKey = "midtrans-" + hex.EncodeToString(hash[:12])
		providerID = "key:" + refundKey
	}
	return refundKey, providerID
}

func (s *Service) RequestMidtransRefund(ctx context.Context, actorUserID uint, orderID string, req dto.AdminRefundRequest) (*dto.AdminRefundResponse, error) {
	if s.midtransClient == nil || !s.midtransClient.IsConfigured() {
		return nil, ErrMidtransNotConfigured
	}
	orderID = strings.TrimSpace(orderID)
	reason := strings.TrimSpace(req.Reason)
	if orderID == "" || reason == "" || req.Amount <= 0 {
		return nil, ErrRefundNotAllowed
	}
	refundKey, err := generateRefundKey()
	if err != nil {
		return nil, err
	}

	var transaction model.PaymentTransaction
	var refund model.PaymentRefund
	if err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		locked, err := s.repo.LockTransactionByOrderID(tx, orderID)
		if err != nil {
			return err
		}
		if locked.PaymentProvider != "midtrans" || locked.Status != model.PaymentStatusPaid || locked.RefundStatus == "refunded" || locked.RefundReconciliationStatus == "pending" {
			return ErrRefundNotAllowed
		}
		reserved, confirmed, err := s.repo.GetPaymentRefundTotalsTx(tx, locked.ID)
		if err != nil {
			return err
		}
		if reserved > confirmed {
			return ErrRefundNotAllowed
		}
		if req.Amount > int64(locked.Amount)-reserved {
			return ErrRefundAmountInvalid
		}
		if locked.ItemType == model.BillingItemTypeTopup {
			pkg, err := s.repo.GetTopupPackageByIDForUpdateTx(tx, locked.ItemID)
			if err != nil {
				return err
			}
			projectedRefund := reserved + req.Amount
			coinsToReverse := proportionalCoinReversal(pkg.TotalCoins(), projectedRefund, locked.Amount)
			alreadyAccounted := locked.CoinsReversed + locked.CoinsWrittenOff
			if coinsToReverse > alreadyAccounted {
				user, err := s.repo.LockUserByID(tx, locked.UserID)
				if err != nil {
					return err
				}
				if user.GoldCoins < coinsToReverse-alreadyAccounted {
					return ErrRefundCoinsInsufficient
				}
			}
		}
		actor := actorUserID
		refund = model.PaymentRefund{
			PaymentTransactionID: locked.ID,
			RefundKey:            refundKey,
			ProviderRefundID:     "key:" + refundKey,
			Amount:               req.Amount,
			Reason:               reason,
			Status:               "requested",
			RequestedBy:          &actor,
			RequestedAt:          time.Now(),
		}
		if err := s.repo.CreatePaymentRefundTx(tx, &refund); err != nil {
			return err
		}
		transaction = *locked
		reserved, confirmed, err = s.repo.GetPaymentRefundTotalsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		s.updateRefundSummary(&transaction, reserved, confirmed)
		if err := tx.Save(&transaction).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	providerResponse, err := s.midtransClient.RefundTransaction(ctx, orderID, MidtransRefundRequest{
		RefundKey: refundKey,
		Amount:    req.Amount,
		Reason:    reason,
	})
	if err != nil {
		var apiErr *MidtransAPIError
		if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != http.StatusNotAcceptable {
			if persistErr := s.persistRefundSubmissionOutcome(ctx, orderID, refundKey, "rejected", "", ""); persistErr != nil {
				return nil, persistErr
			}
			return nil, err
		}
		if persistErr := s.persistRefundSubmissionOutcome(ctx, orderID, refundKey, "pending_confirmation", "", "refund_submission_status_unknown_check_midtrans_dashboard"); persistErr != nil {
			return nil, persistErr
		}
		return nil, ErrRefundSubmissionUnknown
	}

	providerRefundID := ""
	if providerResponse != nil {
		providerRefundID = parseRefundID(providerResponse.RefundChargebackID)
		if strings.TrimSpace(providerResponse.RefundAmount) != "" {
			providerAmount, ok := parseIDRAmount(providerResponse.RefundAmount)
			if !ok || providerAmount != req.Amount {
				if persistErr := s.persistRefundSubmissionOutcome(ctx, orderID, refundKey, "pending_confirmation", providerRefundID, "refund_response_amount_mismatch_check_midtrans_dashboard"); persistErr != nil {
					return nil, persistErr
				}
				return nil, ErrRefundSubmissionUnknown
			}
		}
	}
	if err := s.persistRefundSubmissionOutcome(ctx, orderID, refundKey, "pending_confirmation", providerRefundID, ""); err != nil {
		return nil, err
	}
	return &dto.AdminRefundResponse{
		OrderID:   orderID,
		RefundKey: refundKey,
		Amount:    req.Amount,
		Status:    "pending_confirmation",
		Message:   "Permintaan diterima Midtrans dan menunggu konfirmasi penyedia pembayaran.",
	}, nil
}

func (s *Service) persistRefundSubmissionOutcome(ctx context.Context, orderID, refundKey, status, providerID, reconciliationReason string) error {
	return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		transaction, err := s.repo.LockTransactionByOrderID(tx, orderID)
		if err != nil {
			return err
		}
		refund, err := s.repo.FindPaymentRefundByKeyTx(tx, transaction.ID, refundKey)
		if err != nil {
			return err
		}
		if refund.Status != "confirmed" {
			refund.Status = status
		}
		if providerID != "" && (refund.ProviderRefundID == "" || strings.HasPrefix(refund.ProviderRefundID, "key:")) {
			refund.ProviderRefundID = providerID
		}
		if err := s.repo.SavePaymentRefundTx(tx, refund); err != nil {
			return err
		}
		reserved, confirmed, err := s.repo.GetPaymentRefundTotalsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		s.updateRefundSummary(transaction, reserved, confirmed)
		if reconciliationReason != "" && refund.Status != "confirmed" {
			transaction.RefundReconciliationStatus = "pending"
			transaction.RefundReconciliationReason = reconciliationReason
		}
		return tx.Save(transaction).Error
	})
}

func (s *Service) updateRefundSummary(transaction *model.PaymentTransaction, reserved, confirmed int64) {
	transaction.RefundedAmount = confirmed
	transaction.RefundRequestedAmount = reserved
	switch {
	case confirmed >= int64(transaction.Amount):
		transaction.RefundStatus = "refunded"
		transaction.Status = model.PaymentStatusRefunded
	case confirmed > 0:
		transaction.RefundStatus = "partially_refunded"
	case reserved > 0:
		transaction.RefundStatus = "pending_confirmation"
	default:
		transaction.RefundStatus = "none"
	}
}

func (s *Service) upsertRefundDetail(tx *gorm.DB, transaction *model.PaymentTransaction, detail dto.MidtransRefundDetail) (bool, error) {
	amount, ok := parseIDRAmount(detail.RefundAmount)
	refundKey, providerID := refundIdentifier(detail, transaction.OrderID)
	if !ok || refundKey == "" || providerID == "" {
		return false, ErrWebhookPayloadInvalid
	}
	if len(strings.TrimSpace(detail.Reason)) > 255 || len(refundKey) > 120 || len(providerID) > 120 {
		return false, ErrWebhookPayloadInvalid
	}

	refund, err := s.repo.FindPaymentRefundByKeyTx(tx, transaction.ID, refundKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		refund, err = s.repo.FindPaymentRefundByProviderIDTx(tx, transaction.ID, providerID)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	confirmedAt := parseProviderTimestamp(detail.BankConfirmedAt)
	if refund == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		status := "pending_confirmation"
		if confirmedAt != nil {
			status = "confirmed"
		}
		providerCreatedAt := parseProviderTimestamp(detail.CreatedAt)
		requestedAt := time.Now()
		if providerCreatedAt != nil {
			requestedAt = *providerCreatedAt
		}
		refund = &model.PaymentRefund{
			PaymentTransactionID: transaction.ID,
			RefundKey:            refundKey,
			ProviderRefundID:     providerID,
			Amount:               amount,
			Reason:               strings.TrimSpace(detail.Reason),
			RefundMethod:         strings.TrimSpace(detail.RefundMethod),
			Status:               status,
			RequestedAt:          requestedAt,
			ProviderCreatedAt:    providerCreatedAt,
			BankConfirmedAt:      confirmedAt,
		}
		if err := s.repo.CreatePaymentRefundTx(tx, refund); err != nil {
			return false, err
		}
		return confirmedAt != nil, nil
	}
	if refund.Amount != amount {
		return false, ErrWebhookPayloadInvalid
	}
	if refund.ProviderRefundID == "" || strings.HasPrefix(refund.ProviderRefundID, "key:") {
		refund.ProviderRefundID = providerID
	}
	if refund.RefundKey == "" {
		refund.RefundKey = refundKey
	}
	if refund.Reason == "" {
		refund.Reason = strings.TrimSpace(detail.Reason)
	}
	if refund.RefundMethod == "" {
		refund.RefundMethod = strings.TrimSpace(detail.RefundMethod)
	}
	if refund.ProviderCreatedAt == nil {
		refund.ProviderCreatedAt = parseProviderTimestamp(detail.CreatedAt)
	}
	wasConfirmed := refund.Status == "confirmed"
	if confirmedAt != nil {
		refund.Status = "confirmed"
		refund.BankConfirmedAt = confirmedAt
	} else if refund.Status == "requested" {
		refund.Status = "pending_confirmation"
	}
	if err := s.repo.SavePaymentRefundTx(tx, refund); err != nil {
		return false, err
	}
	return !wasConfirmed && refund.Status == "confirmed", nil
}

func (s *Service) applyConfirmedRefundEffects(tx *gorm.DB, transaction *model.PaymentTransaction, wasFullyRefunded bool) error {
	if transaction.ItemType == model.BillingItemTypeSubscription {
		if transaction.RefundedAmount >= int64(transaction.Amount) {
			if transaction.RefundReconciliationReason == "legacy_subscription_refund_requires_entitlement_review" {
				return nil
			}
			transaction.RefundReconciliationStatus = "not_required"
			transaction.RefundReconciliationReason = ""
			if !wasFullyRefunded {
				return s.reverseSubscriptionPayment(tx, transaction)
			}
			return nil
		}
		if transaction.RefundedAmount > 0 {
			transaction.RefundReconciliationStatus = "pending"
			transaction.RefundReconciliationReason = "partial_subscription_refund_requires_entitlement_review"
		}
		return nil
	}
	if transaction.ItemType != model.BillingItemTypeTopup || transaction.RefundedAmount == 0 {
		return nil
	}
	pkg, err := s.repo.GetTopupPackageByIDTx(tx, transaction.ItemID)
	if err != nil {
		transaction.RefundReconciliationStatus = "pending"
		transaction.RefundReconciliationReason = "topup_package_missing_for_coin_reversal"
		return nil
	}
	desiredReversal := proportionalCoinReversal(pkg.TotalCoins(), transaction.RefundedAmount, transaction.Amount)
	accounted := transaction.CoinsReversed + transaction.CoinsWrittenOff
	if desiredReversal <= accounted {
		if transaction.RefundReconciliationReason == "refund_submission_status_unknown_check_midtrans_dashboard" || transaction.RefundReconciliationReason == "refund_details_missing_from_provider_notification" {
			transaction.RefundReconciliationStatus = "not_required"
			transaction.RefundReconciliationReason = ""
		}
		return nil
	}
	user, err := s.repo.LockUserByID(tx, transaction.UserID)
	if err != nil {
		return err
	}
	coinsToReverse := desiredReversal - accounted
	if user.GoldCoins < coinsToReverse {
		transaction.RefundReconciliationStatus = "pending"
		transaction.RefundReconciliationReason = "insufficient_coin_balance_for_confirmed_refund"
		return nil
	}
	settled, err := s.repo.SettleUserGoldCoinsTx(tx, transaction.UserID, coinsToReverse)
	if err != nil {
		return err
	}
	if !settled {
		transaction.RefundReconciliationStatus = "pending"
		transaction.RefundReconciliationReason = "insufficient_coin_balance_for_confirmed_refund"
		return nil
	}
	transaction.CoinsReversed += coinsToReverse
	transaction.RefundReconciliationStatus = "not_required"
	transaction.RefundReconciliationReason = ""
	return nil
}

func (s *Service) applyRefundWebhook(tx *gorm.DB, transaction *model.PaymentTransaction, payload *dto.MidtransWebhookRequest) error {
	if payload.StatusCode != "200" {
		return ErrWebhookPayloadInvalid
	}
	wasFullyRefunded := transaction.Status == model.PaymentStatusRefunded
	if strings.TrimSpace(payload.RefundAmount) != "" {
		reported, ok := parseIDRAmount(payload.RefundAmount)
		if !ok || reported > int64(transaction.Amount) {
			return ErrWebhookPayloadInvalid
		}
		transaction.ProviderRefundAmountReported = reported
	}
	details := payload.Refunds
	chargeback := strings.Contains(strings.ToLower(payload.TransactionStatus), "chargeback")
	if len(details) == 0 && (parseProviderReference(payload.RefundChargebackID) != "" || strings.TrimSpace(payload.RefundKey) != "") {
		details = []dto.MidtransRefundDetail{{
			RefundChargebackID: payload.RefundChargebackID,
			RefundAmount:       payload.RefundAmount,
			CreatedAt:          payload.TransactionTime,
			Reason:             payload.RefundReason,
			RefundKey:          payload.RefundKey,
			RefundMethod:       payload.RefundMethod,
			BankConfirmedAt:    payload.BankConfirmedAt,
		}}
	}
	partialChargeback := strings.EqualFold(payload.TransactionStatus, "partial_chargeback")
	partialReportedTotal, partialTotalKnown := parseIDRAmount(payload.RefundAmount)
	if chargeback && len(details) == 0 && (!partialChargeback || partialTotalKnown) {
		reportedTotal := int64(transaction.Amount)
		if partialChargeback {
			reportedTotal = partialReportedTotal
			if reportedTotal > int64(transaction.Amount) {
				return ErrWebhookPayloadInvalid
			}
		}
		reserved, confirmed, err := s.repo.GetPaymentRefundTotalsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		if reportedTotal < confirmed {
			return ErrWebhookPayloadInvalid
		}
		chargebackAmount := reportedTotal - confirmed
		if chargebackAmount == 0 {
			s.updateRefundSummary(transaction, reserved, confirmed)
			if err := s.applyConfirmedRefundEffects(tx, transaction, wasFullyRefunded); err != nil {
				return err
			}
			if reserved > confirmed && transaction.RefundReconciliationStatus == "not_required" {
				transaction.RefundReconciliationStatus = "pending"
				transaction.RefundReconciliationReason = "awaiting_midtrans_refund_confirmation"
			}
			return tx.Save(transaction).Error
		}
		refunds, err := s.repo.ListPaymentRefundsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		var pendingRefund *model.PaymentRefund
		pendingRefundCount := 0
		for i := range refunds {
			if refunds[i].Status == "requested" || refunds[i].Status == "pending_confirmation" {
				pendingRefundCount++
				pendingRefund = &refunds[i]
			}
		}
		if pendingRefundCount > 1 || (pendingRefund != nil && pendingRefund.Amount != chargebackAmount) {
			// The provider total cannot be mapped safely to the outstanding local
			// request; leave it for operator review instead of double-counting.
			details = nil
		} else {
			providerID := parseProviderReference(payload.RefundChargebackID)
			refundKey := strings.TrimSpace(payload.RefundKey)
			if pendingRefund != nil {
				refundKey = pendingRefund.RefundKey
			} else if providerID == "" && refundKey == "" {
				identity := payload.TransactionID + ":" + strconv.FormatInt(reportedTotal, 10) + ":" + payload.TransactionTime
				hash := sha256.Sum256([]byte(identity))
				providerID = "chargeback-" + hex.EncodeToString(hash[:12])
				refundKey = providerID
			}
			if providerID == "" {
				providerID = "chargeback-" + refundKey
			}
			if refundKey == "" {
				refundKey = "chargeback-" + providerID
			}
			confirmedAt := payload.BankConfirmedAt
			if confirmedAt == "" {
				confirmedAt = time.Now().Format("2006-01-02 15:04:05")
			}
			details = []dto.MidtransRefundDetail{{
				RefundChargebackID: dto.MidtransProviderID(providerID),
				RefundAmount:       strconv.FormatInt(chargebackAmount, 10),
				CreatedAt:          payload.TransactionTime,
				Reason:             "chargeback",
				RefundKey:          refundKey,
				RefundMethod:       "chargeback",
				BankConfirmedAt:    confirmedAt,
			}}
		}
	}
	if len(details) == 0 {
		refunds, err := s.repo.ListPaymentRefundsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		if strings.HasPrefix(transaction.RefundReconciliationReason, "legacy_") && len(refunds) == 0 {
			// Keep the migration's confirmed amount until provider details or an
			// operator's historical review can account for it.
			return tx.Save(transaction).Error
		}
		transaction.RefundReconciliationStatus = "pending"
		transaction.RefundReconciliationReason = "refund_details_missing_from_provider_notification"
		reserved, confirmed, err := s.repo.GetPaymentRefundTotalsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		s.updateRefundSummary(transaction, reserved, confirmed)
		if transaction.ProviderRefundAmountReported > confirmed && reserved == 0 {
			transaction.RefundStatus = "pending_confirmation"
		}
		return tx.Save(transaction).Error
	}

	for _, detail := range details {
		if chargeback && detail.BankConfirmedAt == "" {
			detail.BankConfirmedAt = time.Now().Format("2006-01-02 15:04:05")
			if detail.RefundMethod == "" {
				detail.RefundMethod = "chargeback"
			}
		}
		if _, err := s.upsertRefundDetail(tx, transaction, detail); err != nil {
			return err
		}
	}
	reserved, confirmed, err := s.repo.GetPaymentRefundTotalsTx(tx, transaction.ID)
	if err != nil {
		return err
	}
	if confirmed > int64(transaction.Amount) || reserved > int64(transaction.Amount) {
		return ErrWebhookPayloadInvalid
	}
	if transaction.ProviderRefundAmountReported > 0 && confirmed > transaction.ProviderRefundAmountReported {
		return ErrWebhookPayloadInvalid
	}
	s.updateRefundSummary(transaction, reserved, confirmed)
	if err := s.applyConfirmedRefundEffects(tx, transaction, wasFullyRefunded); err != nil {
		return err
	}
	if reserved > confirmed && transaction.RefundReconciliationStatus == "not_required" {
		transaction.RefundReconciliationStatus = "pending"
		transaction.RefundReconciliationReason = "awaiting_midtrans_refund_confirmation"
	}
	return tx.Save(transaction).Error
}

func (s *Service) ReconcileRefund(ctx context.Context, actorUserID uint, orderID string, req dto.AdminRefundReconciliationRequest) (*dto.AdminRefundReconciliationResponse, error) {
	orderID = strings.TrimSpace(orderID)
	action := strings.TrimSpace(req.Action)
	note := strings.TrimSpace(req.Note)
	if orderID == "" || note == "" {
		return nil, ErrRefundReconciliationActionInvalid
	}
	var result dto.AdminRefundReconciliationResponse
	err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		transaction, err := s.repo.LockTransactionByOrderID(tx, orderID)
		if err != nil {
			return err
		}
		if transaction.RefundReconciliationStatus != "pending" {
			return ErrRefundReconciliationNotRequired
		}
		if transaction.ProviderRefundAmountReported > transaction.RefundedAmount {
			return ErrRefundReconciliationWaiting
		}
		refunds, err := s.repo.ListPaymentRefundsTx(tx, transaction.ID)
		if err != nil {
			return err
		}
		hasUnconfirmedRefund := false
		for _, refund := range refunds {
			if refund.Status == "requested" || refund.Status == "pending_confirmation" {
				hasUnconfirmedRefund = true
				break
			}
		}
		if hasUnconfirmedRefund && action != "mark_refund_rejected" {
			return ErrRefundReconciliationWaiting
		}
		if transaction.RefundedAmount == 0 && transaction.RefundStatus == "pending_confirmation" && action != "mark_refund_rejected" && action != "complete_manual_review" {
			return ErrRefundReconciliationWaiting
		}
		premiumDaysReduced := 0
		if action == "deduct_remaining_coins" || action == "accept_consumed_coins" {
			if transaction.ItemType != model.BillingItemTypeTopup {
				return ErrRefundReconciliationActionInvalid
			}
			pkg, err := s.repo.GetTopupPackageByIDTx(tx, transaction.ItemID)
			if err != nil {
				return err
			}
			desired := proportionalCoinReversal(pkg.TotalCoins(), transaction.RefundedAmount, transaction.Amount)
			outstanding := desired - transaction.CoinsReversed - transaction.CoinsWrittenOff
			if outstanding <= 0 {
				return ErrRefundReconciliationNotRequired
			}
			if action == "deduct_remaining_coins" {
				if _, err := s.repo.LockUserByID(tx, transaction.UserID); err != nil {
					return err
				}
				settled, err := s.repo.SettleUserGoldCoinsTx(tx, transaction.UserID, outstanding)
				if err != nil {
					return err
				}
				if !settled {
					return ErrRefundCoinsInsufficient
				}
				transaction.CoinsReversed += outstanding
			} else {
				transaction.CoinsWrittenOff += outstanding
			}
		} else if action == "revoke_premium_days" || action == "revoke_refunded_subscription" {
			if transaction.ItemType != model.BillingItemTypeSubscription || transaction.RefundedAmount == 0 {
				return ErrRefundReconciliationActionInvalid
			}
			if action == "revoke_premium_days" && req.PremiumDaysToRevoke <= 0 {
				return ErrRefundReconciliationActionInvalid
			}
			if action == "revoke_refunded_subscription" && transaction.RefundedAmount < int64(transaction.Amount) {
				return ErrRefundReconciliationActionInvalid
			}
			if action == "revoke_refunded_subscription" {
				premiumDaysReduced, err = s.revokeRefundedSubscription(tx, transaction)
			} else {
				premiumDaysReduced, err = s.reduceSubscriptionEntitlement(tx, transaction, req.PremiumDaysToRevoke)
			}
			if err != nil {
				return err
			}
		} else if action == "retain_entitlement" {
			if transaction.ItemType != model.BillingItemTypeSubscription || transaction.RefundedAmount == 0 {
				return ErrRefundReconciliationActionInvalid
			}
		} else if action == "mark_refund_rejected" {
			if !req.ProviderRejectionConfirmed {
				return ErrRefundReconciliationActionInvalid
			}
			if transaction.ProviderRefundAmountReported > transaction.RefundedAmount {
				return ErrRefundReconciliationActionInvalid
			}
			refunds, err := s.repo.ListPaymentRefundsTx(tx, transaction.ID)
			if err != nil {
				return err
			}
			rejected := 0
			for i := range refunds {
				if refunds[i].Status != "requested" && refunds[i].Status != "pending_confirmation" {
					continue
				}
				refunds[i].Status = "rejected"
				if err := s.repo.SavePaymentRefundTx(tx, &refunds[i]); err != nil {
					return err
				}
				rejected++
			}
			if rejected == 0 {
				if transaction.RefundStatus != "pending_confirmation" || transaction.RefundRequestedAmount != 0 {
					return ErrRefundReconciliationNotRequired
				}
				transaction.ProviderRefundAmountReported = 0
			}
			reserved, confirmed, err := s.repo.GetPaymentRefundTotalsTx(tx, transaction.ID)
			if err != nil {
				return err
			}
			s.updateRefundSummary(transaction, reserved, confirmed)
		} else if action == "complete_manual_review" {
			if !req.ManualReviewConfirmed || transaction.ProviderRefundAmountReported > transaction.RefundedAmount {
				return ErrRefundReconciliationActionInvalid
			}
			refunds, err := s.repo.ListPaymentRefundsTx(tx, transaction.ID)
			if err != nil {
				return err
			}
			for _, refund := range refunds {
				if refund.Status == "requested" || refund.Status == "pending_confirmation" {
					return ErrRefundReconciliationWaiting
				}
			}
			if transaction.ItemType == model.BillingItemTypeSubscription && transaction.RefundedAmount > 0 {
				return ErrRefundReconciliationActionInvalid
			}
			if transaction.ItemType == model.BillingItemTypeTopup && transaction.RefundedAmount > 0 {
				pkg, packageErr := s.repo.GetTopupPackageByIDTx(tx, transaction.ItemID)
				if packageErr == nil {
					desired := proportionalCoinReversal(pkg.TotalCoins(), transaction.RefundedAmount, transaction.Amount)
					if desired > transaction.CoinsReversed+transaction.CoinsWrittenOff {
						return ErrRefundReconciliationActionInvalid
					}
				} else if !errors.Is(packageErr, gorm.ErrRecordNotFound) {
					return packageErr
				}
			}
		} else {
			return ErrRefundReconciliationActionInvalid
		}
		transaction.RefundReconciliationStatus = "resolved"
		transaction.RefundReconciliationReason = ""
		if err := tx.Save(transaction).Error; err != nil {
			return err
		}
		actor := actorUserID
		event := &model.PaymentRefundReconciliationEvent{
			PaymentTransactionID: transaction.ID,
			ActorUserID:          &actor,
			Action:               action,
			Note:                 note,
			RefundedAmount:       transaction.RefundedAmount,
			CoinsReversed:        transaction.CoinsReversed,
			CoinsWrittenOff:      transaction.CoinsWrittenOff,
			PremiumDaysReduced:   premiumDaysReduced,
		}
		if err := s.repo.CreateRefundReconciliationEventTx(tx, event); err != nil {
			return err
		}
		result = dto.AdminRefundReconciliationResponse{
			OrderID:                    transaction.OrderID,
			RefundReconciliationStatus: transaction.RefundReconciliationStatus,
			RefundReconciliationReason: transaction.RefundReconciliationReason,
			CoinsReversed:              transaction.CoinsReversed,
			CoinsWrittenOff:            transaction.CoinsWrittenOff,
			PremiumDaysReduced:         premiumDaysReduced,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) SyncMidtransTransactionStatus(ctx context.Context, orderID string) error {
	if s.midtransClient == nil || !s.midtransClient.IsConfigured() {
		return ErrMidtransNotConfigured
	}
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return ErrWebhookPayloadInvalid
	}
	payload, err := s.midtransClient.GetTransactionStatus(ctx, orderID)
	if err != nil {
		return err
	}
	if payload.OrderID != orderID {
		return ErrWebhookPayloadInvalid
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	err = s.HandleMidtransWebhook(ctx, payload, string(rawPayload))
	if errors.Is(err, ErrWebhookDuplicate) {
		return nil
	}
	return err
}

func (s *Service) reduceSubscriptionEntitlement(tx *gorm.DB, transaction *model.PaymentTransaction, days int) (int, error) {
	var source model.UserSubscription
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND source_order_id = ?", transaction.UserID, transaction.OrderID).
		First(&source).Error
	if err != nil {
		return 0, err
	}
	duration := source.EndsAt.Sub(source.StartsAt)
	if duration <= 0 {
		return 0, ErrRefundReconciliationActionInvalid
	}
	reduction := time.Duration(days) * 24 * time.Hour
	if reduction > duration {
		reduction = duration
	}
	source.EndsAt = source.EndsAt.Add(-reduction)
	if !source.EndsAt.After(time.Now()) {
		source.Status = model.SubscriptionStatusCanceled
	}
	if err := tx.Save(&source).Error; err != nil {
		return 0, err
	}

	user, err := s.repo.LockUserByID(tx, transaction.UserID)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	var remaining []model.UserSubscription
	if err := tx.Where("user_id = ? AND status = ? AND ends_at > ?", user.ID, model.SubscriptionStatusActive, now).
		Order("starts_at ASC").Find(&remaining).Error; err != nil {
		return 0, err
	}
	cursor := now
	for i := range remaining {
		subscription := &remaining[i]
		if subscription.StartsAt.After(cursor) {
			remainingDuration := subscription.EndsAt.Sub(subscription.StartsAt)
			subscription.StartsAt = cursor
			subscription.EndsAt = cursor.Add(remainingDuration)
			if err := tx.Save(subscription).Error; err != nil {
				return 0, err
			}
		}
		if subscription.EndsAt.After(cursor) {
			cursor = subscription.EndsAt
		}
	}
	user.IsPremium = cursor.After(now)
	if user.IsPremium {
		user.PremiumExpiresAt = &cursor
	} else {
		user.PremiumExpiresAt = nil
	}
	if err := s.repo.SaveUser(tx, user); err != nil {
		return 0, err
	}
	return int(reduction / (24 * time.Hour)), nil
}

func (s *Service) HandleMidtransWebhook(ctx context.Context, payload *dto.MidtransWebhookRequest, rawPayload string) error {
	if payload == nil {
		return errors.New("empty webhook payload")
	}

	if !s.verifyWebhookSignature(payload) {
		return ErrWebhookSignatureInvalid
	}

	if payload.OrderID == "" || payload.TransactionID == "" || payload.TransactionStatus == "" {
		return ErrWebhookPayloadInvalid
	}
	if rawPayload == "" {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return ErrWebhookPayloadInvalid
		}
		rawPayload = string(encoded)
	}
	payloadHash := sha256.Sum256([]byte(rawPayload))
	eventKey := fmt.Sprintf("midtrans:%s:%s", payload.TransactionID, hex.EncodeToString(payloadHash[:]))
	event := &model.PaymentWebhookEvent{
		Provider:    "midtrans",
		OrderID:     payload.OrderID,
		EventKey:    eventKey,
		Payload:     rawPayload,
		ProcessedAt: time.Now(),
	}

	newStatus := s.mapTransactionStatus(payload.TransactionStatus, payload.FraudStatus)
	paidAt := s.parseSettlementTime(payload.SettlementTime)

	return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.repo.CreateWebhookEventTx(tx, event); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
				return ErrWebhookDuplicate
			}
			return err
		}

		transaction, err := s.repo.LockTransactionByOrderID(tx, payload.OrderID)
		if err != nil {
			return err
		}
		if !webhookAmountMatches(payload.GrossAmount, transaction.Amount) ||
			(transaction.ProviderTransactionID != "" && transaction.ProviderTransactionID != payload.TransactionID) ||
			(newStatus == model.PaymentStatusPaid && payload.StatusCode != "200") {
			return ErrWebhookPayloadInvalid
		}
		if hasMidtransRefundData(payload) {
			if transaction.Status != model.PaymentStatusPaid && transaction.Status != model.PaymentStatusRefunded {
				return ErrWebhookPayloadInvalid
			}
			if transaction.ProviderTransactionID == "" {
				transaction.ProviderTransactionID = payload.TransactionID
			}
			transaction.ProviderPaymentType = payload.PaymentType
			transaction.CallbackPayload = rawPayload
			return s.applyRefundWebhook(tx, transaction, payload)
		}
		if transaction.IsFinalStatus() {
			return nil
		}

		transaction.Status = newStatus
		transaction.ProviderTransactionID = payload.TransactionID
		transaction.ProviderPaymentType = payload.PaymentType
		transaction.CallbackPayload = rawPayload

		if newStatus == model.PaymentStatusPaid {
			if err := s.applySuccessfulPayment(tx, transaction, paidAt); err != nil {
				return err
			}
		} else if newStatus == model.PaymentStatusFailed || newStatus == model.PaymentStatusCanceled || newStatus == model.PaymentStatusExpired {
			transaction.FailureReason = payload.TransactionStatus
		}

		return tx.Save(transaction).Error
	})
}

func isMidtransRefundStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "refund", "partial_refund", "chargeback", "partial_chargeback":
		return true
	default:
		return false
	}
}

func hasMidtransRefundData(payload *dto.MidtransWebhookRequest) bool {
	if payload == nil {
		return false
	}
	if isMidtransRefundStatus(payload.TransactionStatus) || len(payload.Refunds) > 0 ||
		parseProviderReference(payload.RefundChargebackID) != "" || strings.TrimSpace(payload.RefundKey) != "" {
		return true
	}
	amount, ok := parseIDRAmount(payload.RefundAmount)
	return ok && amount > 0
}

func (s *Service) ConsumeChatQuota(ctx context.Context, userID uint) (*entitlement.ChatQuotaResult, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	hasUnlimitedAccess, err := s.hasPremiumEntitlement(ctx, user)
	if err != nil {
		return nil, err
	}

	if hasUnlimitedAccess {
		return &entitlement.ChatQuotaResult{
			Allowed:     true,
			Limit:       0,
			Used:        0,
			Remaining:   0,
			IsUnlimited: true,
		}, nil
	}

	quotaWindow := s.currentChatQuotaWindow(time.Now())
	used, remaining, allowed, err := s.repo.ConsumeFeatureUsage(ctx, userID, featureKeyChatAI, quotaWindow.Start, s.dailyChatLimit)
	if err != nil {
		return nil, err
	}

	if !allowed {
		return &entitlement.ChatQuotaResult{
			Allowed:     false,
			Limit:       s.dailyChatLimit,
			Used:        used,
			Remaining:   remaining,
			IsUnlimited: false,
		}, ErrChatQuotaExceeded
	}

	return &entitlement.ChatQuotaResult{
		Allowed:     true,
		Limit:       s.dailyChatLimit,
		Used:        used,
		Remaining:   remaining,
		IsUnlimited: false,
	}, nil
}

func (s *Service) ListTransactions(ctx context.Context, userID *uint, params TransactionListParams) (*TransactionListResult, error) {
	effectiveUserID := userID
	if effectiveUserID == nil && params.UserID != nil {
		effectiveUserID = params.UserID
	}

	// Validate and normalize pagination params BEFORE query
	page := params.Page
	limit := params.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	filter := infrastructure.TransactionListFilter{
		UserID:                     effectiveUserID,
		Status:                     strings.TrimSpace(params.Status),
		ItemType:                   strings.TrimSpace(params.ItemType),
		RefundReconciliationStatus: strings.TrimSpace(params.RefundReconciliationStatus),
		StartDate:                  params.StartDate,
		EndDate:                    params.EndDate,
		Page:                       page,
		Limit:                      limit,
	}
	transactions, total, err := s.repo.ListTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}

	result := make([]dto.PaymentTransactionDTO, 0, len(transactions))
	for _, transaction := range transactions {
		result = append(result, toTransactionDTO(transaction))
	}

	return &TransactionListResult{
		Transactions: result,
		Total:        total,
		Page:         filter.Page,
		Limit:        filter.Limit,
		TotalPages:   totalPages,
	}, nil
}

func (s *Service) BuildTransactionsCSV(ctx context.Context, params TransactionListParams) (*ExportCSVResult, error) {
	filter := infrastructure.TransactionListFilter{
		UserID:                     params.UserID,
		Status:                     strings.TrimSpace(params.Status),
		ItemType:                   strings.TrimSpace(params.ItemType),
		RefundReconciliationStatus: strings.TrimSpace(params.RefundReconciliationStatus),
		StartDate:                  params.StartDate,
		EndDate:                    params.EndDate,
	}

	transactions, err := s.repo.GetTransactionsForExport(ctx, filter)
	if err != nil {
		return nil, err
	}

	builder := strings.Builder{}
	writer := csv.NewWriter(&builder)

	header := []string{
		"id",
		"order_id",
		"user_id",
		"item_type",
		"item_name",
		"amount",
		"currency",
		"status",
		"payment_provider",
		"provider_transaction_id",
		"provider_payment_type",
		"refund_requested_amount",
		"refunded_amount",
		"provider_refund_amount_reported",
		"refund_status",
		"refund_reconciliation_status",
		"paid_at",
		"created_at",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	for _, txData := range transactions {
		paidAt := ""
		if txData.PaidAt != nil {
			paidAt = txData.PaidAt.Format(time.RFC3339)
		}

		record := []string{
			strconv.FormatUint(uint64(txData.ID), 10),
			txData.OrderID,
			strconv.FormatUint(uint64(txData.UserID), 10),
			string(txData.ItemType),
			txData.ItemName,
			strconv.Itoa(txData.Amount),
			txData.Currency,
			string(txData.Status),
			txData.PaymentProvider,
			txData.ProviderTransactionID,
			txData.ProviderPaymentType,
			strconv.FormatInt(txData.RefundRequestedAmount, 10),
			strconv.FormatInt(txData.RefundedAmount, 10),
			strconv.FormatInt(txData.ProviderRefundAmountReported, 10),
			txData.RefundStatus,
			txData.RefundReconciliationStatus,
			paidAt,
			txData.CreatedAt.Format(time.RFC3339),
		}

		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("billing_transactions_%s.csv", time.Now().Format("20060102_150405"))
	return &ExportCSVResult{Filename: filename, Content: builder.String()}, nil
}

// BuildInvoiceCSV menghasilkan CSV untuk satu transaksi (invoice) berdasarkan
// order ID. Bila [userID] tidak nil, kepemilikan transaksi diverifikasi —
// pengguna hanya boleh mengunduh invoice miliknya sendiri.
func (s *Service) BuildInvoiceCSV(ctx context.Context, orderID string, userID *uint) (*ExportCSVResult, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return nil, errors.New("order ID is required")
	}

	txData, err := s.repo.GetTransactionByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Scope: pastikan transaksi milik pengguna yang meminta.
	if userID != nil && txData.UserID != *userID {
		return nil, infrastructure.ErrTransactionNotFound
	}

	builder := strings.Builder{}
	writer := csv.NewWriter(&builder)

	// Format "field,value" agar mudah dibaca sebagai bukti pembayaran.
	paidAt := ""
	if txData.PaidAt != nil {
		paidAt = txData.PaidAt.Format(time.RFC3339)
	}
	rows := [][]string{
		{"field", "value"},
		{"order_id", txData.OrderID},
		{"item_type", string(txData.ItemType)},
		{"item_name", txData.ItemName},
		{"amount", strconv.Itoa(txData.Amount)},
		{"currency", txData.Currency},
		{"status", string(txData.Status)},
		{"payment_provider", txData.PaymentProvider},
		{"provider_transaction_id", txData.ProviderTransactionID},
		{"provider_payment_type", txData.ProviderPaymentType},
		{"paid_at", paidAt},
		{"created_at", txData.CreatedAt.Format(time.RFC3339)},
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("invoice_%s.csv", txData.OrderID)
	return &ExportCSVResult{Filename: filename, Content: builder.String()}, nil
}

func (s *Service) UpsertPremiumPlan(ctx context.Context, plan *model.PremiumPlan) error {
	if plan == nil {
		return errors.New("plan payload is required")
	}
	if strings.TrimSpace(plan.Code) == "" || strings.TrimSpace(plan.Name) == "" || plan.Price <= 0 || plan.DurationDays <= 0 {
		return errors.New("invalid premium plan payload")
	}

	if plan.ID == 0 {
		return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
			return tx.Create(plan).Error
		})
	}

	if _, err := s.repo.GetPlanByID(ctx, plan.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPremiumPlanNotFound
		}
		return err
	}

	return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.Save(plan).Error
	})
}

func (s *Service) UpsertTopupPackage(ctx context.Context, pkg *model.TopupPackage) error {
	if pkg == nil {
		return errors.New("topup payload is required")
	}
	if strings.TrimSpace(pkg.Code) == "" || strings.TrimSpace(pkg.Name) == "" || pkg.Price <= 0 || pkg.Coins <= 0 {
		return errors.New("invalid topup package payload")
	}

	if pkg.ID == 0 {
		return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
			return tx.Create(pkg).Error
		})
	}

	if _, err := s.repo.GetTopupPackageByID(ctx, pkg.ID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTopupPackageNotFound
		}
		return err
	}

	return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.Save(pkg).Error
	})
}

func (s *Service) GetAllPlans(ctx context.Context, activeOnly bool) ([]model.PremiumPlan, error) {
	if activeOnly {
		return s.repo.GetActivePlans(ctx)
	}

	var plans []model.PremiumPlan
	err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Order("price ASC").Find(&plans).Error
	})
	if err != nil {
		return nil, err
	}
	return plans, nil
}

func (s *Service) GetAllTopupPackages(ctx context.Context, activeOnly bool) ([]model.TopupPackage, error) {
	if activeOnly {
		return s.repo.GetActiveTopupPackages(ctx)
	}

	var packages []model.TopupPackage
	err := s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Order("price ASC").Find(&packages).Error
	})
	if err != nil {
		return nil, err
	}
	return packages, nil
}

func EncodeWebhookPayload(payload *dto.MidtransWebhookRequest) string {
	if payload == nil {
		return "{}"
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
