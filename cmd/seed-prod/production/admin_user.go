package production

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// copyAdminAvatar returns an image URL for the admin
func copyAdminAvatar(filename string) string {
	p := filepath.Join("storage", "images", filename)
	if _, err := os.Stat(p); err == nil {
		uploadDir := filepath.Join("uploads", "images")
		os.MkdirAll(uploadDir, 0755)

		dstPath := filepath.Join(uploadDir, filename)
		if src, err := os.Open(p); err == nil {
			if dst, err := os.Create(dstPath); err == nil {
				io.Copy(dst, src)
				dst.Close()
				src.Close()
				return fmt.Sprintf("/uploads/images/%s", filename)
			}
			src.Close()
		}
	}
	log.Printf("⚠️ Admin avatar image not found for %s", filename)
	return ""
}

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

	// Try to get an avatar for admin
	var adminAvatar string
	if localAsset := copyAdminAvatar("avatar-1.jpg"); localAsset != "" {
		adminAvatar = localAsset
	}

	admin := model.User{
		Name:     adminName,
		Email:    adminEmail,
		Password: string(hashedPassword),
		Role:     model.RoleAdmin,
		Exp:      0,
		Avatar:   adminAvatar,
	}

	return db.Create(&admin).Error
}
