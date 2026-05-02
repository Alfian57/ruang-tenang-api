package presentation

import (
	"errors"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedTestUsers seeds the fixed non-admin presentation accounts.
func SeedTestUsers(db *gorm.DB) error {
	// Hash default password for all users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Test users with varying levels and roles
	testUsers := []struct {
		Name            string
		Email           string
		Role            model.UserRole
		Exp             int64
		GoldCoins       int64
		Avatar          string
		Tagline         string
		Bio             string
		CurrentStreak   int
		LongestStreak   int
		TotalActivities int
		LoginStreak     int
	}{
		{Name: "Mitra Ruang Tenang", Email: presentationMitraEmail, Role: model.RoleMitra, Exp: 2200, GoldCoins: 900, Avatar: "avatar-1.jpg", Tagline: "Mengelola program wellbeing organisasi.", Bio: "Akun mitra tunggal untuk dashboard B2B dan presentasi juri.", CurrentStreak: 10, LongestStreak: 21, TotalActivities: 88, LoginStreak: 10},
		{Name: "Alfian Gading Saputra", Email: presentationGadingEmail, Role: model.RoleUser, Exp: 1600, GoldCoins: 1200, Avatar: "avatar-2.jpg", Tagline: "Mulai dari langkah kecil.", Bio: "User premium personal untuk validasi fitur komunitas dan billing.", CurrentStreak: 6, LongestStreak: 12, TotalActivities: 34, LoginStreak: 6},
		{Name: "Dery Wahyu Perdana", Email: presentationDeryEmail, Role: model.RoleUser, Exp: 980, GoldCoins: 720, Avatar: "avatar-3.jpg", Tagline: "Jalan pelan tapi konsisten.", Bio: "User premium melalui jalur B2B untuk simulasi seat organisasi.", CurrentStreak: 4, LongestStreak: 8, TotalActivities: 24, LoginStreak: 4},
		{Name: "Riki Andhika Kusna Putra", Email: presentationAndhikaEmail, Role: model.RoleUser, Exp: 420, GoldCoins: 180, Avatar: "avatar-4.jpg", Tagline: "Belajar mengenali diri tiap hari.", Bio: "User freemium untuk validasi limit dan onboarding pengguna gratis.", CurrentStreak: 1, LongestStreak: 4, TotalActivities: 10, LoginStreak: 1},
	}

	now := time.Now().UTC()
	for _, u := range testUsers {
		var existing model.User
		err := db.Unscoped().Where("email = ?", u.Email).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Get avatar image
		url := ""
		if avatarURL, ok := placeholderImages[u.Avatar]; ok {
			url = avatarURL
		}
		avatar := getOrDownloadImage(url, u.Avatar)

		updates := map[string]interface{}{
			"name":               u.Name,
			"password":           string(hashedPassword),
			"role":               u.Role,
			"exp":                u.Exp,
			"gold_coins":         u.GoldCoins,
			"avatar":             avatar,
			"profile_theme":      presentationDefaultProfileTheme,
			"tagline":            u.Tagline,
			"bio":                u.Bio,
			"current_streak":     u.CurrentStreak,
			"longest_streak":     u.LongestStreak,
			"total_activities":   u.TotalActivities,
			"login_streak":       u.LoginStreak,
			"is_premium":         false,
			"premium_since":      nil,
			"premium_expires_at": nil,
			"is_blocked":         false,
			"is_forum_blocked":   false,
			"is_banned":          false,
			"suspension_end":     nil,
			"suspension_reason":  "",
			"ban_reason":         "",
			"deleted_at":         nil,
		}

		updates["last_activity_date"] = now
		updates["last_login_date"] = now

		if err == nil {
			if err := db.Unscoped().Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			continue
		}

		user := model.User{
			Name:            u.Name,
			Email:           u.Email,
			Password:        string(hashedPassword),
			Role:            u.Role,
			Exp:             u.Exp,
			GoldCoins:       u.GoldCoins,
			Avatar:          avatar,
			ProfileTheme:    presentationDefaultProfileTheme,
			Tagline:         u.Tagline,
			Bio:             u.Bio,
			CurrentStreak:   u.CurrentStreak,
			LongestStreak:   u.LongestStreak,
			TotalActivities: u.TotalActivities,
			LoginStreak:     u.LoginStreak,
			IsPremium:       false,
		}

		user.LastActivityDate = &now
		user.LastLoginDate = &now

		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}
