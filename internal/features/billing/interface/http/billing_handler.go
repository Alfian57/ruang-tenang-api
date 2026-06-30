package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	billingapp "github.com/Alfian57/ruang-tenang-api/internal/features/billing/application"
	billinginfra "github.com/Alfian57/ruang-tenang-api/internal/features/billing/infrastructure"
	"github.com/Alfian57/ruang-tenang-api/internal/middleware"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/gin-gonic/gin"
)

type BillingHandler struct {
	service *billingapp.Service
}

type premiumPlanPayload struct {
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	Price        int    `json:"price" binding:"required,gt=0"`
	DurationDays int    `json:"duration_days" binding:"required,gt=0"`
	IsActive     *bool  `json:"is_active"`
}

type topupPackagePayload struct {
	Code       string `json:"code" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Coins      int64  `json:"coins" binding:"required,gt=0"`
	BonusCoins int64  `json:"bonus_coins"`
	Price      int    `json:"price" binding:"required,gt=0"`
	IsActive   *bool  `json:"is_active"`
}

func NewBillingHandler(service *billingapp.Service) *BillingHandler {
	return &BillingHandler{service: service}
}

func (h *BillingHandler) requireUserID(c *gin.Context) (uint, bool) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		c.JSON(http.StatusUnauthorized, dto.Response{Success: false, Message: "Unauthorized"})
		return 0, false
	}
	return userID, true
}

func parseDateRange(c *gin.Context) (*time.Time, *time.Time, error) {
	var startDate *time.Time
	var endDate *time.Time

	if rawStart := strings.TrimSpace(c.Query("start_date")); rawStart != "" {
		parsed, err := time.Parse("2006-01-02", rawStart)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD)")
		}
		startDate = &parsed
	}

	if rawEnd := strings.TrimSpace(c.Query("end_date")); rawEnd != "" {
		parsed, err := time.Parse("2006-01-02", rawEnd)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end_date format (expected YYYY-MM-DD)")
		}
		inclusiveEnd := parsed.AddDate(0, 0, 1).Add(-time.Nanosecond)
		endDate = &inclusiveEnd
	}

	return startDate, endDate, nil
}

func parsePagination(c *gin.Context) (int, int, error) {
	page := 1
	limit := 10

	if rawPage := strings.TrimSpace(c.Query("page")); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("invalid page parameter")
		}
		page = parsed
	}

	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("invalid limit parameter")
		}
		limit = parsed
	}

	return page, limit, nil
}

func (h *BillingHandler) parseTransactionListParams(c *gin.Context) (billingapp.TransactionListParams, error) {
	startDate, endDate, err := parseDateRange(c)
	if err != nil {
		return billingapp.TransactionListParams{}, err
	}

	page, limit, err := parsePagination(c)
	if err != nil {
		return billingapp.TransactionListParams{}, err
	}

	return billingapp.TransactionListParams{
		Status:    strings.TrimSpace(c.Query("status")),
		ItemType:  strings.TrimSpace(c.Query("item_type")),
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Limit:     limit,
	}, nil
}

func (h *BillingHandler) GetCatalog(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	catalog, err := h.service.GetCatalog(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load billing catalog"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(catalog, ""))
}

func (h *BillingHandler) GetStatus(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	status, err := h.service.GetStatus(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load billing status"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(status, ""))
}

func (h *BillingHandler) CreateCheckout(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	var req dto.CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.CreateCheckout(ctx, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, billingapp.ErrMidtransNotConfigured):
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse("Payment gateway is not configured"))
		case errors.Is(err, billingapp.ErrItemNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Billing item not found"))
		case errors.Is(err, billingapp.ErrItemNotActive):
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("Billing item is not active"))
		case errors.Is(err, billingapp.ErrPersonalPremiumBlockedByB2B):
			c.JSON(http.StatusConflict, dto.ErrorResponse("Kamu sudah mendapat Premium melalui organisasi B2B, jadi tidak bisa membeli Premium pribadi."))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to create checkout"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, "Checkout created"))
}

func (h *BillingHandler) GetMyTransactions(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	params, err := h.parseTransactionListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	result, err := h.service.ListTransactions(ctx, &userID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load transactions"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

// GetMyTransactionsExport mengekspor riwayat transaksi milik pengguna saat ini
// sebagai CSV, menghormati filter status/item_type/tanggal.
func (h *BillingHandler) GetMyTransactionsExport(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	params, err := h.parseTransactionListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}
	// Selalu paksa scope ke pengguna saat ini (abaikan user_id dari query).
	params.UserID = &userID

	result, err := h.service.BuildTransactionsCSV(ctx, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to export transactions"))
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.Filename))
	c.String(http.StatusOK, result.Content)
}

// GetMyInvoice mengunduh satu invoice (CSV) milik pengguna berdasarkan order ID.
func (h *BillingHandler) GetMyInvoice(c *gin.Context) {
	ctx := c.Request.Context()
	userID, ok := h.requireUserID(c)
	if !ok {
		return
	}

	orderID := strings.TrimSpace(c.Param("orderId"))
	if orderID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("order ID is required"))
		return
	}

	result, err := h.service.BuildInvoiceCSV(ctx, orderID, &userID)
	if err != nil {
		if errors.Is(err, billinginfra.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Invoice tidak ditemukan"))
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to build invoice"))
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.Filename))
	c.String(http.StatusOK, result.Content)
}

func (h *BillingHandler) HandleMidtransWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	rawBody, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid webhook payload"))
		return
	}

	var req dto.MidtransWebhookRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		// Don't echo the unmarshal error to the client (info leak on a public endpoint).
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("Invalid webhook payload"))
		return
	}

	rawPayload := string(rawBody)
	err = h.service.HandleMidtransWebhook(ctx, &req, rawPayload)
	if err != nil {
		switch {
		case errors.Is(err, billingapp.ErrWebhookSignatureInvalid):
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse("Invalid signature"))
		case errors.Is(err, billingapp.ErrWebhookDuplicate):
			c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Webhook already processed"))
		case errors.Is(err, billinginfra.ErrTransactionNotFound):
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Transaction not found"))
		default:
			c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to process webhook"))
		}
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(nil, "Webhook processed"))
}

func (h *BillingHandler) AdminGetTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := h.parseTransactionListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	var userIDPtr *uint
	if rawUserID := strings.TrimSpace(c.Query("user_id")); rawUserID != "" {
		parsed, parseErr := strconv.ParseUint(rawUserID, 10, 32)
		if parseErr != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid user_id parameter"))
			return
		}
		userID := uint(parsed)
		userIDPtr = &userID
	}

	result, err := h.service.ListTransactions(ctx, userIDPtr, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load transactions"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(result, ""))
}

func (h *BillingHandler) AdminExportTransactionsCSV(c *gin.Context) {
	ctx := c.Request.Context()

	params, err := h.parseTransactionListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	if rawUserID := strings.TrimSpace(c.Query("user_id")); rawUserID != "" {
		parsed, parseErr := strconv.ParseUint(rawUserID, 10, 32)
		if parseErr != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid user_id parameter"))
			return
		}
		userID := uint(parsed)
		params.UserID = &userID
	}

	result, err := h.service.BuildTransactionsCSV(ctx, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to export transactions"))
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.Filename))
	c.String(http.StatusOK, result.Content)
}

func (h *BillingHandler) AdminGetPlans(c *gin.Context) {
	ctx := c.Request.Context()
	activeOnly, _ := strconv.ParseBool(c.DefaultQuery("active_only", "false"))

	plans, err := h.service.GetAllPlans(ctx, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load premium plans"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(plans, ""))
}

func (h *BillingHandler) AdminCreatePlan(c *gin.Context) {
	ctx := c.Request.Context()
	var req premiumPlanPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	plan := &model.PremiumPlan{
		Code:         strings.TrimSpace(req.Code),
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		Price:        req.Price,
		DurationDays: req.DurationDays,
		IsActive:     true,
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	if err := h.service.UpsertPremiumPlan(ctx, plan); err != nil {
		if errors.Is(err, billingapp.ErrPremiumPlanNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Premium plan not found"))
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(plan, "Premium plan created"))
}

func (h *BillingHandler) AdminUpdatePlan(c *gin.Context) {
	ctx := c.Request.Context()
	planID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || planID == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid plan id"))
		return
	}

	var req premiumPlanPayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	plan := &model.PremiumPlan{
		ID:           uint(planID),
		Code:         strings.TrimSpace(req.Code),
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		Price:        req.Price,
		DurationDays: req.DurationDays,
		IsActive:     true,
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	if err := h.service.UpsertPremiumPlan(ctx, plan); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(plan, "Premium plan updated"))
}

func (h *BillingHandler) AdminGetTopupPackages(c *gin.Context) {
	ctx := c.Request.Context()
	activeOnly, _ := strconv.ParseBool(c.DefaultQuery("active_only", "false"))

	packages, err := h.service.GetAllTopupPackages(ctx, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse("Failed to load topup packages"))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(packages, ""))
}

func (h *BillingHandler) AdminCreateTopupPackage(c *gin.Context) {
	ctx := c.Request.Context()
	var req topupPackagePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	pkg := &model.TopupPackage{
		Code:       strings.TrimSpace(req.Code),
		Name:       strings.TrimSpace(req.Name),
		Coins:      req.Coins,
		BonusCoins: req.BonusCoins,
		Price:      req.Price,
		IsActive:   true,
	}
	if req.IsActive != nil {
		pkg.IsActive = *req.IsActive
	}

	if err := h.service.UpsertTopupPackage(ctx, pkg); err != nil {
		if errors.Is(err, billingapp.ErrTopupPackageNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse("Topup package not found"))
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse(pkg, "Topup package created"))
}

func (h *BillingHandler) AdminUpdateTopupPackage(c *gin.Context) {
	ctx := c.Request.Context()
	packageID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || packageID == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse("invalid topup package id"))
		return
	}

	var req topupPackagePayload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	pkg := &model.TopupPackage{
		ID:         uint(packageID),
		Code:       strings.TrimSpace(req.Code),
		Name:       strings.TrimSpace(req.Name),
		Coins:      req.Coins,
		BonusCoins: req.BonusCoins,
		Price:      req.Price,
		IsActive:   true,
	}
	if req.IsActive != nil {
		pkg.IsActive = *req.IsActive
	}

	if err := h.service.UpsertTopupPackage(ctx, pkg); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(pkg, "Topup package updated"))
}
