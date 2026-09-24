package presentation

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedStoryCategories seeds the inspiring story categories
func SeedStoryCategories(db *gorm.DB) error {
	categories := []model.StoryCategory{
		{Name: "Perjalanan Pulih", Slug: "recovery-journey", Description: "Refleksi tentang proses pulih yang berbeda bagi setiap orang.", Icon: "🌱", DisplayOrder: 1, IsActive: true},
		{Name: "Depresi dan Dukungan", Slug: "overcoming-depression", Description: "Ruang untuk berbagi tentang depresi dengan dukungan dan moderasi yang aman.", Icon: "☀️", DisplayOrder: 2, IsActive: true},
		{Name: "Mengelola Kecemasan", Slug: "anxiety-management", Description: "Cerita dan refleksi seputar pengalaman menghadapi kecemasan.", Icon: "🧘", DisplayOrder: 3, IsActive: true},
		{Name: "Pemulihan dari Trauma", Slug: "healing-from-trauma", Description: "Cerita sensitif yang memerlukan penanda konten dan moderasi khusus.", Icon: "💚", DisplayOrder: 4, IsActive: true},
		{Name: "Menemukan Harapan", Slug: "finding-hope", Description: "Catatan tentang dukungan, kemungkinan, dan langkah yang terasa berarti.", Icon: "✨", DisplayOrder: 5, IsActive: true},
		{Name: "Merawat Diri", Slug: "self-care-journey", Description: "Pengalaman mencoba merawat kebutuhan dan batas diri.", Icon: "🌸", DisplayOrder: 6, IsActive: true},
		{Name: "Mencari Bantuan Profesional", Slug: "professional-help", Description: "Refleksi tentang mencari dukungan psikologis atau layanan kesehatan.", Icon: "🏥", DisplayOrder: 7, IsActive: true},
		{Name: "Lainnya", Slug: "other", Description: "Topik kesehatan mental lainnya.", Icon: "📝", DisplayOrder: 8, IsActive: true},
	}

	for _, cat := range categories {
		var existing model.StoryCategory
		if db.Where("slug = ?", cat.Slug).First(&existing).RowsAffected == 0 {
			if err := db.Create(&cat).Error; err != nil {
				return err
			}
		} else if err := db.Model(&existing).Updates(map[string]any{
			"name": cat.Name, "description": cat.Description, "icon": cat.Icon,
			"display_order": cat.DisplayOrder, "is_active": cat.IsActive,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}
