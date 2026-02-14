package development

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedTestUsers seeds test users for development
func SeedTestUsers(db *gorm.DB) error {
	// Hash default password
	// Hash default password for all users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Test users with varying levels and roles
	testUsers := []struct {
		Name   string
		Email  string
		Role   model.UserRole
		Exp    int64
		Avatar string
	}{
		// Moderator
		{Name: "Moderator", Email: "moderator@ruang-tenang.com", Role: model.RoleModerator, Exp: 3000, Avatar: "avatar-1.jpg"},

		// Users
		{Name: "Alfian Gading Saputra", Email: "gading@gmail.com", Role: model.RoleMember, Exp: 1200, Avatar: "avatar-2.jpg"},
		{Name: "Dery Wahyu Perdana", Email: "dery@gmail.com", Role: model.RoleMember, Exp: 800, Avatar: "avatar-3.jpg"},
		{Name: "Riki Andhika Kurna Putra", Email: "andhika@gmail.com", Role: model.RoleMember, Exp: 500, Avatar: "avatar-4.jpg"},
	}

	now := time.Now()
	for _, u := range testUsers {
		var existing model.User
		if db.Where("email = ?", u.Email).First(&existing).RowsAffected > 0 {
			continue
		}

		// Get avatar image
		avatar := ""
		if url, ok := placeholderImages[u.Avatar]; ok {
			avatar = getOrDownloadImage(url, u.Avatar)
		}

		user := model.User{
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
