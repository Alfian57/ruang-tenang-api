package production

import (
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// SeedDefaultAccounts seeds default admin/moderator/member accounts for production bootstrap.
func SeedDefaultAccounts(db *gorm.DB) error {
	defaultPassword := os.Getenv("SEED_DEFAULT_USER_PASSWORD")
	if defaultPassword == "" {
		defaultPassword = "password"
	}

	adminEmail := firstNonEmpty(
		os.Getenv("SEED_ADMIN_EMAIL"),
		os.Getenv("ADMIN_EMAIL"),
		"admin@ruang-tenang.com",
	)
	adminName := firstNonEmpty(
		os.Getenv("SEED_ADMIN_NAME"),
		os.Getenv("ADMIN_NAME"),
		"Admin",
	)
	adminPassword := firstNonEmpty(
		os.Getenv("SEED_ADMIN_PASSWORD"),
		os.Getenv("ADMIN_PASSWORD"),
		defaultPassword,
	)

	type accountSeed struct {
		Name     string
		Email    string
		Role     model.UserRole
		Exp      int64
		Password string
	}

	accounts := []accountSeed{
		{Name: adminName, Email: adminEmail, Role: model.RoleAdmin, Exp: 0, Password: adminPassword},
		{Name: "Alfian Gading Saputra", Email: "gading@gmail.com", Role: model.RoleMember, Exp: 1200, Password: defaultPassword},
		{Name: "Dery Wahyu Perdana", Email: "dery@gmail.com", Role: model.RoleMember, Exp: 800, Password: defaultPassword},
		{Name: "Riki Andhika Kurna Putra", Email: "andhika@gmail.com", Role: model.RoleMember, Exp: 500, Password: defaultPassword},
	}

	for _, account := range accounts {
		var existing model.User
		if db.Where("email = ?", account.Email).First(&existing).RowsAffected > 0 {
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		payload := model.User{
			Name:     account.Name,
			Email:    account.Email,
			Password: string(hashedPassword),
			Role:     account.Role,
			Exp:      account.Exp,
		}

		if err := db.Create(&payload).Error; err != nil {
			return err
		}
	}

	return nil
}
