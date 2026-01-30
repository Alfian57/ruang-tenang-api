package development

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedTestUsers seeds test users for development
func SeedTestUsers(db *gorm.DB) error {
	// Hash default password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Test users with varying levels and roles
	testUsers := []struct {
		Name   string
		Email  string
		Role   models.UserRole
		Exp    int64
		Avatar string
	}{
		// Moderator
		{Name: "Moderator Test", Email: "moderator@test.com", Role: models.RoleModerator, Exp: 3000, Avatar: "avatar-1.jpg"},

		// Regular members with varying experience
		{Name: "John Doe", Email: "john@example.com", Role: models.RoleMember, Exp: 150, Avatar: "avatar-2.jpg"},
		{Name: "Jane Smith", Email: "jane@example.com", Role: models.RoleMember, Exp: 500, Avatar: "avatar-3.jpg"},
		{Name: "Alex Johnson", Email: "alex@example.com", Role: models.RoleMember, Exp: 1200, Avatar: "avatar-4.jpg"},
		{Name: "Sarah Wilson", Email: "sarah@example.com", Role: models.RoleMember, Exp: 2500, Avatar: "avatar-1.jpg"},
		{Name: "Michael Brown", Email: "michael@example.com", Role: models.RoleMember, Exp: 50, Avatar: "avatar-2.jpg"},
		{Name: "Emily Davis", Email: "emily@example.com", Role: models.RoleMember, Exp: 800, Avatar: "avatar-3.jpg"},
		{Name: "David Lee", Email: "david@example.com", Role: models.RoleMember, Exp: 1800, Avatar: "avatar-4.jpg"},
		{Name: "Lisa Garcia", Email: "lisa@example.com", Role: models.RoleMember, Exp: 4500, Avatar: "avatar-1.jpg"},
		{Name: "James Martinez", Email: "james@example.com", Role: models.RoleMember, Exp: 6000, Avatar: "avatar-2.jpg"},

		// High-level user
		{Name: "Guardian User", Email: "guardian@test.com", Role: models.RoleMember, Exp: 12000, Avatar: "avatar-3.jpg"},
	}

	now := time.Now()
	for _, u := range testUsers {
		var existing models.User
		if db.Where("email = ?", u.Email).First(&existing).RowsAffected > 0 {
			continue
		}

		// Get avatar image
		avatar := ""
		if url, ok := placeholderImages[u.Avatar]; ok {
			avatar = getOrDownloadImage(url, u.Avatar)
		}

		user := models.User{
			Name:             u.Name,
			Email:            u.Email,
			Password:         string(hashedPassword),
			Role:             u.Role,
			Exp:              u.Exp,
			Avatar:           avatar,
			CurrentStreak:    3,
			LongestStreak:    7,
			LastActivityDate: &now,
			TotalActivities:  10,
		}

		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}
