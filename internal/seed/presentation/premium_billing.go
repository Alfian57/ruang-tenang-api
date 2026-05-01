package presentation

import (
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedPremiumAndTopupData seeds personal premium, topup, transaction, and usage flows.
func SeedPremiumAndTopupData(db *gorm.DB) error {
	if err := ensureBillingCatalog(db); err != nil {
		return err
	}

	emails := []string{
		"gading@gmail.com",
		"dery@gmail.com",
		"andhika@gmail.com",
	}

	var users []model.User
	if err := db.Where("email IN ?", emails).Find(&users).Error; err != nil {
		return err
	}

	userByEmail := make(map[string]model.User, len(users))
	for _, user := range users {
		userByEmail[user.Email] = user
	}

	var yearlyPlan model.PremiumPlan
	if err := db.Where("code = ?", "premium_yearly").First(&yearlyPlan).Error; err != nil {
		return err
	}

	var topup500 model.TopupPackage
	if err := db.Where("code = ?", "topup_500").First(&topup500).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	yearlyStartsAt := now.AddDate(0, -3, 0)
	yearlyEndsAt := yearlyStartsAt.AddDate(0, 0, yearlyPlan.DurationDays)

	if user, ok := userByEmail["gading@gmail.com"]; ok {
		if err := db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"is_premium":         true,
			"premium_since":      yearlyStartsAt,
			"premium_expires_at": yearlyEndsAt,
		}).Error; err != nil {
			return err
		}

		if err := upsertUserSubscription(
			db,
			"DEV-SUB-GADING-YEARLY",
			user.ID,
			yearlyPlan.ID,
			model.SubscriptionStatusActive,
			yearlyStartsAt,
			yearlyEndsAt,
		); err != nil {
			return err
		}

		if err := upsertPaymentTransaction(db, model.PaymentTransaction{
			OrderID:               "DEV-PAY-SUB-GADING-YEARLY",
			UserID:                user.ID,
			ItemType:              model.BillingItemTypeSubscription,
			ItemID:                yearlyPlan.ID,
			ItemName:              yearlyPlan.Name,
			Amount:                yearlyPlan.Price,
			Currency:              "IDR",
			Status:                model.PaymentStatusPaid,
			PaymentProvider:       "midtrans",
			ProviderTransactionID: "TX-SUB-GADING-YEARLY",
			ProviderPaymentType:   "qris",
			SnapToken:             "snap-dev-sub-gading-yearly",
			SnapRedirectURL:       "https://app.midtrans.com/snap/v2/dev-sub-gading-yearly",
			PaidAt:                &yearlyStartsAt,
			ExpiresAt:             &yearlyEndsAt,
		}); err != nil {
			return err
		}

		if err := upsertWebhookEvent(db, model.PaymentWebhookEvent{
			Provider:    "midtrans",
			OrderID:     "DEV-PAY-SUB-GADING-YEARLY",
			EventKey:    "midtrans:DEV-PAY-SUB-GADING-YEARLY:settlement",
			Payload:     `{"transaction_status":"settlement","order_id":"DEV-PAY-SUB-GADING-YEARLY"}`,
			ProcessedAt: now,
		}); err != nil {
			return err
		}
	}

	if user, ok := userByEmail["dery@gmail.com"]; ok {
		if err := db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"is_premium":         false,
			"premium_since":      nil,
			"premium_expires_at": nil,
		}).Error; err != nil {
			return err
		}
	}

	if user, ok := userByEmail["andhika@gmail.com"]; ok {
		if err := db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"is_premium":         false,
			"premium_since":      nil,
			"premium_expires_at": nil,
			"gold_coins":         260,
		}).Error; err != nil {
			return err
		}

		topupPaidAt := now.AddDate(0, 0, -4)
		topupExpiresAt := now.AddDate(0, 0, 3)
		if err := upsertPaymentTransaction(db, model.PaymentTransaction{
			OrderID:               "SEED-PAY-TOPUP-ANDHIKA",
			UserID:                user.ID,
			ItemType:              model.BillingItemTypeTopup,
			ItemID:                topup500.ID,
			ItemName:              topup500.Name,
			Amount:                topup500.Price,
			Currency:              "IDR",
			Status:                model.PaymentStatusPaid,
			PaymentProvider:       "midtrans",
			ProviderTransactionID: "TX-TOPUP-ANDHIKA",
			ProviderPaymentType:   "gopay",
			SnapToken:             "snap-seed-topup-andhika",
			SnapRedirectURL:       "https://app.midtrans.com/snap/v2/seed-topup-andhika",
			PaidAt:                &topupPaidAt,
			ExpiresAt:             &topupExpiresAt,
		}); err != nil {
			return err
		}

		if err := upsertWebhookEvent(db, model.PaymentWebhookEvent{
			Provider:    "midtrans",
			OrderID:     "SEED-PAY-TOPUP-ANDHIKA",
			EventKey:    "midtrans:SEED-PAY-TOPUP-ANDHIKA:settlement",
			Payload:     `{"transaction_status":"settlement","order_id":"SEED-PAY-TOPUP-ANDHIKA"}`,
			ProcessedAt: now,
		}); err != nil {
			return err
		}
	}

	usageDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if user, ok := userByEmail["andhika@gmail.com"]; ok {
		if err := upsertFeatureUsage(db, user.ID, model.FeatureKeyChatAIMessages, usageDate, 24); err != nil {
			return err
		}
	}
	if user, ok := userByEmail["dery@gmail.com"]; ok {
		if err := upsertFeatureUsage(db, user.ID, model.FeatureKeyChatAIMessages, usageDate, 6); err != nil {
			return err
		}
	}

	return nil
}

func ensureBillingCatalog(db *gorm.DB) error {
	plans := []model.PremiumPlan{
		{Code: "premium_monthly", Name: "Premium Monthly", Description: "Akses premium 30 hari untuk personal use.", Price: 29900, DurationDays: 30, IsActive: true},
		{Code: "premium_yearly", Name: "Premium Yearly", Description: "Akses premium 365 hari untuk komitmen jangka panjang.", Price: 279000, DurationDays: 365, IsActive: true},
	}
	for _, plan := range plans {
		var existing model.PremiumPlan
		err := db.Where("code = ?", plan.Code).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&plan).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	pkgs := []model.TopupPackage{
		{Code: "topup_500", Name: "Topup 500 Koin", Coins: 500, BonusCoins: 50, Price: 55000, IsActive: true},
	}
	for _, pkg := range pkgs {
		var existing model.TopupPackage
		err := db.Where("code = ?", pkg.Code).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&pkg).Error; err != nil {
					return err
				}
				continue
			}
			return err
		}
	}

	return nil
}

func upsertUserSubscription(db *gorm.DB, sourceOrderID string, userID, planID uint, status model.SubscriptionStatus, startsAt, endsAt time.Time) error {
	var existing model.UserSubscription
	err := db.Where("source_order_id = ?", sourceOrderID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item := model.UserSubscription{
				UserID:        userID,
				PlanID:        planID,
				SourceOrderID: sourceOrderID,
				Status:        status,
				StartsAt:      startsAt,
				EndsAt:        endsAt,
			}
			return db.Create(&item).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"user_id":   userID,
		"plan_id":   planID,
		"status":    status,
		"starts_at": startsAt,
		"ends_at":   endsAt,
	}).Error
}

func upsertPaymentTransaction(db *gorm.DB, tx model.PaymentTransaction) error {
	var existing model.PaymentTransaction
	err := db.Where("order_id = ?", tx.OrderID).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&tx).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"user_id":                 tx.UserID,
		"item_type":               tx.ItemType,
		"item_id":                 tx.ItemID,
		"item_name":               tx.ItemName,
		"amount":                  tx.Amount,
		"currency":                tx.Currency,
		"status":                  tx.Status,
		"payment_provider":        tx.PaymentProvider,
		"provider_transaction_id": tx.ProviderTransactionID,
		"provider_payment_type":   tx.ProviderPaymentType,
		"snap_token":              tx.SnapToken,
		"snap_redirect_url":       tx.SnapRedirectURL,
		"callback_payload":        tx.CallbackPayload,
		"failure_reason":          tx.FailureReason,
		"paid_at":                 tx.PaidAt,
		"expires_at":              tx.ExpiresAt,
	}).Error
}

func upsertWebhookEvent(db *gorm.DB, event model.PaymentWebhookEvent) error {
	var existing model.PaymentWebhookEvent
	err := db.Where("event_key = ?", event.EventKey).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&event).Error
		}
		return err
	}

	return db.Model(&existing).Updates(map[string]interface{}{
		"provider":     event.Provider,
		"order_id":     event.OrderID,
		"payload":      event.Payload,
		"processed_at": event.ProcessedAt,
	}).Error
}

func upsertFeatureUsage(db *gorm.DB, userID uint, featureKey string, usageDate time.Time, usedCount int) error {
	var existing model.UserFeatureUsage
	usageWindowStart := usageDate.Truncate(time.Second)
	err := db.Where("user_id = ? AND feature_key = ? AND usage_window_start = ?", userID, featureKey, usageWindowStart).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			item := model.UserFeatureUsage{
				UserID:           userID,
				FeatureKey:       featureKey,
				UsageDate:        usageDate,
				UsageWindowStart: usageWindowStart,
				UsedCount:        usedCount,
			}
			return db.Create(&item).Error
		}
		return err
	}

	return db.Model(&existing).Update("used_count", usedCount).Error
}
