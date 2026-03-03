package production

import (
	"os"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedDefaultAccounts seeds default moderator/member accounts for production bootstrap.
func SeedDefaultAccounts(db *gorm.DB) error {
	defaultPassword := os.Getenv("SEED_DEFAULT_USER_PASSWORD")
	if defaultPassword == "" {
		defaultPassword = "password"
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	accounts := []model.User{
		{Name: "Moderator", Email: "moderator@ruang-tenang.com", Role: model.RoleModerator, Exp: 3000},
		{Name: "Alfian Gading Saputra", Email: "gading@gmail.com", Role: model.RoleMember, Exp: 1200},
		{Name: "Dery Wahyu Perdana", Email: "dery@gmail.com", Role: model.RoleMember, Exp: 800},
		{Name: "Riki Andhika Kurna Putra", Email: "andhika@gmail.com", Role: model.RoleMember, Exp: 500},
	}

	for _, account := range accounts {
		var existing model.User
		if db.Where("email = ?", account.Email).First(&existing).RowsAffected > 0 {
			continue
		}

		account.Password = string(hashedPassword)
		if err := db.Create(&account).Error; err != nil {
			return err
		}
	}

	return nil
}
