package production

import (
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAdminUser seeds the default admin user
func SeedAdminUser(db *gorm.DB) error {
	// Get admin email and password from environment (SEED_* preferred, legacy ADMIN_* supported)
	adminEmail := os.Getenv("SEED_ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = os.Getenv("ADMIN_EMAIL")
	}
	if adminEmail == "" {
		adminEmail = "admin@ruang-tenang.com"
	}

	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = os.Getenv("ADMIN_PASSWORD")
	}
	if adminPassword == "" {
		adminPassword = "password"
	}

	adminName := os.Getenv("ADMIN_NAME")
	if adminName == "" {
		adminName = "Admin"
	}

	// Check if admin already exists
	var existing model.User
	if db.Where("email = ?", adminEmail).First(&existing).RowsAffected > 0 {
		return nil // Admin already exists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := model.User{
		Name:     adminName,
		Email:    adminEmail,
		Password: string(hashedPassword),
		Role:     model.RoleAdmin,
		Exp:      0,
	}

	return db.Create(&admin).Error
}
