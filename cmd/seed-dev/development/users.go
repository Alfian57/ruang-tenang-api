package development

import (
	"errors"
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
		Name            string
		Email           string
		Role            model.UserRole
		Exp             int64
		GoldCoins       int64
		Avatar          string
		ProfileTheme    string
		Tagline         string
		Bio             string
		CurrentStreak   int
		LongestStreak   int
		TotalActivities int
		LoginStreak     int
		FreshUser       bool // if true, don't set LastLoginDate/LastActivityDate so daily features trigger
	}{
		// Curated demo users (main + backup) for live showcase flow.
		{Name: "Demo Utama Ruang Tenang", Email: "demo.utama@ruang-tenang.com", Role: model.RoleMember, Exp: 2350, GoldCoins: 640, Avatar: "avatar-2.jpg", ProfileTheme: "ocean_calm", Tagline: "Menyusun ulang pikiran, satu langkah kecil setiap hari.", Bio: "Akun demo utama untuk flow refleksi -> progres -> reward -> komunitas.", CurrentStreak: 9, LongestStreak: 18, TotalActivities: 72, LoginStreak: 9},
		{Name: "Demo Cadangan Ruang Tenang", Email: "demo.cadangan@ruang-tenang.com", Role: model.RoleMember, Exp: 1810, GoldCoins: 430, Avatar: "avatar-3.jpg", ProfileTheme: "forest_zen", Tagline: "Cadangan siap tampil saat demo berpindah jalur.", Bio: "Akun backup dengan state stabil untuk skenario fallback live.", CurrentStreak: 7, LongestStreak: 14, TotalActivities: 56, LoginStreak: 7},

		// Existing member fixtures
		{Name: "Alfian Gading Saputra", Email: "gading@gmail.com", Role: model.RoleMember, Exp: 1200, GoldCoins: 1000, Avatar: "avatar-2.jpg", ProfileTheme: "default", Tagline: "Mulai dari langkah kecil.", Bio: "Pengguna development untuk validasi fitur komunitas.", CurrentStreak: 4, LongestStreak: 9, TotalActivities: 26, LoginStreak: 4},
		{Name: "Dery Wahyu Perdana", Email: "dery@gmail.com", Role: model.RoleMember, Exp: 800, Avatar: "avatar-3.jpg", ProfileTheme: "default", Tagline: "Jalan pelan tapi konsisten.", Bio: "Pengguna development untuk skenario menengah.", CurrentStreak: 3, LongestStreak: 7, TotalActivities: 20, LoginStreak: 3},
		{Name: "Riki Andhika Kurna Putra", Email: "andhika@gmail.com", Role: model.RoleMember, Exp: 500, Avatar: "avatar-4.jpg", ProfileTheme: "default", Tagline: "Belajar mengenali diri tiap hari.", Bio: "Pengguna development untuk state awal.", CurrentStreak: 2, LongestStreak: 5, TotalActivities: 14, LoginStreak: 2},
		{Name: "User Test", Email: "usertest@gmail.com", Role: model.RoleMember, Exp: 1500, GoldCoins: 1000, Avatar: "avatar-3.jpg", ProfileTheme: "sunset_warmth", Tagline: "Ritme pulih versi saya.", Bio: "Akun fresh untuk memicu daily login dan popup mood check-in.", FreshUser: true},
	}

	now := time.Now().UTC()
	for _, u := range testUsers {
		var existing model.User
		err := db.Where("email = ?", u.Email).First(&existing).Error
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
			"name":             u.Name,
			"password":         string(hashedPassword),
			"role":             u.Role,
			"exp":              u.Exp,
			"gold_coins":       u.GoldCoins,
			"avatar":           avatar,
			"profile_theme":    u.ProfileTheme,
			"tagline":          u.Tagline,
			"bio":              u.Bio,
			"current_streak":   u.CurrentStreak,
			"longest_streak":   u.LongestStreak,
			"total_activities": u.TotalActivities,
			"login_streak":     u.LoginStreak,
		}

		if u.FreshUser {
			updates["current_streak"] = 0
			updates["longest_streak"] = 0
			updates["total_activities"] = 0
			updates["login_streak"] = 0
			updates["last_activity_date"] = nil
			updates["last_login_date"] = nil
		} else {
			updates["last_activity_date"] = now
			updates["last_login_date"] = now
		}

		if err == nil {
			if err := db.Model(&existing).Updates(updates).Error; err != nil {
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
			ProfileTheme:    u.ProfileTheme,
			Tagline:         u.Tagline,
			Bio:             u.Bio,
			CurrentStreak:   u.CurrentStreak,
			LongestStreak:   u.LongestStreak,
			TotalActivities: u.TotalActivities,
			LoginStreak:     u.LoginStreak,
		}

		if u.FreshUser {
			// Fresh user: no login date, no activity — so daily login, mood popup, etc. will trigger
			user.CurrentStreak = 0
			user.LongestStreak = 0
			user.LoginStreak = 0
			user.TotalActivities = 0
		} else {
			user.LastActivityDate = &now
			user.LastLoginDate = &now
		}

		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}
