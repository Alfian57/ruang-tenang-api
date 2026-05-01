package presentation

import (
	"errors"
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	presentationAdminEmail   = "admin@ruang-tenang.com"
	presentationMitraEmail   = "mitra@ruang-tenang.com"
	presentationGadingEmail  = "gading@gmail.com"
	presentationDeryEmail    = "dery@gmail.com"
	presentationAndhikaEmail = "andhika@gmail.com"
)

var presentationAccountEmails = []string{
	presentationAdminEmail,
	presentationMitraEmail,
	presentationGadingEmail,
	presentationDeryEmail,
	presentationAndhikaEmail,
}

// SeedDefaultAccounts seeds the single fixed admin account for presentation.
func SeedDefaultAccounts(db *gorm.DB) error {
	if err := cleanupPresentationAccounts(db); err != nil {
		return err
	}

	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = os.Getenv("ADMIN_PASSWORD")
	}
	if adminPassword == "" {
		adminPassword = "password"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		Name:     "Admin",
		Email:    presentationAdminEmail,
		Password: string(hashedPassword),
		Role:     model.RoleAdmin,
		Exp:      0,
	}

	var existing model.User
	err = db.Unscoped().Where("email = ?", admin.Email).First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return db.Create(&admin).Error
		}
		return err
	}

	return db.Unscoped().Model(&existing).Updates(map[string]interface{}{
		"name":                admin.Name,
		"password":            admin.Password,
		"role":                admin.Role,
		"exp":                 admin.Exp,
		"is_premium":          false,
		"premium_since":       nil,
		"premium_expires_at":  nil,
		"gold_coins":          0,
		"is_blocked":          false,
		"is_forum_blocked":    false,
		"is_banned":           false,
		"suspension_end":      nil,
		"suspension_reason":   "",
		"ban_reason":          "",
		"profile_theme":       "default",
		"profile_banner":      "",
		"avatar_border_color": "",
		"tagline":             "",
		"bio":                 "",
		"deleted_at":          nil,
	}).Error
}

func cleanupPresentationAccounts(db *gorm.DB) error {
	obsoleteEmails := []string{
		"demo.utama@ruang-tenang.com",
		"demo.cadangan@ruang-tenang.com",
		"usertest@gmail.com",
		"mitra.invited@ruang-tenang.com",
	}
	if err := db.Where("email IN ?", obsoleteEmails).Delete(&model.User{}).Error; err != nil {
		return err
	}

	presentationRoles := []string{
		string(model.RoleAdmin),
		string(model.RoleMitra),
		string(model.RoleUser),
		"moderator",
		"member",
	}

	return db.
		Where("role IN ? AND email NOT IN ?", presentationRoles, presentationAccountEmails).
		Delete(&model.User{}).
		Error
}
