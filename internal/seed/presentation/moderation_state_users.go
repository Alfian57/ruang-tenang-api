package presentation

import (
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedModerationStateUsers seeds extra non-fixed user accounts in varied
// moderation states (blocked, banned, suspended, and a clean extra active user)
// so the admin & moderation pages have realistic data to test against.
//
// These accounts are separate from the four fixed demo accounts (which are
// always reset to a clean state). Their emails are whitelisted in
// presentationAccountEmails so the cleanup step preserves them.
func SeedModerationStateUsers(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	suspensionEnd := now.Add(5 * 24 * time.Hour)

	type stateUser struct {
		Name             string
		Email            string
		Exp              int64
		GoldCoins        int64
		IsBlocked        bool
		IsBanned         bool
		BanReason        string
		SuspensionEnd    *time.Time
		SuspensionReason string
		IsForumBlocked   bool
	}

	users := []stateUser{
		{
			Name:      "Sinta Wijaya (Aktif)",
			Email:     presentationActiveExtraEmail,
			Exp:       640,
			GoldCoins: 300,
		},
		{
			Name:      "Bayu Pratama (Diblokir)",
			Email:     presentationBlockedEmail,
			Exp:       150,
			GoldCoins: 60,
			IsBlocked: true,
		},
		{
			Name:      "Galih Nugroho (Dibanned)",
			Email:     presentationBannedEmail,
			Exp:       80,
			GoldCoins: 20,
			IsBanned:  true,
			BanReason: "Pelanggaran berulang pedoman komunitas (data contoh).",
		},
		{
			Name:             "Maya Lestari (Disuspend)",
			Email:            presentationSuspendedEmail,
			Exp:              210,
			GoldCoins:        90,
			SuspensionEnd:    &suspensionEnd,
			SuspensionReason: "Suspensi sementara karena spam (data contoh).",
			IsForumBlocked:   true,
		},
	}

	for _, u := range users {
		updates := map[string]interface{}{
			"name":               u.Name,
			"password":           string(hashedPassword),
			"role":               model.RoleUser,
			"exp":                u.Exp,
			"gold_coins":         u.GoldCoins,
			"profile_theme":      presentationDefaultProfileTheme,
			"is_premium":         false,
			"premium_since":      nil,
			"premium_expires_at": nil,
			"is_blocked":         u.IsBlocked,
			"is_forum_blocked":   u.IsForumBlocked,
			"is_banned":          u.IsBanned,
			"ban_reason":         u.BanReason,
			"suspension_end":     u.SuspensionEnd,
			"suspension_reason":  u.SuspensionReason,
			"deleted_at":         nil,
		}

		var existing model.User
		err := db.Unscoped().Where("email = ?", u.Email).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err == nil {
			if err := db.Unscoped().Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			continue
		}

		user := model.User{
			Name:             u.Name,
			Email:            u.Email,
			Password:         string(hashedPassword),
			Role:             model.RoleUser,
			Exp:              u.Exp,
			GoldCoins:        u.GoldCoins,
			ProfileTheme:     presentationDefaultProfileTheme,
			IsBlocked:        u.IsBlocked,
			IsForumBlocked:   u.IsForumBlocked,
			IsBanned:         u.IsBanned,
			BanReason:        u.BanReason,
			SuspensionEnd:    u.SuspensionEnd,
			SuspensionReason: u.SuspensionReason,
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}
