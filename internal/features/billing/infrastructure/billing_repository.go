package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrTransactionNotFound = errors.New("payment transaction not found")

type TransactionListFilter struct {
	UserID                     *uint
	Status                     string
	ItemType                   string
	RefundReconciliationStatus string
	StartDate                  *time.Time
	EndDate                    *time.Time
	Page                       int
	Limit                      int
}

type BillingRepository struct {
	db *gorm.DB
}

func NewBillingRepository(db *gorm.DB) *BillingRepository {
	return &BillingRepository{db: db}
}

func (r *BillingRepository) GetActivePlans(ctx context.Context) ([]model.PremiumPlan, error) {
	var plans []model.PremiumPlan
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("price ASC").
		Find(&plans).Error
	return plans, err
}

func (r *BillingRepository) GetActiveTopupPackages(ctx context.Context) ([]model.TopupPackage, error) {
	var packages []model.TopupPackage
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("price ASC").
		Find(&packages).Error
	return packages, err
}

func (r *BillingRepository) GetPlanByID(ctx context.Context, id uint) (*model.PremiumPlan, error) {
	var plan model.PremiumPlan
	err := r.db.WithContext(ctx).First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *BillingRepository) GetTopupPackageByID(ctx context.Context, id uint) (*model.TopupPackage, error) {
	var pkg model.TopupPackage
	err := r.db.WithContext(ctx).First(&pkg, id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *BillingRepository) CreateTransaction(ctx context.Context, txData *model.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(txData).Error
}

func (r *BillingRepository) UpdateTransaction(ctx context.Context, txData *model.PaymentTransaction) error {
	return r.db.WithContext(ctx).Save(txData).Error
}

func (r *BillingRepository) GetTransactionByOrderID(ctx context.Context, orderID string) (*model.PaymentTransaction, error) {
	var txData model.PaymentTransaction
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&txData).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &txData, nil
}

func (r *BillingRepository) CreateWebhookEvent(ctx context.Context, event *model.PaymentWebhookEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *BillingRepository) CreateWebhookEventTx(tx *gorm.DB, event *model.PaymentWebhookEvent) error {
	return tx.Create(event).Error
}

func (r *BillingRepository) RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

func (r *BillingRepository) LockTransactionByOrderID(tx *gorm.DB, orderID string) (*model.PaymentTransaction, error) {
	var transaction model.PaymentTransaction
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("order_id = ?", orderID).
		First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func (r *BillingRepository) LockUserByID(tx *gorm.DB, userID uint) (*model.User, error) {
	var user model.User
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *BillingRepository) CreateSubscription(tx *gorm.DB, subscription *model.UserSubscription) error {
	return tx.Create(subscription).Error
}

func (r *BillingRepository) GetPlanByIDTx(tx *gorm.DB, id uint) (*model.PremiumPlan, error) {
	var plan model.PremiumPlan
	err := tx.First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *BillingRepository) GetTopupPackageByIDTx(tx *gorm.DB, id uint) (*model.TopupPackage, error) {
	var pkg model.TopupPackage
	err := tx.First(&pkg, id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *BillingRepository) GetTopupPackageByIDForUpdateTx(tx *gorm.DB, id uint) (*model.TopupPackage, error) {
	var pkg model.TopupPackage
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pkg, id).Error
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func (r *BillingRepository) FindPaymentRefundByKeyTx(tx *gorm.DB, transactionID uint, refundKey string) (*model.PaymentRefund, error) {
	var refund model.PaymentRefund
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("payment_transaction_id = ? AND refund_key = ?", transactionID, refundKey).
		First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *BillingRepository) FindPaymentRefundByProviderIDTx(tx *gorm.DB, transactionID uint, providerRefundID string) (*model.PaymentRefund, error) {
	var refund model.PaymentRefund
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("payment_transaction_id = ? AND provider_refund_id = ?", transactionID, providerRefundID).
		First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

func (r *BillingRepository) CreatePaymentRefundTx(tx *gorm.DB, refund *model.PaymentRefund) error {
	return tx.Create(refund).Error
}

func (r *BillingRepository) SavePaymentRefundTx(tx *gorm.DB, refund *model.PaymentRefund) error {
	return tx.Save(refund).Error
}

func (r *BillingRepository) UpdatePaymentRefund(ctx context.Context, refund *model.PaymentRefund) error {
	return r.db.WithContext(ctx).Save(refund).Error
}

func (r *BillingRepository) GetPaymentRefundTotalsTx(tx *gorm.DB, transactionID uint) (reserved int64, confirmed int64, err error) {
	query := tx.Model(&model.PaymentRefund{}).Where("payment_transaction_id = ?", transactionID)
	if err = query.Where("status <> ?", "rejected").Select("COALESCE(SUM(amount), 0)").Scan(&reserved).Error; err != nil {
		return 0, 0, err
	}
	if err = tx.Model(&model.PaymentRefund{}).
		Where("payment_transaction_id = ? AND status = ?", transactionID, "confirmed").
		Select("COALESCE(SUM(amount), 0)").Scan(&confirmed).Error; err != nil {
		return 0, 0, err
	}
	return reserved, confirmed, nil
}

func (r *BillingRepository) ListPaymentRefundsTx(tx *gorm.DB, transactionID uint) ([]model.PaymentRefund, error) {
	var refunds []model.PaymentRefund
	err := tx.Where("payment_transaction_id = ?", transactionID).Order("requested_at ASC, id ASC").Find(&refunds).Error
	return refunds, err
}

func (r *BillingRepository) SettleUserGoldCoinsTx(tx *gorm.DB, userID uint, coins int64) (bool, error) {
	if coins <= 0 {
		return true, nil
	}
	result := tx.Model(&model.User{}).
		Where("id = ? AND gold_coins >= ?", userID, coins).
		Update("gold_coins", gorm.Expr("gold_coins - ?", coins))
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (r *BillingRepository) CreateRefundReconciliationEventTx(tx *gorm.DB, event *model.PaymentRefundReconciliationEvent) error {
	return tx.Create(event).Error
}

func (r *BillingRepository) SaveUser(tx *gorm.DB, user *model.User) error {
	return tx.Save(user).Error
}

func (r *BillingRepository) AddUserGoldCoins(tx *gorm.DB, userID uint, coins int64) error {
	if coins <= 0 {
		return nil
	}

	return tx.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("gold_coins", gorm.Expr("gold_coins + ?", coins)).Error
}

func (r *BillingRepository) GetLatestSubscription(ctx context.Context, userID uint) (*model.UserSubscription, error) {
	var subscription model.UserSubscription
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("user_id = ?", userID).
		Order("ends_at DESC").
		First(&subscription).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &subscription, nil
}

func (r *BillingRepository) FindUserByID(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *BillingRepository) ListTransactions(ctx context.Context, filter TransactionListFilter) ([]model.PaymentTransaction, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	query := r.db.WithContext(ctx).Model(&model.PaymentTransaction{})
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ItemType != "" {
		query = query.Where("item_type = ?", filter.ItemType)
	}
	if filter.RefundReconciliationStatus != "" {
		query = query.Where("refund_reconciliation_status = ?", filter.RefundReconciliationStatus)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var transactions []model.PaymentTransaction
	err := query.
		Order("created_at DESC").
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

func (r *BillingRepository) GetTransactionsForExport(ctx context.Context, filter TransactionListFilter) ([]model.PaymentTransaction, error) {
	query := r.db.WithContext(ctx).Model(&model.PaymentTransaction{})
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ItemType != "" {
		query = query.Where("item_type = ?", filter.ItemType)
	}
	if filter.RefundReconciliationStatus != "" {
		query = query.Where("refund_reconciliation_status = ?", filter.RefundReconciliationStatus)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}

	var transactions []model.PaymentTransaction
	err := query.Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func normalizeDate(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func normalizeWindowStart(t time.Time) time.Time {
	return t.Truncate(time.Second)
}

func (r *BillingRepository) GetDailyFeatureUsage(ctx context.Context, userID uint, featureKey string, day time.Time) (int, error) {
	return r.GetFeatureUsage(ctx, userID, featureKey, normalizeDate(day))
}

func (r *BillingRepository) GetFeatureUsage(ctx context.Context, userID uint, featureKey string, windowStart time.Time) (int, error) {
	var usage model.UserFeatureUsage
	normalizedWindowStart := normalizeWindowStart(windowStart)
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND feature_key = ? AND usage_window_start = ?", userID, featureKey, normalizedWindowStart).
		First(&usage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return usage.UsedCount, nil
}

func (r *BillingRepository) ConsumeDailyFeatureUsage(ctx context.Context, userID uint, featureKey string, day time.Time, limit int) (int, int, bool, error) {
	return r.ConsumeFeatureUsage(ctx, userID, featureKey, normalizeDate(day), limit)
}

func (r *BillingRepository) ConsumeFeatureUsage(ctx context.Context, userID uint, featureKey string, windowStart time.Time, limit int) (int, int, bool, error) {
	normalizedWindowStart := normalizeWindowStart(windowStart)
	normalizedDate := normalizeDate(normalizedWindowStart)
	if limit <= 0 {
		return 0, 0, false, nil
	}
	// The unique window key serializes concurrent first messages and the WHERE
	// clause prevents increments beyond the free limit in a single statement.
	const query = `INSERT INTO user_feature_usages
		(user_id, feature_key, usage_date, usage_window_start, used_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, NOW(), NOW())
		ON CONFLICT ON CONSTRAINT uq_user_feature_usages_window
		DO UPDATE SET used_count = user_feature_usages.used_count + 1, updated_at = NOW()
		WHERE user_feature_usages.used_count < ?
		RETURNING used_count`
	var used int
	err := r.db.WithContext(ctx).Raw(query, userID, featureKey, normalizedDate, normalizedWindowStart, limit).Row().Scan(&used)
	consumed := err == nil
	if errors.Is(err, sql.ErrNoRows) {
		used, err = r.GetFeatureUsage(ctx, userID, featureKey, normalizedWindowStart)
	}
	if err != nil {
		return 0, 0, false, err
	}

	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	return used, remaining, consumed, nil
}
