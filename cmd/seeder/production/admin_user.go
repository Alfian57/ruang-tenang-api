package production

import (
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedAdminUser seeds the default admin user
func SeedAdminUser(db *gorm.DB) error {
	// Get admin email and password from environment or use defaults
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@ruangtenang.id"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123" // Default for development, should be changed in production
	}

	adminName := os.Getenv("ADMIN_NAME")
	if adminName == "" {
		adminName = "Admin"
	}

	// Check if admin already exists
	var existing models.User
	if db.Where("email = ?", adminEmail).First(&existing).RowsAffected > 0 {
		return nil // Admin already exists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Name:     adminName,
		Email:    adminEmail,
		Password: string(hashedPassword),
		Role:     models.RoleAdmin,
		Exp:      0,
	}

	return db.Create(&admin).Error
}
