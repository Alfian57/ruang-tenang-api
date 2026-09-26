package application

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
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
	ErrItemNotFound            = errors.New("billing item not found")
	ErrItemNotActive           = errors.New("billing item is not active")
	ErrDuitkuNotConfigured     = errors.New("duitku is not configured")
	ErrWebhookSignatureInvalid = errors.New("invalid webhook signature")
	ErrWebhookPayloadInvalid   = errors.New("invalid webhook payment details")
	ErrWebhookDuplicate        = errors.New("webhook event already processed")
	ErrChatQuotaExceeded       = errors.New("chat quota exceeded")
	ErrPremiumPlanNotFound     = errors.New("premium plan not found")
	ErrTopupPackageNotFound    = errors.New("topup package not found")
	// ErrPersonalPremiumBlockedByB2B is returned when a user covered by an active
	// B2B premium seat tries to purchase a personal premium subscription.
	ErrPersonalPremiumBlockedByB2B = errors.New("personal premium is unavailable while you have active B2B premium access")
)

type ServiceConfig struct {
	DuitkuMerchantCode string
	DuitkuAPIKey       string
	DuitkuCallbackURL  string
	FrontendURL        string
	DefaultDailyLimit  int
	ResetInterval      string
}

type Service struct {
	repo               *infrastructure.BillingRepository
	duitkuClient       DuitkuClient
	merchantCode       string
	apiKey             string
	callbackURL        string
	frontendURL        string
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
	duitkuClient DuitkuClient,
	cfg ServiceConfig,
) *Service {
	limit := cfg.DefaultDailyLimit
	if limit <= 0 {
		limit = defaultDailyChatMessageLimit
	}
	resetInterval := parseChatQuotaResetInterval(cfg.ResetInterval)

	return &Service{
		repo:               repo,
		duitkuClient:       duitkuClient,
		merchantCode:       strings.TrimSpace(cfg.DuitkuMerchantCode),
		apiKey:             strings.TrimSpace(cfg.DuitkuAPIKey),
		callbackURL:        strings.TrimRight(strings.TrimSpace(cfg.DuitkuCallbackURL), "/"),
		frontendURL:        strings.TrimRight(strings.TrimSpace(cfg.FrontendURL), "/"),
		dailyChatLimit:     limit,
		quotaResetInterval: resetInterval,
	}
}

type TransactionListParams struct {
	UserID    *uint
	Status    string
	ItemType  string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	Limit     int
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
		ID:                    tx.ID,
		OrderID:               tx.OrderID,
		UserID:                tx.UserID,
		ItemType:              string(tx.ItemType),
		ItemID:                tx.ItemID,
		ItemName:              tx.ItemName,
		Amount:                tx.Amount,
		Currency:              tx.Currency,
		Status:                string(tx.Status),
		PaymentProvider:       tx.PaymentProvider,
		ProviderTransactionID: tx.ProviderTransactionID,
		ProviderPaymentType:   tx.ProviderPaymentType,
		FailureReason:         tx.FailureReason,
		ProviderReference:     tx.ProviderReference,
		PaymentURL:            tx.PaymentURL,
		PaidAt:                tx.PaidAt,
		CreatedAt:             tx.CreatedAt,
		UpdatedAt:             tx.UpdatedAt,
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
	if s.duitkuClient == nil || !s.duitkuClient.IsConfigured() || s.callbackURL == "" || s.frontendURL == "" {
		return nil, ErrDuitkuNotConfigured
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
		PaymentProvider: "duitku",
	}

	var amount int
	var itemName string

	switch itemType {
	case model.BillingItemTypeSubscription:
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
	}

	transaction.Amount = amount
	transaction.ItemName = itemName
	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return nil, err
	}

	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	returnURL := s.frontendURL + "/payment/success"
	customerName := truncateRunes(strings.TrimSpace(user.Name), 20)
	if customerName == "" {
		customerName = "Ruang Tenang"
	}
	invoice, err := s.duitkuClient.CreateInvoice(ctx, DuitkuInvoiceRequest{
		PaymentAmount:    amount,
		MerchantOrderID:  orderID,
		ProductDetails:   truncateRunes("Ruang Tenang - "+itemName, 255),
		Email:            user.Email,
		MerchantUserInfo: user.Email,
		CustomerVaName:   customerName,
		ItemDetails: []DuitkuItemDetail{{
			Name:     truncateRunes(itemName, 50),
			Price:    amount,
			Quantity: 1,
		}},
		CustomerDetail: DuitkuCustomerDetail{
			FirstName: truncateRunes(strings.TrimSpace(user.Name), 50),
			Email:     user.Email,
		},
		CallbackURL: s.callbackURL,
		ReturnURL:   returnURL,
	})
	if err != nil {
		_ = s.repo.SetCheckoutFailure(ctx, orderID, err.Error())
		return nil, err
	}

	if err := s.repo.UpdateCheckoutLink(ctx, orderID, invoice.Reference, invoice.PaymentURL); err != nil {
		return nil, err
	}
	transaction, err = s.repo.GetTransactionByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return &dto.CreateCheckoutResponse{
		TransactionID:     transaction.ID,
		OrderID:           transaction.OrderID,
		ItemType:          string(transaction.ItemType),
		ItemID:            transaction.ItemID,
		ItemName:          transaction.ItemName,
		Amount:            transaction.Amount,
		Currency:          transaction.Currency,
		Status:            string(transaction.Status),
		ProviderReference: transaction.ProviderReference,
		PaymentURL:        transaction.PaymentURL,
		ExpiresAt:         transaction.ExpiresAt,
	}, nil
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func (s *Service) verifyWebhookSignature(payload *dto.DuitkuWebhookRequest) bool {
	if payload == nil || s.apiKey == "" || s.merchantCode == "" || strings.TrimSpace(payload.MerchantCode) != s.merchantCode {
		return false
	}

	stringToSign := payload.MerchantCode + payload.Amount + payload.MerchantOrderID
	mac := hmac.New(sha256.New, []byte(s.apiKey))
	_, _ = mac.Write([]byte(stringToSign))
	expected := hex.EncodeToString(mac.Sum(nil))
	actual := strings.ToLower(strings.TrimSpace(payload.Signature))
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func webhookAmountMatches(rawAmount string, amount int) bool {
	value, err := strconv.ParseInt(strings.TrimSpace(rawAmount), 10, 64)
	return err == nil && value > 0 && value == int64(amount)
}

func mapDuitkuResultCode(resultCode string) (model.PaymentStatus, bool) {
	switch strings.TrimSpace(resultCode) {
	case "00":
		return model.PaymentStatusPaid, true
	case "01":
		return model.PaymentStatusFailed, true
	default:
		return model.PaymentStatusPending, false
	}
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

func (s *Service) HandleDuitkuWebhook(ctx context.Context, payload *dto.DuitkuWebhookRequest, rawPayload string) error {
	if payload == nil {
		return errors.New("empty webhook payload")
	}
	if !s.verifyWebhookSignature(payload) {
		return ErrWebhookSignatureInvalid
	}
	if strings.TrimSpace(payload.MerchantOrderID) == "" || strings.TrimSpace(payload.Reference) == "" ||
		strings.TrimSpace(payload.PaymentCode) == "" || strings.TrimSpace(payload.Amount) == "" {
		return ErrWebhookPayloadInvalid
	}
	amount, amountErr := strconv.ParseInt(strings.TrimSpace(payload.Amount), 10, 64)
	if amountErr != nil || amount <= 0 {
		return ErrWebhookPayloadInvalid
	}
	newStatus, validResult := mapDuitkuResultCode(payload.ResultCode)
	if !validResult {
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
	event := &model.PaymentWebhookEvent{
		Provider:    "duitku",
		OrderID:     payload.MerchantOrderID,
		EventKey:    fmt.Sprintf("duitku:%s:%s", payload.MerchantOrderID, hex.EncodeToString(payloadHash[:])),
		Payload:     rawPayload,
		ProcessedAt: time.Now(),
	}

	return s.repo.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.repo.CreateWebhookEventTx(tx, event); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
				return ErrWebhookDuplicate
			}
			return err
		}

		transaction, err := s.repo.LockTransactionByOrderID(tx, payload.MerchantOrderID)
		if err != nil {
			return err
		}
		if transaction.PaymentProvider != "duitku" || !webhookAmountMatches(payload.Amount, transaction.Amount) ||
			(transaction.ProviderReference != "" && transaction.ProviderReference != payload.Reference) {
			return ErrWebhookPayloadInvalid
		}
		if transaction.IsFinalStatus() {
			return nil
		}

		transaction.Status = newStatus
		transaction.ProviderReference = payload.Reference
		transaction.ProviderPaymentType = payload.PaymentCode
		transaction.ProviderTransactionID = payload.PublisherOrderID
		if transaction.ProviderTransactionID == "" {
			transaction.ProviderTransactionID = payload.Reference
		}
		transaction.CallbackPayload = rawPayload
		if newStatus == model.PaymentStatusPaid {
			if err := s.applySuccessfulPayment(tx, transaction, nil); err != nil {
				return err
			}
		} else {
			transaction.FailureReason = "Duitku result code " + payload.ResultCode
		}
		return tx.Save(transaction).Error
	})
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
		UserID:    effectiveUserID,
		Status:    strings.TrimSpace(params.Status),
		ItemType:  strings.TrimSpace(params.ItemType),
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
		Page:      page,
		Limit:     limit,
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
		UserID:    params.UserID,
		Status:    strings.TrimSpace(params.Status),
		ItemType:  strings.TrimSpace(params.ItemType),
		StartDate: params.StartDate,
		EndDate:   params.EndDate,
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
