package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedStoryCategories seeds the inspiring story categories
func SeedStoryCategories(db *gorm.DB) error {
	categories := []model.StoryCategory{
		{Name: "Recovery Journey", Slug: "recovery-journey", Description: "Share your path to recovery and healing", Icon: "🌱", DisplayOrder: 1, IsActive: true},
		{Name: "Overcoming Depression", Slug: "overcoming-depression", Description: "Stories about battling and overcoming depression", Icon: "☀️", DisplayOrder: 2, IsActive: true},
		{Name: "Anxiety Management", Slug: "anxiety-management", Description: "Experiences with managing anxiety", Icon: "🧘", DisplayOrder: 3, IsActive: true},
		{Name: "Healing from Trauma", Slug: "healing-from-trauma", Description: "Journey of healing from traumatic experiences", Icon: "💚", DisplayOrder: 4, IsActive: true},
		{Name: "Finding Hope", Slug: "finding-hope", Description: "Stories about finding hope in difficult times", Icon: "✨", DisplayOrder: 5, IsActive: true},
		{Name: "Self-Care Journey", Slug: "self-care-journey", Description: "Personal self-care practices and discoveries", Icon: "🌸", DisplayOrder: 6, IsActive: true},
		{Name: "Professional Help Experience", Slug: "professional-help", Description: "Experiences with therapy, counseling, or treatment", Icon: "🏥", DisplayOrder: 7, IsActive: true},
		{Name: "Other", Slug: "other", Description: "Other mental health related stories", Icon: "📝", DisplayOrder: 8, IsActive: true},
	}

	for _, cat := range categories {
		var existing model.StoryCategory
		if db.Where("slug = ?", cat.Slug).First(&existing).RowsAffected == 0 {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
