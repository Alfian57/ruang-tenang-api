package main

import (
	"log"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"github.com/Alfian57/ruang-tenang-api/pkg/utils"
	"gorm.io/gorm"
)

func seedUsers(db *gorm.DB) {
	log.Println("📝 Seeding users...")
	adminPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	memberPassword, err := utils.HashPassword("member123")
	if err != nil {
		log.Fatalf("Failed to hash member password: %v", err)
	}

	users := []models.User{
		{Name: "Admin", Email: "admin@ruangtenang.id", Password: adminPassword, Role: models.RoleAdmin, Exp: 0},
		{Name: "John Doe", Email: "john@example.com", Password: memberPassword, Role: models.RoleMember, Exp: 850},
		{Name: "Alfian Gading Saputra", Email: "alfian@gmail.com", Password: memberPassword, Role: models.RoleMember, Exp: 1200},
		{Name: "Dery Wahyu", Email: "dery@gmail.com", Password: memberPassword, Role: models.RoleMember, Exp: 2300},
		{Name: "Andhika Khusna", Email: "andhika@gmail.com", Password: memberPassword, Role: models.RoleMember, Exp: 450},
	}

	for _, user := range users {
		var existing models.User
		result := db.Where("email = ?", user.Email).First(&existing)

		if result.RowsAffected == 0 {
			if err := db.Create(&user).Error; err != nil {
				log.Printf("  ❌ Failed to create user %s: %v", user.Email, err)
			} else {
				log.Printf("  ✓ Created user: %s", user.Email)
			}
		} else {
			if err := db.Model(&existing).Updates(map[string]interface{}{
				"password": user.Password,
				"exp":      user.Exp,
			}).Error; err != nil {
				log.Printf("  ❌ Failed to update user %s: %v", user.Email, err)
			} else {
				log.Printf("  ✓ Updated user: %s", user.Email)
			}
		}
	}
}
